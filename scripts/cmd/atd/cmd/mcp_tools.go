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
func RegisterMCPTools(r *mcp.Registry) {
	r.Register(mcp.Tool{
		Name:        "atd_query",
		Description: "Search ATD atoms by frontmatter field value. Returns JSON array of matching atoms.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"field":  map[string]any{"type": "string", "description": "Frontmatter field to search (e.g. 'type', 'status', 'id')."},
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
		Name:        "atd_crawl",
		Description: "Crawl ATD docs and source code. Returns a dependency graph JSON. Set gaps=true to list STABLE atoms with no implementations.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"src":  map[string]any{"type": "string", "description": "Path to source code directory (optional)."},
				"gaps": map[string]any{"type": "boolean", "description": "If true, return only orphaned STABLE atoms."},
				"docs": map[string]any{"type": "string", "description": "Override docs directory path."},
			},
		},
	}, func(args map[string]any) (string, error) {
		src := argString(args, "src", "")
		docs := argString(args, "docs", config.DocsDir())
		gaps := argBool(args, "gaps")
		return runCrawl(src, docs, gaps)
	})

	r.Register(mcp.Tool{
		Name:        "atd_weave",
		Description: "Populate the dependents[] array in ATD atoms by scanning parents references. Bi-directional link weaving.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runWeave(config.DocsDir())
	})

	r.Register(mcp.Tool{
		Name:        "atd_update",
		Description: "Surgically update fields in an ATD atom file without rewriting it. Pass set as 'key=value' pairs. File and filter are mutually exclusive.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file":           map[string]any{"type": "string", "description": "Absolute or relative path to the .atom.md file. Optional if filter is provided."},
				"filter":         map[string]any{"type": "string", "description": "Filter atoms to update instead of a single file (e.g. 'status=DRAFT,type=RULE'). Optional if file is provided."},
				"set":            map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Frontmatter edits as 'key=value' strings, e.g. [\"status=STABLE\",\"priority=CORE\"]."},
				"intent":         map[string]any{"type": "string", "description": "New INTENT section text."},
				"logic":          map[string]any{"type": "string", "description": "New THE RULE / LOGIC section text."},
				"interface":      map[string]any{"type": "string", "description": "New TECHNICAL INTERFACE section text."},
				"expectation":    map[string]any{"type": "string", "description": "New EXPECTATION section text."},
				"spec_link":      map[string]any{"type": "string", "description": "Atom ID to prepend as @spec-link in a source file (requires spec_link_file)."},
				"spec_link_file": map[string]any{"type": "string", "description": "Source file path for --spec-link injection."},
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
		Name:        "atd_roadmap",
		Description: "Scan a source directory and build a complexity roadmap JSON identifying high-density files.",
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

	// @spec-link [[atd_serve_stats]]
	r.Register(mcp.Tool{
		Name:        "atd_stats",
		Description: "Produce quantitative documentation health metrics: total atoms, atoms by type, status, domain, coverage ratio, and orphan count.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"src":  map[string]any{"type": "string", "description": "Path to source code directory (optional)."},
				"docs": map[string]any{"type": "string", "description": "Override docs directory path."},
			},
		},
	}, func(args map[string]any) (string, error) {
		src := argString(args, "src", ".")
		docs := argString(args, "docs", config.DocsDir())
		return runStats(src, docs)
	})

	r.Register(mcp.Tool{
		Name:        "atd_verify",
		Description: "Run git diff, extract @spec-link tags, and produce an audit prompt for the IDE Agent.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runVerify(config.DocsDir())
	})

	r.Register(mcp.Tool{
		Name:        "atd_assemble",
		Description: "Stitch ATD atoms together into a single narrative or technical document.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"starts":   map[string]any{"type": "string", "description": "Comma-separated list of Root Atom IDs (source atoms)."},
				"purpose":  map[string]any{"type": "string", "description": "Optional purpose to wrap the output in <System Objective> tags."},
				"snapshot": map[string]any{"type": "boolean", "description": "If true, delegate narrative generation to IDE Agent (returns a task ID)."},
				"theme":    map[string]any{"type": "string", "description": "Theme for snapshot narrative (defaults to 'Executive Summary')."},
				"docs":     map[string]any{"type": "string", "description": "Override docs directory path."},
			},
			"required": []string{"starts"},
		},
	}, func(args map[string]any) (string, error) {
		starts := argString(args, "starts", "")
		purpose := argString(args, "purpose", "")
		snapshot := argBool(args, "snapshot")
		theme := argString(args, "theme", "Executive Summary")
		docs := argString(args, "docs", config.DocsDir())
		return runAssemble(starts, purpose, snapshot, theme, docs)
	})

	r.Register(mcp.Tool{
		Name:        "atd_test_links",
		Description: "Audit @test-link tags in source code to find which atoms are verified by which tests.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"src":  map[string]any{"type": "string", "description": "Path to source code to scan."},
				"atom": map[string]any{"type": "string", "description": "Optional: Filter for a specific Atom ID."},
				"docs": map[string]any{"type": "string", "description": "Override docs directory path."},
			},
		},
	}, func(args map[string]any) (string, error) {
		src := argString(args, "src", ".")
		atomID := argString(args, "atom", "")
		docs := argString(args, "docs", config.DocsDir())
		return runTestLinks(src, atomID, docs)
	})

	// --- LLM tools added as part of MCP preamble ---

	r.Register(mcp.Tool{
		Name:        "atd_dissect",
		Description: "Dissect a source code or documentation file into atomic boundaries. Set llm=true to route through Ollama; false returns the prompt for IDE Agent passthrough.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{"type": "string", "description": "Path to the file to dissect."},
				"llm":  map[string]any{"type": "boolean", "description": "If true, route through tiered Ollama provider. Defaults to false (IDE Agent passthrough)."},
			},
			"required": []string{"file"},
		},
	}, func(args map[string]any) (string, error) {
		file := argString(args, "file", "")
		useLLM := argBool(args, "llm")
		return runDissect(file, useLLM)
	})

	r.Register(mcp.Tool{
		Name:        "atd_index",
		Description: "Build a semantic vector index of source code and/or ATD documents using nomic-embed-text. Requires a local or remote Ollama provider.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"dir":  map[string]any{"type": "string", "description": "Directory to crawl and index. Defaults to current directory."},
				"db":   map[string]any{"type": "string", "description": "Path to SQLite database. Defaults to <docs_path>/.atd_index.db."},
				"mode": map[string]any{"type": "string", "description": "What to index: 'code', 'docs', or 'all'. Defaults to 'code'."},
			},
		},
	}, func(args map[string]any) (string, error) {
		dir := argString(args, "dir", ".")
		db := argString(args, "db", "")
		mode := argString(args, "mode", "code")
		if db == "" {
			db = config.DocsDir() + "/.atd_index.db"
		}
		return captureStdout(func() error {
			return runIndex(dir, db, mode)
		})
	})

	r.Register(mcp.Tool{
		Name:        "atd_search",
		Description: "Search the indexed codebase semantically (requires a built index) or via keyword grep.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"query": map[string]any{"type": "string", "description": "Semantic search query (uses Nomic embeddings)."},
				"grep":  map[string]any{"type": "string", "description": "Literal keyword search across project files."},
				"db":    map[string]any{"type": "string", "description": "Path to SQLite index database."},
				"limit": map[string]any{"type": "integer", "description": "Number of semantic results to return. Defaults to 5."},
				"scope": map[string]any{"type": "string", "description": "Search scope: 'code', 'docs', or 'all'. Defaults to 'all'."},
			},
		},
	}, func(args map[string]any) (string, error) {
		query := argString(args, "query", "")
		grep := argString(args, "grep", "")
		db := argString(args, "db", "")
		scope := argString(args, "scope", "all")
		limitRaw, _ := args["limit"]
		limit := 5
		if f, ok := limitRaw.(float64); ok {
			limit = int(f)
		}
		if db == "" {
			db = config.DocsDir() + "/.atd_index.db"
		}
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
		Name:        "atd_audit",
		Description: "Audit ATD atoms for documentation bloat and semantic collisions. Can also check code compliance against a specific atom.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"docs":      map[string]any{"type": "string", "description": "Override docs directory path."},
				"threshold": map[string]any{"type": "number", "description": "Cosine similarity threshold for collision detection (0.0–1.0). Defaults to config value."},
				"code":      map[string]any{"type": "string", "description": "Path to code file for compliance mode (requires 'atom')."},
				"atom":      map[string]any{"type": "string", "description": "Path to atom file for compliance mode (requires 'code')."},
			},
		},
	}, func(args map[string]any) (string, error) {
		docs := argString(args, "docs", config.DocsDir())
		code := argString(args, "code", "")
		atom := argString(args, "atom", "")
		thresholdRaw, _ := args["threshold"]
		threshold := 0.0
		if f, ok := thresholdRaw.(float64); ok {
			threshold = f
		}
		if threshold <= 0 {
			threshold = config.ActiveConfig.DiffSimilarityThreshold
			if threshold <= 0 {
				threshold = 0.85
			}
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
		Name:        "atd_recon",
		Description: "Semantic archaeology: validate whether a candidate source file implements a specific ATD atom.",
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
		Name:        "atd_discover",
		Description: "Extract architectural intent from an undocumented source file, search the ATD index, and recommend @spec-link tags to apply.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{"type": "string", "description": "Path to the undocumented source file."},
				"docs": map[string]any{"type": "string", "description": "Override docs directory path."},
			},
			"required": []string{"file"},
		},
	}, func(args map[string]any) (string, error) {
		file := argString(args, "file", "")
		docs := argString(args, "docs", config.DocsDir())
		return captureStdout(func() error {
			out, err := runDiscover(file, docs)
			if err != nil {
				return err
			}
			fmt.Println(out)
			return nil
		})
	})

	r.Register(mcp.Tool{
		Name:        "atd_check",
		Description: "Check ATD configuration and model availability across all providers.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runCheck()
	})

	r.Register(mcp.Tool{
		Name:        "atd_config",
		Description: "View or modify .atd configuration. Use 'list':true for full config, or 'bloating_factor' to get bloating factor.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task":            map[string]any{"type": "string", "description": "Task name to update (optional)."},
				"model":           map[string]any{"type": "string", "description": "Model name to assign to the task (required with task)."},
				"list":            map[string]any{"type": "boolean", "description": "List current config as JSON."},
				"bloating_factor": map[string]any{"type": "string", "description": "Atom type to get bloating factor for (e.g. 'REQUIREMENT')."},
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

	// @spec-link [[atd_serve_lint]]
	r.Register(mcp.Tool{
		Name:        "atd_lint",
		Description: "Perform deterministic structural validation on ATD atoms.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"docs": map[string]any{"type": "string", "description": "Override docs directory path. Defaults to configured docs path."},
			},
		},
	}, func(args map[string]any) (string, error) {
		docs := argString(args, "docs", config.DocsDir())
		return runLint(docs)
	})
}
