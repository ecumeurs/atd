package exploration

// Shared helpers for this package's WP-7 unit back-fill (test_atd_07_26.md
// §3.1 item 1, §6 WP-7). Split out so links_test.go / resolve_test.go /
// crawl_test.go / orphan_test.go / trace_test.go can all reach for the same
// small set of fixtures without duplicating setup.

import (
	"testing"

	"atd-tools/pkg/atom"
	"atd-tools/pkg/testutil"
)

// newLinkTestExplorer builds a standalone (no workspace) Explorer whose
// graph contains exactly one bare AtomData per given id, mirroring what
// Load() would have populated after parsing docs/*.atom.md. Resolver stays
// nil, which is the standalone-project code path extractLinks/ResolveAtom
// take when no .atd.workspace is found above ProjectRoot (see
// pkg/exploration/explorer.go's NewExplorerWithConfig).
func newLinkTestExplorer(ids ...string) *Explorer {
	atoms := make(map[string]*atom.AtomData, len(ids))
	for _, id := range ids {
		atoms[id] = &atom.AtomData{ID: id}
	}
	return &Explorer{Graph: &DependencyGraph{Atoms: atoms}}
}

// loadExplorerFromSandbox loads config.ActiveConfig from sb's root (under
// testutil's mutex-protected Run, so this stays safe alongside other
// t.Parallel() scenario tests) and returns a fully Load()-ed Explorer over
// it -- the same construction path `atd check`/`atd trace` use in
// production (NewExplorer + Load(true)), just invoked directly against the
// sandbox instead of through a cmd/ entry point.
func loadExplorerFromSandbox(t *testing.T, sb *testutil.SB) *Explorer {
	t.Helper()
	var explorer *Explorer
	res := sb.Run(func() (string, error) {
		explorer = NewExplorer(sb.Root, sb.DocsDir)
		return "", explorer.Load(true)
	})
	if res.Err != nil {
		t.Fatalf("loadExplorerFromSandbox: %v", res.Err)
	}
	return explorer
}
