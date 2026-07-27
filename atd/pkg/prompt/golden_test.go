package prompt

// WP-6 prompt snapshot goldens (test_atd_07_26.md §3.5 #4, §6 WP-6
// deliverable "prompt goldens"): golden-file every assembled prompt (one
// per task type) plus its JSON format schema, under
// testdata/golden/prompts/. Prompt drift silently changes product behavior;
// a golden diff makes any change a reviewed decision.
//
// Regenerate with `go test ./pkg/prompt/ -update`. All Build functions
// serialize a Go map via encoding/json (which sorts map keys), so output is
// deterministic and two consecutive -update runs produce no diff (WP-6 AC).

import (
	"encoding/json"
	"testing"

	"atd-tools/pkg/testutil"
)

// golden wraps (*SB).Golden with an empty sandbox root — prompt output
// embeds no temp paths or timestamps, so no normalization root is needed.
// findTestdataGoldenDir walks upward from this package's source dir and
// lands on atd/testdata/golden/, the same directory the WP-2 CLI goldens
// use.
func golden(t *testing.T, name, got string) {
	t.Helper()
	sb := &testutil.SB{}
	sb.Golden(t, "prompts/"+name, got)
}

// schemaJSON renders a Format() schema deterministically for golding.
func schemaJSON(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshaling format schema: %v", err)
	}
	return string(b) + "\n"
}

func TestGoldenPrompt_Recon(t *testing.T) {
	golden(t, "recon.txt", ReconBuild("## INTENT\nzzfix atom intent.", "package zzfix\nfunc Beta() {}\n"))
	golden(t, "recon.schema.json", schemaJSON(t, ReconFormat()))
}

func TestGoldenPrompt_AuditBloat(t *testing.T) {
	golden(t, "audit_bloat.txt", AuditBloatBuild("Architectural Linter", "One rule that does X and also Y.", 0.8))
	golden(t, "audit_bloat.schema.json", schemaJSON(t, AuditBloatFormat()))
}

func TestGoldenPrompt_AuditCode(t *testing.T) {
	atom := CuratedAuditAtom{
		ID:          "zzfix_rule",
		Type:        "RULE",
		Layer:       "ARCHITECTURE",
		Intent:      "zzfix intent.",
		Logic:       "zzfix logic.",
		Expectation: "zzfix expectation.",
	}
	golden(t, "audit_code_pm.txt", AuditCodeBuild(PersonaPM, atom, "package zzfix"))
	golden(t, "audit_code_techlead.txt", AuditCodeBuild(PersonaTechLead, atom, "package zzfix"))
	golden(t, "audit_code.schema.json", schemaJSON(t, AuditCodeFormat()))
}

func TestGoldenPrompt_IntentExtract(t *testing.T) {
	golden(t, "intent_extract.txt", IntentExtractBuild("package zzfix\nfunc Beta() {}\n"))
	golden(t, "intent_extract.schema.json", schemaJSON(t, IntentExtractFormat()))
}

func TestGoldenPrompt_DiscoverLinks(t *testing.T) {
	golden(t, "discover_links.txt", DiscoverLinksBuild("- [[api_zzfix_beta]]: beta api\n", "package zzfix\nfunc Beta() {}\n"))
	golden(t, "discover_links.schema.json", schemaJSON(t, DiscoverLinksFormat()))
}

func TestGoldenPrompt_Reconcile(t *testing.T) {
	golden(t, "reconcile.txt", ReconcileBuild("# existing zzfix atoms", "# inbound zzfix edits"))
	golden(t, "reconcile.schema.json", schemaJSON(t, ReconcileFormat()))
}

func TestGoldenPrompt_Compare(t *testing.T) {
	a := AtomData{HumanName: "zzfix A", Type: "RULE", Intent: "intent a", Logic: "logic a"}
	b := AtomData{HumanName: "zzfix B", Type: "RULE", Intent: "intent b", Logic: "logic b"}
	golden(t, "compare.txt", CompareBuild(a, b))
	golden(t, "compare.schema.json", schemaJSON(t, CompareFormat()))
}

func TestGoldenPrompt_Congruence(t *testing.T) {
	golden(t, "congruence.txt", CongruenceBuild("## zzfix target atom", "## zzfix sibling specs"))
	golden(t, "congruence.schema.json", schemaJSON(t, CongruenceFormat()))
}

func TestGoldenPrompt_FixSplit(t *testing.T) {
	golden(t, "fix_split.txt", FixSplitBuild("## THE RULE / LOGIC\nDoes X and also Y."))
	golden(t, "fix_split.schema.json", schemaJSON(t, FixSplitFormat()))
}

func TestGoldenPrompt_TraceSummary(t *testing.T) {
	ctx := TraceSummaryContext{
		Target: AtomBrief{
			ID: "mech_zzfix_gamma", HumanName: "zzfix Gamma", Type: "MECHANIC",
			Layer: "IMPLEMENTATION", Intent: "gamma intent", Logic: "gamma logic",
		},
		Ancestry: []AtomBrief{
			{ID: "api_zzfix_beta", HumanName: "zzfix Beta", Type: "API", Layer: "ARCHITECTURE", Intent: "beta intent", Logic: "beta logic"},
		},
		Dependents: []AtomBrief{},
	}
	golden(t, "trace_summary.txt", TraceSummaryBuild(ctx))
	golden(t, "trace_summary.schema.json", schemaJSON(t, TraceSummaryFormat()))
}

func TestGoldenPrompt_Assemble(t *testing.T) {
	golden(t, "assemble.txt", AssembleBuild("summarize zzfix", "short", "fragment one\nfragment two"))
	golden(t, "assemble_layer_pass.txt", LayerPassBuild("BUSINESS", "short", "fragment one\nfragment two"))
	golden(t, "assemble_final.txt", FinalAssembleBuild("summarize zzfix", "short", "biz summary", "arch summary", "impl summary"))
	golden(t, "assemble.schema.json", schemaJSON(t, AssembleFormat()))
}
