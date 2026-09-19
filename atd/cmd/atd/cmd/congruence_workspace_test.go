package cmd

// Pins the FIX for Problem 2 of
// failures/20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md
// ("no --workspace flag"). `congruence` (and `trace`) have no --workspace
// bool flag today, unlike audit/crawl/stats/weave (see e.g. weave.go's
// `weaveCmd.Flags().Bool("workspace", false, ...)`), and congruence.go's
// atom lookup is a flat map keyed by bare filename-derived id read from a
// single --docs directory -- a workspace-qualified target like
// "project:atom_id" cannot resolve at all.
//
// These tests use testdata/fixture_workspace (two projects, zzfix_a and
// zzfix_b, joined by an .atd.workspace) via pkg/testutil.Sandbox, the same
// fixture pkg/exploration/crawl_test.go's
// TestCrawlWorkspaceDocs_AggregatesAcrossProjects and
// cmd/atd/cmd/scenario_test.go's testScenarioS1Workspace already exercise
// for workspace-aware resolution.

import (
	"testing"

	"atd-tools/pkg/testutil"
	"atd-tools/pkg/testutil/fakeprovider"
)

// TestCongruenceCmd_HasWorkspaceFlag pins that congruenceCmd registers a
// --workspace bool flag, matching the pattern audit/crawl/stats/weave
// already follow. Fails today: no such flag exists.
func TestCongruenceCmd_HasWorkspaceFlag(t *testing.T) {
	flag := congruenceCmd.Flags().Lookup("workspace")
	if flag == nil {
		t.Fatal("expected congruenceCmd to register a --workspace bool flag (matching audit/crawl/stats/weave), got none")
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("expected --workspace flag to be a bool flag, got type %q", flag.Value.Type())
	}
}

// TestTraceCmd_HasWorkspaceFlag pins the same expectation for traceCmd
// ("ideally trace" per the field report -- trace has the identical
// single-docs-dir limitation). Fails today: no such flag exists.
func TestTraceCmd_HasWorkspaceFlag(t *testing.T) {
	flag := traceCmd.Flags().Lookup("workspace")
	if flag == nil {
		t.Fatal("expected traceCmd to register a --workspace bool flag (matching audit/crawl/stats/weave), got none")
	}
	if flag.Value.Type() != "bool" {
		t.Errorf("expected --workspace flag to be a bool flag, got type %q", flag.Value.Type())
	}
}

// TestCongruenceRun_WorkspaceQualifiedTarget_ResolvesAcrossProjects pins
// that, with --workspace set, a workspace-qualified target id
// ("project:atom_id") resolves against THAT project's docs dir, not only
// the current project's --docs/config.DocsDir(). The fixture's
// zzfix_a:req_zzfix_ws_alpha atom lives in zzfix_a/docs, and this test
// invokes congruence from zzfix_b's project context (config loaded from
// zzfix_b's root, so config.DocsDir() would only see zzfix_b/docs on its
// own) -- so a correct resolution requires actually consulting the
// workspace's other project(s), not just the active one.
//
// Fails today for two compounding reasons: (1) --workspace does not exist
// on congruenceCmd at all (Set below errors), and (2) even if the flag were
// silently ignored, congruence.go's atomMap is built from a single --docs
// directory and has no concept of a "project:atom_id" qualified lookup, so
// the target would report "not found".
func TestCongruenceRun_WorkspaceQualifiedTarget_ResolvesAcrossProjects(t *testing.T) {
	ws := testutil.Sandbox(t, "fixture_workspace")
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"is_congruent": true, "audit_report": "no contradictions found", "findings": []}`)

	if err := congruenceCmd.Flags().Set("workspace", "true"); err != nil {
		t.Fatalf("expected --workspace flag to exist on congruenceCmd, got error setting it: %v", err)
	}
	if err := congruenceCmd.Flags().Set("target", "zzfix_a:req_zzfix_ws_alpha"); err != nil {
		t.Fatalf("setting --target: %v", err)
	}
	if err := congruenceCmd.Flags().Set("docs", ""); err != nil {
		t.Fatalf("setting --docs: %v", err)
	}
	t.Cleanup(func() {
		congruenceCmd.Flags().Set("workspace", "false")
		congruenceCmd.Flags().Set("target", "")
		congruenceCmd.Flags().Set("docs", "")
	})

	b := ws.Sub("zzfix_b")
	res := b.Run(func() (string, error) {
		return captureStdout(t, func() error {
			return congruenceCmd.RunE(congruenceCmd, []string{})
		})
	})

	if res.Err != nil {
		t.Fatalf("expected congruence --workspace to resolve the cross-project target \"zzfix_a:req_zzfix_ws_alpha\" from zzfix_b's context, got error: %v", res.Err)
	}
	if fake.Calls() == nil || len(fake.Calls()) == 0 {
		t.Error("expected the LLM to be queried once resolution succeeded, but it was never called")
	}
}
