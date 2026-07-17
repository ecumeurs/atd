package exploration

// Load()/ListFiles() unit back-fill (test_atd_07_26.md §3.1 item 1, §6 WP-7).
// These use a hand-built temp tree rather than testdata/fixture_project
// because the case under test -- unsupported extensions and ignored
// directories being excluded from the walk -- isn't expressed by the
// fixture (per the report's guidance: construct minimal temp trees only
// when the fixture doesn't cover the case).

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/atom"
)

func writeTestFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatalf("mkdir for %s: %v", rel, err)
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// TestLoad_ParsesDocsAndFiltersCodeByExtensionAndPath exercises the full
// Load() pipeline (docs parse, then code-link extraction gated by
// cfg.SupportedExtensions) against a temp tree with one implemented atom
// and several decoys that must NOT contribute a link: an unsupported
// extension, and files sitting in vendor/, node_modules/, dist/, build/.
func TestLoad_ParsesDocsAndFiltersCodeByExtensionAndPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTestFile(t, root, "docs/zz_load_atom.atom.md", "---\nid: zz_load_atom\nhuman_name: \"Load Test Atom\"\n---\n")
	writeTestFile(t, root, "src/main.go", "package p\n\n// @spec-link [[zz_load_atom]]\nfunc F() {}\n")
	writeTestFile(t, root, "src/main_test.go", "package p\n\nimport \"testing\"\n\n// @test-link [[zz_load_atom]]\nfunc TestF(t *testing.T) {}\n")
	writeTestFile(t, root, "src/notes.txt", "@spec-link [[zz_load_atom]] -- unsupported extension, must be ignored\n")
	writeTestFile(t, root, "vendor/ignored.go", "// @spec-link [[zz_load_atom]] -- inside vendor/, must be ignored\n")
	writeTestFile(t, root, "node_modules/ignored.go", "// @spec-link [[zz_load_atom]] -- inside node_modules/, must be ignored\n")
	writeTestFile(t, root, "dist/ignored.go", "// @spec-link [[zz_load_atom]] -- inside dist/, must be ignored\n")
	writeTestFile(t, root, "build/ignored.go", "// @spec-link [[zz_load_atom]] -- inside build/, must be ignored\n")

	cfg := &config.Config{SupportedExtensions: map[string]bool{".go": true}}
	explorer := NewExplorerWithConfig(root, filepath.Join(root, "docs"), cfg)
	if explorer.Resolver != nil {
		t.Fatal("expected no workspace resolver for an isolated temp tree with no .atd.workspace above it")
	}

	if err := explorer.Load(false); err != nil {
		t.Fatalf("Load: %v", err)
	}

	node, ok := explorer.Graph.Atoms["zz_load_atom"]
	if !ok {
		t.Fatal("expected zz_load_atom to be parsed from docs/")
	}
	if len(node.Implementations) != 1 {
		t.Errorf("expected exactly 1 impl link (only src/main.go's .go file counts), got %d: %v", len(node.Implementations), node.Implementations)
	}
	if !node.HasTests {
		t.Error("expected HasTests from src/main_test.go's @test-link")
	}
	if len(explorer.SpecLinks) != 1 {
		t.Errorf("expected exactly 1 SpecLink recorded across the whole tree, got %d: %v", len(explorer.SpecLinks), explorer.SpecLinks)
	}

	// Load(false) is memoized once Graph is non-nil: mutate it and confirm a
	// second Load(false) call is a true no-op.
	explorer.Graph.Atoms["zz_extra_marker"] = &atom.AtomData{ID: "zz_extra_marker"}
	if err := explorer.Load(false); err != nil {
		t.Fatalf("Load(false) second call: %v", err)
	}
	if _, ok := explorer.Graph.Atoms["zz_extra_marker"]; !ok {
		t.Error("expected Load(false) to be a no-op and preserve the marker atom added after the first Load")
	}

	// Load(true) forces a rebuild, discarding anything not actually on disk.
	if err := explorer.Load(true); err != nil {
		t.Fatalf("Load(true): %v", err)
	}
	if _, ok := explorer.Graph.Atoms["zz_extra_marker"]; ok {
		t.Error("expected Load(true) to rebuild the graph from scratch, discarding the marker atom")
	}
}

