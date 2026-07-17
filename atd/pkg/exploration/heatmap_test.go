package exploration

// heatmap.go unit tests. Everything here except CalculateUpdateHeat is pure
// (no LLM, no filesystem beyond reading a.Implementations paths already in
// memory); CalculateUpdateHeat shells out to `git log`, exercised against a
// real throwaway git repo. This was 0%-covered headroom used to help this
// package clear the WP-7 coverage floor (test_atd_07_26.md §6 WP-7:
// pkg/exploration >= 60%) alongside the extractLinks/ResolveAtom/crawl/
// orphan/trace back-fill this WP was primarily scoped to.

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"atd-tools/pkg/atom"
)

func TestCalculateDependencyHeat(t *testing.T) {
	e := &Explorer{}
	cases := []struct {
		name string
		node *atom.AtomData
		want HeatState
	}{
		{"zero_parents_zero_dependents_is_cold", &atom.AtomData{Layer: "ARCHITECTURE"}, HeatCold},
		{"one_dependent_is_optimal", &atom.AtomData{Layer: "ARCHITECTURE", Dependents: []string{"a"}}, HeatOptimal},
		{"four_parents_is_warm", &atom.AtomData{Layer: "ARCHITECTURE", Parents: []string{"a", "b", "c", "d"}}, HeatWarm},
		{"ten_dependents_is_hot", &atom.AtomData{Layer: "ARCHITECTURE", Dependents: make([]string, 10)}, HeatHot},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := e.CalculateDependencyHeat(c.node); got != c.want {
				t.Errorf("CalculateDependencyHeat(%+v) = %v, want %v", c.node, got, c.want)
			}
		})
	}
}

func TestGetCodeFileHeatAndCalculateCodeHeat(t *testing.T) {
	e := &Explorer{SpecLinks: []SpecLink{
		{AtomID: "zz_a", FilePath: "hot.go"},
		{AtomID: "zz_a", FilePath: "hot.go"},
	}}
	for i := 0; i < 8; i++ {
		e.SpecLinks = append(e.SpecLinks, SpecLink{AtomID: "zz_a", FilePath: "hot.go"})
	}
	e.SpecLinks = append(e.SpecLinks, SpecLink{AtomID: "zz_b", FilePath: "cold.go"})

	if got := e.GetCodeFileHeat("hot.go"); got != HeatHot {
		t.Errorf("GetCodeFileHeat(hot.go) [10 links] = %v, want %v", got, HeatHot)
	}
	if got := e.GetCodeFileHeat("cold.go"); got != HeatCold {
		t.Errorf("GetCodeFileHeat(cold.go) [1 link] = %v, want %v", got, HeatCold)
	}
	if got := e.GetCodeFileHeat("never-seen.go"); got != HeatCold {
		t.Errorf("GetCodeFileHeat(unknown file) = %v, want %v", got, HeatCold)
	}

	noImpl := &atom.AtomData{}
	if got := e.CalculateCodeHeat(noImpl); got != HeatCold {
		t.Errorf("CalculateCodeHeat(no implementations) = %v, want %v", got, HeatCold)
	}

	// Takes the MAX heat across all of the atom's implementation files.
	mixed := &atom.AtomData{Implementations: []string{"cold.go:1", "hot.go:5"}}
	if got := e.CalculateCodeHeat(mixed); got != HeatHot {
		t.Errorf("CalculateCodeHeat(mixed cold+hot files) = %v, want %v (max heat)", got, HeatHot)
	}
}

// gitCommitsTouching initializes a git repo at dir and creates n commits,
// each appending a line to relFile so `git log` sees n revisions of it.
func gitCommitsTouching(t *testing.T, dir, relFile string, n int) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=atd-testutil", "GIT_AUTHOR_EMAIL=atd-testutil@example.com",
			"GIT_COMMITTER_NAME=atd-testutil", "GIT_COMMITTER_EMAIL=atd-testutil@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	full := filepath.Join(dir, relFile)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		content, _ := os.ReadFile(full)
		content = append(content, []byte("line\n")...)
		if err := os.WriteFile(full, content, 0644); err != nil {
			t.Fatal(err)
		}
		run("add", relFile)
		run("commit", "-q", "-m", "commit")
	}
}

