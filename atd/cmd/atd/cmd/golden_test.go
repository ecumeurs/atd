package cmd

// Golden-file tests for report-shaped output (test_atd_07_26.md §3.2, §6
// WP-2 deliverable #5). Run with `go test ./... -update` to (re)generate
// testdata/golden/*; two consecutive -update runs must produce byte-
// identical files (WP-2 AC) -- which is exactly what the row/key
// normalization below exists to guarantee against the real, observed
// nondeterminism in the underlying report builders (see comments inline).

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"atd-tools/pkg/testutil"
)

// TestGolden_CheckFull goldens `check --full` against fixture_project.
//
// check --full's row order comes directly from ranging over
// explorer.Graph.Atoms (a Go map) in coverage.GenerateReport's `full`
// branch -- unsorted, so raw report text is NOT stable run to run even
// though every row's content is. Sorting rows by AtomID before golding
// is required for the "-update twice produces no diff" AC; it also makes
// the golden diffable/meaningful (a real regression shows as a content
// change, not a reshuffle).
func TestGolden_CheckFull(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	res := sb.Run(func() (string, error) {
		return runCoverageCheck("full", "", "", sb.DocsDir, true, false, nil)
	})
	if res.Err != nil {
		t.Fatalf("check --full: %v", res.Err)
	}

	rows := parseCheckReport(t, res.Output)
	sort.Slice(rows, func(i, j int) bool { return rows[i].AtomID < rows[j].AtomID })

	var out strings.Builder
	out.WriteString("# check --full (sorted by atom id for golden stability)\n")
	for _, r := range rows {
		fmt.Fprintf(&out, "%-24s impl=%d tests=%d status=%s\n", r.AtomID, r.Impl, r.Tests, r.Status)
	}

	sb.Golden(t, "check_full.txt", out.String())
}

// TestGolden_Lint goldens `lint` against fixture_project, which is
// deliberately lint-clean (S7's rename-propagation scenario depends on
// that baseline). runLint returns ("", nil) on success -- the CLI's
// "All atoms passed structural validation." message is printed by the
// cobra command itself, not returned by runLint -- so the golden pins
// that same success message for readability.
func TestGolden_Lint(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	res := sb.Run(func() (string, error) { return runLint(sb.DocsDir) })

	out := res.Output
	if res.Err == nil {
		out = "OK: All atoms passed structural validation.\n"
	}
	sb.Golden(t, "lint.txt", out)
}

// TestGolden_Stats goldens `stats` against fixture_project.
//
// Regression test for test_atd_07_26.md §8.3 defect #7: runStats
// (cmd/atd/cmd/stats.go) used to build its explorer via
// exploration.NewExplorer(".", docsDir) -- the literal string ".", ignoring
// both the srcPath parameter and config.ProjectRoot(). Since Explorer.Load
// walks ProjectRoot for every file (atoms and links alike, see
// pkg/exploration/explorer.go), calling runStats from a test process used
// to scan the test binary's own working directory instead of any sandbox
// passed via parameters -- a read-only cousin of the I-1 cwd-anchoring
// hazard (test_atd_07_26.md §2.1), just without the destructive write. A
// prior version of this test worked around the bug with an explicit
// os.Chdir(sb.Root)/restore dance (mirroring
// cmd/atd/cmd/workspace_integration_test.go). Now that runStats passes
// srcPath straight through to NewExplorer -- which falls back to
// config.ProjectRoot() when srcPath is "", exactly what sb.Run's
// config.LoadFromDir(sb.Root) sets up -- the chdir workaround is gone and
// this test is back to the plain Sandbox/Run/Golden shape the other golden
// tests use. KNOWN DEFECT expectation changed: if this test starts failing
// because runStats silently resolves against the process cwd again,
// runStats has regressed to hardcoding "." (or an equivalent) and must be
// fixed instead of reintroducing the chdir workaround.
func TestGolden_Stats(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	res := sb.Run(func() (string, error) { return runStats("", sb.DocsDir, false) })
	if res.Err != nil {
		t.Fatalf("stats: %v", res.Err)
	}
	sb.Golden(t, "stats.json", res.Output)
}
