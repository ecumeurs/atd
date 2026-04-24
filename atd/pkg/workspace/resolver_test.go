package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolver(t *testing.T) {
	// Setup a temporary workspace
	tempDir, err := os.MkdirTemp("", "atd-workspace-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create structure:
	// /projA/docs/atom1.atom.md
	// /projB/docs/atom2.atom.md
	// /docs/shared1.atom.md
	
	setupFile := func(path string) {
		dir := filepath.Dir(path)
		os.MkdirAll(dir, 0755)
		os.WriteFile(path, []byte("---\nid: "+strings.TrimSuffix(filepath.Base(path), ".atom.md")+"\n---"), 0644)
	}

	setupFile(filepath.Join(tempDir, "projA/docs/atom1.atom.md"))
	setupFile(filepath.Join(tempDir, "projB/docs/atom2.atom.md"))
	setupFile(filepath.Join(tempDir, "docs/shared1.atom.md"))

	ws := &Workspace{
		WorkspaceConfig: WorkspaceConfig{
			WorkspaceRoot: tempDir,
			Projects: []Project{
				{Name: "projA", Path: "projA", FullDocsPath: filepath.Join(tempDir, "projA/docs")},
				{Name: "projB", Path: "projB", FullDocsPath: filepath.Join(tempDir, "projB/docs")},
			},
		},
	}

	idx, err := ws.BuildIndex()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("LocalResolve", func(t *testing.T) {
		r := NewResolver(ws, idx, "projA")
		parsed, err := r.Resolve("[[atom1]]")
		if err != nil {
			t.Errorf("Expected to find atom1, got %v", err)
		}
		if parsed.Type != ReferenceLocal {
			t.Errorf("Expected ReferenceLocal, got %v", parsed.Type)
		}
	})

	t.Run("CrossProjectResolve", func(t *testing.T) {
		r := NewResolver(ws, idx, "projA")
		parsed, err := r.Resolve("[[atom2]]")
		if err != nil {
			t.Errorf("Expected to find atom2, got %v", err)
		}
		if parsed.Type != ReferenceCrossProject {
			t.Errorf("Expected ReferenceCrossProject, got %v", parsed.Type)
		}
		if parsed.Project != "projB" {
			t.Errorf("Expected project projB, got %s", parsed.Project)
		}
	})

	t.Run("SharedResolve", func(t *testing.T) {
		r := NewResolver(ws, idx, "projA")
		parsed, err := r.Resolve("[[shared1]]")
		if err != nil {
			t.Errorf("Expected to find shared1, got %v", err)
		}
		if parsed.Project != "shared" {
			t.Errorf("Expected project shared, got %s", parsed.Project)
		}
	})

	t.Run("ShouldUpdate", func(t *testing.T) {
		r := NewResolver(ws, idx, "projA")
		shouldUpdate, canonical := r.ShouldUpdate("[[atom2]]")
		if !shouldUpdate {
			t.Error("Expected shouldUpdate to be true for cross-project reference")
		}
		if canonical != "[[projB:atom2]]" {
			t.Errorf("Expected [[projB:atom2]], got %s", canonical)
		}
	})
}
