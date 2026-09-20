package cmd

// KNOWN DEFECT pins (now partly FIXED) for
// failures/20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md.
//
// The report claimed two problems: (1) `atd congruence` accepts a bare,
// title-only "audit_report" as a complete result even when is_congruent is
// false, giving the caller nothing to act on, and (2) both a bare-verdict
// run and a target-not-found run "exited 0 despite failing." Investigation
// against the then-current source (congruence.go) confirmed (1) but not (2):
// see TestCongruenceRun_TargetNotFound_ReturnsError below for why the exit-0
// claim did not reproduce from this file as written at the time.
//
// FIXED (1): congruence.go's RunE now parses resp.Response and rejects
// (non-nil error) an is_congruent:false response whose findings array is
// missing or empty, per the schema pkg/prompt/congruence.go's
// CongruenceFormat() now declares (pkg/prompt/congruence_findings_test.go)
// and the RunE-level contract cmd/atd/cmd/congruence_findings_test.go pins.
// TestCongruenceLLM_KnownDefect_BareVerdictAccepted below -- which pinned
// the old accept-a-bare-verdict behavior -- is flipped to
// TestCongruenceLLM_BareVerdictRejected_Fixed, asserting the new rejection,
// per this file's own original instruction to flip rather than delete once
// validation landed.

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

// TestCongruenceLLM_BareVerdictRejected_Fixed replaces
// TestCongruenceLLM_KnownDefect_BareVerdictAccepted now that RunE validates
// resp.Response: a title-only "audit_report" string with is_congruent:
// false and no "findings" key at all -- no named contradicting atom, no
// section reference -- must now be REJECTED (non-nil error), not printed
// and accepted as a complete result. See the FIXED note in this file's
// header and cmd/atd/cmd/congruence_findings_test.go's
// TestCongruenceRun_RejectsIncongruentWithoutFindings, which pins the same
// contract with additional cases.
func TestCongruenceLLM_BareVerdictRejected_Fixed(t *testing.T) {
	docsDir := congruenceSandboxDocs(t, "req_bare_verdict_target")
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"audit_report": "System Congruence Verification", "is_congruent": false}`)

	congruenceCmd.Flags().Set("target", "req_bare_verdict_target")
	congruenceCmd.Flags().Set("docs", docsDir)
	t.Cleanup(func() {
		congruenceCmd.Flags().Set("target", "")
		congruenceCmd.Flags().Set("docs", "")
	})

	_, runErr := captureStdout(t, func() error {
		return congruenceCmd.RunE(congruenceCmd, []string{})
	})

	if runErr == nil {
		t.Fatal("expected RunE to reject a bare title-only audit_report with no findings (is_congruent: false), got nil error")
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
