package cmd

// WP-6 (test_atd_07_26.md §3.5, §6, S12): deterministic logic + parser-
// robustness tests for `atd map`'s three LLM-backed modes (confirm/recon,
// propose, discover), using pkg/testutil/fakeprovider instead of a live
// Ollama. Named map_llm_test.go to sit alongside the existing map_test.go
// (which only exercises flag/routing/no-LLM paths) without colliding.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/testutil/fakeprovider"
)

// mapSandboxDocs loads config from a fresh temp project dir (a real .atd
// file plus a docs/ dir) so resolveAtomPath's "no workspace: try local
// docs" branch (cmd/atd/cmd/map.go) finds atoms via config.DocsDir(). This
// mirrors the existing map_test.go idiom (TestMapNewFlagProducesSkeleton) of
// anchoring config to t.TempDir() rather than relying on ambient state.
func mapSandboxDocs(t *testing.T) (root, docsDir string) {
	t.Helper()
	saved := config.Snapshot()
	t.Cleanup(func() { config.Restore(saved) })

	root = t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	docsDir = filepath.Join(root, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := config.LoadFromDir(root); err != nil {
		t.Fatalf("config.LoadFromDir: %v", err)
	}
	return root, docsDir
}

// ── recon / map --atom (confirm mode) ──────────────────────────────────────

func writeConfirmAtom(t *testing.T, docsDir, atomID string) {
	t.Helper()
	content := `---
id: ` + atomID + `
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
Some intent for the recon target.

## THE RULE / LOGIC
Some logic for the recon target.

## EXPECTATION
n/a
`
	if err := os.WriteFile(filepath.Join(docsDir, atomID+".atom.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestRunMapConfirm_ValidResponse pins the happy path: a schema-valid recon
// verdict with a real mismatch comes back as pretty-printed JSON.
func TestRunMapConfirm_ValidResponse(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)
	writeConfirmAtom(t, docsDir, "zzfix_recon_target")

	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"Confidence": 70, "Mismatches": [{"aspect": "locking", "expected": "mutex", "found": "none"}]}`)

	out, err := runMapConfirm("src/candidate.go", "package main", "zzfix_recon_target")
	if err != nil {
		t.Fatalf("runMapConfirm: %v", err)
	}
	if !strings.Contains(out, `"Confidence": 70`) {
		t.Errorf("expected the parsed verdict to be printed, got: %s", out)
	}
}

// TestRunMapConfirm_DegenerateZeroConfidence pins the existing graceful
// handling of Confidence:0 with no rationale (a schema-valid but
// unhelpful response) — must produce an explicit "inconclusive" message,
// never a bare zero-value JSON blob that reads as a confident "no match".
func TestRunMapConfirm_DegenerateZeroConfidence(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)
	writeConfirmAtom(t, docsDir, "zzfix_recon_target")

	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"Confidence": 0, "Mismatches": []}`)

	out, err := runMapConfirm("src/candidate.go", "package main", "zzfix_recon_target")
	if err != nil {
		t.Fatalf("runMapConfirm: %v", err)
	}
	if !strings.Contains(out, "inconclusive") {
		t.Errorf("expected an explicit 'inconclusive' message for Confidence:0 with no rationale, got: %s", out)
	}
}

// TestRunMapConfirm_MalformedResponses is the S12 suite for the recon
// parser (pkg/prompt/recon.go's ReconResult, parsed in
// cmd/atd/cmd/map.go's runMapConfirm). Every case here already salvages
// correctly on the current code: json.Unmarshal fails (never silently
// succeeds with a partial/zero-value ReconResult), and runMapConfirm's
// `if err != nil { return resp.Response, nil }` branch surfaces the raw
// response rather than swallowing it. This pins that behavior rather than
// flagging a defect.
func TestRunMapConfirm_MalformedResponses(t *testing.T) {
	cases := []struct {
		name string
		resp string
	}{
		{"markdown_fenced", "```json\n{\"Confidence\": 80, \"Mismatches\": []}\n```"},
		{"trailing_prose", `Here is my verdict: {"Confidence": 80, "Mismatches": []}`},
		{"truncated", `{"Confidence": 80, "Mismatches": [{"aspect": "x"`},
		// The real-world field-report case (test_atd_07_26.md §3.5): a
		// string where the schema declares an array.
		{"wrong_type_mismatches_as_string", `{"Confidence": 0, "Mismatches": "Yes, there are several problems"}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, docsDir := mapSandboxDocs(t)
			writeConfirmAtom(t, docsDir, "zzfix_recon_target")

			fake := fakeprovider.InstallOllama(t)
			fake.SetJSON(c.resp)

			out, err := runMapConfirm("src/candidate.go", "package main", "zzfix_recon_target")
			if err != nil {
				t.Fatalf("runMapConfirm must salvage malformed JSON rather than error, got err: %v", err)
			}
			if out == "" {
				t.Fatalf("malformed recon response (%s) must never produce an empty (zero-value) result", c.name)
			}
			if out != c.resp {
				t.Errorf("expected the raw malformed response to be salvaged verbatim, got: %s", out)
			}
		})
	}
}

// ── map --new (propose mode) ────────────────────────────────────────────

// TestRunMapPropose_ValidResponse pins the happy path.
func TestRunMapPropose_ValidResponse(t *testing.T) {
	mapSandboxDocs(t)
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"intent": "Handles user login validation."}`)

	out, err := runMapPropose("src/login.go", "package main")
	if err != nil {
		t.Fatalf("runMapPropose: %v", err)
	}
	if !strings.Contains(out, "Handles user login validation.") {
		t.Errorf("expected the extracted intent in the proposed skeleton, got: %s", out)
	}
}

// TestRunMapPropose_MalformedResponses is the S12 case for propose mode.
// Before WP-6's surgical fix, runMapPropose discarded the intent-extraction
// Unmarshal error entirely (`json.Unmarshal([]byte(resp.Response),
// &intentResult)` with no `if err != nil` check), so a malformed response
// silently produced an EMPTY intent and the function returned a
// successful-looking skeleton with a blank "**Intent:**" line — a
// zero-value success. The fix makes this loud instead.
func TestRunMapPropose_MalformedResponses(t *testing.T) {
	cases := []struct {
		name string
		resp string
	}{
		{"markdown_fenced", "```json\n{\"intent\": \"x\"}\n```"},
		{"trailing_prose", `Here is the intent: {"intent": "x"}`},
		{"truncated", `{"intent": "handles user log`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mapSandboxDocs(t)
			fake := fakeprovider.InstallOllama(t)
			fake.SetJSON(c.resp)

			out, err := runMapPropose("src/login.go", "package main")
			if err == nil {
				t.Fatalf("KNOWN DEFECT expectation changed: runMapPropose used to silently succeed with a blank intent on malformed JSON (%s); it now must error loudly per the WP-6 surgical fix. Got output instead: %q", c.name, out)
			}
			if !strings.Contains(err.Error(), "malformed JSON") {
				t.Errorf("expected an error naming the malformed JSON, got: %v", err)
			}
		})
	}
}

// ── map (discover mode) — intent-extraction leg only ─────────────────────
//
// Full discover-mode success also requires a semantic search index
// (exploration.Search against a real sqlite db); that plumbing is exercised
// by the scenario/fixture-project suite (WP-2), not here. This test isolates
// the first LLM call (intent extraction) and its parser, matching the
// scope of the other S12 cases in this file.

// TestRunMapDiscover_MalformedIntentResponse pins the same defect class as
// propose mode, for discover mode's first Generate call. Before WP-6's
// surgical fix, the discarded Unmarshal error here let a malformed response
// silently become codeIntent = "" and drove a semantic search on an empty
// string — a zero-value success masquerading as a completed discover step.
func TestRunMapDiscover_MalformedIntentResponse(t *testing.T) {
	_, docsDir := mapSandboxDocs(t)

	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"intent": "truncated`)

	_, err := runMapDiscover("src/foo.go", "package main", docsDir)
	if err == nil {
		t.Fatal("KNOWN DEFECT expectation changed: runMapDiscover used to silently proceed with an empty intent on malformed JSON; it now must error loudly per the WP-6 surgical fix")
	}
	if !strings.Contains(err.Error(), "malformed JSON") {
		t.Errorf("expected an error naming the malformed JSON, got: %v", err)
	}
}
