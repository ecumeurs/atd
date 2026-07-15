package atom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
)

// TestUpdateLinksScopeGuardBlocksMisanchoredConfig pins P-1 (test_atd_07_26.md
// §3.6): a rename request whose docsPath sits outside the active
// config.ProjectRoot() must never rewrite files under that unrelated root.
// This is the direct regression test for incident I-1, where a leaked,
// cwd-fallback config caused UpdateLinks' second (source-file) walk to
// rewrite the test suite's own source directory.
//
// otherRoot is given a real .atd (non-fallback) specifically so this test
// isolates the "docsPath outside ProjectRoot" branch of the guard from the
// separate loadedFromFallback branch (covered by the sibling test below).
func TestUpdateLinksScopeGuardBlocksMisanchoredConfig(t *testing.T) {
	saved := config.Snapshot()
	t.Cleanup(func() { config.Restore(saved) })

	// "otherRoot" stands in for the real project/source tree that a leaked
	// config might anchor to (e.g. the test binary's own package dir).
	otherRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(otherRoot, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(otherRoot, "victim.go")
	if err := os.WriteFile(srcFile, []byte("// see [[old_id]] for details\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Deliberately mis-anchor the active config: ProjectRoot() points at
	// otherRoot, but the rename we're about to run scopes its docsPath to a
	// completely unrelated temp dir.
	config.ActiveConfig = config.Config{}
	if err := config.LoadFromDir(otherRoot); err != nil {
		t.Fatalf("config.LoadFromDir failed: %v", err)
	}
	if config.LoadedFromFallback() {
		t.Fatal("expected a real, non-fallback-anchored config (a .atd file is present in otherRoot)")
	}
	if config.ProjectRoot() != otherRoot {
		t.Fatalf("expected ProjectRoot %s, got %s", otherRoot, config.ProjectRoot())
	}

	docsPath := t.TempDir() // unrelated to otherRoot
	numUpdates := UpdateLinks(docsPath, "old_id", "new_id")

	// The docs walk found nothing (docsPath is empty), so the only thing
	// this call could have mutated is otherRoot's source file — which the
	// scope guard must have refused to touch.
	if numUpdates != 0 {
		t.Errorf("expected 0 updates (docsPath has no atom files), got %d", numUpdates)
	}
	after, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "[[old_id]]") || strings.Contains(string(after), "[[new_id]]") {
		t.Errorf("scope guard failed: file outside docsPath was rewritten:\n%s", after)
	}
}

// TestUpdateLinksScopeGuardBlocksFallbackAnchoredConfig pins the second half
// of P-1: when no real .atd was found and the config fell back to cwd
// (loadedFromFallback), the source-file walk must refuse even if docsPath
// happens to resolve inside that fallback-anchored root.
func TestUpdateLinksScopeGuardBlocksFallbackAnchoredConfig(t *testing.T) {
	saved := config.Snapshot()
	t.Cleanup(func() { config.Restore(saved) })

	fallbackRoot := t.TempDir()
	srcFile := filepath.Join(fallbackRoot, "victim.go")
	if err := os.WriteFile(srcFile, []byte("// see [[old_id]] for details\n"), 0644); err != nil {
		t.Fatal(err)
	}
	docsDir := filepath.Join(fallbackRoot, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	savedWD, _ := os.Getwd()
	if err := os.Chdir(fallbackRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(savedWD) })

	// No .atd anywhere above fallbackRoot (t.TempDir() lives under the OS
	// temp root, which has none) — LoadFromDirLegacy must fall back to cwd.
	config.ActiveConfig = config.Config{}
	if err := config.LoadFromDir(fallbackRoot); err != nil {
		t.Fatalf("config.LoadFromDir failed: %v", err)
	}
	if !config.LoadedFromFallback() {
		t.Fatal("expected config to be fallback-anchored for this test to be meaningful")
	}
	if config.ProjectRoot() != fallbackRoot {
		t.Fatalf("expected ProjectRoot %s, got %s", fallbackRoot, config.ProjectRoot())
	}

	// docsDir IS inside the fallback-anchored ProjectRoot, but the fallback
	// check must still refuse the source-file walk.
	UpdateLinks(docsDir, "old_id", "new_id")

	after, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), "[[old_id]]") || strings.Contains(string(after), "[[new_id]]") {
		t.Errorf("fallback guard failed: file was rewritten despite fallback-anchored config:\n%s", after)
	}
}

