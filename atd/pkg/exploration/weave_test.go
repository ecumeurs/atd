package exploration

// weave.go unit back-fill. Weave is the one function in this package that
// REWRITES atom files on disk (the same rewrite machinery family as incident
// I-1's UpdateLinks walk, test_atd_07_26.md §2.1), so every test here runs
// against a testutil.Sandbox copy or a t.TempDir() tree — never the
// checked-in fixture. Deliberately NOT exercised: Weave()'s nil-Workspace
// discovery branch (os.Getwd + workspace.LoadWorkspace(cwd)); on a machine
// with an .atd.workspace anywhere above the test binary's cwd it would
// route the weave at that real workspace's files. weaveSingleProject and
// weaveWorkspace are driven directly instead.

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"atd-tools/pkg/testutil"
	"atd-tools/pkg/workspace"
)

func TestRenderRefBlock(t *testing.T) {
	t.Parallel()

	if got := renderRefBlock("dependents", nil); !reflect.DeepEqual(got, []string{"dependents: []"}) {
		t.Errorf("empty refs: got %v, want [dependents: []]", got)
	}

	// Bare ids and already-bracketed refs must both come out single-wrapped.
	got := renderRefBlock("parents", []string{"bare_id", "[[bracketed_id]]"})
	want := []string{"parents:", "  - [[bare_id]]", "  - [[bracketed_id]]"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRewriteAtomLinks(t *testing.T) {
	t.Parallel()

	write := func(t *testing.T, content string) string {
		t.Helper()
		p := filepath.Join(t.TempDir(), "a.atom.md")
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	read := func(t *testing.T, p string) string {
		t.Helper()
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	t.Run("rewrites_existing_dependents_block", func(t *testing.T) {
		t.Parallel()
		p := write(t, "---\nid: zz_rw\ndependents:\n  - [[old_dep]]\n---\n# Body\n")
		changed, err := rewriteAtomLinks(p, nil, []string{"new_dep"}, "")
		if err != nil || !changed {
			t.Fatalf("changed=%v err=%v, want true, nil", changed, err)
		}
		got := read(t, p)
		if !strings.Contains(got, "  - [[new_dep]]") || strings.Contains(got, "old_dep") {
			t.Errorf("dependents block not rewritten:\n%s", got)
		}
		if !strings.Contains(got, "# Body") {
			t.Errorf("body lost during rewrite:\n%s", got)
		}
	})

	t.Run("inserts_dependents_when_field_missing", func(t *testing.T) {
		t.Parallel()
		p := write(t, "---\nid: zz_rw\n---\n# Body\n")
		changed, err := rewriteAtomLinks(p, nil, []string{"added_dep"}, "")
		if err != nil || !changed {
			t.Fatalf("changed=%v err=%v, want true, nil", changed, err)
		}
		got := read(t, p)
		if !strings.Contains(got, "dependents:\n  - [[added_dep]]") {
			t.Errorf("dependents block not inserted into frontmatter:\n%s", got)
		}
		// The insertion must land INSIDE the frontmatter, before the closing ---.
		if idx := strings.Index(got, "dependents:"); idx > strings.LastIndex(got, "---") {
			t.Errorf("dependents block inserted after frontmatter end:\n%s", got)
		}
	})

	t.Run("updates_parents_when_provided", func(t *testing.T) {
		t.Parallel()
		p := write(t, "---\nid: zz_rw\nparents:\n  - [[old_parent]]\ndependents: []\n---\n")
		changed, err := rewriteAtomLinks(p, []string{"[[proj:new_parent]]"}, nil, "")
		if err != nil || !changed {
			t.Fatalf("changed=%v err=%v, want true, nil", changed, err)
		}
		got := read(t, p)
		if !strings.Contains(got, "  - [[proj:new_parent]]") || strings.Contains(got, "old_parent") {
			t.Errorf("parents block not rewritten:\n%s", got)
		}
	})

	t.Run("no_frontmatter_is_a_silent_noop", func(t *testing.T) {
		t.Parallel()
		content := "# Just a markdown file, no frontmatter\n"
		p := write(t, content)
		changed, err := rewriteAtomLinks(p, nil, []string{"dep"}, "")
		if err != nil || changed {
			t.Fatalf("changed=%v err=%v, want false, nil (files without frontmatter are skipped, not errored)", changed, err)
		}
		if got := read(t, p); got != content {
			t.Errorf("file without frontmatter was modified:\n%s", got)
		}
	})

	t.Run("identical_content_reports_unchanged", func(t *testing.T) {
		t.Parallel()
		p := write(t, "---\nid: zz_rw\ndependents:\n  - [[dep_a]]\n---\n")
		changed, err := rewriteAtomLinks(p, nil, []string{"dep_a"}, "")
		if err != nil || changed {
			t.Fatalf("changed=%v err=%v, want false, nil (rewrite producing identical bytes must not report a change)", changed, err)
		}
	})

	t.Run("missing_file_errors", func(t *testing.T) {
		t.Parallel()
		if _, err := rewriteAtomLinks(filepath.Join(t.TempDir(), "absent.atom.md"), nil, nil, ""); err == nil {
			t.Error("expected an error for a nonexistent file")
		}
	})
}

// TestWeaveSingleProject_IdempotentOnConsistentFixture pins that
// fixture_project's checked-in frontmatter is already fully woven: a weave
// pass over an untouched sandbox copy must rewrite zero files. This is the
// property that makes the fixture safe to share across the suite — any
// change to the fixture or to weave's rendering that breaks it shows up
// here, not as mysterious churn in unrelated tests.
func TestWeaveSingleProject_IdempotentOnConsistentFixture(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")
	explorer := loadExplorerFromSandbox(t, sb)

	res := sb.Run(explorer.weaveSingleProject)
	if res.Err != nil {
		t.Fatalf("weaveSingleProject: %v", res.Err)
	}
	if !strings.Contains(res.Output, "Updated 0 files") {
		t.Errorf("expected a no-op weave over the already-consistent fixture, got: %q", res.Output)
	}
}

// TestWeaveSingleProject_PropagatesNewParentIntoDependents adds a new atom
// whose parents: names api_zzfix_beta, then weaves: api_zzfix_beta's file
// must gain the new atom in its dependents block (alongside the existing
// mech_zzfix_gamma), and exactly one file must be reported updated.
func TestWeaveSingleProject_PropagatesNewParentIntoDependents(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	newAtom := "---\nid: mech_zzfix_delta\nhuman_name: \"zzfix Delta Mechanic\"\ntype: MECHANIC\nstatus: DRAFT\nlayer: IMPLEMENTATION\nparents:\n  - [[api_zzfix_beta]]\ndependents: []\n---\n\n# zzfix Delta Mechanic\n"
	if err := os.WriteFile(filepath.Join(sb.DocsDir, "mech_zzfix_delta.atom.md"), []byte(newAtom), 0644); err != nil {
		t.Fatal(err)
	}

	explorer := loadExplorerFromSandbox(t, sb)
	res := sb.Run(explorer.weaveSingleProject)
	if res.Err != nil {
		t.Fatalf("weaveSingleProject: %v", res.Err)
	}
	if !strings.Contains(res.Output, "Updated 1 files") {
		t.Errorf("expected exactly api_zzfix_beta's file to be rewritten, got: %q", res.Output)
	}

	beta, err := os.ReadFile(filepath.Join(sb.DocsDir, "api_zzfix_beta.atom.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"  - [[mech_zzfix_delta]]", "  - [[mech_zzfix_gamma]]"} {
		if !strings.Contains(string(beta), want) {
			t.Errorf("api_zzfix_beta's dependents missing %q after weave:\n%s", want, beta)
		}
	}

	// In-memory graph must reflect the same propagation.
	if deps := explorer.Graph.Atoms["api_zzfix_beta"].Dependents; len(deps) != 2 {
		t.Errorf("expected 2 in-memory dependents on api_zzfix_beta after weave, got %v", deps)
	}
}

// TestWeaveWorkspace_CanonicalizesCrossProjectRefs drives the
// workspace-aware path with a pre-attached temp workspace (never cwd
// discovery -- see the file doc comment): a child atom in project zzweave_b
// declares a BARE parent ref to an atom living in project zzweave_a. The
// weave must (a) rewrite the child's parent ref to the canonical
// [[zzweave_a:...]] form and (b) give the parent's file a cross-project
// dependents entry naming the child.
func TestWeaveWorkspace_CanonicalizesCrossProjectRefs(t *testing.T) {
	testutil.SnapshotConfigLocked(t)
	root := t.TempDir()

	mkAtom := func(rel, id, parentsBlock string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		content := fmt.Sprintf("---\nid: %s\nhuman_name: %q\n%sdependents: []\n---\n\n# %s\n", id, id, parentsBlock, id)
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mkAtom("zzweave_a/docs/parent_zzweave.atom.md", "parent_zzweave", "parents: []\n")
	mkAtom("zzweave_b/docs/child_zzweave.atom.md", "child_zzweave", "parents:\n  - [[parent_zzweave]]\n")

	wsJSON := `{
		"workspace_name": "zzweave-ws",
		"projects": [
			{"name": "zzweave_a", "path": "./zzweave_a"},
			{"name": "zzweave_b", "path": "./zzweave_b"}
		]
	}`
	if err := os.WriteFile(filepath.Join(root, ".atd.workspace"), []byte(wsJSON), 0644); err != nil {
		t.Fatal(err)
	}

	ws, err := workspace.LoadWorkspace(root)
	if err != nil {
		t.Fatalf("LoadWorkspace: %v", err)
	}
	idx, err := ws.BuildIndex()
	if err != nil {
		t.Fatalf("BuildIndex: %v", err)
	}

	e := &Explorer{ProjectRoot: root, Workspace: ws, Index: idx}
	out, err := e.Weave() // Workspace+Index pre-attached: routes straight to weaveWorkspace
	if err != nil {
		t.Fatalf("Weave: %v", err)
	}
	if !strings.Contains(out, "Workspace Link Weaving Complete") {
		t.Errorf("expected the workspace-mode completion message, got: %q", out)
	}

	child, err := os.ReadFile(filepath.Join(root, "zzweave_b/docs/child_zzweave.atom.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(child), "  - [[zzweave_a:parent_zzweave]]") {
		t.Errorf("child's bare cross-project parent ref was not canonicalized:\n%s", child)
	}

	parent, err := os.ReadFile(filepath.Join(root, "zzweave_a/docs/parent_zzweave.atom.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(parent), "  - [[zzweave_b:child_zzweave]]") {
		t.Errorf("parent's dependents did not gain the cross-project child entry:\n%s", parent)
	}
}
