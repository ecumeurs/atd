package exploration

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAtom(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atd-workspace-test")
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
human_name: "Atom A"
---
`
	os.WriteFile(filepath.Join(projA, "docs", "atom-a.atom.md"), []byte(atomA), 0644)

	wsConfig := `{
		"workspace_name": "test-ws",
		"projects": [
			{"name": "project-a", "path": "./project-a"},
			{"name": "project-b", "path": "./project-b"}
		]
	}`
	os.WriteFile(filepath.Join(tmpDir, ".atd.workspace"), []byte(wsConfig), 0644)

	// Configure active workspace
	config.ActiveConfig.Workspace = &config.WorkspaceConfig{
		WorkspaceName: "test-ws",
		LoadedFrom:    tmpDir,
		Projects: []config.ProjectConfig{
			{Name: "project-a", Path: "./project-a"},
			{Name: "project-b", Path: "./project-b"},
		},
	}

	explorer := NewExplorer(projB, filepath.Join(projB, "docs"))
	explorer.Graph = &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}

	// Test cross-project bare resolution (auto-discover)
	_, err = explorer.ResolveAtom("atom-a")
	if err != nil {
		t.Errorf("expected success for auto-discovered cross-project atom, got error: %v", err)
	}

	// Test cross-project resolution
	node, err := explorer.ResolveAtom("project-a:atom-a")
	if err != nil {
		t.Fatalf("failed to resolve project-a:atom-a: %v", err)
	}
	if node.ID != "atom-a" {
		t.Errorf("expected atom-a, got %s", node.ID)
	}

	// Test non-existent project
	_, err = explorer.ResolveAtom("project-c:atom-c")
	if err == nil {
		t.Errorf("expected error for non-existent project")
	}
}