// TestListFiles_SkipsIgnoredDirectories pins ListFiles' path-based ignore
// filter (explorer.go): vendor/, node_modules/, dist/, build/ are matched
// as plain substrings anywhere in the relative path, so they're skipped
// whether they sit at the tree root or nested. The dotdir check is
// different and narrower: it looks for the literal substring "/." (a slash
// immediately before a dot), which only matches a dotdir *nested* under
// something else -- a dotdir sitting directly at ProjectRoot (e.g. a
// top-level .git/) has no leading "/" before its own name and is NOT
// filtered. This test pins both halves, including that asymmetry.
func TestListFiles_SkipsIgnoredDirectories(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTestFile(t, root, "src/keep.go", "x")
	writeTestFile(t, root, "vendor/skip.go", "x")
	writeTestFile(t, root, "node_modules/skip.go", "x")
	writeTestFile(t, root, "dist/skip.go", "x")
	writeTestFile(t, root, "build/skip.go", "x")
	writeTestFile(t, root, "sub/.git/skip", "x") // nested dotdir: rel contains "/." -> filtered
	writeTestFile(t, root, ".git/topskip", "x")  // KNOWN GAP: see doc comment above -- top-level dotdir is NOT filtered

	e := &Explorer{ProjectRoot: root}
	files, err := e.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	sort.Strings(files)

	want := []string{".git/topskip", "src/keep.go"}
	if !reflect.DeepEqual(files, want) {
		t.Errorf("ListFiles() = %v, want %v", files, want)
	}
}

// TestGetGraph_GetTestLinks pins the two trivial accessor methods
// (explorer.go) that exist purely so callers outside the package don't
// reach into Explorer's exported-but-conventionally-internal fields
// directly.
func TestGetGraph_GetTestLinks(t *testing.T) {
	t.Parallel()
	graph := &DependencyGraph{Atoms: map[string]*atom.AtomData{"zz_a": {ID: "zz_a"}}}
	e := &Explorer{
		Graph:     graph,
		TestLinks: []TestLink{{AtomID: "zz_a", TestFile: "f_test.go", Line: 3}},
	}

	if got := e.GetGraph(); got != graph {
		t.Errorf("GetGraph() returned a different pointer than the one set")
	}
	links := e.GetTestLinks()
	if len(links) != 1 || links[0].AtomID != "zz_a" {
		t.Errorf("GetTestLinks() = %+v, want the one TestLink set on the Explorer", links)
	}
}

// TestLoadWorkspace_NoWorkspaceErrors and
// TestLoadWorkspace_AggregatesProjectSpecLinks cover the workspace-wide
// Load path (explorer.go's LoadWorkspace/loadFileLinks), which is separate
// from the single-project Load() the rest of this file exercises: it reads
// e.Index (already built by NewExplorerWithConfig) rather than walking
// ProjectRoot's own docs/ directory, and fans out across every project's
// CodePaths via loadFileLinks.
func TestLoadWorkspace_NoWorkspaceErrors(t *testing.T) {
	t.Parallel()
	e := &Explorer{}
	if err := e.LoadWorkspace(false); err == nil {
		t.Error("expected an error calling LoadWorkspace with no active workspace")
	}
}

func TestLoadWorkspace_AggregatesProjectSpecLinks(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeTestFile(t, root, "project-a/docs/zz_ws_atom.atom.md", "---\nid: zz_ws_atom\nhuman_name: \"WS Atom\"\n---\n")
	writeTestFile(t, root, "project-a/src/impl.go", "package p\n\n// @spec-link [[zz_ws_atom]]\nfunc F() {}\n")
	writeTestFile(t, root, ".atd.workspace", `{
		"workspace_name": "loadws-test",
		"projects": [{"name": "project-a", "path": "./project-a"}]
	}`)

	projA := filepath.Join(root, "project-a")
	explorer := NewExplorerWithConfig(projA, filepath.Join(projA, "docs"), &config.Config{})
	if explorer.Workspace == nil {
		t.Fatal("expected a workspace to be detected above project-a")
	}

	if err := explorer.LoadWorkspace(false); err != nil {
		t.Fatalf("LoadWorkspace: %v", err)
	}

	if _, ok := explorer.Graph.Atoms["zz_ws_atom"]; !ok {
		t.Fatal("expected zz_ws_atom to be present in the workspace-wide graph")
	}
	if len(explorer.SpecLinks) != 1 || explorer.SpecLinks[0].AtomID != "zz_ws_atom" {
		t.Errorf("expected exactly 1 SpecLink for zz_ws_atom aggregated from project-a's code paths, got %+v", explorer.SpecLinks)
	}
}
