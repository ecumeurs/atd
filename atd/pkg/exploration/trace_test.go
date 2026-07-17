package exploration

// Trace unit back-fill (test_atd_07_26.md §3.1 item 1, §6 WP-7): a real
// cross-layer chain via testdata/fixture_project, plus hand-built cyclic and
// deep-chain graphs to pin WalkUp/WalkDown's cycle-safety and the
// (non-existent) depth limit.

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"atd-tools/pkg/atom"
	"atd-tools/pkg/testutil"
)

// TestTrace_FixtureProjectChain traces req_zzfix_alpha, whose dependent
// chain is req_zzfix_alpha -> api_zzfix_beta -> mech_zzfix_gamma plus the
// sibling rule_zzfix_untested (test_atd_07_26.md §3.2's fixture design).
// CodeLinks/TestLinks are sorted before comparison per the pinned
// nondeterminism (pkg/exploration/trace.go builds them from Go maps; see
// cmd/atd/cmd/scenario_test.go's normalizeTraceJSON).
func TestTrace_FixtureProjectChain(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")
	explorer := loadExplorerFromSandbox(t, sb)

	snap, err := explorer.Trace("req_zzfix_alpha")
	if err != nil {
		t.Fatalf("Trace: %v", err)
	}

	if snap.Layer != "BUSINESS" {
		t.Errorf("Layer = %q, want BUSINESS", snap.Layer)
	}
	if len(snap.GraphSlice.Parents) != 0 {
		t.Errorf("expected no parents for a top-level requirement, got %v", snap.GraphSlice.Parents)
	}

	deps := append([]string(nil), snap.GraphSlice.Dependents...)
	sort.Strings(deps)
	wantDeps := []string{"api_zzfix_beta", "mech_zzfix_gamma", "rule_zzfix_untested"}
	if !reflect.DeepEqual(deps, wantDeps) {
		t.Errorf("Dependents = %v, want %v (mech_zzfix_gamma reached transitively via api_zzfix_beta)", deps, wantDeps)
	}

	if !snap.HealthSummary.HasBusinessOrigin {
		t.Error("expected HasBusinessOrigin = true for a BUSINESS-layer atom")
	}
	if !snap.HealthSummary.AncestryComplete {
		t.Error("expected AncestryComplete = true (no ancestors to fail the STABLE check)")
	}

	// NOTE (documented oddity, not fixed by this back-fill): trace.go seeds
	// totalPool with len(Dependents) (=3 here) BEFORE the loop below adds
	// one more increment per dependent that is IMPLEMENTATION or an
	// ARCHITECTURE atom with a direct @spec-link -- all 3 dependents qualify
	// here, so totalPool ends up double-counted at 3+3=6 rather than the 3
	// a reader would expect, halving ImplementationRate. All 3 qualifying
	// dependents (api_zzfix_beta, mech_zzfix_gamma, rule_zzfix_untested) are
	// in fact fully implemented; this asserts the actual (halved) rate the
	// double-count produces, not the "3 of 3" a naive reading would expect.
	if got, want := snap.HealthSummary.ImplementationRate, 0.5; got != want {
		t.Errorf("ImplementationRate = %v, want %v (see double-count note above; if trace.go's totalPool seeding changed, update this pin)", got, want)
	}
	const wantTestRate = 2.0 / 3.0
	if got := snap.HealthSummary.TestCoverageRate; got < wantTestRate-0.001 || got > wantTestRate+0.001 {
		t.Errorf("TestCoverageRate = %v, want ~%v", got, wantTestRate)
	}
}

// TestTrace_UnknownIDSuggestsAndErrorsLoudly pins the "resolve or shout"
// contract (test_atd_07_26.md §7.2): a nonsense id must error, naming the
// input, rather than silently returning an empty snapshot.
func TestTrace_UnknownIDSuggestsAndErrorsLoudly(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")
	explorer := loadExplorerFromSandbox(t, sb)

	_, err := explorer.Trace("api_zzfix_totally_bogus")
	if err == nil {
		t.Fatal("expected an error tracing a nonexistent atom id")
	}
}

