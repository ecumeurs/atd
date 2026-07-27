package exploration

import (
	"atd-tools/pkg/atom"
	"fmt"
	"sort"
	"strings"
)

func (e *Explorer) Trace(targetID string) (*TraceSnapshot, error) {
	canonicalID, err := e.CanonicalAtomID(targetID)
	if err != nil {
		if suggestion := e.SuggestAtomID(targetID); suggestion != "" {
			return nil, fmt.Errorf("atom '%s' not found (did you mean '%s'?)", targetID, suggestion)
		}
		return nil, fmt.Errorf("atom '%s' not found", targetID)
	}
	targetID = canonicalID

	target, err := e.ResolveAtom(targetID)
	if err != nil {
		return nil, fmt.Errorf("atom not found: %s", targetID)
	}

	snap := &TraceSnapshot{
		TargetID: targetID,
		Layer:    target.Layer,
		GraphSlice: TraceGraphSlice{
			Parents:    []string{},
			Dependents: []string{},
			CodeLinks:  []string{},
			TestLinks:  []string{},
		},
		Context:  make(map[string]AtomBrief),
		Warnings: []string{},
	}

	addContext := func(id string, node *atom.AtomData) {
		if _, ok := snap.Context[id]; !ok {
			snap.Context[id] = AtomBrief{
				ID:        node.ID,
				HumanName: node.HumanName,
				Type:      node.Type,
				Layer:     node.Layer,
				Intent:    node.Intent,
				Logic:     node.Logic,
				FilePath:  node.FilePath,
			}
		}
	}

	addContext(targetID, target)

	ancestryComplete := true
	visitedUp := make(map[string]bool)
	e.WalkUp(targetID, visitedUp, func(id string) {
		if id != targetID {
			snap.GraphSlice.Parents = append(snap.GraphSlice.Parents, id)
			node, err := e.ResolveAtom(id)
			if err == nil {
				addContext(id, node)
				if node.Status != "STABLE" {
					ancestryComplete = false
				}
			}
		}
	})

	visitedDown := make(map[string]bool)
	e.WalkDown(targetID, visitedDown, func(id string) {
		if id != targetID {
			snap.GraphSlice.Dependents = append(snap.GraphSlice.Dependents, id)
			if node, err := e.ResolveAtom(id); err == nil {
				addContext(id, node)
			}
		}
	})

	codeFiles := make(map[string]bool)
	testFiles := make(map[string]bool)
	uniqueCodeForTargetOrDescendants := make(map[string]bool)

	processNode := func(id string) {
		for _, sl := range e.SpecLinksForAtom(id) {
			codeFiles[sl.FilePath] = true
			uniqueCodeForTargetOrDescendants[sl.FilePath] = true
		}
		for _, tl := range e.TestLinksForAtom(id) {
			testFiles[tl.TestFile] = true
			uniqueCodeForTargetOrDescendants[tl.TestFile] = true
		}
	}

	processNode(targetID)
	for _, id := range snap.GraphSlice.Dependents {
		processNode(id)
	}

	foundArchAnc := false
	foundCustAnc := false
	for _, p := range snap.GraphSlice.Parents {
		if node, ok := e.Graph.Atoms[p]; ok {
			layer := node.Layer
			if layer == "ARCHITECTURE" {
				foundArchAnc = true
			}
			if layer == "BUSINESS" {
				foundCustAnc = true
			}
		}
	}

	foundArchDesc := false
	foundImplDesc := false
	for _, d := range snap.GraphSlice.Dependents {
		if node, ok := e.Graph.Atoms[d]; ok {
			layer := node.Layer
			if layer == "ARCHITECTURE" {
				foundArchDesc = true
			}
			if layer == "IMPLEMENTATION" {
				foundImplDesc = true
			}
		}
	}

	switch target.Layer {
	case "BUSINESS":
		if !foundArchDesc { snap.Warnings = append(snap.Warnings, "Business atom has no Architecture dependents") }
		if !foundImplDesc { snap.Warnings = append(snap.Warnings, "Business atom has no Implementation dependents") }
	case "ARCHITECTURE":
		if !foundImplDesc && len(e.SpecLinksForAtom(targetID)) == 0 {
			snap.Warnings = append(snap.Warnings, "Architecture atom has no Implementation dependents or direct @spec-link")
		}
		if !foundCustAnc { snap.Warnings = append(snap.Warnings, "Architecture atom has no Business origin") }
	case "IMPLEMENTATION":
		if !foundArchAnc { snap.Warnings = append(snap.Warnings, "Implementation atom has no Architecture origin") }
		if len(uniqueCodeForTargetOrDescendants) == 0 { snap.Warnings = append(snap.Warnings, "Implementation atom has no linked code") }
	}

	for f := range codeFiles {
		snap.GraphSlice.CodeLinks = append(snap.GraphSlice.CodeLinks, f)
	}
	sort.Strings(snap.GraphSlice.CodeLinks)
	for f := range testFiles {
		snap.GraphSlice.TestLinks = append(snap.GraphSlice.TestLinks, f)
	}
	sort.Strings(snap.GraphSlice.TestLinks)

	// totalPool/implementedCount only ever grow via the target-itself checks
	// below and the per-dependent loop further down -- it must NOT also be
	// seeded from len(Dependents), since that double-counts every dependent
	// that qualifies (IMPLEMENTATION, or ARCHITECTURE with a direct
	// @spec-link) as both a pool slot from the seed AND a pool slot from the
	// loop, halving ImplementationRate for a fully-covered chain.
	totalPool := 0
	implementedCount := 0
	testedCount := 0

	checkHealth := func(id string) (bool, bool) {
		impl := len(e.SpecLinksForAtom(id)) > 0
		test := len(e.TestLinksForAtom(id)) > 0
		return impl, test
	}

	if target.Layer == "IMPLEMENTATION" {
		totalPool++
		impl, test := checkHealth(targetID)
		if impl {
			implementedCount++
			if test { testedCount++ }
		}
	}

	if target.Layer == "ARCHITECTURE" && len(e.SpecLinksForAtom(targetID)) > 0 {
		totalPool++
		implementedCount++
		_, test := checkHealth(targetID)
		if test { testedCount++ }
	}

	for _, id := range snap.GraphSlice.Dependents {
		if node, ok := e.Graph.Atoms[id]; ok {
			if node.Layer == "IMPLEMENTATION" {
				totalPool++
				impl, test := checkHealth(id)
				if impl {
					implementedCount++
					if test { testedCount++ }
				}
			}
			if node.Layer == "ARCHITECTURE" && len(e.SpecLinksForAtom(id)) > 0 {
				totalPool++
				impl, test := checkHealth(id)
				if impl {
					implementedCount++
					if test { testedCount++ }
				}
			}
		}
	}

	snap.HealthSummary.AncestryComplete = ancestryComplete
	snap.HealthSummary.HasBusinessOrigin = foundCustAnc || target.Layer == "BUSINESS"
	if totalPool > 0 {
		snap.HealthSummary.ImplementationRate = float64(implementedCount) / float64(totalPool)
	}
	if implementedCount > 0 {
		snap.HealthSummary.TestCoverageRate = float64(testedCount) / float64(implementedCount)
	}

	snap.Metrics.TotalDependents = len(snap.GraphSlice.Dependents)
	snap.Metrics.ImplementedDependents = implementedCount
	snap.Metrics.TotalCodeFiles = len(codeFiles)
	snap.Metrics.TotalTests = len(testFiles)

	return snap, nil
}

