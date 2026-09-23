package audit

// WP-6 (test_atd_07_26.md §3.5, §6): deterministic logic tests for
// RunFullAudit's two phases (bloat detection, collision detection) using
// pkg/testutil/fakeprovider instead of a live Ollama, plus S12
// parser-robustness cases for the bloat-detection JSON parse.
//
// pkg/audit had zero test files before this change (test_atd_07_26.md §1.1
// lists it under "Zero tests" alongside pkg/indexer, pkg/llmservice,
// pkg/mcp, pkg/chat).

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/testutil/fakeprovider"
)

func writeFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestRunFullAudit_BloatDetection drives RunFullAudit's Phase 1 with a fake
// Ollama that reports one atom bloated and one clean, and asserts the
// per-atom classification the report text carries — pinning both branches
// deterministically instead of the previous zero test files.
func TestRunFullAudit_BloatDetection(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)

	dir := t.TempDir()
	writeFile(t, dir, "zzfix_bloated.atom.md", `---
id: zzfix_bloated
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Bloated intent that does X and also Y.

## THE RULE / LOGIC
Bloated logic that does X and also Y.

## EXPECTATION
n/a
`)
	writeFile(t, dir, "zzfix_clean.atom.md", `---
id: zzfix_clean
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
One clean rule.

## THE RULE / LOGIC
One clean rule's logic.

## EXPECTATION
n/a
`)

	// Order of processing is deterministic (filepath.Glob sorts lexically),
	// so "zzfix_bloated..." precedes "zzfix_clean...". Each atom triggers two
	// Generate calls (intent judge, then logic judge); the embed call
	// (Phase 2's collision input) is a distinct seam and RunFullAudit
	// tolerates its error (`emb, _ := ollama.QueryEmbed(...)`), so leaving
	// EmbedFunc unconfigured is fine for this Phase-1-only test.
	fake.SetSequence(
		`{"is_bloated": true, "reason": "compound rule"}`, // bloated: intent
		`{"is_bloated": true, "reason": "compound rule"}`, // bloated: logic
		`{"is_bloated": false, "reason": "single rule"}`,  // clean: intent
		`{"is_bloated": false, "reason": "single rule"}`,  // clean: logic
	)

	report, err := RunFullAudit(dir, 0.85, false, 1)
	if err != nil {
		t.Fatalf("RunFullAudit: %v", err)
	}

	if !strings.Contains(report.Text, "Auditing: zzfix_bloated.atom.md ... [BLOATED]") {
		t.Errorf("expected zzfix_bloated to be classified BLOATED, got:\n%s", report.Text)
	}
	if !strings.Contains(report.Text, "Auditing: zzfix_clean.atom.md ... [PASS]") {
		t.Errorf("expected zzfix_clean to be classified PASS, got:\n%s", report.Text)
	}
}

// TestRunFullAudit_CollisionDetection drives Phase 2 with two atoms of the
// same type, unrelated (no shared parent, no ancestor relationship), and a
// fake embedder returning near-identical fixed vectors for both — this must
// surface as a [COLLISION] line. Bloat checking is disabled per atom
// (`bloating: off`) so the test isolates the collision path from Phase 1.
func TestRunFullAudit_CollisionDetection(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)

	dir := t.TempDir()
	atomA := `---
id: zzfix_collide_a
type: RULE
status: DRAFT
bloating: off
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Collision candidate A.

## THE RULE / LOGIC
Collision candidate A logic.

## EXPECTATION
n/a
`
	atomB := `---
id: zzfix_collide_b
type: RULE
status: DRAFT
bloating: off
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Collision candidate B.

## THE RULE / LOGIC
Collision candidate B logic.

## EXPECTATION
n/a
`
	writeFile(t, dir, "zzfix_collide_a.atom.md", atomA)
	writeFile(t, dir, "zzfix_collide_b.atom.md", atomB)

	// Embed keys are the full raw file content (audit.go: `content, _ :=
	// os.ReadFile(f); emb, _ = ollama.QueryEmbed(string(content))`) — key on
	// the exact bytes so the fake fails loudly (not silently) if the file
	// content ever drifts from what this test expects.
	fake.SetEmbedTable(map[string][]float32{
		atomA: {1, 0, 0},
		atomB: {1, 0, 0.001}, // near-identical, unrelated, same type -> collision
	})

	report, err := RunFullAudit(dir, 0.85, false, 1)
	if err != nil {
		t.Fatalf("RunFullAudit: %v", err)
	}

	if !strings.Contains(report.Text, "[COLLISION] zzfix_collide_a.atom.md <--> zzfix_collide_b.atom.md") &&
		!strings.Contains(report.Text, "[COLLISION] zzfix_collide_b.atom.md <--> zzfix_collide_a.atom.md") {
		t.Errorf("expected a [COLLISION] line between the two near-identical atoms, got:\n%s", report.Text)
	}
}

