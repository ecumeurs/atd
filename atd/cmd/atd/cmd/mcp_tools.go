package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"atd-tools/config"
	"atd-tools/pkg/exploration"
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

	// @spec-link [[api_atd_serve_query]]
	r.Register(mcp.Tool{
		Name: "atd_query",
		Description: `Search ATD atoms by frontmatter field value (e.g. type, status, id, layer, tags).
Use during PLAN stage to find existing atoms before creating new ones, or to locate all atoms matching a criteria (e.g. all STABLE atoms, all RULE types, atoms tagged 'auth').
Returns a JSON array of matching atoms with full frontmatter.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"field":       map[string]any{"type": "string", "description": "Frontmatter field to search (e.g. 'type', 'status', 'id', 'layer', 'tags'). Omit to search all fields."},
				"search":      map[string]any{"type": "string", "description": "Value to match (case-insensitive substring)."},
				"paths_only":  map[string]any{"type": "boolean", "description": "If true, return only a JSON array of absolute file paths."},
			},
			"required": []string{"search"},
		},
	}, func(args map[string]any) (string, error) {
		field := argString(args, "field", "")
		search := argString(args, "search", "")
		pathsOnly := argBool(args, "paths_only")
		return runQuery(field, search, pathsOnly)
	})

	// @spec-link [[api_atd_serve_crawl]]
	r.Register(mcp.Tool{
		Name: "atd_crawl",
		Description: `Build a dependency graph of ATD atoms and their @spec-link connections to source code.
Use during EVOLVE stage before modifying a high-level atom to understand ripple effects (blast radius analysis).
Set gaps=true during VERIFY stage to find STABLE atoms with no code implementations (orphan detection).`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"gaps":      map[string]any{"type": "boolean", "description": "If true, return only STABLE atoms with zero code implementations (orphan detection)."},
				"workspace": map[string]any{"type": "boolean", "description": "If true, crawl the entire workspace."},
			},
		},
	}, func(args map[string]any) (string, error) {
		gaps := argBool(args, "gaps")
		workspace := argBool(args, "workspace")
		return runCrawl(".", config.DocsDir(), gaps, workspace)
	})

	// @spec-link [[api_atd_serve_weave]]
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
		explorer := exploration.NewExplorer("", config.DocsDir())
		return explorer.Weave()
	})

	// @spec-link [[api_atd_serve_update]]
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

	// @spec-link [[api_atd_serve_roadmap]]
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
			"type": "object",
			"properties": map[string]any{
				"workspace": map[string]any{"type": "boolean", "description": "If true, aggregate stats from all projects in the workspace."},
			},
		},
	}, func(args map[string]any) (string, error) {
		workspace := argBool(args, "workspace")
		return runStats(".", config.DocsDir(), workspace)
	})

	// @spec-link [[api_atd_serve_verify]]
	r.Register(mcp.Tool{
		Name: "atd_check",
		Description: `Unified coverage report: lists impl links (@spec-link) and test links (@test-link) for every atom touched by the current diff or the full project.

Default: audits uncommitted changes (git diff). Pass base/target to compare commits. Pass full:true to audit the entire project regardless of changes.

When semantic:true is added, each impl link is also checked for LLM compliance — the linked code is compared against the atom specification and returns PASS/FAIL per link. This consumes tokens via the configured LLM provider.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"base":     map[string]any{"type": "string", "description": "Optional: base commit/ref to compare from (e.g. 'HEAD~5')."},
				"target":   map[string]any{"type": "string", "description": "Optional: target commit/ref to compare to (defaults to working tree)."},
				"full":     map[string]any{"type": "boolean", "description": "Optional: audit the entire project instead of just the diff."},
				"file":     map[string]any{"type": "string", "description": "Optional: target a specific file for verification."},
				"line":     map[string]any{"type": "integer", "description": "Optional: target a specific line for verification (requires 'file')."},
				"semantic": map[string]any{"type": "boolean", "description": "Optional: add LLM compliance check per impl link (consumes tokens). Returns PASS/FAIL per @spec-link."},
			},
		},
	}, func(args map[string]any) (string, error) {
		base := argString(args, "base", "")
		target := argString(args, "target", "")
		full := argBool(args, "full")
		file := argString(args, "file", "")
		semantic := argBool(args, "semantic")

		verifyArgs := []string{}
		if base != "" {
			verifyArgs = append(verifyArgs, base)
		}
		if target != "" {
			verifyArgs = append(verifyArgs, target)
		}
		return runCoverageCheck("diff", "", file, config.DocsDir(), full, semantic, verifyArgs)
	})

	// @spec-link [[api_atd_serve_assemble]]
	r.Register(mcp.Tool{
		Name: "atd_assemble",
		Description: `Stitch atoms together into a cohesive document by walking the dependency graph from root atoms.
Use during PLAN stage for onboarding documents, architecture overviews, or executive summaries.
Supports structured layer-by-layer summarization by the LLM by passing structured=true.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"starts":     map[string]any{"type": "string", "description": "Comma-separated list of root Atom IDs to begin assembly from."},
				"intent":          map[string]any{"type": "string", "description": "The intent the LLM should focus on (e.g., summarize, executive summary). Defaults to 'Executive Summary'."},
				"length":          map[string]any{"type": "string", "description": "Length constraint: 'short', 'default', 'extended', 'long'."},
				"structured":      map[string]any{"type": "boolean", "description": "If true, group atoms by layer and perform multi-pass summarization."},
				"json":            map[string]any{"type": "boolean", "description": "If true, outputs the result as a structured JSON object along with involved atoms metadata."},
				"only_parents":    map[string]any{"type": "boolean", "description": "If true, restricts assembly to the target atom's ancestry only."},
				"only_dependents": map[string]any{"type": "boolean", "description": "If true, restricts assembly to the target atom's descendants only."},
			},
			"required": []string{"starts"},
		},
	}, func(args map[string]any) (string, error) {
		starts := argString(args, "starts", "")
		intent := argString(args, "intent", "Executive Summary")
		length := argString(args, "length", "default")
		structured := argBool(args, "structured")
		asJSON := argBool(args, "json")
		onlyParents := argBool(args, "only_parents")
		onlyDependents := argBool(args, "only_dependents")
		return runAssemble(starts, intent, length, structured, asJSON, onlyParents, onlyDependents, config.DocsDir())
	})

	// @spec-link [[api_atd_serve_trace]]
	r.Register(mcp.Tool{
		Name: "atd_trace",
		Description: `Get a structured Health Snapshot JSON for a specific atom by traversing its graph ancestry and descendants. Includes warnings for layer compliance and metrics for testing and implementation coverage.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"atom":    map[string]any{"type": "string", "description": "Target ID of the atom to trace."},
				"summary": map[string]any{"type": "boolean", "description": "If true, returns a narrative contextual summary instead of raw JSON."},
			},
			"required": []string{"atom"},
		},
	}, func(args map[string]any) (string, error) {
		atomID, _ := args["atom"].(string)
		summary, _ := args["summary"].(bool)
		if atomID == "" {
			return "", fmt.Errorf("atom is required")
		}
		return runTrace(atomID, config.DocsDir(), ".", summary)
	})

	// @spec-link [[api_atd_serve_test_links]]
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
		return runCoverageCheck("atom", atomID, "", config.DocsDir(), false, false, nil)
	})

	// ── LLM-Backed Tools (require Ollama or IDE Agent fallback) ──────────

	// @spec-link [[api_atd_serve_dissect]]
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

	// @spec-link [[api_atd_serve_index]]
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
		db := config.IndexDBPath(config.DocsDir())
		return captureStdout(func() error {
			return runIndex(".", db, "all")
		})
	})

	// @spec-link [[api_atd_serve_search]]
	r.Register(mcp.Tool{
		Name: "atd_search",
		Description: `Search the project semantically or by keyword.
Semantic mode (query): embeds the query via Nomic and finds the most similar code/doc chunks by cosine similarity. Requires a built index (run atd_index first).
Keyword mode (grep): literal string search across all project files.
Use during PLAN stage to find related code or atoms by meaning, or to locate implementations when atom IDs are unknown.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query":      map[string]any{"type": "string", "description": "Semantic search query (uses Nomic embeddings). Provide this OR grep, not both."},
				"grep":       map[string]any{"type": "string", "description": "Literal keyword search across project files. Provide this OR query, not both."},
				"scope":      map[string]any{"type": "string", "description": "Search scope: 'code' (source files only), 'docs' (ATD atoms only), or 'all' (both). Defaults to 'all'."},
				"limit":      map[string]any{"type": "integer", "description": "Number of semantic results to return. Defaults to 5."},
				"paths_only": map[string]any{"type": "boolean", "description": "If true, return only a list of unique absolute file paths."},
			},
		},
	}, func(args map[string]any) (string, error) {
		query := argString(args, "query", "")
		grep := argString(args, "grep", "")
		scope := argString(args, "scope", "all")
		pathsOnly := argBool(args, "paths_only")
		limitRaw, _ := args["limit"]
		limit := 5
		if f, ok := limitRaw.(float64); ok {
			limit = int(f)
		}
		db := config.IndexDBPath(config.DocsDir())
		if grep != "" {
			return captureStdout(func() error {
				return runGrepSearch(grep, pathsOnly)
			})
		}
		return captureStdout(func() error {
			return runSemanticSearch(query, db, limit, scope, pathsOnly)
		})
	})

	// @spec-link [[api_atd_serve_audit]]
	r.Register(mcp.Tool{
		Name: "atd_audit",
		Description: `Audit ATD atoms for documentation quality issues.
Default mode: detect bloated atoms and semantic collisions (duplicate/overlapping atoms) across the entire docs directory.
Use during PLAN stage after creating new atoms to check for overlap.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		docs := config.DocsDir()
		threshold := config.ActiveConfig.DiffSimilarityThreshold
		if threshold <= 0 {
			threshold = 0.85
		}
		return captureStdout(func() error {
			return runFullAudit(docs, threshold, false)
		})
	})

	// @spec-link [[api_atd_serve_recon]]
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
		return runMap(candidate, atom, config.DocsDir(), false)
	})

	// @spec-link [[api_atd_serve_discover]]
	r.Register(mcp.Tool{
		Name: "atd_map",
		Description: `Three-mode tool for linking source code to ATD atoms.

Default mode (file only): extracts architectural intent from an undocumented file and recommends @spec-link tags to apply. Follows surgical attachment rules (no global headers, logic-boundary placement).

Confirm mode (file + atom): validates whether a specific file implements the given atom. Returns a confidence score and rationale. Shorthand for atd_recon without requiring explicit atom file path.

Propose mode (file + new:true): treats the file as entirely undocumented and returns a proposed new atom skeleton (id, type, layer, intent, logic) ready to be passed to atd_update.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{"type": "string", "description": "Path to the source file to analyse."},
				"atom": map[string]any{"type": "string", "description": "Confirm mode: atom ID or path to validate against the file."},
				"new":  map[string]any{"type": "boolean", "description": "Propose mode: return a new atom skeleton for the file instead of searching existing atoms."},
			},
			"required": []string{"file"},
		},
	}, func(args map[string]any) (string, error) {
		file := argString(args, "file", "")
		atom := argString(args, "atom", "")
		isNew := argBool(args, "new")
		return runMap(file, atom, config.DocsDir(), isNew)
	})

	// ── Configuration & Diagnostics ──────────────────────────────────────

	r.Register(mcp.Tool{
		Name: "atd_env",
		Description: `Unified environment smoke test: validates .atd config, checks provider connectivity, and verifies model availability.
Use to diagnose 'Connection Refused' or 'Model Not Found' errors, or to verify a new provider/model is correctly configured.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runCheck(false)
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

	r.Register(mcp.Tool{
		Name: "atd_workspace_list",
		Description: `List all projects in the current ATD workspace.
Use to discover available projects when working in a monorepo.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		if config.ActiveConfig.Workspace == nil {
			return "No workspace active", nil
		}
		out, _ := json.MarshalIndent(config.ActiveConfig.Workspace.Projects, "", "  ")
		return fmt.Sprintf("Workspace: %s\nActive Project: %s\nProjects:\n%s", 
			config.ActiveConfig.Workspace.WorkspaceName, config.ActiveConfig.ActiveProject, string(out)), nil
	})

	r.Register(mcp.Tool{
		Name: "atd_workspace_use",
		Description: `Switch the active project context in the current workspace.
Subsequent tool calls will be scoped to this project.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project": map[string]any{"type": "string", "description": "Name of the project to switch to."},
			},
			"required": []string{"project"},
		},
	}, func(args map[string]any) (string, error) {
		project := argString(args, "project", "")
		if err := config.SetProject(project); err != nil {
			return "", err
		}
		return fmt.Sprintf("Switched to project: %s", project), nil
	})

	r.Register(mcp.Tool{
		Name: "atd_workspace_stats",
		Description: `Aggregate documentation health metrics across all projects in the workspace.`,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runStats(".", config.DocsDir(), true)
	})

	r.Register(mcp.Tool{
		Name: "atd_heatmap",
		Description: `Get heat map metrics for a specific atom. 
Layers: dependency (coupling), code (implementation density), updates (instability).`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"atom": map[string]any{"type": "string", "description": "Atom ID or file path."},
			},
			"required": []string{"atom"},
		},
	}, func(args map[string]any) (string, error) {
		atomID := argString(args, "atom", "")
		return runHeatmapAtom(atomID)
	})

	r.Register(mcp.Tool{
		Name: "atd_heatmap_code",
		Description: `Get heat map metrics for a specific source file based on @spec-link density.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{"type": "string", "description": "Source file path."},
			},
			"required": []string{"file"},
		},
	}, func(args map[string]any) (string, error) {
		file := argString(args, "file", "")
		return runHeatmapCode(file)
	})

	r.Register(mcp.Tool{
		Name: "atd_heatmap_project",
		Description: `Get a project-wide heat map summary.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"layer": map[string]any{"type": "string", "description": "Heat layer: dependency, code, updates, all (default)."},
			},
		},
	}, func(args map[string]any) (string, error) {
		layer := argString(args, "layer", "all")
		return runHeatmapProject(layer)
	})
}
