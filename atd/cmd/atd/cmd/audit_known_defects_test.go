package cmd

// FIXED regression test for issue C (failures/20260917_atd_audit_never_returns_docs_and_atom_code_scopes.md,
// failures/20260917_atd_audit_docs_exits_zero_with_no_report.md,
// failures/20260916_atd_audit_workspace_no_return.md): three field reports
// described `atd audit` appearing to hang or return nothing across
// `--workspace`, `--docs`, and a narrower `--atom --code` "compliance check"
// invocation. One of those reports later self-corrected: the observed hangs
// were mostly self-inflicted concurrent-process contention against a shared
// local Ollama backend, not a proven defect in `atd` itself, and
// pkg/audit.RunFullAudit (see pkg/audit/audit_test.go) already returns a
// well-formed, non-empty report on every code path it exercises, including a
// clean pass ("No critical semantic collisions detected." / "=== AUDIT
// COMPLETE ==="). So there is no reproducible "silently empty on success"
// bug in the library, and cmd/atd/cmd/audit.go's RunE unconditionally prints
// whatever it gets back.
//
// The actual, concrete, deterministic defect that explained "even the
// narrowest possible scope never returned promptly" was: the CLI registered
// --atom and --code flags documented as "Atom path for compliance check" /
// "Snippet path for compliance check" (a narrower, single-atom-vs-single-file
// mode, distinct from the full docsDir sweep), but auditCmd's RunE never
// read either flag — it unconditionally called audit.RunFullAudit(docsDir,
// threshold, workspace), which globs and scores every *.atom.md file in
// docsDir. A caller asking for a "narrow" compliance check silently got the
// exact same full-directory sweep as a plain `atd audit --docs <dir>` call.
//
// FIXED: auditCmd's RunE now reads --atom, and when it is set, calls the new
// audit.RunScopedAudit(atomPath, threshold) instead of RunFullAudit — that
// scopes the bloat/collision analysis to exactly the one named atom file
// rather than globbing docsDir. (--code is not yet wired to a real
// atom-vs-code compliance comparison because pkg/audit has no such
// capability today; RunE prints a note rather than silently dropping it.)
//
// ADDENDUM (2026-09-18): the "mostly self-inflicted concurrent-process
// contention, not a proven defect" verdict above turned out to be
// incomplete. pkg/ollama's generateHTTP/embedHTTP called http.Post with no
// deadline at all — a genuinely unbounded wait, not just contention — so a
// slow or stalled Ollama backend really could hang `atd audit` forever.
// Fixed by adding config.LLMConfig.GenerateTimeoutMs (.atd key
// "llm.generate_timeout_ms", default 120s, config.DefaultGenerateTimeoutMs)
// and threading it into a bounded http.Client in both functions; a timeout
// now surfaces as a clear "timed out after Nms ... see llm.generate_timeout_ms
// in .atd" error instead of hanging (see pkg/ollama/client_timeout_test.go).
// Separately, pkg/audit.RunFullAudit's bloat-check and embed error paths
// used to fall through silently (leaving bloatResult at its "PASS" default,
// or leaving collision detection to just skip an atom) on any query/embed
// error — including this timeout — which is how a hang could also present
// as a quiet, misleadingly clean pass rather than a loud failure. Fixed to
// log an explicit [ERROR] line per failed atom and emit a guaranteed
// "Summary: N atom(s) scanned, ..." line on every run (see
// pkg/audit/audit_test.go's TestRunFullAudit_QueryErrorIsLoudNotSilentPass
// and TestRunFullAudit_GuaranteedSummaryLine_* tests). `atd init` / `atd
// init --upgrade` now also write llm.generate_timeout_ms into the nearest
// .atd at its default value so the bound is visible, not an invisible
// Go-side-only default (see init_generate_timeout_backfill_test.go).
// Together these close out failures/20260916_atd_audit_workspace_no_return.md
// and failures/20260917_atd_audit_docs_exits_zero_with_no_report.md for real.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/testutil/fakeprovider"
)

