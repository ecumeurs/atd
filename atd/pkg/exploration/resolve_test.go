package exploration

// ResolveAtom/CanonicalAtomID/SuggestAtomID unit back-fill (test_atd_07_26.md
// §3.1 item 1, §6 WP-7, §8.3 #1/#11): bare id, project:-prefixed,
// TYPE_-prefixed, ambiguous cross-project match, and complete miss.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/testutil"
	"atd-tools/pkg/workspace"
)

// TestResolveAtom_StandaloneBareIDAndKnownDefect covers the bare-id and
// complete-miss cases against the standalone fixture_project (no
// .atd.workspace above it, so NewExplorerWithConfig leaves Resolver nil --
// see explorer.go), plus the fix for the matching cmd/atd/cmd/scenario_test.go
// testScenarioS1Standalone pin at the CLI layer (test_atd_07_26.md §8.3 #1):
// ResolveAtom/CanonicalAtomID now retry a redundant TYPE_-prefixed id by
// stripping the known prefix and re-checking the graph, generalizing the
// same trick workspace.Resolver already applied, rather than failing exactly
// like a genuinely nonexistent id.
func TestResolveAtom_StandaloneBareIDAndKnownDefect(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")
	explorer := loadExplorerFromSandbox(t, sb)

	if explorer.Resolver != nil {
		t.Fatal("expected no workspace resolver for a standalone fixture_project sandbox")
	}

	node, err := explorer.ResolveAtom("api_zzfix_beta")
	if err != nil {
		t.Fatalf("bare id resolve failed: %v", err)
	}
	if node.ID != "api_zzfix_beta" {
		t.Errorf("got %q, want api_zzfix_beta", node.ID)
	}

	// FIXED (test_atd_07_26.md §8.3 #1): pkg/workspace's StripKnownPrefix
	// retry is no longer workspace-only -- ResolveAtom/CanonicalAtomID apply
	// the same strip-and-retry in the standalone (no e.Resolver) path, so a
	// redundant TYPE_-prefixed id ("api_api_zzfix_beta" -> strip "api_" ->
	// "api_zzfix_beta") now resolves. See
	// TestResolveAtom_WorkspaceProjectPrefixedTypePrefixedAndAmbiguous below
	// for the workspace-resolver twin of this case.
	node, err = explorer.ResolveAtom("api_api_zzfix_beta")
	if err != nil {
		t.Errorf("expected TYPE_-prefixed id to resolve via the standalone strip-and-retry, got: %v", err)
	} else if node.ID != "api_zzfix_beta" {
		t.Errorf("got %q, want api_zzfix_beta", node.ID)
	}

	nonsense := "api_zzfix_does_not_exist_at_all"
	if _, err := explorer.ResolveAtom(nonsense); err == nil {
		t.Error("expected an error resolving a nonexistent atom id")
	} else if !strings.Contains(err.Error(), nonsense) {
		t.Errorf("expected the error to name the input id %q, got: %v", nonsense, err)
	}

	// CanonicalAtomID mirrors ResolveAtom's resolution logic and must agree.
	canon, err := explorer.CanonicalAtomID("api_zzfix_beta")
	if err != nil || canon != "api_zzfix_beta" {
		t.Errorf("CanonicalAtomID(bare) = (%q, %v), want (api_zzfix_beta, nil)", canon, err)
	}
	canon, err = explorer.CanonicalAtomID("api_api_zzfix_beta")
	if err != nil || canon != "api_zzfix_beta" {
		t.Errorf("CanonicalAtomID(TYPE_-prefixed) = (%q, %v), want (api_zzfix_beta, nil)", canon, err)
	}
	if _, err := explorer.CanonicalAtomID(nonsense); err == nil {
		t.Error("expected CanonicalAtomID to error on a nonexistent id")
	} else if !strings.Contains(err.Error(), nonsense) {
		t.Errorf("expected CanonicalAtomID's error to name the input id %q, got: %v", nonsense, err)
	}
}

