package exploration

// ResolveAtom/CanonicalAtomID/SuggestAtomID unit back-fill (test_atd_07_26.md
// §3.1 item 1, §6 WP-7): bare id, project:-prefixed, TYPE_-prefixed,
// ambiguous cross-project match, and complete miss.

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/testutil"
)

// TestResolveAtom_StandaloneBareIDAndKnownDefect covers the bare-id and
// complete-miss cases against the standalone fixture_project (no
// .atd.workspace above it, so NewExplorerWithConfig leaves Resolver nil --
// see explorer.go), plus the same KNOWN DEFECT
// cmd/atd/cmd/scenario_test.go's testScenarioS1Standalone pins at the CLI
// layer: a redundant TYPE_-prefixed id has no strip-and-retry mechanism to
// fall back on outside workspace.Resolver, so it fails exactly like a
// genuinely nonexistent id.
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

	// KNOWN DEFECT: pkg/workspace/resolver.go's strip-known-prefix retry
	// (lets "requirement_req_x" resolve to "req_x") lives only on
	// workspace.Resolver. A standalone project never constructs one, so
	// CanonicalAtomID/ResolveAtom fall back to an exact map lookup with no
	// retry -- a redundant TYPE_-prefixed id fails identically to a
	// genuinely nonsense one. See TestResolveAtom_WorkspaceProjectPrefixed
	// below for proof the underlying mechanism does work once a workspace
	// resolver is active, and cmd/atd/cmd/scenario_test.go's
	// testScenarioS1Standalone for the CLI-level twin of this pin.
	if _, err := explorer.ResolveAtom("api_api_zzfix_beta"); err == nil {
		t.Error("KNOWN DEFECT expectation changed: TYPE_-prefixed id unexpectedly resolved in a standalone project; if this now passes, generalize the fix and update this pin")
	}

	if _, err := explorer.ResolveAtom("api_zzfix_does_not_exist_at_all"); err == nil {
		t.Error("expected an error resolving a nonexistent atom id")
	}

	// CanonicalAtomID mirrors ResolveAtom's resolution logic and must agree.
	canon, err := explorer.CanonicalAtomID("api_zzfix_beta")
	if err != nil || canon != "api_zzfix_beta" {
		t.Errorf("CanonicalAtomID(bare) = (%q, %v), want (api_zzfix_beta, nil)", canon, err)
	}
	if _, err := explorer.CanonicalAtomID("api_zzfix_does_not_exist_at_all"); err == nil {
		t.Error("expected CanonicalAtomID to error on a nonexistent id")
	}
}

// TestResolveAtom_WorkspaceProjectPrefixedTypePrefixedAndAmbiguous builds a
// hand-rolled three-project workspace (testdata/fixture_workspace has no
// duplicate-id case, so this constructs the minimal tree needed) to cover:
// project:-prefixed resolution, TYPE_-prefixed resolution succeeding via
// workspace.Resolver's strip-and-retry (in contrast to the standalone KNOWN
// DEFECT above), and ambiguous bare-id resolution when the SAME atom id
// exists in two different non-current projects.
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
		// In direct contrast to the standalone KNOWN DEFECT
		// (TestResolveAtom_StandaloneBareIDAndKnownDefect): with a workspace
		// resolver active, "requirement_req_local" strips the known
		// "requirement" prefix and retries as "req_local", succeeding.
		node, err := explorer.ResolveAtom("requirement_req_local")
		if err != nil {
			t.Fatalf("expected TYPE_-prefixed id to resolve via the workspace resolver's strip-and-retry, got: %v", err)
		}
		if node.ID != "req_local" {
			t.Errorf("got %q, want req_local", node.ID)
		}
	})

	t.Run("ambiguous_bare_id_silently_picks_first_registered_project", func(t *testing.T) {
		// KNOWN (documented) BEHAVIOR, not an error: resolving the bare id
		// "amb_atom" from project-c -- which has no local copy -- falls
		// through to AtomIndex.FindAtom, whose ByID map was built by
		// AtomIndex.BuildIndex's "first one wins" loop over
		// ws.Projects (pkg/workspace/registry.go). Since project-a is
		// listed first in .atd.workspace, it silently wins with zero
		// indication to the caller that the id was ambiguous.
		node, err := explorer.ResolveAtom("amb_atom")
		if err != nil {
			t.Fatalf("expected the ambiguous bare id to resolve silently (first-project-wins), got error: %v", err)
		}
		if node.HumanName != "Project A's amb_atom" {
			t.Errorf("expected the ambiguous match to silently resolve to project-a's (first-listed) copy, got %+v -- if AtomIndex.BuildIndex's tie-break changed, update this pin", node)
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