// TestRunFullAudit_BloatParser_MalformedJSON is the S12 parser-robustness
// suite for pkg/audit's bloat-detection parser (test_atd_07_26.md §3.5,
// S12): feed the fake malformed LLM output and assert the parser's handling
// is loud (never a silent zero-value "PASS").
func TestRunFullAudit_BloatParser_MalformedJSON(t *testing.T) {
	cases := []struct {
		name string
		resp string
	}{
		{"markdown_fenced", "```json\n{\"is_bloated\": true, \"reason\": \"x\"}\n```"},
		{"trailing_prose", "Here is my analysis: {\"is_bloated\": true, \"reason\": \"x\"}"},
		{"truncated", `{"is_bloated": true, "reason": "x`},
		{"wrong_field_type", `{"is_bloated": "yes", "reason": "x"}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fake := fakeprovider.InstallOllama(t)

			dir := t.TempDir()
			writeFile(t, dir, "zzfix_malformed.atom.md", `---
id: zzfix_malformed
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Some intent.

## THE RULE / LOGIC
Some logic.

## EXPECTATION
n/a
`)

			// Both the intent and logic Query calls get the same malformed
			// response; either is enough to trip the parser.
			fake.SetJSON(c.resp)

			report, err := RunFullAudit(dir, 0.85, false, 1)
			if err != nil {
				t.Fatalf("RunFullAudit: %v", err)
			}

			// pkg/audit/audit.go already handles this correctly: on
			// json.Unmarshal error it writes an [ERROR] line and `continue`s
			// rather than defaulting bloatResult to "PASS". Pin that
			// behavior: the atom must NOT show up as silently PASSed, and an
			// [ERROR] line must be present.
			if strings.Contains(report.Text, "Auditing: zzfix_malformed.atom.md ... [PASS]") {
				t.Errorf("malformed LLM response (%s) must never silently classify as PASS:\n%s", c.name, report.Text)
			}
			if !strings.Contains(report.Text, "[ERROR] Failed to parse") {
				t.Errorf("expected a loud [ERROR] line for malformed response (%s), got:\n%s", c.name, report.Text)
			}
		})
	}
}

// TestRunFullAudit_BloatParser_MissingKey pins the fix for a genuine defect:
// unlike syntactically-invalid JSON (handled loudly above), a syntactically
// VALID JSON object that simply omits "is_bloated" used to unmarshal with no
// error at all — Go's encoding/json does not require declared fields to be
// present — leaving bI.IsBloated at its zero value (false) and silently
// classifying the atom PASS even though the model's response was unusable.
//
// FIXED (test_atd_07_26.md §8.3 #9a): audit.go's hasKey helper probes the raw
// response into a map[string]json.RawMessage before trusting the typed
// Unmarshal's silence, and treats a missing "is_bloated" key the same as
// malformed JSON — a loud [ERROR] line, never a silent PASS.
func TestRunFullAudit_BloatParser_MissingKey(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)

	dir := t.TempDir()
	writeFile(t, dir, "zzfix_missingkey.atom.md", `---
id: zzfix_missingkey
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Some intent.

## THE RULE / LOGIC
Some logic.

## EXPECTATION
n/a
`)

	// Valid JSON, but missing the "is_bloated" key the parser depends on.
	fake.SetJSON(`{"reason": "the model forgot the required field"}`)

	report, err := RunFullAudit(dir, 0.85, false, 1)
	if err != nil {
		t.Fatalf("RunFullAudit: %v", err)
	}

	if strings.Contains(report.Text, "Auditing: zzfix_missingkey.atom.md ... [PASS]") {
		t.Errorf("a valid-JSON-but-missing-key bloat response must never silently classify as PASS:\n%s", report.Text)
	}
	if !strings.Contains(report.Text, "[ERROR]") || !strings.Contains(report.Text, "is_bloated") {
		t.Errorf("expected a loud [ERROR] line naming the missing \"is_bloated\" key, got:\n%s", report.Text)
	}
}

// TestRunFullAudit_VanishedFile pins the fix for a nil-pointer panic: a file
// that matches the docsDir glob but is gone by the time os.Stat runs on it
// (deleted between glob and stat, a dangling symlink, a permission race,
// etc.) must degrade to a loud [ERROR] line and move on, never dereference
// the nil *FileInfo os.Stat returns alongside its error.
func TestRunFullAudit_VanishedFile(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"is_bloated": false}`)

	dir := t.TempDir()
	writeFile(t, dir, "zzfix_present.atom.md", `---
id: zzfix_present
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Present and valid.

## THE RULE / LOGIC
Present and valid.

## EXPECTATION
n/a
`)

	// A dangling symlink matches the "*.atom.md" glob (directory listing
	// only) but os.Stat on it fails because the target doesn't exist --
	// the same failure shape as a file deleted mid-audit.
	danglingLink := filepath.Join(dir, "zzfix_vanished.atom.md")
	if err := os.Symlink(filepath.Join(dir, "does_not_exist"), danglingLink); err != nil {
		t.Fatalf("failed to create dangling symlink: %v", err)
	}

	report, err := RunFullAudit(dir, 0.85, false, 1)
	if err != nil {
		t.Fatalf("RunFullAudit panicked or errored instead of skipping the vanished file: %v", err)
	}

	if !strings.Contains(report.Text, "Auditing: zzfix_vanished.atom.md ... [ERROR:") {
		t.Errorf("expected a loud [ERROR] line for the vanished file, got:\n%s", report.Text)
	}
	if !strings.Contains(report.Text, "zzfix_present.atom.md") {
		t.Errorf("expected the still-present atom to be audited normally alongside the vanished one, got:\n%s", report.Text)
	}
}

// TestRunFullAudit_QueryErrorIsLoudNotSilentPass regression-tests
// failures/20260917_atd_audit_docs_exits_zero_with_no_report.md: a generic
// (non-IDE-fallback) LLM error from the bloat-check query -- a timeout,
// connection refused, whatever -- used to fall through the if/else-if chain
// with no final else, leaving the atom at its pre-set "PASS" default and
// zero trace in the report. It must now classify as ERROR with a loud
// [ERROR] line instead.
func TestRunFullAudit_QueryErrorIsLoudNotSilentPass(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetError(fmt.Errorf("connection refused"))

	dir := t.TempDir()
	writeFile(t, dir, "zzfix_queryerr.atom.md", `---
id: zzfix_queryerr
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Some intent.

## THE RULE / LOGIC
Some logic.

## EXPECTATION
n/a
`)

	report, err := RunFullAudit(dir, 0.85, false, 1)
	if err != nil {
		t.Fatalf("RunFullAudit: %v", err)
	}

	if strings.Contains(report.Text, "Auditing: zzfix_queryerr.atom.md ... [PASS]") {
		t.Errorf("a generic LLM query error must never silently classify as PASS:\n%s", report.Text)
	}
	if !strings.Contains(report.Text, "Auditing: zzfix_queryerr.atom.md ... [ERROR]") {
		t.Errorf("expected the atom to be classified ERROR, got:\n%s", report.Text)
	}
	if !strings.Contains(report.Text, "[ERROR] LLM bloat check failed for zzfix_queryerr") {
		t.Errorf("expected a loud [ERROR] line naming the failed atom, got:\n%s", report.Text)
	}
	if !strings.Contains(report.Text, "connection refused") {
		t.Errorf("expected the underlying error to be surfaced, got:\n%s", report.Text)
	}
}

// TestRunFullAudit_GuaranteedSummaryLine_CleanPass regression-tests
// failures/20260916_atd_audit_workspace_no_return.md and
// failures/20260917_atd_audit_docs_exits_zero_with_no_report.md: a clean run
// must always end with a "Summary:" line stating counts, so a genuinely
// clean 0-findings pass is never indistinguishable from a silently broken
// one that produced no report at all.
func TestRunFullAudit_GuaranteedSummaryLine_CleanPass(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"is_bloated": false}`)
	fake.SetEmbedVector([]float32{0.1, 0.2, 0.3})

	dir := t.TempDir()
	writeFile(t, dir, "zzfix_clean1.atom.md", `---
id: zzfix_clean1
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Clean intent one.

## THE RULE / LOGIC
Clean logic one.

## EXPECTATION
n/a
`)

	report, err := RunFullAudit(dir, 0.85, false, 1)
	if err != nil {
		t.Fatalf("RunFullAudit: %v", err)
	}

	if !strings.Contains(report.Text, "Summary: 1 atom(s) scanned, 0 bloated, 0 collision(s), 0 LLM error(s)") {
		t.Errorf("expected a guaranteed clean-pass summary line, got:\n%s", report.Text)
	}
}

// TestRunFullAudit_GuaranteedSummaryLine_CountsErrors pins the same
// guaranteed summary line, but with a failing bloat-check query: the summary
// must report a non-zero error count rather than silently reading like a
// clean pass.
func TestRunFullAudit_GuaranteedSummaryLine_CountsErrors(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetError(fmt.Errorf("connection refused"))

	dir := t.TempDir()
	writeFile(t, dir, "zzfix_erratom.atom.md", `---
id: zzfix_erratom
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Some intent.

## THE RULE / LOGIC
Some logic.

## EXPECTATION
n/a
`)

	report, err := RunFullAudit(dir, 0.85, false, 1)
	if err != nil {
		t.Fatalf("RunFullAudit: %v", err)
	}

	if !strings.Contains(report.Text, "Summary: 1 atom(s) scanned, 0 bloated, 0 collision(s), 2 LLM error(s) (1 bloat-check, 1 embedding)") {
		t.Errorf("expected the summary line to count both the bloat-check and embedding errors, got:\n%s", report.Text)
	}
}

// TestRunFullAudit_ConcurrencyDeterminism pins the bounded-worker-pool
// scheduling change to Phase 1: running the same fixed set of atom files
// through RunFullAudit with concurrency=1 (effectively sequential) versus
// concurrency=4 (several files in flight at once) must produce
// byte-for-byte identical report text. Report lines are written into a
// per-file result slot indexed by each file's position in the glob's sorted
// file list and only merged into the shared report after every worker
// finishes, so output ordering must never depend on which goroutine happens
// to finish first.
//
// Each atom is audited in its own TempDir (a fresh, empty audit cache) so
// the two runs can't observe each other's cached results and diverge on
// that basis alone -- the comparison isolates the effect of concurrency.
func TestRunFullAudit_ConcurrencyDeterminism(t *testing.T) {
	buildFixtureDir := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		for _, id := range []string{"zzfix_conc_a", "zzfix_conc_b", "zzfix_conc_c", "zzfix_conc_d", "zzfix_conc_e", "zzfix_conc_f"} {
			writeFile(t, dir, id+".atom.md", fmt.Sprintf(`---
id: %s
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Intent for %s.

## THE RULE / LOGIC
Logic for %s.

## EXPECTATION
n/a
`, id, id, id))
		}
		return dir
	}

	runWithConcurrency := func(t *testing.T, concurrency int) string {
		t.Helper()
		fake := fakeprovider.InstallOllama(t)
		// A fixed, content-independent response for every bloat-check call
		// and a fixed embedding vector for every file -- neither depends on
		// call order, so any nondeterminism in the report text can only come
		// from the worker-pool scheduling itself, not from the fake.
		fake.SetJSON(`{"is_bloated": false}`)
		fake.SetEmbedVector([]float32{0.1, 0.2, 0.3})

		dir := buildFixtureDir(t)
		report, err := RunFullAudit(dir, 0.85, false, concurrency)
		if err != nil {
			t.Fatalf("RunFullAudit(concurrency=%d): %v", concurrency, err)
		}
		return report.Text
	}

	sequential := runWithConcurrency(t, 1)
	parallel := runWithConcurrency(t, 4)

	if sequential != parallel {
		t.Errorf("report text differs between concurrency=1 and concurrency=4 runs:\n--- concurrency=1 ---\n%s\n--- concurrency=4 ---\n%s", sequential, parallel)
	}
}
