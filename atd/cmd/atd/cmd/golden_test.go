package cmd

// Golden-file tests for report-shaped output (test_atd_07_26.md §3.2, §6
// WP-2 deliverable #5). Run with `go test ./... -update` to (re)generate
// testdata/golden/*; two consecutive -update runs must produce byte-
// identical files (WP-2 AC) -- which is exactly what the row/key
// normalization below exists to guarantee against the real, observed
// nondeterminism in the underlying report builders (see comments inline).

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"atd-tools/pkg/testutil"
)

// TestGolden_CheckFull goldens `check --full` against fixture_project.
//
// coverage.GenerateReport's `full` branch now sorts atom ids itself
// (test_atd_07_26.md §8.3 #5 -- row order used to come directly from
// ranging over explorer.Graph.Atoms, a Go map, so raw report text was NOT
// stable run to run even though every row's content was). The sort here is
// therefore redundant with the source but kept as a defensive belt-and-
// braces for the "-update twice produces no diff" AC, and it keeps the
// golden diffable/meaningful (a real regression shows as a content change,
// not a reshuffle).
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
// KNOWN DEFECT: runStats (cmd/atd/cmd/stats.go) builds its explorer via
// exploration.NewExplorer(".", docsDir) -- the root is the literal string
// ".", ignoring both the srcPath parameter and config.ProjectRoot(). Since
// Explorer.Load walks ProjectRoot for *every* file (atoms and links alike,
// see pkg/exploration/explorer.go), calling runStats from a test process
// scans the test binary's own working directory, not any sandbox passed
// via parameters -- a read-only cousin of the I-1 cwd-anchoring hazard
// (test_atd_07_26.md §2.1), just without the destructive write. The only
// way to exercise it against a sandbox is to chdir into it first, exactly
// as cmd/atd/cmd/workspace_integration_test.go already does for a
// different command -- so, like that test, this one is NOT t.Parallel().
func TestGolden_Stats(t *testing.T) {
	sb := testutil.Sandbox(t, "fixture_project")

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(sb.Root); err != nil {
		t.Fatal(err)
	}
	res := sb.Run(func() (string, error) { return runStats("", sb.DocsDir, false) })
	// Restore cwd immediately -- (*SB).Golden below walks upward from the
	// current working directory to find testdata/golden, which must be the
	// test package's own directory, not the sandbox we chdir'd into above.
	if err := os.Chdir(oldWD); err != nil {
		t.Fatal(err)
	}

	if res.Err != nil {
		t.Fatalf("stats: %v", res.Err)
	}
	sb.Golden(t, "stats.json", res.Output)
}
