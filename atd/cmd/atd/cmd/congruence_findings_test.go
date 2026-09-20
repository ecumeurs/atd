package cmd

// Pins the FIX for Problem 1 of
// failures/20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md
// ("bare/unusable verdict, no validation"), at the RunE layer. The schema
// side of this contract is pinned in pkg/prompt/congruence_findings_test.go
// (CongruenceFormat() must declare a `findings` array of
// {atom_id, section, contradiction} objects); this file pins that RunE must
// actually validate a parsed response against that shape instead of
// unconditionally printing resp.Response and returning nil, which is what
// congruence_known_defects_test.go's TestCongruenceLLM_KnownDefect_BareVerdictAccepted
// pins as today's (broken) behavior.
//
// Desired RunE behavior:
//   - is_congruent: false with a missing or empty `findings` array -> RunE
//     returns a non-nil error (nothing useful to act on).
//   - is_congruent: false with `findings` populated (each entry naming an
//     atom_id, section, and contradiction) -> RunE returns nil and the
//     finding details are visible in stdout.
//
// KNOWN DEFECT note: once this fix lands, TestCongruenceLLM_KnownDefect_BareVerdictAccepted
// in congruence_known_defects_test.go will itself start failing (by its own
// design/comment) and must be flipped by whoever implements the fix — not by
// this file.

import (
	"strings"
	"testing"

	"atd-tools/pkg/testutil/fakeprovider"
)

// setCongruenceFlags sets --target/--docs on congruenceCmd and registers a
// cleanup that resets them, mirroring the idiom in
// congruence_known_defects_test.go.
func setCongruenceFlags(t *testing.T, target, docsDir string) {
	t.Helper()
	if err := congruenceCmd.Flags().Set("target", target); err != nil {
		t.Fatalf("setting --target: %v", err)
	}
	if err := congruenceCmd.Flags().Set("docs", docsDir); err != nil {
		t.Fatalf("setting --docs: %v", err)
	}
	t.Cleanup(func() {
		congruenceCmd.Flags().Set("target", "")
		congruenceCmd.Flags().Set("docs", "")
	})
}

// TestCongruenceRun_RejectsIncongruentWithoutFindings pins that RunE must
// reject (non-nil error) an is_congruent:false response that carries no
// structured findings -- whether the "findings" key is absent entirely (the
// exact known-defect payload) or present but empty. This currently fails:
// today's RunE is `fmt.Println(resp.Response); return nil` unconditionally.
func TestCongruenceRun_RejectsIncongruentWithoutFindings(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{
			name: "findings key absent",
			json: `{"audit_report": "System Congruence Verification", "is_congruent": false}`,
		},
		{
			name: "findings present but empty",
			json: `{"audit_report": "System Congruence Verification", "is_congruent": false, "findings": []}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			docsDir := congruenceSandboxDocs(t, "req_no_findings_target")
			fake := fakeprovider.InstallOllama(t)
			fake.SetJSON(tc.json)

			setCongruenceFlags(t, "req_no_findings_target", docsDir)

			_, runErr := captureStdout(t, func() error {
				return congruenceCmd.RunE(congruenceCmd, []string{})
			})

			if runErr == nil {
				t.Fatalf("expected RunE to reject an is_congruent:false response with no findings (%s), got nil error", tc.name)
			}
		})
	}
}

// TestCongruenceRun_AcceptsIncongruentWithPopulatedFindings pins that a
// well-formed is_congruent:false response -- findings populated with an
// atom_id, section, and contradiction per entry -- is accepted (RunE returns
// nil) and that the finding details are surfaced in stdout, not just a bare
// title/verdict.
func TestCongruenceRun_AcceptsIncongruentWithPopulatedFindings(t *testing.T) {
	docsDir := congruenceSandboxDocs(t, "req_findings_target")
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{
		"is_congruent": false,
		"audit_report": "System Congruence Verification",
		"findings": [
			{
				"atom_id": "req_findings_target_sibling",
				"section": "LOGIC",
				"contradiction": "sibling atom requires the opposite ordering of steps"
			}
		]
	}`)

	setCongruenceFlags(t, "req_findings_target", docsDir)

	stdout, runErr := captureStdout(t, func() error {
		return congruenceCmd.RunE(congruenceCmd, []string{})
	})

	if runErr != nil {
		t.Fatalf("expected RunE to accept a populated-findings response, got error: %v", runErr)
	}
	for _, want := range []string{"req_findings_target_sibling", "LOGIC", "sibling atom requires the opposite ordering of steps"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("expected stdout to surface finding detail %q, got: %q", want, stdout)
		}
	}
}
