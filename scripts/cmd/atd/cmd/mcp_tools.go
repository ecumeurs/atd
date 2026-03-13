package cmd

import (
	"atd-tools/config"
	"atd-tools/pkg/mcp"
)

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
			"type": "object",
			"properties": map[string]any{},
		},
	}, func(args map[string]any) (string, error) {
		return runWeave(config.DocsDir())
	})

	r.Register(mcp.Tool{
		Name:        "atd_update",
		Description: "Surgically update fields in an ATD atom file without rewriting it. Pass set as 'key=value' pairs.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file":      map[string]any{"type": "string", "description": "Absolute or relative path to the .atom.md file."},
				"set":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Frontmatter edits as 'key=value' strings, e.g. [\"status=STABLE\",\"priority=CORE\"]."},
				"intent":    map[string]any{"type": "string", "description": "New INTENT section text."},
				"logic":     map[string]any{"type": "string", "description": "New THE RULE / LOGIC section text."},
				"interface": map[string]any{"type": "string", "description": "New TECHNICAL INTERFACE section text."},
				"spec_link": map[string]any{"type": "string", "description": "Atom ID to prepend as @spec-link in a source file (requires spec_link_file)."},
				"spec_link_file": map[string]any{"type": "string", "description": "Source file path for --spec-link injection."},
			},
			"required": []string{"file"},
		},
	}, func(args map[string]any) (string, error) {
		file := argString(args, "file", "")
		intent := argString(args, "intent", "")
		logic := argString(args, "logic", "")
		iface := argString(args, "interface", "")
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

		return runUpdate(file, setPairs, intent, logic, iface, specLink, specFile)
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

	r.Register(mcp.Tool{
		Name:        "atd_verify",
		Description: "Run git diff, extract @spec-link tags, and produce an audit prompt for the IDE Agent.",
		InputSchema: map[string]any{
			"type": "object",
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
}
