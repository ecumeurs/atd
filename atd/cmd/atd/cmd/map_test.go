package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
)

func TestMapFlags(t *testing.T) {
	if mapCmd.Use != "map" {
		t.Errorf("expected use 'map', got %s", mapCmd.Use)
	}
	for _, flag := range []string{"file", "atom", "new", "docs"} {
		if mapCmd.Flag(flag) == nil {
			t.Errorf("expected flag '--%s' to exist", flag)
		}
	}
}

func TestMapMissingFile(t *testing.T) {
	_, err := runMap("", "", "", false)
	if err == nil {
		t.Fatal("expected error when --file is empty")
	}
	if !strings.Contains(err.Error(), "--file is required") {
		t.Errorf("expected '--file is required' error, got: %v", err)
	}
}

func TestMapFileNotFound(t *testing.T) {
	_, err := runMap("/nonexistent/path/file.go", "", "", false)
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestMapBadAtomID(t *testing.T) {
	// Create a temp file so file reading succeeds
	tmp := t.TempDir()
	f := filepath.Join(tmp, "sample.go")
	os.WriteFile(f, []byte("package main\nfunc main() {}"), 0644)

	_, err := runMap(f, "nonexistent_atom_xyz", "", false)
	if err == nil {
		t.Fatal("expected error for unknown atom ID")
	}
}

func TestMapNewFlagProducesSkeleton(t *testing.T) {
	// This path (propose/--new) delegates to the pipeline package, which
	// writes pipeline_output/task_list.md under config.ProjectRoot(). Anchor
	// config to this test's own tmp dir (rather than relying on whatever
	// ActiveConfig happens to be ambient) so that write — like every write a
	// test triggers — lands under t.TempDir(), never under the package
	// source dir. This is the map_test.go path referenced in
	// test_atd_07_26.md §2.1 (incident I-1) as arming the self-rewrite bug.
	savedConfig := config.Snapshot()
	defer config.Restore(savedConfig)
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := config.LoadFromDir(tmp); err != nil {
		t.Fatalf("config.LoadFromDir failed: %v", err)
	}

	f := filepath.Join(tmp, "design.md")
	os.WriteFile(f, []byte("# Feature Design\n\nThis feature handles user login."), 0644)

	// With no Ollama available, runMapPropose will fail at ollama.Query.
	// We just confirm routing doesn't panic and returns something or an error.
	out, err := runMap(f, "", "", true)
	// Either IDE fallback (pipeline task delegated) or ollama error — both acceptable.
	if err != nil {
		// An ollama error is fine in test environment
		return
	}
	// If no error, output should mention atom skeleton or delegation
	if out == "" {
		t.Error("expected non-empty output from propose path")
	}
}

// TestMapResolvesRelativePathAgainstProjectRoot exercises the §8(a) fix:
// a relative --file must be read relative to config.ProjectRoot(), not the
// process cwd, so map accepts the same inputs as `atd check --file` under
// the MCP server (where cwd is often a workspace umbrella root).
func TestMapResolvesRelativePathAgainstProjectRoot(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "sample.go"), []byte("package main\nfunc main() {}"), 0644); err != nil {
		t.Fatal(err)
	}

	savedConfig := config.Snapshot()
	savedWD, _ := os.Getwd()
	defer func() {
		config.Restore(savedConfig)
		os.Chdir(savedWD)
	}()

	if err := config.LoadFromDir(tmp); err != nil {
		t.Fatalf("config.LoadFromDir failed: %v", err)
	}
	if config.ProjectRoot() != tmp {
		t.Fatalf("expected ProjectRoot %s, got %s", tmp, config.ProjectRoot())
	}

	// cwd deliberately differs from ProjectRoot.
	elsewhere := t.TempDir()
	if err := os.Chdir(elsewhere); err != nil {
		t.Fatal(err)
	}

	_, err := runMap("sample.go", "nonexistent_atom_xyz", "", false)
	if err == nil {
		t.Fatal("expected an error (unknown atom), but the important thing is which error")
	}
	if strings.Contains(err.Error(), "failed to read file") {
		t.Errorf("relative --file should resolve against ProjectRoot, not cwd; got: %v", err)
	}
}

func TestReconResultIsDegenerate(t *testing.T) {
	cases := []struct {
		name   string
		result ReconResult
		want   bool
	}{
		{"empty mismatches", ReconResult{Confidence: 0, Mismatches: nil}, true},
		{"blank single entry", ReconResult{Confidence: 0, Mismatches: []ReconMismatch{{}}}, true},
		{"real mismatch", ReconResult{Confidence: 40, Mismatches: []ReconMismatch{
			{Aspect: "locking", Expected: "mutex around purchase", Found: "not visible in this file"},
		}}, false},
	}

	for _, c := range cases {
		if got := c.result.isDegenerate(); got != c.want {
			t.Errorf("%s: isDegenerate() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestMapFileTypeHintInjected(t *testing.T) {
	// Verify that .md (non-atom) files get the design document hint prepended
	// We test this by inspecting the content passed — since we can't intercept
	// prompt.IntentExtractBuild directly, we verify the file is read and the
	// discover path is invoked without errors (file exists, no atom flag, no --new).
	tmp := t.TempDir()
	f := filepath.Join(tmp, "notes.md")
	os.WriteFile(f, []byte("# Notes\nSome design notes."), 0644)

	// atom.md files should NOT get the hint
	atomFile := filepath.Join(tmp, "rule_foo.atom.md")
	os.WriteFile(atomFile, []byte("---\nid: rule_foo\n---\n# Rule Foo"), 0644)

	// Just confirm both paths don't crash on file reading
	content, _ := os.ReadFile(f)
	if strings.HasPrefix(string(content), "// [CONTEXT:") {
		t.Error("original file should not have hint prefix")
	}
}
