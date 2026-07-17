package workspace

import (
	"errors"
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

	t.Run("StripKnownPrefixRetry", func(t *testing.T) {
		r := NewResolver(ws, idx, "projA")
		// "atom1" exists in projA; a bogus "requirement_atom1" reference should
		// resolve by stripping the known "requirement" type/layer prefix.
		parsed, err := r.Resolve("[[requirement_atom1]]")
		if err != nil {
			t.Fatalf("Expected strip-and-retry to resolve requirement_atom1, got %v", err)
		}
		if parsed.AtomID != "atom1" {
			t.Errorf("Expected canonical AtomID atom1, got %s", parsed.AtomID)
		}
		if parsed.Type != ReferenceLocal {
			t.Errorf("Expected ReferenceLocal, got %v", parsed.Type)
		}
	})

	t.Run("StripKnownPrefixRetryCrossProject", func(t *testing.T) {
		r := NewResolver(ws, idx, "projA")
		// "atom2" lives in projB; "api_atom2" should strip "api" and resolve cross-project.
		parsed, err := r.Resolve("[[api_atom2]]")
		if err != nil {
			t.Fatalf("Expected strip-and-retry to resolve api_atom2, got %v", err)
		}
		if parsed.AtomID != "atom2" {
			t.Errorf("Expected canonical AtomID atom2, got %s", parsed.AtomID)
		}
		if parsed.Project != "projB" {
			t.Errorf("Expected project projB, got %s", parsed.Project)
		}
	})

	t.Run("UnknownPrefixStaysUnresolved", func(t *testing.T) {
		r := NewResolver(ws, idx, "projA")
		// "bogus" is not a known type/layer token, so no retry should happen
		// and resolution should fail loudly.
		_, err := r.Resolve("[[bogus_atom1]]")
		if err != ErrAtomNotFound {
			t.Errorf("Expected ErrAtomNotFound for unknown-prefix id, got %v", err)
		}
	})

	t.Run("NoMatchEvenAfterStrip", func(t *testing.T) {
		r := NewResolver(ws, idx, "projA")
		_, err := r.Resolve("[[requirement_does_not_exist]]")
		if err != ErrAtomNotFound {
			t.Errorf("Expected ErrAtomNotFound when stripped id still doesn't exist, got %v", err)
		}
	})
}

// TestResolver_AmbiguousBareID pins the fix for test_atd_07_26.md §8.3 #11:
// a bare id that exists identically in more than one non-current project
// used to silently resolve to whichever project BuildIndex's "first one
// wins" loop registered first, with no signal to the caller that the id was
// ambiguous. It must now return a loud error wrapping ErrAmbiguousAtom that
// names every candidate project and the "project:id" syntax.
func TestResolver_AmbiguousBareID(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "atd-workspace-ambiguous-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	setupFile := func(path, id string) {
		dir := filepath.Dir(path)
		os.MkdirAll(dir, 0755)
		os.WriteFile(path, []byte("---\nid: "+id+"\n---"), 0644)
	}

	// projA and projB both declare "dup_atom"; projC (where resolution
	// happens from) has no local copy, so there is no local-project
	// tie-breaker to fall back on.
	setupFile(filepath.Join(tempDir, "projA/docs/dup_atom.atom.md"), "dup_atom")
	setupFile(filepath.Join(tempDir, "projB/docs/dup_atom.atom.md"), "dup_atom")
	setupFile(filepath.Join(tempDir, "projC/docs/own_atom.atom.md"), "own_atom")

	ws := &Workspace{
		WorkspaceConfig: WorkspaceConfig{
			WorkspaceRoot: tempDir,
			Projects: []Project{
				{Name: "projA", Path: "projA", FullDocsPath: filepath.Join(tempDir, "projA/docs")},
				{Name: "projB", Path: "projB", FullDocsPath: filepath.Join(tempDir, "projB/docs")},
				{Name: "projC", Path: "projC", FullDocsPath: filepath.Join(tempDir, "projC/docs")},
			},
		},
	}

	idx, err := ws.BuildIndex()
	if err != nil {
		t.Fatal(err)
	}

	r := NewResolver(ws, idx, "projC")
	_, err = r.Resolve("[[dup_atom]]")
	if err == nil {
		t.Fatal("expected an ambiguity error resolving a bare id that exists identically in two projects")
	}
	if !errors.Is(err, ErrAmbiguousAtom) {
		t.Errorf("expected the error to wrap ErrAmbiguousAtom, got: %v", err)
	}
	if !strings.Contains(err.Error(), "dup_atom") {
		t.Errorf("expected the error to name the ambiguous id, got: %v", err)
	}
	if !strings.Contains(err.Error(), "projA") || !strings.Contains(err.Error(), "projB") {
		t.Errorf("expected the error to name both candidate projects, got: %v", err)
	}
	if !strings.Contains(err.Error(), "projA:dup_atom") {
		t.Errorf("expected the error to suggest the \"project:id\" syntax, got: %v", err)
	}

	// A local copy in the resolving project always wins deterministically --
	// no ambiguity signal needed, since that's an intentional local shadow,
	// not an accidental collision.
	rLocal := NewResolver(ws, idx, "projA")
	parsed, err := rLocal.Resolve("[[dup_atom]]")
	if err != nil {
		t.Fatalf("expected projA's own local copy to resolve without ambiguity, got: %v", err)
	}
	if parsed.Type != ReferenceLocal {
		t.Errorf("expected ReferenceLocal for projA's own copy, got %v", parsed.Type)
	}
}