func writeAuditKnownDefectFixture(t *testing.T, dir, filename, id string) {
	t.Helper()
	content := "---\nid: " + id + "\ntype: RULE\nstatus: DRAFT\nparents: []\ndependents: []\nlayer: BUSINESS\n---\n\n" +
		"## INTENT\nSome intent for " + id + ".\n\n" +
		"## THE RULE / LOGIC\nSome logic for " + id + ".\n\n" +
		"## EXPECTATION\nn/a\n"
	if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestAuditCmd_AtomFlagScopesAudit_Fixed pins the fix for the dead-flag
// defect described above: with --atom pointed at exactly one atom and --code
// left unset (a single-atom scope, per the --atom flag's own description),
// the CLI must make exactly 2 Generate calls (intent + logic judge for that
// one atom) — not 4, the full 2-atom docsDir sweep it used to make while
// --atom was silently ignored.
func TestAuditCmd_AtomFlagScopesAudit_Fixed(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"is_bloated": false}`)

	dir := t.TempDir()
	writeAuditKnownDefectFixture(t, dir, "zzfix_defect_a.atom.md", "zzfix_defect_a")
	writeAuditKnownDefectFixture(t, dir, "zzfix_defect_b.atom.md", "zzfix_defect_b")

	origDocs, _ := auditCmd.Flags().GetString("docs")
	origAtom, _ := auditCmd.Flags().GetString("atom")
	t.Cleanup(func() {
		auditCmd.Flags().Set("docs", origDocs)
		auditCmd.Flags().Set("atom", origAtom)
	})

	if err := auditCmd.Flags().Set("docs", dir); err != nil {
		t.Fatal(err)
	}
	// A caller asking for a narrow single-atom scope, per the CLI's own
	// --atom flag description — deliberately naming only ONE of the two
	// fixture atoms.
	if err := auditCmd.Flags().Set("atom", filepath.Join(dir, "zzfix_defect_a.atom.md")); err != nil {
		t.Fatal(err)
	}

	if err := auditCmd.RunE(auditCmd, []string{}); err != nil {
		t.Fatalf("auditCmd.RunE: %v", err)
	}

	calls := fake.Calls()
	if len(calls) != 2 {
		t.Errorf("got %d Generate call(s) with --atom set to a single atom, want 2 (intent + logic judge for that one atom only) — --atom must scope the audit to exactly that file, not sweep the whole --docs directory", len(calls))
	}
}

// TestAuditCmd_AtomAndCodeReturnsError pins the current, intentional
// --atom+--code behavior: audit has no atom-vs-code compliance comparison
// capability (no prompt or scoring path in pkg/audit takes a code snippet as
// input), so setting both flags together must fail loudly with a returned
// error naming the actual command for that job ("atd map --atom ... --file
// ..."), instead of silently falling back to a single-atom scope and
// printing an easy-to-miss note.
func TestAuditCmd_AtomAndCodeReturnsError(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"is_bloated": false}`)

	dir := t.TempDir()
	writeAuditKnownDefectFixture(t, dir, "zzfix_defect_a.atom.md", "zzfix_defect_a")

	origDocs, _ := auditCmd.Flags().GetString("docs")
	origAtom, _ := auditCmd.Flags().GetString("atom")
	origCode, _ := auditCmd.Flags().GetString("code")
	t.Cleanup(func() {
		auditCmd.Flags().Set("docs", origDocs)
		auditCmd.Flags().Set("atom", origAtom)
		auditCmd.Flags().Set("code", origCode)
	})

	if err := auditCmd.Flags().Set("docs", dir); err != nil {
		t.Fatal(err)
	}
	if err := auditCmd.Flags().Set("atom", filepath.Join(dir, "zzfix_defect_a.atom.md")); err != nil {
		t.Fatal(err)
	}
	if err := auditCmd.Flags().Set("code", "some/unrelated/file.go"); err != nil {
		t.Fatal(err)
	}

	err := auditCmd.RunE(auditCmd, []string{})
	if err == nil {
		t.Fatal("expected auditCmd.RunE to return an error when both --atom and --code are set, got nil")
	}
	if !strings.Contains(err.Error(), "atd map --atom") {
		t.Errorf("expected the error to name \"atd map --atom ... --file ...\" as the correct command, got: %v", err)
	}

	if calls := fake.Calls(); len(calls) != 0 {
		t.Errorf("expected no Generate calls when --atom+--code is rejected before any audit runs, got %d", len(calls))
	}
}
