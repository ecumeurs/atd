package cmd

import (
	"encoding/json"
	"fmt"

	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/mcp"
)

// mcpArgs wraps a single MCP tool call's raw arguments map and enforces
// that declared parameters are well-typed: a value present under the
// wrong JSON/Go type is a loud error rather than a silently-coerced zero
// value (test_atd_07_26.md §3.3 #2 / §8.3 item 4 -- the argBool/argString/
// argInt KNOWN DEFECT: e.g. atd_check(full="true") used to be silently
// treated as full=false, producing a diff-mode report with no error at
// all). Every handler below reads its args through the accessors here,
// then checks Err() once before dispatching to its run* function --
// "required" presence is enforced earlier still, centrally, by
// pkg/mcp.Registry.Call via each Tool's declared InputSchema "required".
type mcpArgs struct {
	tool string
	args map[string]any
	err  error
}

func newMCPArgs(tool string, args map[string]any) *mcpArgs {
	return &mcpArgs{tool: tool, args: args}
}

func (a *mcpArgs) typeErr(key, want string, got any) {
	if a.err == nil {
		a.err = fmt.Errorf("%s: argument %q must be a %s, got %T", a.tool, key, want, got)
	}
}

func (a *mcpArgs) String(key, fallback string) string {
	v, ok := a.args[key]
	if !ok || v == nil {
		return fallback
	}
	s, ok := v.(string)
	if !ok {
		a.typeErr(key, "string", v)
		return fallback
	}
	return s
}

func (a *mcpArgs) Bool(key string, fallback bool) bool {
	v, ok := a.args[key]
	if !ok || v == nil {
		return fallback
	}
	b, ok := v.(bool)
	if !ok {
		a.typeErr(key, "boolean", v)
		return fallback
	}
	return b
}

func (a *mcpArgs) Int(key string, fallback int) int {
	v, ok := a.args[key]
	if !ok || v == nil {
		return fallback
	}
	switch val := v.(type) {
	case int:
		return val
	case int32:
		return int(val)
	case int64:
		return int(val)
	case float64:
		return int(val)
	}
	a.typeErr(key, "number", v)
	return fallback
}

func (a *mcpArgs) Float64(key string, fallback float64) float64 {
	v, ok := a.args[key]
	if !ok || v == nil {
		return fallback
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	}
	a.typeErr(key, "number", v)
	return fallback
}

