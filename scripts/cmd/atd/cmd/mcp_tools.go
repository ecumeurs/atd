package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"atd-tools/config"
	"atd-tools/pkg/mcp"
)

// captureStdout runs fn() while redirecting os.Stdout to a buffer.
// Returns the captured output as a string.
// This is necessary because several run* functions print progress directly
// to os.Stdout. In MCP stdio mode, os.Stdout is the JSON-RPC channel and
// must not receive unstructured output.
func captureStdout(fn func() error) (string, error) {
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w

	copyDone := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		copyDone <- buf.String()
	}()

	runErr := fn()
	w.Close()
	os.Stdout = orig
	captured := <-copyDone
	r.Close()

	if runErr != nil {
		return captured, runErr
	}
	return captured, nil
}

func argString(args map[string]any, key, fallback string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}

func argBool(args map[string]any, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// RegisterMCPTools registers all ATD subcommands as MCP tools.
// All tools auto-configure from the .atd project configuration.
// The Agent LLM should never need to provide internal paths (docs, src, db) or thresholds.
func RegisterMCPTools(r *mcp.Registry) {

	// ── Deterministic Tools (no LLM, fast, token-free) ───────────────────

	r.Register(mcp.Tool{
		Name: "atd_query",
		Description: `Search ATD atoms by frontmatter field value (e.g. type, status, id, layer, tags).
Use during PLAN stage to find existing atoms before creating new ones, or to locate all atoms matching a criteria (e.g. all STABLE atoms, all RULE types, atoms tagged 'auth').
Returns a JSON array of matching atoms with full frontmatter.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"field":  map[string]any{"type": "string", "description": "Frontmatter field to search (e.g. 'type', 'status', 'id', 'layer', 'tags'). Omit to search all fields."},
				"search": map[string]any{"type": "string", "description": "Value to match (case-insensitive substring)."},
			},
			"required": []string{"search"},
		},
	}, func(args map[string]any) (string, error) {
		field := argString(args, "field", "id")
		search := argString(args, "search", "")
		return runQuery(config.DocsDir(), field, search)
	})

	r.Register(mcp.Tool{
		Name: "atd_crawl",
		Description: `Build a dependency graph of ATD atoms and their @spec-link connections to source code.
Use during EVOLVE stage before modifying a high-level atom to understand ripple effects (blast radius analysis).
Set gaps=true during VERIFY stage to find STABLE atoms with no code implementations (orphan detection).`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gaps": map[string]any{"type": "boolean", "description": "If true, return only STABLE atoms with zero code implementations (orphan detection)."},
			},
		},
	}, func(args map[string]any) (string, error) {
		gaps := argBool(args, "gaps")
		return runCrawl(".", config.DocsDir(), gaps)
	})

	r.Register(mcp.Tool{
		Name: "atd_weave",
		Description: `Synchronize the bidirectional atom graph by populating dependents[] from parents[] references.
Run after creating or modifying atoms (especially parents fields) during the SPECIFY stage.
This is mandatory after any atom creation to keep the dependency graph consistent.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runWeave(config.DocsDir())
	})

	r.Register(mcp.Tool{
		Name: "atd_update",
		Description: `Surgically modify ATD atom files — the ONLY correct way to edit .atom.md files.
Use for: creating new atoms (provide file path + all required fields), changing status/priority/layer, editing H2 sections (intent, logic, interface, expectation), injecting @spec-link tags into source code.
For batch operations, use 'filter' instead of 'file' to update all matching atoms in one call.
NEVER rewrite an entire .atom.md file manually — always use this tool.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file":           map[string]any{"type": "string", "description": "Path to the .atom.md file. Required unless 'filter' is provided."},
				"filter":         map[string]any{"type": "string", "description": "Filter atoms to update in batch (e.g. 'status=DRAFT,type=RULE'). Mutually exclusive with 'file'."},
				"set":            map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Frontmatter edits as 'key=value' strings, e.g. [\"status=STABLE\",\"priority=3\"]."},
				"intent":         map[string]any{"type": "string", "description": "New INTENT section text."},
				"logic":          map[string]any{"type": "string", "description": "New THE RULE / LOGIC section text."},
				"interface":      map[string]any{"type": "string", "description": "New TECHNICAL INTERFACE section text."},
				"expectation":    map[string]any{"type": "string", "description": "New EXPECTATION section text."},
				"spec_link":      map[string]any{"type": "string", "description": "Atom ID to inject as @spec-link in a source file (requires spec_link_file)."},
				"spec_link_file": map[string]any{"type": "string", "description": "Source file path for @spec-link injection."},
			},
		},
	}, func(args map[string]any) (string, error) {
		file := argString(args, "file", "")
		filter := argString(args, "filter", "")
		intent := argString(args, "intent", "")
		logic := argString(args, "logic", "")
		iface := argString(args, "interface", "")
		expectation := argString(args, "expectation", "")
		specLink := argString(args, "spec_link", "")
		specFile := argString(args, "spec_link_file", "")

		var setPairs []string
		if raw, ok := args["set"]; ok {
			if arr, ok := raw.([]any); ok {
				for _, item := range arr {
					if s, ok := item.(string); ok {
						setPairs = append(setPairs, s)
					}
				}
			}
		}

		if filter != "" {
			if file != "" {
				return "", fmt.Errorf("cannot use both 'file' and 'filter'")
			}
			return runBatchUpdate(filter, setPairs, intent, logic, iface, expectation, specLink, specFile)
		}

		if file == "" {
			return "", fmt.Errorf("either 'file' or 'filter' is required")
		}

		return runUpdate(file, setPairs, intent, logic, iface, expectation, specLink, specFile)
	})

	r.Register(mcp.Tool{
		Name: "atd_roadmap",
		Description: `Scan a source directory and produce a complexity map ranking files by density (lines, cyclomatic complexity, function count).
Use during cold-start PLAN stage to prioritize which files to dissect first.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"dir": map[string]any{"type": "string", "description": "Directory to scan."},
				"out": map[string]any{"type": "string", "description": "Optional output file path for roadmap.json."},
			},
			"required": []string{"dir"},
		},
	}, func(args map[string]any) (string, error) {
		dir := argString(args, "dir", ".")
		out := argString(args, "out", "roadmap.json")
		return runRoadmap(dir, out)
	})

	// @spec-link [[api_atd_serve_stats]]
	r.Register(mcp.Tool{
		Name: "atd_stats",
		Description: `Produce quantitative documentation health metrics: total atoms, atoms by type/status/layer, @spec-link coverage ratio, and orphan count.
Use during VERIFY stage to assess overall documentation quality, or in CI to generate health badges.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runStats(".", config.DocsDir())
	})

	r.Register(mcp.Tool{
		Name: "atd_verify",
		Description: `Run git diff, extract impacted @spec-link tags, and produce a structured audit prompt.
Use during VERIFY stage (pre-commit or CI) to check whether code changes still comply with their linked atom specifications.
Returns the changed code alongside each atom it is linked to, for compliance review.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runVerify(config.DocsDir())
	})

	r.Register(mcp.Tool{
		Name: "atd_assemble",
		Description: `Stitch atoms together into a cohesive narrative document by walking the dependency graph from root atoms.
Use during PLAN stage for onboarding documents, architecture overviews, or executive summaries.
Set snapshot=true to delegate narrative generation to the IDE Agent for polished output.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"starts":   map[string]any{"type": "string", "description": "Comma-separated list of root Atom IDs to begin assembly from."},
				"purpose":  map[string]any{"type": "string", "description": "Optional purpose to wrap the output in <System Objective> tags for LLM consumption."},
				"snapshot": map[string]any{"type": "boolean", "description": "If true, delegate narrative generation to IDE Agent (returns a task ID). Uses LLM."},
				"theme":    map[string]any{"type": "string", "description": "Theme for snapshot narrative (defaults to 'Executive Summary')."},
			},
			"required": []string{"starts"},
		},
	}, func(args map[string]any) (string, error) {
		starts := argString(args, "starts", "")
		purpose := argString(args, "purpose", "")
		snapshot := argBool(args, "snapshot")
		theme := argString(args, "theme", "Executive Summary")
		return runAssemble(starts, purpose, snapshot, theme, config.DocsDir())
	})

	// @spec-link [[api_atd_serve_trace]]
	r.Register(mcp.Tool{
		Name: "atd_trace",
		Description: `Get a structured Health Snapshot JSON for a specific atom by traversing its graph ancestry and descendants. Includes warnings for layer compliance and metrics for testing and implementation coverage.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"atom": map[string]any{"type": "string", "description": "Target ID of the atom to trace."},
			},
			"required": []string{"atom"},
		},
	}, func(args map[string]any) (string, error) {
		atomID, ok := args["atom"].(string)
		if !ok || atomID == "" {
			return "", fmt.Errorf("atom is required")
		}
		return runTrace(atomID, config.DocsDir(), ".")
	})

	r.Register(mcp.Tool{
		Name: "atd_test_links",
		Description: `Audit @test-link [[ATOM_ID]] tags in source code to map atoms to their verification tests.
Use during VERIFY stage to confirm test coverage per atom, or before modifying an atom to identify which tests need re-running.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"atom": map[string]any{"type": "string", "description": "Optional: filter results for a specific Atom ID."},
			},
		},
	}, func(args map[string]any) (string, error) {
		atomID := argString(args, "atom", "")
		return runTestLinks(".", atomID, config.DocsDir())
	})

	// ── LLM-Backed Tools (require Ollama or IDE Agent fallback) ──────────

	r.Register(mcp.Tool{
		Name: "atd_dissect",
		Description: `Dissect a source code or documentation file into proposed atomic boundaries (IDs, types, line ranges).
Use during cold-start to break down undocumented files into atomic units, or when onboarding legacy code.
The tool uses the LLM provider configured in .atd; if no provider is available, it returns a structured prompt for the IDE Agent to process.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{"type": "string", "description": "Path to the source or documentation file to dissect."},
			},
			"required": []string{"file"},
		},
	}, func(args map[string]any) (string, error) {
		file := argString(args, "file", "")
		return runDissect(file, true)
	})

	r.Register(mcp.Tool{
		Name: "atd_index",
		Description: `Build or refresh the semantic vector index of all source code and ATD documents.
Uses nomic-embed-text to generate embeddings stored in a SQLite database. Files unchanged since last indexing are automatically skipped (mtime-based caching).
Run before using atd_search (semantic mode), or after significant code/documentation changes to keep the index fresh.
This tool takes no parameters — it indexes the entire project using the .atd configuration.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		db := config.DocsDir() + "/.atd_index.db"
		return captureStdout(func() error {
			return runIndex(".", db, "all")
		})
	})

	r.Register(mcp.Tool{
		Name: "atd_search",
		Description: `Search the project semantically or by keyword.
Semantic mode (query): embeds the query via Nomic and finds the most similar code/doc chunks by cosine similarity. Requires a built index (run atd_index first).
Keyword mode (grep): literal string search across all project files.
Use during PLAN stage to find related code or atoms by meaning, or to locate implementations when atom IDs are unknown.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Semantic search query (uses Nomic embeddings). Provide this OR grep, not both."},
				"grep":  map[string]any{"type": "string", "description": "Literal keyword search across project files. Provide this OR query, not both."},
				"scope": map[string]any{"type": "string", "description": "Search scope: 'code' (source files only), 'docs' (ATD atoms only), or 'all' (both). Defaults to 'all'."},
				"limit": map[string]any{"type": "integer", "description": "Number of semantic results to return. Defaults to 5."},
			},
		},
	}, func(args map[string]any) (string, error) {
		query := argString(args, "query", "")
		grep := argString(args, "grep", "")
		scope := argString(args, "scope", "all")
		limitRaw, _ := args["limit"]
		limit := 5
		if f, ok := limitRaw.(float64); ok {
			limit = int(f)
		}
		db := config.DocsDir() + "/.atd_index.db"
		if grep != "" {
			return captureStdout(func() error {
				return runGrepSearch(grep)
			})
		}
		return captureStdout(func() error {
			return runSemanticSearch(query, db, limit, scope)
		})
	})

	r.Register(mcp.Tool{
		Name: "atd_audit",
		Description: `Audit ATD atoms for documentation quality issues.
Default mode: detect bloated atoms and semantic collisions (duplicate/overlapping atoms) across the entire docs directory.
Compliance mode (provide both atom + code): validate whether a specific code file conforms to a specific atom's specification.
Use during VERIFY stage after creating new atoms to check for overlap, or to validate code-spec alignment.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"code": map[string]any{"type": "string", "description": "Path to code file for compliance mode. Must be provided together with 'atom'."},
				"atom": map[string]any{"type": "string", "description": "Path to atom file for compliance mode. Must be provided together with 'code'."},
			},
		},
	}, func(args map[string]any) (string, error) {
		docs := config.DocsDir()
		code := argString(args, "code", "")
		atom := argString(args, "atom", "")
		threshold := config.ActiveConfig.DiffSimilarityThreshold
		if threshold <= 0 {
			threshold = 0.85
		}
		if code != "" && atom != "" {
			return captureStdout(func() error {
				return runCodeAudit(code, atom)
			})
		}
		return captureStdout(func() error {
			return runFullAudit(docs, threshold)
		})
	})

	r.Register(mcp.Tool{
		Name: "atd_recon",
		Description: `Validate whether a candidate source file implements a specific ATD atom (semantic archaeology).
Use during cold-start to verify discovered file-atom links before applying @spec-link tags, or to audit existing links after refactoring.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"atom":      map[string]any{"type": "string", "description": "Path to the target .atom.md file."},
				"candidate": map[string]any{"type": "string", "description": "Path to the candidate source code file to validate."},
			},
			"required": []string{"atom", "candidate"},
		},
	}, func(args map[string]any) (string, error) {
		atom := argString(args, "atom", "")
		candidate := argString(args, "candidate", "")
		return runRecon(atom, candidate)
	})

	r.Register(mcp.Tool{
		Name: "atd_discover",
		Description: `Extract architectural intent from an undocumented source file and recommend @spec-link tags to apply.
Searches the ATD index for matching atoms and suggests placements following surgical attachment rules (no global headers, logic-boundary placement).
Use during IMPLEMENT stage to ensure new files are linked to the appropriate atoms.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{"type": "string", "description": "Path to the undocumented source file."},
			},
			"required": []string{"file"},
		},
	}, func(args map[string]any) (string, error) {
		file := argString(args, "file", "")
		return captureStdout(func() error {
			out, err := runDiscover(file, config.DocsDir())
			if err != nil {
				return err
			}
			fmt.Println(out)
			return nil
		})
	})

	// ── Configuration & Diagnostics ──────────────────────────────────────

	r.Register(mcp.Tool{
		Name: "atd_check",
		Description: `Unified environment smoke test: validates .atd config, checks provider connectivity, and verifies model availability.
Use to diagnose 'Connection Refused' or 'Model Not Found' errors, or to verify a new provider/model is correctly configured.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runCheck()
	})

	r.Register(mcp.Tool{
		Name: "atd_config",
		Description: `View or modify the .atd project configuration.
Use 'list':true to see the full config. Use 'bloating_factor' to check the granularity tolerance for a specific atom type before creating atoms.
Use 'task'+'model' to reassign which LLM model handles a specific task type (e.g. dissect, embed, audit_bloat).`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"list":            map[string]any{"type": "boolean", "description": "If true, return the full .atd configuration as JSON."},
				"bloating_factor": map[string]any{"type": "string", "description": "Atom type to query for its bloating factor (e.g. 'RULE', 'USECASE'). Check this before creating atoms."},
				"task":            map[string]any{"type": "string", "description": "Task name to reassign (requires 'model'). E.g. 'dissect', 'embed', 'audit_bloat'."},
				"model":           map[string]any{"type": "string", "description": "Model name to assign to the task (requires 'task')."},
			},
		},
	}, func(args map[string]any) (string, error) {
		list := argBool(args, "list")
		if list {
			out, _ := json.MarshalIndent(config.ActiveConfig, "", "  ")
			return string(out), nil
		}
		atomType := argString(args, "bloating_factor", "")
		if atomType != "" {
			return runConfigGetBloating(atomType)
		}
		task := argString(args, "task", "")
		model := argString(args, "model", "")
		if task != "" && model != "" {
			return runConfigUpdate(task, model)
		}
		return "", fmt.Errorf("must provide 'list':true, 'bloating_factor', or both 'task' and 'model'")
	})

	// @spec-link [[api_atd_serve_lint]]
	r.Register(mcp.Tool{
		Name: "atd_lint",
		Description: `Perform fast, deterministic structural validation on all ATD atoms (no LLM, no tokens).
Catches: missing mandatory fields, malformed [[id]] references, broken parent/dependent links, empty H2 sections.
Use during VERIFY stage as a cheap first-pass before running the heavier atd_audit.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runLint(config.DocsDir())
	})
}