// TestUpdateLinksPropagatesWithProperlyAnchoredConfig is the positive
// counterpart (S7 in test_atd_07_26.md §4): with a real .atd anchoring
// ProjectRoot and a docsPath genuinely nested inside it, rename propagation
// into source files must still work — the scope guard must not break the
// legitimate feature.
func TestUpdateLinksPropagatesWithProperlyAnchoredConfig(t *testing.T) {
	saved := config.Snapshot()
	t.Cleanup(func() { config.Restore(saved) })

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	docsDir := filepath.Join(projectDir, "docs")
	srcDir := filepath.Join(projectDir, "src")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	docFile := filepath.Join(docsDir, "ref.atom.md")
	docContent := "---\nid: ref_atom\nparents:\n  - [[old_id]]\n---\nSee [[old_id]] for details."
	if err := os.WriteFile(docFile, []byte(docContent), 0644); err != nil {
		t.Fatal(err)
	}

	srcFile := filepath.Join(srcDir, "beta.go")
	srcContent := "package src\n\n// @spec-link [[old_id]]\nfunc Beta() {}\n"
	if err := os.WriteFile(srcFile, []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	// A test file and a testdata fixture, each containing the literal —
	// P-2 says neither should be touched by the propagation walk.
	testFile := filepath.Join(srcDir, "beta_test.go")
	testContent := "package src\n\n// fixture: [[old_id]]\n"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatal(err)
	}
	testdataDir := filepath.Join(srcDir, "testdata")
	if err := os.MkdirAll(testdataDir, 0755); err != nil {
		t.Fatal(err)
	}
	testdataFile := filepath.Join(testdataDir, "fixture.go")
	if err := os.WriteFile(testdataFile, []byte("// [[old_id]]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	config.ActiveConfig = config.Config{}
	if err := config.LoadFromDir(projectDir); err != nil {
		t.Fatalf("config.LoadFromDir failed: %v", err)
	}
	if config.LoadedFromFallback() {
		t.Fatal("expected a real, non-fallback-anchored config (a .atd file is present)")
	}
	if config.ProjectRoot() != projectDir {
		t.Fatalf("expected ProjectRoot %s, got %s", projectDir, config.ProjectRoot())
	}

	numUpdates := UpdateLinks(docsDir, "old_id", "new_id")
	if numUpdates < 1 {
		t.Fatalf("expected at least 1 file updated, got %d", numUpdates)
	}

	updatedDoc, _ := os.ReadFile(docFile)
	if !strings.Contains(string(updatedDoc), "[[new_id]]") {
		t.Errorf("expected doc ref rewritten to [[new_id]], got:\n%s", updatedDoc)
	}

	updatedSrc, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updatedSrc), "[[new_id]]") {
		t.Errorf("expected legitimate source-file propagation to rewrite [[old_id]] -> [[new_id]], got:\n%s", updatedSrc)
	}

	// P-2: _test.go and testdata/ must be left untouched.
	untouchedTest, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(untouchedTest), "[[old_id]]") {
		t.Errorf("P-2 violated: _test.go file was rewritten:\n%s", untouchedTest)
	}
	untouchedTestdata, err := os.ReadFile(testdataFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(untouchedTestdata), "[[old_id]]") {
		t.Errorf("P-2 violated: testdata/ file was rewritten:\n%s", untouchedTestdata)
	}
}
