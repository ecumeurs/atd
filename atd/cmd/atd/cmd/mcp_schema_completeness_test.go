package cmd

// Schema-completeness sweep (test_atd_07_26.md §3.3 #2's "greppable via a
// table the test maintains" half of the schema-honesty requirement). The
// with/without spot-checks in mcp_schema_honesty_test.go's paramCases
// demonstrate the METHOD on a representative sample; this file makes the
// COVERAGE exhaustive and self-enforcing: registeredParamsByTool below is
// the full, hand-audited set of every InputSchema property every tool
// declares today, and TestMCPContract_SchemaCompleteness fails the moment
// any registered tool's actual schema drifts from that table in either
// direction -- catching a newly added, unaccounted-for param the instant it
// ships, without needing a paramCase written for it first (test_atd_07_26.md
// §6 WP-4 AC: "adding an undocumented schema param to one tool turns the
// schema-honesty or equality test red").
import (
	"testing"

	"atd-tools/pkg/mcp"
)

// registeredParamsByTool is the exhaustive, hand-audited set of every
// InputSchema.properties key every tool in RegisterMCPTools (mcp_tools.go)
// declares as of this WP. Adding a new param to any tool's schema without
// updating this table (and ideally adding a paramCase to
// mcp_schema_honesty_test.go demonstrating it is actually read) is exactly
// the drift this test exists to catch.
var registeredParamsByTool = map[string][]string{
	"atd_query":           {"field", "search", "paths_only"},
	"atd_crawl":           {"src", "docs", "gaps", "workspace"},
	"atd_weave":           {},
	"atd_update":          {"file", "filter", "set", "intent", "logic", "interface", "expectation", "spec_link", "spec_link_file", "force"},
	"atd_roadmap":         {"dir", "out"},
	"atd_stats":           {"src", "docs", "workspace"},
	"atd_check":           {"base", "target", "full", "file", "semantic"},
	"atd_assemble":        {"starts", "intent", "length", "structured", "json", "only_parents", "only_dependents"},
	"atd_trace":           {"atom", "summary"},
	"atd_test_links":      {"atom", "docs"},
	"atd_index":           {"dir", "db", "mode"},
	"atd_search":          {"query", "grep", "scope", "limit", "paths_only"},
	"atd_audit":           {"docs", "threshold", "atom", "code", "concurrency"},
	"atd_recon":           {"atom", "candidate"},
	"atd_map":             {"file", "atom", "new"},
	"atd_env":             {},
	"atd_config":          {"list", "bloating_factor", "task", "model"},
	"atd_lint":            {},
	"atd_workspace_list":  {},
	"atd_workspace_use":   {"project"},
	"atd_workspace_stats": {},
	"atd_heatmap":         {"atom"},
	"atd_heatmap_code":    {"file"},
	"atd_heatmap_project": {"layer"},
}

// TestMCPContract_SchemaCompleteness pins test_atd_07_26.md §3.3 #2: every
// registered tool's InputSchema.properties must exactly match
// registeredParamsByTool -- no unaccounted-for additions, no stale entries
// for params that were removed. This is the mechanical, exhaustive half of
// schema honesty; paramCases in mcp_schema_honesty_test.go is the semantic
// half (does the param actually change behavior).
func TestMCPContract_SchemaCompleteness(t *testing.T) {
	r := mcp.NewRegistry()
	RegisterMCPTools(r)

	seenTools := make(map[string]bool, len(registeredParamsByTool))

	for _, tool := range r.List() {
		seenTools[tool.Name] = true

		known, hasEntry := registeredParamsByTool[tool.Name]
		if !hasEntry {
			t.Errorf("registered tool %q has no entry in registeredParamsByTool -- add one listing every InputSchema property it declares (this table must stay exhaustive)", tool.Name)
			continue
		}
		knownSet := make(map[string]bool, len(known))
		for _, k := range known {
			knownSet[k] = true
		}

		props, _ := tool.InputSchema["properties"].(map[string]any)
		for p := range props {
			if !knownSet[p] {
				t.Errorf("%s declares schema param %q that registeredParamsByTool does not account for -- add it to the table, and either add a paramCase in mcp_schema_honesty_test.go proving the handler reads it, or pin it as a new KNOWN DEFECT if it is dead", tool.Name, p)
			}
		}
		for _, k := range known {
			if _, ok := props[k]; !ok {
				t.Errorf("registeredParamsByTool claims %s declares %q but its actual InputSchema no longer does -- prune the stale entry", tool.Name, k)
			}
		}
	}

	for name := range registeredParamsByTool {
		if !seenTools[name] {
			t.Errorf("registeredParamsByTool has an entry for %q but it is not a currently registered tool -- prune it", name)
		}
	}
}
