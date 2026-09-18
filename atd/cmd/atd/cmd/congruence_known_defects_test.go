package cmd

// KNOWN DEFECT pins for failures/20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md.
//
// The report claimed two problems: (1) `atd congruence` accepts a bare,
// title-only "audit_report" as a complete result even when is_congruent is
// false, giving the caller nothing to act on, and (2) both a bare-verdict
// run and a target-not-found run "exited 0 despite failing." Investigation
// against the current source (congruence.go) confirms (1) but not (2): see
// TestCongruenceRun_TargetNotFound_ReturnsError below for why the exit-0
// claim does not reproduce from this file as written today.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/testutil/fakeprovider"
)

// congruenceSandboxDocs creates a throwaway docs dir with a single target
// atom file, mirroring map_llm_test.go's mapSandboxDocs idiom. congruence.go
// reads atoms by raw filename/content (no atom.Parse), so the fixture only
// needs to look enough like an atom.md for the file's own link/tag regexes.
func congruenceSandboxDocs(t *testing.T, targetID string) (docsDir string) {
	t.Helper()
	saved := config.Snapshot()
	t.Cleanup(func() { config.Restore(saved) })

	root := t.TempDir()
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

	content := "---\nid: " + targetID + "\ntype: RULE\nlayer: BUSINESS\nparents: []\ntags: []\n---\n\n## INTENT\nSome intent.\n\n## THE RULE / LOGIC\nSome logic.\n"
	if err := os.WriteFile(filepath.Join(docsDir, targetID+".atom.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return docsDir
}

// captureStdout is defined in reconcile_llm_test.go (same package) and
// reused here rather than duplicated.

// TestCongruenceLLM_KnownDefect_BareVerdictAccepted pins the report's core
// finding: congruence.go's RunE (congruence.go:126-149) never parses or
// validates resp.Response at all — it does `fmt.Println(resp.Response)` and
// returns nil unconditionally. A title-only "audit_report" string with
// is_congruent: false — no findings, no named contradicting atom, no
// section reference — is therefore accepted as a complete, successful
// result exactly as readily as a real findings body would be.
//
// KNOWN DEFECT: if this test starts failing because RunE now returns a
// non-nil error (or prints a warning) for a congruent-false response with an
// empty/title-only audit_report, that means validation was added — flip
// this pin to a positive assertion instead of relaxing it blindly.
func TestCongruenceLLM_KnownDefect_BareVerdictAccepted(t *testing.T) {
	docsDir := congruenceSandboxDocs(t, "req_bare_verdict_target")
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"audit_report": "System Congruence Verification", "is_congruent": false}`)

	congruenceCmd.Flags().Set("target", "req_bare_verdict_target")
	congruenceCmd.Flags().Set("docs", docsDir)
	t.Cleanup(func() {
		congruenceCmd.Flags().Set("target", "")
		congruenceCmd.Flags().Set("docs", "")
	})

	stdout, runErr := captureStdout(t, func() error {
		return congruenceCmd.RunE(congruenceCmd, []string{})
	})

	if runErr != nil {
		t.Fatalf("KNOWN DEFECT expectation changed: RunE now rejects a bare title-only audit_report (err: %v) — validation must have been added; flip this pin to assert the error instead", runErr)
	}
	if !strings.Contains(stdout, `"is_congruent": false`) {
		t.Fatalf("expected the bare verdict JSON to be printed verbatim to stdout, got: %q", stdout)
	}
	if strings.Contains(stdout, "INTENT") || strings.Contains(stdout, "LOGIC") || strings.Contains(stdout, "contradict") {
		t.Fatalf("KNOWN DEFECT expectation changed: output now contains findings detail beyond the bare title (got: %q) — RunE must be validating/enriching the response now; flip this pin", stdout)
	}
}

// TestCongruenceRun_TargetNotFound_ReturnsError checks the report's second
// claim — that a target-atom-not-found run "exited 0 despite failing" — at
// the one layer this test binary can actually observe: congruenceCmd.RunE's
// return value (matching the existing TestCongruenceRun idiom in
// congruence_test.go).
//
// This is deliberately NOT a KNOWN DEFECT pin: reading congruence.go:53-56
// shows the not-found path returns a genuine non-nil error
// (`fmt.Errorf("target atom '%s' not found in %s", ...)`), and root.go's
// Execute() (cmd/atd/cmd/root.go:60-65) calls os.Exit(1) whenever
// rootCmd.Execute() returns a non-nil error — so on current source, a
// real `atd congruence --target <missing>` invocation exits 1, not 0. The
// field report's "exited 0" observation for this specific command does not
// reproduce here; if it is genuine, the cause is not in this file (e.g. a
// harness/shell artifact around how the exit code was read), and this test
// exists to keep that distinction on record precisely so nobody "fixes" a
// not-found path that already errors correctly.
func TestCongruenceRun_TargetNotFound_ReturnsError(t *testing.T) {
	docsDir := congruenceSandboxDocs(t, "req_present_atom")

	congruenceCmd.Flags().Set("target", "req_this_atom_does_not_exist")
	congruenceCmd.Flags().Set("docs", docsDir)
	t.Cleanup(func() {
		congruenceCmd.Flags().Set("target", "")
		congruenceCmd.Flags().Set("docs", "")
	})

	err := congruenceCmd.RunE(congruenceCmd, []string{})
	if err == nil {
		t.Fatal("expected a non-nil error for an unresolved --target, got nil")
	}
	if !strings.Contains(err.Error(), "not found in") {
		t.Errorf("expected error to mention 'not found in', got: %v", err)
	}
}
