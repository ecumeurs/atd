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

// TestWeaveSingleProject_StripsGovernanceLinks pins weave's repair half of the
// governance graph-isolation rule (ATD.md §1.4, the rule `atd lint` reports on):
// a CONTRACT/VISION named as someone's parent is dropped from that parents:
// block, a governance atom's own parents: are emptied, and neither side is left
// with a dependents: entry for the removed edge. Every removal is named in the
// result text — weave rewrites files in place, so nothing is dropped silently.
// @test-link [[rule_atd_governance_graph_isolation]]
func TestWeaveSingleProject_StripsGovernanceLinks(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	mk := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(sb.DocsDir, name), []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// contract_zzgov itself points UP at an ordinary atom (forbidden), and
	// rule_zzgov points at the contract (also forbidden).
	mk("contract_zzgov.atom.md", "---\nid: contract_zzgov\nhuman_name: \"zzgov Contract\"\ntype: CONTRACT\nstatus: STABLE\nlayer: BUSINESS\nparents:\n  - [[api_zzfix_beta]]\ndependents: []\n---\n\n# zzgov Contract\n")
	mk("rule_zzgov.atom.md", "---\nid: rule_zzgov\nhuman_name: \"zzgov Rule\"\ntype: RULE\nstatus: DRAFT\nlayer: BUSINESS\nparents:\n  - [[contract_zzgov]]\n  - [[api_zzfix_beta]]\ndependents: []\n---\n\n# zzgov Rule\n")

	explorer := loadExplorerFromSandbox(t, sb)
	res := sb.Run(explorer.weaveSingleProject)
	if res.Err != nil {
		t.Fatalf("weaveSingleProject: %v", res.Err)
	}

	read := func(name string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(sb.DocsDir, name))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	contract := read("contract_zzgov.atom.md")
	if !strings.Contains(contract, "parents: []") {
		t.Errorf("governance atom's own parents: were not emptied:\n%s", contract)
	}
	if !strings.Contains(contract, "dependents: []") {
		t.Errorf("governance atom gained dependents from the stripped edge:\n%s", contract)
	}

	rule := read("rule_zzgov.atom.md")
	if strings.Contains(rule, "contract_zzgov") {
		t.Errorf("parent ref to the governance atom was not stripped:\n%s", rule)
	}
	// The legitimate parent must survive the surgery untouched.
	if !strings.Contains(rule, "  - [[api_zzfix_beta]]") {
		t.Errorf("non-governance parent was lost during the strip:\n%s", rule)
	}

	// api_zzfix_beta keeps rule_zzgov (a real edge) but must NOT list the
	// contract, whose forbidden parents: entry was dropped.
	beta := read("api_zzfix_beta.atom.md")
	if !strings.Contains(beta, "  - [[rule_zzgov]]") {
		t.Errorf("legitimate dependent missing from api_zzfix_beta:\n%s", beta)
	}
	if strings.Contains(beta, "contract_zzgov") {
		t.Errorf("api_zzfix_beta gained a dependents entry for the governance atom:\n%s", beta)
	}

	for _, want := range []string{
		"Removed 2 forbidden governance link(s)",
		"contract_zzgov (CONTRACT) parents: [[api_zzfix_beta]]",
		"rule_zzgov parents: [[contract_zzgov]]",
	} {
		if !strings.Contains(res.Output, want) {
			t.Errorf("weave output missing %q, got:\n%s", want, res.Output)
		}
	}
}

// TestWeaveSingleProject_LeavesCleanParentsUntouched guards the conservative
// half of the strip: when no forbidden link exists, weave must not touch any
// parents: block at all (it passes nil through to rewriteAtomLinks), so an
// already-consistent corpus does not churn its parent ordering on every run.
// @test-link [[rule_atd_governance_graph_isolation]]
func TestWeaveSingleProject_LeavesCleanParentsUntouched(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	contract := "---\nid: contract_zzclean\nhuman_name: \"zzclean Contract\"\ntype: CONTRACT\nstatus: STABLE\nlayer: BUSINESS\nparents: []\ndependents: []\n---\n\n# zzclean Contract\n"
	if err := os.WriteFile(filepath.Join(sb.DocsDir, "contract_zzclean.atom.md"), []byte(contract), 0644); err != nil {
		t.Fatal(err)
	}

	explorer := loadExplorerFromSandbox(t, sb)
	res := sb.Run(explorer.weaveSingleProject)
	if res.Err != nil {
		t.Fatalf("weaveSingleProject: %v", res.Err)
	}
	if !strings.Contains(res.Output, "Updated 0 files") {
		t.Errorf("a governance-clean corpus must weave as a no-op, got: %q", res.Output)
	}
	if strings.Contains(res.Output, "Removed") {
		t.Errorf("nothing should have been stripped, got: %q", res.Output)
	}
}

// TestWeaveWorkspace_StripsCrossProjectGovernanceParent pins the same rule on
// the workspace path: a governance atom is graph-isolated across project
// boundaries too, so a bare cross-project parent ref naming one is dropped
// rather than canonicalized into [[project:id]] form.
// @test-link [[rule_atd_governance_graph_isolation]]
func TestWeaveWorkspace_StripsCrossProjectGovernanceParent(t *testing.T) {
	testutil.SnapshotConfigLocked(t)
	root := t.TempDir()

	mkAtom := func(rel, id, atomType, parentsBlock string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		content := fmt.Sprintf("---\nid: %s\nhuman_name: %q\ntype: %s\n%sdependents: []\n---\n\n# %s\n", id, id, atomType, parentsBlock, id)
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mkAtom("zzgov_a/docs/vision_zzgov.atom.md", "vision_zzgov", "VISION", "parents: []\n")
	mkAtom("zzgov_a/docs/req_zzgov.atom.md", "req_zzgov", "REQUIREMENT", "parents: []\n")
	mkAtom("zzgov_b/docs/child_zzgov.atom.md", "child_zzgov", "RULE", "parents:\n  - [[vision_zzgov]]\n  - [[req_zzgov]]\n")

	wsJSON := `{
		"workspace_name": "zzgov-ws",
		"projects": [
			{"name": "zzgov_a", "path": "./zzgov_a"},
			{"name": "zzgov_b", "path": "./zzgov_b"}
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
	out, err := e.Weave()
	if err != nil {
		t.Fatalf("Weave: %v", err)
	}
	if !strings.Contains(out, "zzgov_b:child_zzgov parents: [[vision_zzgov]]") {
		t.Errorf("expected the stripped cross-project governance parent to be reported, got:\n%s", out)
	}

	child, err := os.ReadFile(filepath.Join(root, "zzgov_b/docs/child_zzgov.atom.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(child), "vision_zzgov") {
		t.Errorf("cross-project governance parent was canonicalized instead of stripped:\n%s", child)
	}
	if !strings.Contains(string(child), "  - [[zzgov_a:req_zzgov]]") {
		t.Errorf("legitimate cross-project parent was lost or left uncanonicalized:\n%s", child)
	}

	vision, err := os.ReadFile(filepath.Join(root, "zzgov_a/docs/vision_zzgov.atom.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(vision), "dependents: []") {
		t.Errorf("governance atom gained a dependents entry for the stripped edge:\n%s", vision)
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
