package cmd

// TestAtdAuditMCP_AtomScopesToSingleFile and
// TestAtdAuditMCP_AtomAndCodeReturnsError cover MCP/CLI parity for the
// atd_audit MCP tool's "atom" and "code" parameters: before this change, the
// MCP tool only exposed "docs"/"threshold" and always ran a full docsDir
// sweep, with no way to scope to a single atom at all (unlike the CLI's
// --atom flag). These pin that the MCP handler now shares the exact same
// atom/code dispatch decision as the CLI, through dispatchAudit.

import (
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/mcp"
	"atd-tools/pkg/testutil/fakeprovider"
)

// TestAtdAuditMCP_AtomScopesToSingleFile drives the atd_audit MCP tool with
// "atom" set to exactly one of two atom files in the docs directory, and
// asserts the report text only mentions that one file -- proof the MCP
// handler actually calls the scoped path (audit.RunScopedAudit) rather than
// a full sweep of "docs".
func TestAtdAuditMCP_AtomScopesToSingleFile(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"is_bloated": false}`)

	dir := t.TempDir()
	writeAuditKnownDefectFixture(t, dir, "zzfix_mcp_a.atom.md", "zzfix_mcp_a")
	writeAuditKnownDefectFixture(t, dir, "zzfix_mcp_b.atom.md", "zzfix_mcp_b")

	r := mcp.NewRegistry()
	RegisterMCPTools(r)

	out, err := r.Call("atd_audit", map[string]any{
		"docs": dir,
		"atom": filepath.Join(dir, "zzfix_mcp_a.atom.md"),
	})
	if err != nil {
		t.Fatalf("atd_audit MCP call: %v", err)
	}

	if !strings.Contains(out, "zzfix_mcp_a.atom.md") {
		t.Errorf("expected the scoped atom to appear in the report, got:\n%s", out)
	}
	if strings.Contains(out, "zzfix_mcp_b.atom.md") {
		t.Errorf("atom-scoped atd_audit must not touch the other atom in the docs directory, got:\n%s", out)
	}

	if calls := len(fake.Calls()); calls != 2 {
		t.Errorf("got %d Generate call(s) with atom scoped to a single file, want 2 (intent + logic judge for that one atom only)", calls)
	}
}

// TestAtdAuditMCP_AtomAndCodeReturnsError pins MCP/CLI parity for the
// --atom+--code rejection: setting both "atom" and "code" on the atd_audit
// MCP tool must return the same error the CLI does (compliance checking is
// atd_recon / "atd map --atom ... --file ...", not audit), not silently
// scope to atom alone or sweep the whole docs directory.
func TestAtdAuditMCP_AtomAndCodeReturnsError(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"is_bloated": false}`)

	dir := t.TempDir()
	writeAuditKnownDefectFixture(t, dir, "zzfix_mcp_c.atom.md", "zzfix_mcp_c")

	r := mcp.NewRegistry()
	RegisterMCPTools(r)

	_, err := r.Call("atd_audit", map[string]any{
		"docs": dir,
		"atom": filepath.Join(dir, "zzfix_mcp_c.atom.md"),
		"code": "some/unrelated/file.go",
	})
	if err == nil {
		t.Fatal("expected atd_audit MCP call to error when both \"atom\" and \"code\" are set, got nil")
	}
	if !strings.Contains(err.Error(), "atd map --atom") {
		t.Errorf("expected the error to name \"atd map --atom ... --file ...\" as the correct command, got: %v", err)
	}

	if calls := len(fake.Calls()); calls != 0 {
		t.Errorf("expected no Generate calls when atom+code is rejected before any audit runs, got %d", calls)
	}
}
