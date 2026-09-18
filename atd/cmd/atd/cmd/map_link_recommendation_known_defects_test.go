package cmd

// Regression tests for `atd map` discover mode's link-recommendation leg
// (the second LLM call in runMapDiscover, cmd/atd/cmd/map.go, which builds
// the "### Recommended Atom Links" section). Distinct from the already-fixed
// intent-extraction leg covered by TestRunMapPropose_MalformedResponses /
// TestRunMapDiscover_MalformedIntentResponse in map_llm_test.go.
//
// These started life as KNOWN DEFECT pins. Both defects are now fixed:
// candidates are validated against the atom-ID naming convention and the real
// registry (cmd/atd/cmd/map_link_validation.go), and a truncated response is
// salvaged instead of crashing, with pkg/prompt/discover_links.go emitting the
// structured field before the free-form rationale so there is something left to
// salvage. The pins below now assert the fixed behavior.
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
// The atom it writes (rule_seed) doubles as the one genuinely valid
// recommendation candidate available in these sandboxes.
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

// TestRunMapDiscover_LinkRecommendation_HallucinationDropped covers the defect
// from failures/20260901_atd_map_hallucinated_code_values.md and its 20260917
// recurrence: runMapDiscover used to write every entry of the model's
// `recommendations` array straight into the output as `- [[<entry>]]`, with no
// naming-convention check and no registry cross-check, so a code-review
// sentence, a refusal or a fabricated URL was presented exactly like a genuine
// atom ID.
//
// Post-fix: non-atom-ID entries are dropped (and reported as discarded
// suggestions, never as links), while a real registry atom passes through.
func TestRunMapDiscover_LinkRecommendation_HallucinationDropped(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)
	fake := fakeprovider.InstallOllama(t)
	seedOneAtomIndex(t, docsDir, fake)

	hallucinated := []string{
		"Add comments to explain complex logic",
		"https://github.com/UPILab/UPILab/blob/master/commit/1234567890abcdef#L1",
		"I don't recommend any IDs from the provided registry.",
	}
	fake.SetSequence(
		`{"intent": "does something with the file"}`,
		`{"recommendations": ["`+hallucinated[0]+`", "`+hallucinated[1]+`", "`+hallucinated[2]+`", "rule_seed"], "rationale": "some prose"}`,
	)

	out, err := runMapDiscover("src/foo.go", "package main", docsDir)
	if err != nil {
		t.Fatalf("runMapDiscover: unexpected error: %v", err)
	}

	for _, h := range hallucinated {
		if strings.Contains(out, "[["+h+"]]") {
			t.Errorf("non-atom-ID recommendation %q was rendered as an atom link; output:\n%s", h, out)
		}
	}
	if !strings.Contains(out, "- [[rule_seed]]") {
		t.Errorf("the one real registry atom should survive validation; output:\n%s", out)
	}
	if !strings.Contains(out, "Discarded Suggestions") {
		t.Errorf("expected a discarded-suggestions note explaining the drops; output:\n%s", out)
	}
}

// TestRunMapDiscover_LinkRecommendation_UnknownAtomIDDropped covers the second
// half of the validation mandate: an entry can be perfectly atom-ID shaped and
// still be a hallucination. It must be cross-checked against the project's real
// atom registry (reusing exploration.Explorer's resolution, the same lookup
// behind `atd query --field id`) and dropped when it resolves to nothing.
func TestRunMapDiscover_LinkRecommendation_UnknownAtomIDDropped(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)
	fake := fakeprovider.InstallOllama(t)
	seedOneAtomIndex(t, docsDir, fake)

	fake.SetSequence(
		`{"intent": "does something with the file"}`,
		`{"recommendations": ["mech_totally_made_up_atom"], "rationale": "confident nonsense"}`,
	)

	out, err := runMapDiscover("src/foo.go", "package main", docsDir)
	if err != nil {
		t.Fatalf("runMapDiscover: unexpected error: %v", err)
	}
	if strings.Contains(out, "[[mech_totally_made_up_atom]]") {
		t.Errorf("a well-shaped but nonexistent atom ID must not be recommended; output:\n%s", out)
	}
	if !strings.Contains(out, "no such atom") {
		t.Errorf("expected the drop reason to name the failed registry lookup; output:\n%s", out)
	}
}

// TestRunMapDiscover_LinkRecommendation_AllInvalid pins the behavior when
// nothing survives validation: an explicit empty recommendation list, no
// crash, no substituted hallucination — and the rationale suppressed, since
// with no atom to anchor it, it is exactly the unverifiable free-form prose the
// 20260901 report found to be confidently wrong.
func TestRunMapDiscover_LinkRecommendation_AllInvalid(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)
	fake := fakeprovider.InstallOllama(t)
	seedOneAtomIndex(t, docsDir, fake)

	fake.SetSequence(
		`{"intent": "does something with the file"}`,
		`{"recommendations": ["Use more descriptive variable names"], "rationale": "The Movement property is 5/5 here."}`,
	)

	out, err := runMapDiscover("src/foo.go", "package main", docsDir)
	if err != nil {
		t.Fatalf("runMapDiscover must degrade to an empty recommendation list, not error: %v", err)
	}
	if !strings.Contains(out, "### Recommended Atom Links") {
		t.Errorf("expected the recommendations section to still be rendered; output:\n%s", out)
	}
	if !strings.Contains(out, "(none") {
		t.Errorf("expected an explicit empty-recommendations marker; output:\n%s", out)
	}
	if strings.Contains(out, "### Rationale") {
		t.Errorf("rationale must be suppressed when no recommendation survived; output:\n%s", out)
	}
	if strings.Contains(out, "[[Use more descriptive") {
		t.Errorf("the hallucinated entry leaked into the links; output:\n%s", out)
	}
}