// configuredMaxDepth returns the max traversal depth (in hops from the walk's
// starting atom) that WalkUp/WalkDown should honor, or 0 for "unlimited" --
// e.Config is nil for hand-built Explorers in tests, and a non-positive
// MaxDepth is treated the same as unconfigured rather than as "walk nothing".
func (e *Explorer) configuredMaxDepth() int {
	if e.Config == nil || e.Config.MaxDepth <= 0 {
		return 0
	}
	return e.Config.MaxDepth
}

func (e *Explorer) WalkUp(id string, visited map[string]bool, onVisit func(string)) {
	e.walkUp(id, visited, onVisit, 0)
}

func (e *Explorer) walkUp(id string, visited map[string]bool, onVisit func(string), depth int) {
	if visited[id] {
		return
	}
	if maxDepth := e.configuredMaxDepth(); maxDepth > 0 && depth > maxDepth {
		return
	}
	visited[id] = true
	node, err := e.ResolveAtom(id)
	if err != nil {
		return
	}
	onVisit(id)
	for _, p := range node.Parents {
		e.walkUp(strings.TrimSpace(p), visited, onVisit, depth+1)
	}
}

func (e *Explorer) WalkDown(id string, visited map[string]bool, onVisit func(string)) {
	e.walkDown(id, visited, onVisit, 0)
}

func (e *Explorer) walkDown(id string, visited map[string]bool, onVisit func(string), depth int) {
	if visited[id] {
		return
	}
	if maxDepth := e.configuredMaxDepth(); maxDepth > 0 && depth > maxDepth {
		return
	}
	visited[id] = true
	node, err := e.ResolveAtom(id)
	if err != nil {
		return
	}
	onVisit(id)
	for _, d := range node.Dependents {
		e.walkDown(strings.TrimSpace(d), visited, onVisit, depth+1)
	}
}