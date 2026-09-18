package cmd

// KNOWN DEFECT pins for `atd map` discover mode's link-recommendation leg
// (the second LLM call in runMapDiscover, cmd/atd/cmd/map.go, which builds
// the "### Recommended Atom Links" section). Distinct from the already-fixed
// intent-extraction leg covered by TestRunMapPropose_MalformedResponses /
// TestRunMapDiscover_MalformedIntentResponse in map_llm_test.go.
//
// Field reports: failures/20260901_atd_map_hallucinated_code_values.md,
// failures/20260901_atd_map_malformed_json_from_link_recommender.md,
// failures/20260902_atd_map_malformed_json_link_recommender_recurrence.md,
// failures/20260917_atd_map_hallucinated_links_and_malformed_json_recurrence.md.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/indexer"
	"atd-tools/pkg/testutil/fakeprovider"
)

// seedOneAtomIndex writes a single real atom into docsDir and builds a
// semantic index over it, so runMapDiscover's exploration.Search step (which
// errors loudly on a missing/empty index — see exploration.ErrIndexMissing)
// succeeds and the discover flow reaches the link-recommendation LLM call.
func seedOneAtomIndex(t *testing.T, docsDir string, embed *fakeprovider.Ollama) {
	t.Helper()

	atomContent := `---
id: rule_seed
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Seed atom so the semantic index has at least one row.

## THE RULE / LOGIC
n/a

## EXPECTATION
n/a
`
	if err := os.WriteFile(filepath.Join(docsDir, "rule_seed.atom.md"), []byte(atomContent), 0644); err != nil {
		t.Fatalf("writing seed atom: %v", err)
	}

	embed.SetEmbedVector([]float32{0.1, 0.2, 0.3})

	dbPath := config.IndexDBPath(docsDir)
	if _, err := indexer.Index(docsDir, dbPath, "docs"); err != nil {
		t.Fatalf("seeding semantic index: %v", err)
	}
}

// TestRunMapDiscover_LinkRecommendation_HallucinationNotValidated pins the
// KNOWN DEFECT from failures/20260901_atd_map_hallucinated_code_values.md
// and its 20260917 recurrence: runMapDiscover's link-recommendation call
// (cmd/atd/cmd/map.go ~line 307-324) unmarshals `{"recommendations": [...],
// "rationale": "..."}` and then, for every entry in Recommendations, writes
// it straight into the output as `- [[<entry>]]` with NO validation against
// the project's atom-ID naming convention (snake_case `<type>_<slug>`) and
// NO cross-check against the real atom registry (e.g. via
// `atd query --field id --search <candidate>`, as the field reports
// suggest). A prose sentence, a refusal, or a fabricated URL passes through
// exactly as if it were a genuine atom ID.
//
// If this test ever starts failing (i.e. the hallucinated string above no
// longer appears verbatim in the output), treat it as a signal the
// recommendation path gained validation and relax/remove this pin.
func TestRunMapDiscover_LinkRecommendation_HallucinationNotValidated(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)
	fake := fakeprovider.InstallOllama(t)
	seedOneAtomIndex(t, docsDir, fake)

	const hallucinated = "Add comments to explain complex logic"
	fake.SetSequence(
		`{"intent": "does something with the file"}`,
		`{"recommendations": ["`+hallucinated+`"], "rationale": "n/a"}`,
	)

	out, err := runMapDiscover("src/foo.go", "package main", docsDir)
	if err != nil {
		t.Fatalf("runMapDiscover: unexpected error: %v", err)
	}

	want := "- [[" + hallucinated + "]]"
	if !strings.Contains(out, want) {
		t.Errorf("KNOWN DEFECT expectation changed: non-atom-ID recommendation %q no longer passed through unvalidated into the output (got: %s) — the link-recommendation leg must have gained naming-convention/registry validation; if intentional, relax this pin", want, out)
	}
}

// TestRunMapDiscover_LinkRecommendation_MalformedJSON confirms, at the unit
// level, what actually happens when the link-recommendation LLM call
// returns truncated/malformed JSON — the failure mode described in
// failures/20260901_atd_map_malformed_json_from_link_recommender.md,
// failures/20260902_atd_map_malformed_json_link_recommender_recurrence.md,
// and failures/20260917_atd_map_hallucinated_links_and_malformed_json_recurrence.md
// (the battle.go case).
//
// Ground truth from reading cmd/atd/cmd/map.go: the malformed-JSON branch
// (~line 311-318) already returns a wrapped error
// ("link recommendation returned malformed JSON: ..."), and
// cmd/atd/cmd/root.go's Execute() calls os.Exit(1) on any RunE error — so at
// THIS layer the command does NOT swallow the failure into exit 0. The field
// reports' "exit code 0 despite the functional failure" observation was
// captured from a batched/backgrounded multi-command shell run, not a
// single `atd map` invocation, so it most likely reflects the wrapping
// shell/harness's own exit status rather than a defect in this function or
// in Execute(). This test pins the correct current behavior at the layer
// that can be tested deterministically, so a future regression that starts
// swallowing this error is caught.
func TestRunMapDiscover_LinkRecommendation_MalformedJSON(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)
	fake := fakeprovider.InstallOllama(t)
	seedOneAtomIndex(t, docsDir, fake)

	fake.SetSequence(
		`{"intent": "does something with the file"}`,
		`{"rationale": "The provided Go code defines a`, // truncated mid-string, unparseable
	)

	out, err := runMapDiscover("src/foo.go", "package main", docsDir)
	if err == nil {
		t.Fatalf("expected an error for malformed link-recommendation JSON, got output: %q", out)
	}
	if !strings.Contains(err.Error(), "link recommendation returned malformed JSON") {
		t.Errorf("expected error naming the malformed link-recommendation JSON, got: %v", err)
	}
}
