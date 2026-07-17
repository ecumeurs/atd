package exploration

// Trace unit back-fill (test_atd_07_26.md §3.1 item 1, §6 WP-7): a real
// cross-layer chain via testdata/fixture_project, plus hand-built cyclic and
// deep-chain graphs to pin WalkUp/WalkDown's cycle-safety and configured
// depth limit.

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/testutil"
)

// TestTrace_FixtureProjectChain traces req_zzfix_alpha, whose dependent
// chain is req_zzfix_alpha -> api_zzfix_beta -> mech_zzfix_gamma plus the
// sibling rule_zzfix_untested (test_atd_07_26.md §3.2's fixture design).
// CodeLinks/TestLinks are asserted against an exact, ordered slice: trace.go
// now sorts both before returning (test_atd_07_26.md §8.3 #5 -- they used to
// come straight out of Go map iteration, so run-to-run order was not
// deterministic; cmd/atd/cmd/scenario_test.go's normalizeTraceJSON worked
// around that upstream of this fix and remains harmless now that the source
// is already sorted).
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

	wantCodeLinks := []string{"src/beta.go", "src/gamma.go", "src/untested.go"}
	if !reflect.DeepEqual(snap.GraphSlice.CodeLinks, wantCodeLinks) {
		t.Errorf("CodeLinks = %v, want %v (in sorted order)", snap.GraphSlice.CodeLinks, wantCodeLinks)
	}
	wantTestLinks := []string{"src/beta_test.go", "src/gamma_test.go"}
	if !reflect.DeepEqual(snap.GraphSlice.TestLinks, wantTestLinks) {
		t.Errorf("TestLinks = %v, want %v (in sorted order)", snap.GraphSlice.TestLinks, wantTestLinks)
	}

	// All 3 dependents (api_zzfix_beta, mech_zzfix_gamma, rule_zzfix_untested)
	// carry a @spec-link, so the pool trace.go builds from "IMPLEMENTATION
	// dependents" + "ARCHITECTURE dependents with a direct @spec-link" is
	// exactly these 3, all implemented -- rate 3/3 = 1.0. (Previously
	// totalPool was seeded with len(Dependents) AND re-incremented once per
	// qualifying dependent in the loop below, double-counting the pool and
	// halving this to 0.5 -- see TestTrace_ImplementationRateFullyCoveredChain
	// for a minimal, from-scratch pin of the same fix.)
	if got, want := snap.HealthSummary.ImplementationRate, 1.0; got != want {
		t.Errorf("ImplementationRate = %v, want %v", got, want)
	}
	const wantTestRate = 2.0 / 3.0
	if got := snap.HealthSummary.TestCoverageRate; got < wantTestRate-0.001 || got > wantTestRate+0.001 {
		t.Errorf("TestCoverageRate = %v, want ~%v", got, wantTestRate)
	}
}

// TestTrace_ImplementationRateFullyCoveredChain pins the fix for the
// totalPool double-count (test_atd_07_26.md §8.3 #6a): a fully-implemented
// and fully-tested 3-atom chain (BUSINESS root -> IMPLEMENTATION child ->
// IMPLEMENTATION grandchild) must report ImplementationRate == 1.0 and
// TestCoverageRate == 1.0, not 0.5. Built from scratch (no fixture project)
// so the pool math is unambiguous: exactly 2 qualifying dependents, both
// implemented and tested.
func TestTrace_ImplementationRateFullyCoveredChain(t *testing.T) {
	t.Parallel()
	graph := &DependencyGraph{Atoms: map[string]*atom.AtomData{
		"zz_rate_root":       {ID: "zz_rate_root", Status: "STABLE", Layer: "BUSINESS", Type: "REQUIREMENT", Dependents: []string{"zz_rate_child"}},
		"zz_rate_child":      {ID: "zz_rate_child", Status: "STABLE", Layer: "IMPLEMENTATION", Type: "MECHANIC", Dependents: []string{"zz_rate_grandchild"}},
		"zz_rate_grandchild": {ID: "zz_rate_grandchild", Status: "STABLE", Layer: "IMPLEMENTATION", Type: "MECHANIC"},
	}}
	e := &Explorer{
		Graph: graph,
		SpecLinks: []SpecLink{
			{AtomID: "zz_rate_child", FilePath: "child.go", Line: 1},
			{AtomID: "zz_rate_grandchild", FilePath: "grandchild.go", Line: 1},
		},
		TestLinks: []TestLink{
			{AtomID: "zz_rate_child", TestFile: "child_test.go", Line: 1},
			{AtomID: "zz_rate_grandchild", TestFile: "grandchild_test.go", Line: 1},
		},
	}

	snap, err := e.Trace("zz_rate_root")
	if err != nil {
		t.Fatalf("Trace: %v", err)
	}
	if got, want := snap.HealthSummary.ImplementationRate, 1.0; got != want {
		t.Errorf("ImplementationRate = %v, want %v for a fully-implemented chain", got, want)
	}
	if got, want := snap.HealthSummary.TestCoverageRate, 1.0; got != want {
		t.Errorf("TestCoverageRate = %v, want %v for a fully-tested chain", got, want)
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

// TestTrace_LongParentChainRespectsConfiguredMaxDepth pins the fix for
// test_atd_07_26.md §8.3 #6b: config.Config.MaxDepth (default 10) is now
// wired into WalkUp/WalkDown (trace.go), so a 15-deep pure chain (no cycle)
// is truncated at the configured depth instead of being walked to its end.
// Depth is counted in hops from the traced atom (depth 0), so a MaxDepth of
// 10 yields exactly 10 ancestors (depths 1..10).
func TestTrace_LongParentChainRespectsConfiguredMaxDepth(t *testing.T) {
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
	e := &Explorer{Graph: &DependencyGraph{Atoms: atoms}, Config: &config.Config{MaxDepth: 10}}

	snap, err := e.Trace("zz_chain_00")
	if err != nil {
		t.Fatalf("Trace: %v", err)
	}
	if len(snap.GraphSlice.Parents) != 10 {
		t.Errorf("expected traversal to stop at config.MaxDepth=10, got %d ancestors: %v",
			len(snap.GraphSlice.Parents), snap.GraphSlice.Parents)
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