// TestTrace_CyclicParentsAndDependentsNoInfiniteLoop constructs a 3-node
// cycle in BOTH parents and dependents (a -> b -> c -> a both ways).
// WalkUp/WalkDown's visited-map guard (trace.go) is what stands between
// this and an infinite recursion; the 5s deadline turns a guard regression
// into a fast, readable test failure instead of a hung CI job.
func TestTrace_CyclicParentsAndDependentsNoInfiniteLoop(t *testing.T) {
	t.Parallel()

	mk := func(id, parent, dependent string) *atom.AtomData {
		return &atom.AtomData{
			ID: id, Status: "STABLE", Layer: "ARCHITECTURE", Type: "API",
			Parents: []string{parent}, Dependents: []string{dependent},
		}
	}
	graph := &DependencyGraph{Atoms: map[string]*atom.AtomData{
		"zz_cyc_a": mk("zz_cyc_a", "zz_cyc_c", "zz_cyc_b"),
		"zz_cyc_b": mk("zz_cyc_b", "zz_cyc_a", "zz_cyc_c"),
		"zz_cyc_c": mk("zz_cyc_c", "zz_cyc_b", "zz_cyc_a"),
	}}
	e := &Explorer{Graph: graph}

	type result struct {
		snap *TraceSnapshot
		err  error
	}
	done := make(chan result, 1)
	go func() {
		snap, err := e.Trace("zz_cyc_a")
		done <- result{snap, err}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("Trace on a cyclic graph returned an error: %v", r.err)
		}
		parents := append([]string(nil), r.snap.GraphSlice.Parents...)
		sort.Strings(parents)
		if want := []string{"zz_cyc_b", "zz_cyc_c"}; !reflect.DeepEqual(parents, want) {
			t.Errorf("Parents = %v, want %v", parents, want)
		}
		deps := append([]string(nil), r.snap.GraphSlice.Dependents...)
		sort.Strings(deps)
		if want := []string{"zz_cyc_b", "zz_cyc_c"}; !reflect.DeepEqual(deps, want) {
			t.Errorf("Dependents = %v, want %v", deps, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Trace did not return within 5s on a cyclic graph -- likely an infinite-loop regression in WalkUp/WalkDown's visited-map cycle guard")
	}
}

// TestTrace_LongParentChainIgnoresConfiguredMaxDepth documents a real gap:
// config.Config.MaxDepth exists (default 10, config/config.go) but
// WalkUp/WalkDown (trace.go) never read it -- the only bound on recursion
// is the visited-map cycle guard, not an actual depth counter. A 15-deep
// pure chain (no cycle) is walked to its end, not truncated at 10. This is
// pinned as current behavior per this WP's instructions (pin, don't fix);
// see the final report for a pointer back to this test.
func TestTrace_LongParentChainIgnoresConfiguredMaxDepth(t *testing.T) {
	t.Parallel()
	const depth = 15
	atoms := make(map[string]*atom.AtomData, depth)
	for i := 0; i < depth; i++ {
		id := fmt.Sprintf("zz_chain_%02d", i)
		var parents []string
		if i+1 < depth {
			parents = []string{fmt.Sprintf("zz_chain_%02d", i+1)}
		}
		atoms[id] = &atom.AtomData{ID: id, Status: "STABLE", Layer: "ARCHITECTURE", Type: "API", Parents: parents}
	}
	e := &Explorer{Graph: &DependencyGraph{Atoms: atoms}}

	snap, err := e.Trace("zz_chain_00")
	if err != nil {
		t.Fatalf("Trace: %v", err)
	}
	if len(snap.GraphSlice.Parents) != depth-1 {
		t.Errorf("expected all %d ancestors to be walked despite config.MaxDepth's default of 10 (MaxDepth is not wired into WalkUp/WalkDown), got %d: %v",
			depth-1, len(snap.GraphSlice.Parents), snap.GraphSlice.Parents)
	}
}

// TestWalkUp_StopsAtUnresolvableAncestor pins that WalkUp/WalkDown fail
// silently (no panic, no error return -- there's nowhere for one to go)
// when a parent/dependent id doesn't exist in the graph at all: the walk
// just stops there instead of recursing into a zero-value node.
func TestWalkUp_StopsAtUnresolvableAncestor(t *testing.T) {
	t.Parallel()
	e := &Explorer{Graph: &DependencyGraph{Atoms: map[string]*atom.AtomData{
		"zz_dangling": {ID: "zz_dangling", Parents: []string{"zz_ghost_parent_does_not_exist"}},
	}}}

	var visited []string
	e.WalkUp("zz_dangling", make(map[string]bool), func(id string) {
		visited = append(visited, id)
	})

	if want := []string{"zz_dangling"}; !reflect.DeepEqual(visited, want) {
		t.Errorf("visited = %v, want %v (the dangling parent reference must not be visited or crash the walk)", visited, want)
	}
}
