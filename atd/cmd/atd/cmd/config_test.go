package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/testutil"
)

// TestRunConfigUpdateRefusesFallbackAnchoredConfig pins the P-1 write guard
// added to runConfigUpdate (test_atd_07_26.md §3.6, §8.3 defect #7): the
// atd_config MCP tool's task+model branch calls a bare config.Load(), which
// re-resolves the active config from the process's cwd. With no real .atd
// anywhere above cwd, config.Load() falls back to cwd and
// config.LoadedFromFallback() reports true -- and the write must now be
// refused rather than silently creating/clobbering a ".atd" file wherever
// the process happened to be invoked from. This is the write-side
// counterpart of incident I-1 (§2.1), where a leaked, cwd-fallback config
// let a different destructive operation (rename propagation) run against
// the wrong directory.
//
// This test is deliberately NOT t.Parallel(): it chdirs the whole process,
// like cmd/atd/cmd/workspace_integration_test.go does for the same reason.
func TestRunConfigUpdateRefusesFallbackAnchoredConfig(t *testing.T) {
	saved := config.Snapshot()
	t.Cleanup(func() { config.Restore(saved) })

	// t.TempDir() lives under the OS temp root, which has no .atd anywhere
	// above it -- config.Load() (bare, cwd-based) must fall back to cwd
	// here, exactly as pkg/atom/update_scope_test.go's fallback test
	// assumes for the same reason.
	fallbackRoot := t.TempDir()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(fallbackRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(oldWD) })

	out, err := runConfigUpdate("audit_code", "some-model")
	if err == nil {
		t.Fatalf("expected runConfigUpdate to refuse a fallback-anchored write, got output: %q", out)
	}
	if !strings.Contains(err.Error(), "fallback-anchored") {
		t.Errorf("expected a loud, actionable fallback-anchored error, got: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(fallbackRoot, ".atd")); !os.IsNotExist(statErr) {
		t.Errorf("expected no .atd to be written under the fallback-anchored root, stat error: %v", statErr)
	}
}

// TestRunConfigUpdateWritesWhenProperlyAnchored is the positive counterpart:
// with a real .atd anchoring the active config, the legitimate task/model
// reassignment must still work -- the P-1 guard must not break the feature.
func TestRunConfigUpdateWritesWhenProperlyAnchored(t *testing.T) {
	sb := testutil.Sandbox(t, "fixture_project")

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(sb.Root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(oldWD) })

	out, err := runConfigUpdate("audit_code", "some-model")
	if err != nil {
		t.Fatalf("expected a properly-anchored config to accept the write, got error: %v", err)
	}
	if !strings.Contains(out, "some-model") {
		t.Errorf("expected confirmation output to mention the assigned model, got: %q", out)
	}

	data, err := os.ReadFile(filepath.Join(sb.Root, ".atd"))
	if err != nil {
		t.Fatalf("expected .atd to have been written: %v", err)
	}
	var written config.Config
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("written .atd is not valid JSON: %v", err)
	}
	mc, ok := written.LLM.Models["some-model"]
	if !ok {
		t.Fatalf("expected .atd to record model 'some-model', got: %s", data)
	}
	found := false
	for _, task := range mc.Tasks {
		if task == "audit_code" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected model 'some-model' to be assigned task 'audit_code', got tasks: %v", mc.Tasks)
	}
}