func TestCalculateUpdateHeat(t *testing.T) {
	t.Run("not_a_git_repo_is_stable", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "f.go"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		e := &Explorer{ProjectRoot: root}
		state, count, last := e.CalculateUpdateHeat(&atom.AtomData{FilePath: filepath.Join(root, "f.go")})
		if state != HeatStable || count != 0 || last != "" {
			t.Errorf("got (%v, %d, %q), want (HeatStable, 0, \"\")", state, count, last)
		}
	})

	t.Run("one_commit_is_optimal", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		gitCommitsTouching(t, root, "f.go", 1)
		e := &Explorer{ProjectRoot: root}
		state, count, last := e.CalculateUpdateHeat(&atom.AtomData{FilePath: filepath.Join(root, "f.go")})
		if state != HeatOptimal || count != 1 || last == "" {
			t.Errorf("got (%v, %d, %q), want (HeatOptimal, 1, non-empty)", state, count, last)
		}
	})

	t.Run("four_commits_is_warm", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		gitCommitsTouching(t, root, "f.go", 4)
		e := &Explorer{ProjectRoot: root}
		state, count, _ := e.CalculateUpdateHeat(&atom.AtomData{FilePath: filepath.Join(root, "f.go")})
		if state != HeatWarm || count != 4 {
			t.Errorf("got (%v, %d), want (HeatWarm, 4)", state, count)
		}
	})

	t.Run("seven_commits_is_hot", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		gitCommitsTouching(t, root, "f.go", 7)
		e := &Explorer{ProjectRoot: root}
		state, count, _ := e.CalculateUpdateHeat(&atom.AtomData{FilePath: filepath.Join(root, "f.go")})
		if state != HeatHot || count != 7 {
			t.Errorf("got (%v, %d), want (HeatHot, 7)", state, count)
		}
	})
}

func TestGetHeatMapResult(t *testing.T) {
	t.Run("unknown_atom_errors", func(t *testing.T) {
		e := &Explorer{Graph: &DependencyGraph{Atoms: map[string]*atom.AtomData{}}}
		if _, err := e.GetHeatMapResult("zz_does_not_exist"); err == nil {
			t.Error("expected an error for an atom absent from the graph")
		}
	})

	t.Run("hot_dependency_recommends_splitting", func(t *testing.T) {
		root := t.TempDir()
		a := &atom.AtomData{ID: "zz_hot_dep", Layer: "ARCHITECTURE", Dependents: make([]string, 10), FilePath: filepath.Join(root, "a.atom.md")}
		e := &Explorer{Graph: &DependencyGraph{Atoms: map[string]*atom.AtomData{"zz_hot_dep": a}}, ProjectRoot: root}

		res, err := e.GetHeatMapResult("zz_hot_dep")
		if err != nil {
			t.Fatalf("GetHeatMapResult: %v", err)
		}
		if res.DependencyState != HeatHot {
			t.Errorf("DependencyState = %v, want %v", res.DependencyState, HeatHot)
		}
		found := false
		for _, r := range res.Recommendations {
			if r != "" && (r == "Excessive coupling detected. Consider splitting this atom.") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected a splitting recommendation, got %v", res.Recommendations)
		}
	})

	t.Run("heatmap_override_none_forces_optimal_and_stable", func(t *testing.T) {
		root := t.TempDir()
		a := &atom.AtomData{
			ID: "zz_override", Layer: "ARCHITECTURE", HeatMap: "none",
			Dependents: make([]string, 10), FilePath: filepath.Join(root, "a.atom.md"),
		}
		e := &Explorer{Graph: &DependencyGraph{Atoms: map[string]*atom.AtomData{"zz_override": a}}, ProjectRoot: root}

		res, err := e.GetHeatMapResult("zz_override")
		if err != nil {
			t.Fatalf("GetHeatMapResult: %v", err)
		}
		// @spec-link [[rule_atd_atom_overrides]]
		if res.DependencyState != HeatOptimal {
			t.Errorf("expected heatmap override 'none' to force DependencyState=Optimal despite 10 dependents, got %v", res.DependencyState)
		}
		if res.CodeState != HeatOptimal {
			t.Errorf("expected heatmap override 'none' to force CodeState=Optimal, got %v", res.CodeState)
		}
		if res.UpdateState != HeatStable {
			t.Errorf("expected heatmap override 'none' to force UpdateState=Stable, got %v", res.UpdateState)
		}
	})
}