func (a *mcpArgs) Err() error { return a.err }

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
				"field":      map[string]any{"type": "string", "description": "Frontmatter field to search (e.g. 'type', 'status', 'id', 'layer', 'tags'). Omit to search all fields."},
				"search":     map[string]any{"type": "string", "description": "Value to match (case-insensitive substring)."},
				"paths_only": map[string]any{"type": "boolean", "description": "If true, return only a JSON array of absolute file paths."},
			},
			"required": []string{"search"},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_query", args)
		field := a.String("field", "")
		search := a.String("search", "")
		pathsOnly := a.Bool("paths_only", false)
		if err := a.Err(); err != nil {
			return "", err
		}
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
				"src":       map[string]any{"type": "string", "description": "Path to source code directory. Defaults to the current directory."},
				"docs":      map[string]any{"type": "string", "description": "Override docs directory path."},
				"gaps":      map[string]any{"type": "boolean", "description": "If true, return only STABLE atoms with zero code implementations (orphan detection)."},
				"workspace": map[string]any{"type": "boolean", "description": "If true, crawl the entire workspace."},
			},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_crawl", args)
		// Default to the loaded config's project root, not the process cwd
		// (test_atd_07_26.md §2.1/§8.3 defect #7) -- an I-1-shaped hazard for
		// a long-lived MCP server process whose cwd need not match the
		// active project.
		src := a.String("src", config.ProjectRoot())
		docs := a.String("docs", config.DocsDir())
		gaps := a.Bool("gaps", false)
		workspace := a.Bool("workspace", false)
		if err := a.Err(); err != nil {
			return "", err
		}
		return runCrawl(src, docs, gaps, workspace)
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
		explorer := exploration.NewExplorerWithConfig("", config.DocsDir(), &config.ActiveConfig)
		return explorer.Weave()
	})

	// @spec-link [[api_atd_serve_update]]
	r.Register(mcp.Tool{
		Name: "atd_update",
		Description: `Surgically modify ATD atom files — the ONLY correct way to edit .atom.md files.
Use for: creating new atoms (provide file path + all required fields), changing status/priority/layer, editing H2 sections (intent, logic, interface, expectation), injecting @spec-link tags into source code.
For batch operations, use 'filter' instead of 'file' to update all matching atoms in one call.
Pass force:true to override the STABLE+BUSINESS governance guard (mirrors the CLI's --force flag) -- otherwise updating a STABLE atom whose layer is BUSINESS is refused.
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
				"force":          map[string]any{"type": "boolean", "description": "If true, override the STABLE+BUSINESS governance guard and confirm the modification (mirrors the CLI's --force flag)."},
			},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_update", args)
		file := a.String("file", "")
		filter := a.String("filter", "")
		intent := a.String("intent", "")
		logic := a.String("logic", "")
		iface := a.String("interface", "")
		expectation := a.String("expectation", "")
		specLink := a.String("spec_link", "")
		specFile := a.String("spec_link_file", "")
		force := a.Bool("force", false)

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

		if err := a.Err(); err != nil {
			return "", err
		}

		if filter != "" {
			if file != "" {
				return "", fmt.Errorf("cannot use both 'file' and 'filter'")
			}
			return runBatchUpdate(filter, setPairs, intent, logic, iface, expectation, specLink, specFile, force)
		}

		if file == "" {
			return "", fmt.Errorf("either 'file' or 'filter' is required")
		}

		return runUpdate(file, setPairs, intent, logic, iface, expectation, specLink, specFile, force)
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
		a := newMCPArgs("atd_roadmap", args)
		// Fall back to the loaded config's project root, not the process cwd
		// (test_atd_07_26.md §2.1/§8.3 defect #7), if the caller omits "dir"
		// despite it being declared required.
		dir := a.String("dir", config.ProjectRoot())
		out := a.String("out", "roadmap.json")
		if err := a.Err(); err != nil {
			return "", err
		}
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
				"src":       map[string]any{"type": "string", "description": "Path to source code directory. Defaults to the current directory."},
				"docs":      map[string]any{"type": "string", "description": "Override docs directory path."},
				"workspace": map[string]any{"type": "boolean", "description": "If true, aggregate stats from all projects in the workspace."},
			},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_stats", args)
		// Default to the loaded config's project root, not the process cwd
		// (test_atd_07_26.md §2.1/§8.3 defect #7): runStats now honors this
		// argument (see cmd/atd/cmd/stats.go), so passing the literal "."
		// here would have reintroduced the same cwd-anchoring hazard the
		// stats.go fix just closed.
		src := a.String("src", config.ProjectRoot())
		docs := a.String("docs", config.DocsDir())
		workspace := a.Bool("workspace", false)
		if err := a.Err(); err != nil {
			return "", err
		}
		return runStats(src, docs, workspace)
	})

	// @spec-link [[api_atd_serve_check]]
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
				"semantic": map[string]any{"type": "boolean", "description": "Optional: add LLM compliance check per impl link (consumes tokens). Returns PASS/FAIL per @spec-link."},
			},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_check", args)
		base := a.String("base", "")
		target := a.String("target", "")
		full := a.Bool("full", false)
		file := a.String("file", "")
		semantic := a.Bool("semantic", false)
		if err := a.Err(); err != nil {
			return "", err
		}

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
				"starts":          map[string]any{"type": "string", "description": "Comma-separated list of root Atom IDs to begin assembly from."},
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
		a := newMCPArgs("atd_assemble", args)
		starts := a.String("starts", "")
		intent := a.String("intent", "Executive Summary")
		length := a.String("length", "default")
		structured := a.Bool("structured", false)
		asJSON := a.Bool("json", false)
		onlyParents := a.Bool("only_parents", false)
		onlyDependents := a.Bool("only_dependents", false)
		if err := a.Err(); err != nil {
			return "", err
		}
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
		a := newMCPArgs("atd_trace", args)
		atomID := a.String("atom", "")
		summary := a.Bool("summary", false)
		if err := a.Err(); err != nil {
			return "", err
		}
		// Anchor to the loaded config's project root, not the process cwd
		// (test_atd_07_26.md §2.1/§8.3 defect #7): the CLI's own `trace`
		// command passes "" here (its --src flag defaults to ""), which lets
		// exploration.NewExplorer fall back to config.ProjectRoot() -- the
		// literal "." this handler passed instead skipped that fallback and
		// anchored the walk to the MCP server process's cwd.
		return runTrace(atomID, config.DocsDir(), config.ProjectRoot(), summary)
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
				"docs": map[string]any{"type": "string", "description": "Override docs directory path."},
			},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_test_links", args)
		atomID := a.String("atom", "")
		docs := a.String("docs", config.DocsDir())
		if err := a.Err(); err != nil {
			return "", err
		}
		return runCoverageCheck("atom", atomID, "", docs, false, false, nil)
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
				"llm":  map[string]any{"type": "boolean", "description": "If true (default), route through the tiered Ollama provider. If false, return the raw prompt for IDE Agent passthrough."},
			},
			"required": []string{"file"},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_dissect", args)
		file := a.String("file", "")
		useLLM := a.Bool("llm", true)
		if err := a.Err(); err != nil {
			return "", err
		}
		return runDissect(file, useLLM)
	})

	// @spec-link [[api_atd_serve_index]]
	r.Register(mcp.Tool{
		Name: "atd_index",
		Description: `Build or refresh the semantic vector index of all source code and ATD documents.
Uses nomic-embed-text to generate embeddings stored in a SQLite database. Files unchanged since last indexing are automatically skipped (mtime-based caching).
Run before using atd_search (semantic mode), or after significant code/documentation changes to keep the index fresh.
Defaults to indexing the entire project ('all' mode) using the .atd configuration.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"dir":  map[string]any{"type": "string", "description": "Directory to crawl and index. Defaults to the current directory."},
				"db":   map[string]any{"type": "string", "description": "Path to the SQLite index database. Defaults to <docs_path>/.atd_index.db."},
				"mode": map[string]any{"type": "string", "description": "What to index: 'code', 'docs', or 'all'. Defaults to 'all'."},
			},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_index", args)
		dir := a.String("dir", ".")
		db := a.String("db", config.IndexDBPath(config.DocsDir()))
		mode := a.String("mode", "all")
		if err := a.Err(); err != nil {
			return "", err
		}
		return runIndex(dir, db, mode)
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
		a := newMCPArgs("atd_search", args)
		query := a.String("query", "")
		grep := a.String("grep", "")
		scope := a.String("scope", "all")
		pathsOnly := a.Bool("paths_only", false)
		limit := a.Int("limit", 5)
		if err := a.Err(); err != nil {
			return "", err
		}
		db := config.IndexDBPath(config.DocsDir())
		if grep != "" {
			return runGrepSearch(grep, pathsOnly)
		}
		return runSemanticSearch(query, db, limit, scope, pathsOnly)
	})

	// @spec-link [[api_atd_serve_audit]]
	r.Register(mcp.Tool{
		Name: "atd_audit",
		Description: `Audit ATD atoms for documentation quality issues.
Default mode: detect bloated atoms and semantic collisions (duplicate/overlapping atoms) across the entire docs directory.
Use during PLAN stage after creating new atoms to check for overlap.`,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"docs":      map[string]any{"type": "string", "description": "Override docs directory path."},
				"threshold": map[string]any{"type": "number", "description": "Cosine similarity threshold for collision detection (0.0-1.0). Defaults to the configured diff_similarity_threshold (or 0.85)."},
			},
		},
	}, func(args map[string]any) (string, error) {
		a := newMCPArgs("atd_audit", args)
		defaultThreshold := config.ActiveConfig.DiffSimilarityThreshold
		if defaultThreshold <= 0 {
			defaultThreshold = 0.85
		}
		docs := a.String("docs", config.DocsDir())
		threshold := a.Float64("threshold", defaultThreshold)
		if err := a.Err(); err != nil {
			return "", err
		}
		return runFullAudit(docs, threshold, false)
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
		a := newMCPArgs("atd_recon", args)
		atom := a.String("atom", "")
		candidate := a.String("candidate", "")
		if err := a.Err(); err != nil {
			return "", err
		}
		return runMap(candidate, atom, config.DocsDir(), false)
	})

	// @spec-link [[api_atd_serve_map]]
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
		a := newMCPArgs("atd_map", args)
		file := a.String("file", "")
		atom := a.String("atom", "")
		isNew := a.Bool("new", false)
		if err := a.Err(); err != nil {
			return "", err
		}
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
		a := newMCPArgs("atd_config", args)
		list := a.Bool("list", false)
		atomType := a.String("bloating_factor", "")
		task := a.String("task", "")
		model := a.String("model", "")
		if err := a.Err(); err != nil {
			return "", err
		}
		if list {
			out, _ := json.MarshalIndent(config.ActiveConfig, "", "  ")
			return string(out), nil
		}
		if atomType != "" {
			return runConfigGetBloating(atomType)
		}
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
		a := newMCPArgs("atd_workspace_use", args)
		project := a.String("project", "")
		if err := a.Err(); err != nil {
			return "", err
		}
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
		// Anchor to the loaded config's project root, not the process cwd --
		// see the atd_stats handler above for why the literal "." is unsafe
		// now that runStats honors this argument (test_atd_07_26.md §8.3
		// defect #7).
		return runStats(config.ProjectRoot(), config.DocsDir(), true)
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
		a := newMCPArgs("atd_heatmap", args)
		atomID := a.String("atom", "")
		if err := a.Err(); err != nil {
			return "", err
		}
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
		a := newMCPArgs("atd_heatmap_code", args)
		file := a.String("file", "")
		if err := a.Err(); err != nil {
			return "", err
		}
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
		a := newMCPArgs("atd_heatmap_project", args)
		layer := a.String("layer", "all")
		if err := a.Err(); err != nil {
			return "", err
		}
		return runHeatmapProject(layer)
	})
}
