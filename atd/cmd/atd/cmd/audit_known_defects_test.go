package cmd

// KNOWN DEFECT pin for issue C (failures/20260917_atd_audit_never_returns_docs_and_atom_code_scopes.md,
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
// whatever RunFullAudit returns.
//
// Reading cmd/atd/cmd/audit.go turns up the actual, concrete, deterministic
// defect that explains the "even the narrowest possible scope never
// returned" observation: the CLI registers --atom and --code flags
// documented as "Atom path for compliance check" / "Snippet path for
// compliance check" (a narrower, single-atom-vs-single-file mode, distinct
// from the full docsDir sweep), but auditCmd's RunE never reads either flag
// — it unconditionally calls audit.RunFullAudit(docsDir, threshold,
// workspace), which globs and scores every *.atom.md file in docsDir. A
// caller asking for a "narrow" compliance check silently gets the exact same
// full-directory sweep as a plain `atd audit --docs <dir>` call, with no
// indication their --atom/--code flags did anything at all.

import (
	"os"
	"path/filepath"
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

// TestAuditCmd_AtomCodeFlagsAreIgnored_KnownDefect pins the dead-flag defect
// above: with --atom pointed at exactly one atom and --code at an unrelated
// path (a single-atom, single-file compliance check per the flags' own
// descriptions), the CLI must have made exactly 2 Generate calls (intent +
// logic judge for that one atom) if --atom/--code actually scoped the audit.
// Today it makes 4 — the full 2-atom docsDir sweep — because the flags are
// never consulted.
func TestAuditCmd_AtomCodeFlagsAreIgnored_KnownDefect(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"is_bloated": false}`)

	dir := t.TempDir()
	writeAuditKnownDefectFixture(t, dir, "zzfix_defect_a.atom.md", "zzfix_defect_a")
	writeAuditKnownDefectFixture(t, dir, "zzfix_defect_b.atom.md", "zzfix_defect_b")

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
	// A caller asking for a narrow single-atom/single-file compliance check,
	// per the CLI's own --atom/--code flag descriptions ("Atom path for
	// compliance check", "Snippet path for compliance check") — deliberately
	// naming only ONE of the two fixture atoms.
	if err := auditCmd.Flags().Set("atom", filepath.Join(dir, "zzfix_defect_a.atom.md")); err != nil {
		t.Fatal(err)
	}
	if err := auditCmd.Flags().Set("code", "some/unrelated/file.go"); err != nil {
		t.Fatal(err)
	}

	if err := auditCmd.RunE(auditCmd, []string{}); err != nil {
		t.Fatalf("auditCmd.RunE: %v", err)
	}

	calls := fake.Calls()
	if len(calls) != 4 {
		t.Errorf("KNOWN DEFECT expectation changed: got %d Generate call(s) with --atom/--code set to a single atom/file, want 4 (the full 2-atom docsDir sweep this defect currently produces regardless of --atom/--code) — --atom/--code may have started actually scoping the audit; if intentional, relax this pin to assert 2 calls instead", len(calls))
	}
}