// TestRunMapDiscover_LinkRecommendation_MalformedJSON confirms what happens
// when the link-recommendation LLM call returns a payload with nothing
// salvageable — the failure mode described in
// failures/20260901_atd_map_malformed_json_from_link_recommender.md,
// failures/20260902_atd_map_malformed_json_link_recommender_recurrence.md,
// and failures/20260917_atd_map_hallucinated_links_and_malformed_json_recurrence.md
// (the battle.go case): the response was cut off inside the free-form
// `rationale`, which the old schema emitted BEFORE the structured result.
//
// Ground truth from reading cmd/atd/cmd/map.go: the malformed-JSON branch
// already returns a wrapped error ("link recommendation returned malformed
// JSON: ..."), and cmd/atd/cmd/root.go's Execute() calls os.Exit(1) on any RunE
// error — so at THIS layer the command does NOT swallow the failure into exit
// 0. The field reports' "exit code 0 despite the functional failure"
// observation was captured from a batched/backgrounded multi-command shell
// run, not a single `atd map` invocation, so it most likely reflects the
// wrapping shell/harness's own exit status rather than a defect in this
// function or in Execute(). This test pins the correct current behavior at the
// layer that can be tested deterministically, so a future regression that
// starts swallowing this error is caught.
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

// TestRunMapDiscover_LinkRecommendation_TruncatedAfterRecommendations is the
// payoff of the schema reordering in pkg/prompt/discover_links.go: with
// `recommendations` emitted first, a response truncated inside the trailing
// rationale still carries a complete, usable atom-ID list. It must be salvaged
// (and flagged as truncated) rather than crashing the whole `atd map` call.
func TestRunMapDiscover_LinkRecommendation_TruncatedAfterRecommendations(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)
	fake := fakeprovider.InstallOllama(t)
	seedOneAtomIndex(t, docsDir, fake)

	fake.SetSequence(
		`{"intent": "does something with the file"}`,
		`{"recommendations": ["rule_seed"], "rationale": "The provided Go code defines a`,
	)

	out, err := runMapDiscover("src/foo.go", "package main", docsDir)
	if err != nil {
		t.Fatalf("a truncated rationale must not fail the call once the ids are complete: %v", err)
	}
	if !strings.Contains(out, "- [[rule_seed]]") {
		t.Errorf("expected the salvaged atom ID in the output; got:\n%s", out)
	}
	if !strings.Contains(out, "truncated") {
		t.Errorf("expected the output to flag the truncated response; got:\n%s", out)
	}
}

// ── unit-level coverage for the validation/salvage helpers ────────────────

func TestIsAtomIDShaped(t *testing.T) {
	valid := []string{
		"rule_seed",
		"mech_action_economy_action_cost_rules",
		"api_auth_login",
		"upsilonapi:rule_matchmaking_single_queue",
		"shared:uc_match_resolution",
		"mech_go_battle_v2",
	}
	for _, s := range valid {
		if !isAtomIDShaped(s) {
			t.Errorf("expected %q to be accepted as atom-ID shaped", s)
		}
	}

	invalid := []string{
		"",
		"Add comments to explain complex logic",
		"Use more descriptive variable names",
		"https://github.com/UPILab/UPILab/blob/master/commit/1234567890abcdef#L1",
		"I don't recommend any IDs from the provided registry.",
		"RULE_SEED",
		"noseparator",
		"trailing_",
		strings.Repeat("a_b", 60), // absurdly long "sentence_without_spaces"
	}
	for _, s := range invalid {
		if isAtomIDShaped(s) {
			t.Errorf("expected %q to be rejected as an atom ID", s)
		}
	}
}

func TestNormalizeCandidate(t *testing.T) {
	cases := map[string]string{
		"  rule_seed ":    "rule_seed",
		"[[rule_seed]]":   "rule_seed",
		"- [[rule_seed]]": "rule_seed",
		"[[ rule_seed ]]": "rule_seed",
	}
	for in, want := range cases {
		if got := normalizeCandidate(in); got != want {
			t.Errorf("normalizeCandidate(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSalvageRecommendations(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "truncated_in_rationale",
			raw:  `{"recommendations": ["rule_a", "mech_b"], "rationale": "The provided Go code def`,
			want: []string{"rule_a", "mech_b"},
		},
		{
			name: "truncated_mid_element_drops_partial",
			raw:  `{"recommendations": ["rule_a", "mech_`,
			want: []string{"rule_a"},
		},
		{
			name: "rationale_first_truncated_before_ids",
			raw:  `{"rationale": "The provided Go code defines a`,
			want: nil,
		},
		{
			name: "empty_array",
			raw:  `{"recommendations": [`,
			want: nil,
		},
		{
			name: "word_only_inside_prose",
			raw:  `{"rationale": "I will list recommendations for this file`,
			want: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := salvageRecommendations(c.raw)
			if len(got) != len(c.want) {
				t.Fatalf("salvageRecommendations(%q) = %v, want %v", c.raw, got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("salvageRecommendations(%q) = %v, want %v", c.raw, got, c.want)
				}
			}
		})
	}
}

func TestSalvageRationale(t *testing.T) {
	if got := salvageRationale(`{"recommendations": ["rule_a"], "rationale": "Half a senten`); got != "Half a senten" {
		t.Errorf("expected the truncated rationale text, got %q", got)
	}
	if got := salvageRationale(`{"recommendations": ["rule_a"`); got != "" {
		t.Errorf("expected no rationale, got %q", got)
	}
}