// TestResolveAtom_WorkspaceProjectPrefixedTypePrefixedAndAmbiguous builds a
// hand-rolled three-project workspace (testdata/fixture_workspace has no
// duplicate-id case, so this constructs the minimal tree needed) to cover:
// project:-prefixed resolution, TYPE_-prefixed resolution succeeding via
// workspace.Resolver's strip-and-retry (the same mechanism generalized to
// the standalone path above), and ambiguous bare-id resolution when the SAME
// atom id exists in two different non-current projects (test_atd_07_26.md
// §8.3 #11).
func TestResolveAtom_WorkspaceProjectPrefixedTypePrefixedAndAmbiguous(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	writeAtomFile := func(rel, id, humanName string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		content := fmt.Sprintf("---\nid: %s\nhuman_name: %q\n---\n", id, humanName)
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// project-a and project-b both declare an atom with the SAME id -- the
	// ambiguous-match case. project-c is where resolution happens FROM (it
	// has no local atom of that name), and also carries a TYPE_-prefixed
	// reference to its own atom to exercise the strip-and-retry path.
	writeAtomFile("project-a/docs/amb_atom.atom.md", "amb_atom", "Project A's amb_atom")
	writeAtomFile("project-b/docs/amb_atom.atom.md", "amb_atom", "Project B's amb_atom")
	writeAtomFile("project-c/docs/req_local.atom.md", "req_local", "Project C's own atom")

	wsConfig := `{
		"workspace_name": "amb-ws",
		"projects": [
			{"name": "project-a", "path": "./project-a"},
			{"name": "project-b", "path": "./project-b"},
			{"name": "project-c", "path": "./project-c"}
		]
	}`
	if err := os.WriteFile(filepath.Join(root, ".atd.workspace"), []byte(wsConfig), 0644); err != nil {
		t.Fatal(err)
	}

	projC := filepath.Join(root, "project-c")
	explorer := NewExplorerWithConfig(projC, filepath.Join(projC, "docs"), &config.Config{})
	explorer.Graph = &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}

	if explorer.Resolver == nil {
		t.Fatal("expected a workspace resolver to be active for a project nested under .atd.workspace")
	}

	t.Run("project_prefixed", func(t *testing.T) {
		node, err := explorer.ResolveAtom("project-a:amb_atom")
		if err != nil {
			t.Fatalf("project-prefixed resolve failed: %v", err)
		}
		if node.HumanName != "Project A's amb_atom" {
			t.Errorf("expected project-a's copy, got %+v", node)
		}
	})

	t.Run("type_prefixed_resolves_via_workspace_resolver", func(t *testing.T) {
		// Same strip-and-retry mechanism as
		// TestResolveAtom_StandaloneBareIDAndKnownDefect's standalone case,
		// exercised here via the workspace resolver: "requirement_req_local"
		// strips the known "requirement" prefix and retries as "req_local",
		// succeeding.
		node, err := explorer.ResolveAtom("requirement_req_local")
		if err != nil {
			t.Fatalf("expected TYPE_-prefixed id to resolve via the workspace resolver's strip-and-retry, got: %v", err)
		}
		if node.ID != "req_local" {
			t.Errorf("got %q, want req_local", node.ID)
		}
	})

	t.Run("ambiguous_bare_id_errors_naming_candidate_projects", func(t *testing.T) {
		// FIXED (test_atd_07_26.md §8.3 #11): resolving the bare id
		// "amb_atom" from project-c -- which has no local copy -- used to
		// fall through to AtomIndex.FindAtom and silently return whichever
		// project was registered first (project-a), with zero indication to
		// the caller that the id was ambiguous. AtomIndex now tracks every
		// project that declares an id (AtomIndex.Duplicates), and the
		// resolver checks that before falling back to FindAtom, returning a
		// loud ErrAmbiguousAtom naming every candidate project and the
		// "project:id" syntax needed to disambiguate.
		_, err := explorer.ResolveAtom("amb_atom")
		if err == nil {
			t.Fatal("expected an ambiguity error resolving a bare id that exists identically in two projects")
		}
		if !errors.Is(err, workspace.ErrAmbiguousAtom) {
			t.Errorf("expected the error to wrap workspace.ErrAmbiguousAtom, got: %v", err)
		}
		if !strings.Contains(err.Error(), "amb_atom") {
			t.Errorf("expected the error to name the ambiguous id, got: %v", err)
		}
		if !strings.Contains(err.Error(), "project-a") || !strings.Contains(err.Error(), "project-b") {
			t.Errorf("expected the error to name both candidate projects, got: %v", err)
		}
		if !strings.Contains(err.Error(), "project-a:amb_atom") {
			t.Errorf("expected the error to suggest the \"project:id\" syntax, got: %v", err)
		}
	})

	t.Run("complete_miss", func(t *testing.T) {
		if _, err := explorer.ResolveAtom("totally_bogus_atom_id_xyz"); err == nil {
			t.Error("expected an error resolving a nonexistent atom id")
		}
	})

	t.Run("canonical_atom_id_cross_project_and_local", func(t *testing.T) {
		canon, err := explorer.CanonicalAtomID("project-a:amb_atom")
		if err != nil || canon != "project-a:amb_atom" {
			t.Errorf("CanonicalAtomID(project-prefixed) = (%q, %v), want (project-a:amb_atom, nil)", canon, err)
		}

		// req_local is local to project-c (the resolving context), so its
		// canonical form carries no project prefix even though a workspace
		// resolver is active.
		localCanon, err := explorer.CanonicalAtomID("req_local")
		if err != nil || localCanon != "req_local" {
			t.Errorf("CanonicalAtomID(local bare id) = (%q, %v), want (req_local, nil)", localCanon, err)
		}

		if _, err := explorer.CanonicalAtomID("totally_bogus_atom_id_xyz"); err == nil {
			t.Error("expected CanonicalAtomID to error on a nonexistent id via the workspace resolver")
		}
	})
}

// TestSuggestAtomID pins SuggestAtomID's narrow fallback: with no
// workspace.AtomIndex active, it only helps when the bare id being searched
// for is the TAIL of some "project:bare_id"-shaped key already sitting in
// the graph (e.g. left there by a prior CrawlWorkspaceDocs pass) -- it does
// NOT do fuzzy matching or handle the TYPE_-prefixed case at all.
func TestSuggestAtomID(t *testing.T) {
	e := newLinkTestExplorer()
	e.Graph.Atoms["project-a:atom_x"] = &atom.AtomData{ID: "atom_x"}
	e.Graph.Atoms["atom_y"] = &atom.AtomData{ID: "atom_y"}

	if got := e.SuggestAtomID("atom_x"); got != "project-a:atom_x" {
		t.Errorf("SuggestAtomID(atom_x) = %q, want project-a:atom_x", got)
	}
	if got := e.SuggestAtomID("does_not_exist_anywhere"); got != "" {
		t.Errorf("SuggestAtomID(unknown) = %q, want empty", got)
	}
	// A bare id that IS already a key gets no suggestion from this path --
	// SuggestAtomID is a "did you mean" helper for the miss case, not a
	// general lookup (ResolveAtom/CanonicalAtomID own the direct-hit path).
	if got := e.SuggestAtomID("atom_y"); got != "" {
		t.Errorf("SuggestAtomID(exact existing bare id) = %q, want empty (not this function's job)", got)
	}
}
