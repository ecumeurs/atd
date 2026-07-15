package exploration

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkspaceSearchAndAssemble(t *testing.T) {
	saved := config.Snapshot()
	defer config.Restore(saved)

	tmpDir, err := os.MkdirTemp("", "atd-search-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create workspace structure
	projA := filepath.Join(tmpDir, "project-a")
	projB := filepath.Join(tmpDir, "project-b")
	os.MkdirAll(filepath.Join(projA, "docs"), 0755)
	os.MkdirAll(filepath.Join(projB, "docs"), 0755)

	atomA := `---
id: atom-a
human_name: "Unique Search Term"
---
`
	atomB := `---
id: atom-b
human_name: "Another Atom"
parents:
  - [[project-a:atom-a]]
---
`
	os.WriteFile(filepath.Join(projA, "docs", "atom-a.atom.md"), []byte(atomA), 0644)
	os.WriteFile(filepath.Join(projB, "docs", "atom-b.atom.md"), []byte(atomB), 0644)

	// Configure active workspace
	config.ActiveConfig.Workspace = &config.WorkspaceConfig{
		WorkspaceName: "test-ws",
		LoadedFrom:    tmpDir,
		Projects: []config.ProjectConfig{
			{Name: "project-a", Path: "./project-a"},
			{Name: "project-b", Path: "./project-b"},
		},
	}

	// 1. Test WorkspaceSearch
	opts := SearchOptions{
		Query:     "Unique",
		Workspace: true,
	}
	results, err := WorkspaceSearch(opts)
	if err != nil {
		t.Fatalf("WorkspaceSearch failed: %v", err)
	}

	found := false
	for _, r := range results {
		if r.Project == "project-a" && strings.Contains(r.FilePath, "atom-a.atom.md") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find atom-a in project-a results")
	}

	// 2. Test Workspace Assemble
	// This tests if CrawlWorkspaceDocs correctly tags atoms for GetProject()
	graph := &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	err = CrawlWorkspaceDocs(graph)
	if err != nil {
		t.Fatalf("CrawlWorkspaceDocs failed: %v", err)
	}

	nodeA, ok := graph.Atoms["atom-a"]
	if !ok {
		t.Fatalf("atom-a not found in graph")
	}
	if nodeA.GetProject() != "project-a" {
		t.Errorf("expected project-a, got %s", nodeA.GetProject())
	}

	nodeB, ok := graph.Atoms["atom-b"]
	if !ok {
		t.Fatalf("atom-b not found in graph")
	}
	if nodeB.GetProject() != "project-b" {
		t.Errorf("expected project-b, got %s", nodeB.GetProject())
	}
}
