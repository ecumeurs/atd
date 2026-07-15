package coverage

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
)

func TestCheckRowStatusLogic(t *testing.T) {
	tests := []struct {
		impl     int
		test     int
		expected string
	}{
		{0, 0, "NO_IMPL"},
		{1, 0, "NO_TESTS"},
		{2, 1, "OK"},
		{3, 2, "OK"},
	}

	for _, tt := range tests {
		row := CheckAtomRow{ImplLinks: tt.impl, TestLinks: tt.test}
		switch {
		case row.ImplLinks == 0:
			row.Status = "NO_IMPL"
		case row.TestLinks == 0:
			row.Status = "NO_TESTS"
		default:
			row.Status = "OK"
		}
		if row.Status != tt.expected {
			t.Errorf("impl=%d test=%d: expected status %q, got %q", tt.impl, tt.test, tt.expected, row.Status)
		}
	}
}

func TestFormatCheckReport(t *testing.T) {
	r := &CheckReport{
		Mode: "atom",
		Rows: []CheckAtomRow{
			{AtomID: "rule_foo", ImplLinks: 2, TestLinks: 1, Semantic: "-", Status: "OK"},
			{AtomID: "rule_bar", ImplLinks: 0, TestLinks: 0, Semantic: "-", Status: "NO_IMPL"},
		},
		Summary: CheckSummary{Total: 2, WithImpl: 1, WithTests: 1},
	}

	out := formatCheckReport(r)

	if !strings.Contains(out, "Atom ID") {
		t.Error("expected 'Atom ID' header in output")
	}
	if !strings.Contains(out, "rule_foo") {
		t.Error("expected 'rule_foo' in output")
	}
	if !strings.Contains(out, "rule_bar") {
		t.Error("expected 'rule_bar' in output")
	}
	if !strings.Contains(out, "NO_IMPL") {
		t.Error("expected 'NO_IMPL' status in output")
	}
	if !strings.Contains(out, "OK") {
		t.Error("expected 'OK' status in output")
	}
	if !strings.Contains(out, "2 atoms") {
		t.Error("expected '2 atoms' in summary")
	}
}

// TestDiffChangedFilesRebasesOntoProjectRoot exercises the §1 fix: when
// config.ProjectRoot() is a subdirectory of the git repo that owns it (e.g. a
// nested project, or a submodule accessed via an umbrella working dir), `git
// diff --name-only` paths must be rebased onto ProjectRoot so they match
// SpecLink.FilePath (which the crawler always stores relative to
// ProjectRoot). Before the fix, diffChangedFiles ran `git diff` with no
// working directory at all, so it used whatever repo owned the process cwd.
func TestDiffChangedFilesRebasesOntoProjectRoot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	projDir := filepath.Join(repoDir, "proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = repoDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")

	fooPath := filepath.Join(projDir, "foo.go")
	if err := os.WriteFile(fooPath, []byte("package proj\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "init")

	// Uncommitted change — this is what diff mode should surface.
	if err := os.WriteFile(fooPath, []byte("package proj\n// touched\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Give proj/ its own .atd so config.LoadFromDir anchors ProjectRoot there,
	// deliberately below repoDir (the actual git toplevel) — mirroring how a
	// submodule/nested project's root differs from the repo `git diff`
	// resolves paths against.
	if err := os.WriteFile(filepath.Join(projDir, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	savedConfig := config.ActiveConfig
	defer func() { config.ActiveConfig = savedConfig }()

	if err := config.LoadFromDir(projDir); err != nil {
		t.Fatalf("config.LoadFromDir failed: %v", err)
	}
	if config.ProjectRoot() != projDir {
		t.Fatalf("expected ProjectRoot %s, got %s", projDir, config.ProjectRoot())
	}

	codeFiles, atomFiles, err := diffChangedFiles(nil)
	if err != nil {
		t.Fatalf("diffChangedFiles failed: %v", err)
	}
	if len(atomFiles) != 0 {
		t.Errorf("expected no atom files, got %v", atomFiles)
	}
	if len(codeFiles) != 1 || codeFiles[0] != "foo.go" {
		t.Errorf("expected codeFiles=[foo.go] (rebased onto ProjectRoot), got %v", codeFiles)
	}
}
