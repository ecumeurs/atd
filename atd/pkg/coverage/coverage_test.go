package coverage

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/testutil"
)

func TestCheckRowStatusLogic(t *testing.T) {
	tests := []struct {
		impl     int
		test     int
		expected string
	}{
		{0, 0, "NO_IMPL"},
		{1, 0, "NO_TESTS"},
		{2, 1, "OK"},
		{3, 2, "OK"},
	}

	for _, tt := range tests {
		row := CheckAtomRow{ImplLinks: tt.impl, TestLinks: tt.test}
		switch {
		case row.ImplLinks == 0:
			row.Status = "NO_IMPL"
		case row.TestLinks == 0:
			row.Status = "NO_TESTS"
		default:
			row.Status = "OK"
		}
		if row.Status != tt.expected {
			t.Errorf("impl=%d test=%d: expected status %q, got %q", tt.impl, tt.test, tt.expected, row.Status)
		}
	}
}

func TestFormatCheckReport(t *testing.T) {
	r := &CheckReport{
		Mode: "atom",
		Rows: []CheckAtomRow{
			{AtomID: "rule_foo", ImplLinks: 2, TestLinks: 1, Semantic: "-", Status: "OK"},
			{AtomID: "rule_bar", ImplLinks: 0, TestLinks: 0, Semantic: "-", Status: "NO_IMPL"},
		},
		Summary: CheckSummary{Total: 2, WithImpl: 1, WithTests: 1},
	}

	out := formatCheckReport(r)

	if !strings.Contains(out, "Atom ID") {
		t.Error("expected 'Atom ID' header in output")
	}
	if !strings.Contains(out, "rule_foo") {
		t.Error("expected 'rule_foo' in output")
	}
	if !strings.Contains(out, "rule_bar") {
		t.Error("expected 'rule_bar' in output")
	}
	if !strings.Contains(out, "NO_IMPL") {
		t.Error("expected 'NO_IMPL' status in output")
	}
	if !strings.Contains(out, "OK") {
		t.Error("expected 'OK' status in output")
	}
	if !strings.Contains(out, "2 atoms") {
		t.Error("expected '2 atoms' in summary")
	}
}

// TestDiffChangedFilesRebasesOntoProjectRoot exercises the §1 fix: when
// config.ProjectRoot() is a subdirectory of the git repo that owns it (e.g. a
// nested project, or a submodule accessed via an umbrella working dir), `git
// diff --name-only` paths must be rebased onto ProjectRoot so they match
// SpecLink.FilePath (which the crawler always stores relative to
// ProjectRoot). Before the fix, diffChangedFiles ran `git diff` with no
// working directory at all, so it used whatever repo owned the process cwd.
func TestDiffChangedFilesRebasesOntoProjectRoot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	repoDir := t.TempDir()
	projDir := filepath.Join(repoDir, "proj")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatal(err)
	}

	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = repoDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")

	fooPath := filepath.Join(projDir, "foo.go")
	if err := os.WriteFile(fooPath, []byte("package proj\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "init")

	// Uncommitted change — this is what diff mode should surface.
	if err := os.WriteFile(fooPath, []byte("package proj\n// touched\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Give proj/ its own .atd so config.LoadFromDir anchors ProjectRoot there,
	// deliberately below repoDir (the actual git toplevel) — mirroring how a
	// submodule/nested project's root differs from the repo `git diff`
	// resolves paths against.
	if err := os.WriteFile(filepath.Join(projDir, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	savedConfig := config.Snapshot()
	defer config.Restore(savedConfig)

	if err := config.LoadFromDir(projDir); err != nil {
		t.Fatalf("config.LoadFromDir failed: %v", err)
	}
	if config.ProjectRoot() != projDir {
		t.Fatalf("expected ProjectRoot %s, got %s", projDir, config.ProjectRoot())
	}

	codeFiles, atomFiles, err := diffChangedFiles(nil)
	if err != nil {
		t.Fatalf("diffChangedFiles failed: %v", err)
	}
	if len(atomFiles) != 0 {
		t.Errorf("expected no atom files, got %v", atomFiles)
	}
	if len(codeFiles) != 1 || codeFiles[0] != "foo.go" {
		t.Errorf("expected codeFiles=[foo.go] (rebased onto ProjectRoot), got %v", codeFiles)
	}
}

// ─────────────────────────────────────────────────────────────────────────
// WP-7 unit back-fill (test_atd_07_26.md §3.1 item 2 / §6 WP-7): GenerateReport
// per mode (atom/file/full/diff) against testutil.Sandbox's fixture_project,
// the file-mode dedup case, and the id-canonicalization path. These sit
// below cmd/atd/cmd/scenario_test.go's S1/S2 (which drive the identical
// behavior through the CLI-facing runCoverageCheck entry point): these call
// coverage.GenerateReport directly. Since this file lives in package
// coverage (not coverage_test), reportRows below parses the same tabular
// text format scenario_test.go's checkRow/parseCheckReport parses — the
// text IS the contract every caller (CLI, MCP) actually gets back.
// ─────────────────────────────────────────────────────────────────────────

// reportRow is a parsed row of formatCheckReport's table, mirroring
// cmd/atd/cmd/scenario_test.go's checkRow/parseCheckReport.
type reportRow struct {
	AtomID string
	Impl   int
	Tests  int
	Status string
}

func parseReportRows(t *testing.T, text string) []reportRow {
	t.Helper()
	var rows []reportRow
	sepCount := 0
	inTable := false
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "----") {
			sepCount++
			if sepCount == 1 {
				inTable = true
				continue
			}
			if sepCount == 2 {
				inTable = false
				continue
			}
		}
		if !inTable {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		impl, _ := strconv.Atoi(fields[len(fields)-4])
		tests, _ := strconv.Atoi(fields[len(fields)-3])
		status := fields[len(fields)-1]
		atomID := strings.Join(fields[:len(fields)-4], " ")
		rows = append(rows, reportRow{AtomID: atomID, Impl: impl, Tests: tests, Status: status})
	}
	return rows
}

// TestGenerateReport_AtomMode exercises the `--atom` mode against three of
// fixture_project's pinned coverage states: fully covered (impl+test),
// implemented-but-untested, and not-implemented-at-all.
func TestGenerateReport_AtomMode(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	tests := []struct {
		name       string
		atomID     string
		wantImpl   int
		wantTests  int
		wantStatus string
	}{
		{"covered", "api_zzfix_beta", 2, 1, "OK"},
		{"implemented_untested", "rule_zzfix_untested", 1, 0, "NO_TESTS"},
		{"no_impl", "req_zzfix_draft", 0, 0, "NO_IMPL"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res := sb.Run(func() (string, error) {
				report, err := GenerateReport("atom", tt.atomID, "", sb.DocsDir, false, false, nil)
				if err != nil {
					return "", err
				}
				return report.Text, nil
			})
			if res.Err != nil {
				t.Fatalf("GenerateReport(atom=%s): %v", tt.atomID, res.Err)
			}
			if !strings.Contains(res.Output, "mode: atom") {
				t.Errorf("expected mode: atom in header, got:\n%s", res.Output)
			}
			rows := parseReportRows(t, res.Output)
			if len(rows) != 1 {
				t.Fatalf("expected exactly 1 row for %s, got %d: %+v", tt.atomID, len(rows), rows)
			}
			row := rows[0]
			if row.AtomID != tt.atomID {
				t.Errorf("expected atom id %s, got %s", tt.atomID, row.AtomID)
			}
			if row.Impl != tt.wantImpl {
				t.Errorf("%s: expected %d impl links, got %d", tt.atomID, tt.wantImpl, row.Impl)
			}
			if row.Tests != tt.wantTests {
				t.Errorf("%s: expected %d test links, got %d", tt.atomID, tt.wantTests, row.Tests)
			}
			if row.Status != tt.wantStatus {
				t.Errorf("%s: expected status %s, got %s", tt.atomID, tt.wantStatus, row.Status)
			}
		})
	}
}

// TestGenerateReport_AtomMode_NonsenseID pins the E1 "resolve or shout"
// contract at the GenerateReport (not just cmd's runCoverageCheck) level: an
// id that resolves to nothing must error loudly, naming the offending id,
// rather than silently returning an empty report.
func TestGenerateReport_AtomMode_NonsenseID(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	const nonsense = "api_zzfix_does_not_exist_at_all"
	res := sb.Run(func() (string, error) {
		report, err := GenerateReport("atom", nonsense, "", sb.DocsDir, false, false, nil)
		if err != nil {
			return "", err
		}
		return report.Text, nil
	})
	if res.Err == nil {
		t.Fatal("expected a loud error for a nonsense atom id, got none")
	}
	if !strings.Contains(res.Err.Error(), nonsense) {
		t.Errorf("expected error to name the input id %q, got: %v", nonsense, res.Err)
	}
}

// TestGenerateReport_FileModeDedup pins scenario S2 (test_atd_07_26.md §4) at
// the unit level: src/beta.go tags api_zzfix_beta via two separate
// @spec-link sites (Beta and BetaHelper), and `check --file` must list it
// exactly once, not twice, while still counting both impl links.
func TestGenerateReport_FileModeDedup(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	res := sb.Run(func() (string, error) {
		report, err := GenerateReport("file", "", "src/beta.go", sb.DocsDir, false, false, nil)
		if err != nil {
			return "", err
		}
		return report.Text, nil
	})
	if res.Err != nil {
		t.Fatalf("GenerateReport(file=src/beta.go): %v", res.Err)
	}
	if !strings.Contains(res.Output, "mode: file") {
		t.Errorf("expected mode: file in header, got:\n%s", res.Output)
	}

	rows := parseReportRows(t, res.Output)
	if len(rows) != 1 {
		t.Fatalf("expected exactly 1 row for src/beta.go (api_zzfix_beta tagged twice), got %d: %+v", len(rows), rows)
	}
	if rows[0].AtomID != "api_zzfix_beta" {
		t.Errorf("expected row api_zzfix_beta, got %q", rows[0].AtomID)
	}
	if rows[0].Impl != 2 {
		t.Errorf("expected 2 impl links (Beta + BetaHelper both tag api_zzfix_beta), got %d", rows[0].Impl)
	}
	if rows[0].Tests != 1 {
		t.Errorf("expected 1 test link, got %d", rows[0].Tests)
	}
}

// TestGenerateReport_FullMode exercises `--full` mode: every atom in the
// fixture's docs/ must appear exactly once, sorted by atom id, with the
// expected impl/test counts (this mirrors testdata/golden/check_full.txt's
// content).
//
// Full mode's row order used to come directly from ranging over
// explorer.Graph.Atoms, a Go map, so row order was not deterministic run to
// run (test_atd_07_26.md §8.3 #5). GenerateReport now sorts atom ids before
// building the report, so rows are asserted in that exact order below --
// no test-side sort needed.
func TestGenerateReport_FullMode(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	res := sb.Run(func() (string, error) {
		report, err := GenerateReport("", "", "", sb.DocsDir, true, false, nil)
		if err != nil {
			return "", err
		}
		return report.Text, nil
	})
	if res.Err != nil {
		t.Fatalf("GenerateReport(full=true): %v", res.Err)
	}
	if !strings.Contains(res.Output, "mode: full") {
		t.Errorf("expected mode: full in header, got:\n%s", res.Output)
	}

	rows := parseReportRows(t, res.Output)

	want := []reportRow{
		{"api_zzfix_beta", 2, 1, "OK"},
		{"contract_zzfix", 0, 0, "NO_IMPL"},
		{"mech_zzfix_gamma", 1, 1, "OK"},
		{"mech_zzfix_orphan", 0, 0, "NO_IMPL"},
		{"req_zzfix_alpha", 0, 0, "NO_IMPL"},
		{"req_zzfix_draft", 0, 0, "NO_IMPL"},
		{"req_zzfix_tech_debt_backlog", 0, 0, "NO_IMPL"},
		{"rule_zzfix_untested", 1, 0, "NO_TESTS"},
		{"vision_zzfix", 0, 0, "NO_IMPL"},
	}
	if len(rows) != len(want) {
		t.Fatalf("expected %d rows in full mode, got %d: %+v", len(want), len(rows), rows)
	}
	for i, w := range want {
		if rows[i] != w {
			t.Errorf("row %d: expected %+v, got %+v", i, w, rows[i])
		}
	}
}

// TestGenerateReport_FullModeScopedToDocsDir pins the fix for
// test_atd_07_26.md §8.3 #10: full mode used to walk the entire project
// root regardless of the configured docs path, so nested test fixtures
// living elsewhere under the project root (e.g. the real repo's
// tests/trace/docs/ or test-workspace/) leaked into `check --full` as
// extra rows. This reproduces that shape by dropping an atom file in a
// nested docs/ directory outside the sandbox's configured DocsDir, and
// asserts it does NOT appear in the full-mode report.
func TestGenerateReport_FullModeScopedToDocsDir(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	nestedDocsDir := filepath.Join(sb.Root, "tests", "trace", "docs")
	if err := os.MkdirAll(nestedDocsDir, 0o755); err != nil {
		t.Fatalf("creating nested docs dir: %v", err)
	}
	nestedAtom := `---
id: zzfix_nested_outside_docs
human_name: "zzfix Nested Outside Docs"
type: REQUIREMENT
version: 1.0
status: STABLE
priority: 3
tags: [zzfix]
parents: []
dependents: []
layer: BUSINESS
---

# zzfix Nested Outside Docs

## INTENT
To live outside the sandbox's configured docs/ (under tests/trace/docs/ instead), so full mode's docs-path scoping has a real nested fixture to exclude.

## THE RULE / LOGIC
Must never appear in ` + "`check --full`" + `'s report for this sandbox.

## TECHNICAL INTERFACE
- **Code Tag:** none.

## EXPECTATION
` + "`check --full`" + ` scoped to sb.DocsDir must not list zzfix_nested_outside_docs.
`
	nestedFile := filepath.Join(nestedDocsDir, "zzfix_nested_outside_docs.atom.md")
	if err := os.WriteFile(nestedFile, []byte(nestedAtom), 0o644); err != nil {
		t.Fatalf("writing nested atom file: %v", err)
	}

	res := sb.Run(func() (string, error) {
		report, err := GenerateReport("", "", "", sb.DocsDir, true, false, nil)
		if err != nil {
			return "", err
		}
		return report.Text, nil
	})
	if res.Err != nil {
		t.Fatalf("GenerateReport(full=true): %v", res.Err)
	}

	rows := parseReportRows(t, res.Output)
	for _, r := range rows {
		if r.AtomID == "zzfix_nested_outside_docs" {
			t.Errorf("full mode leaked an atom from outside the configured docs path: %+v", r)
		}
	}
	if len(rows) != 9 {
		t.Errorf("expected the 9 real fixture atoms only, got %d: %+v", len(rows), rows)
	}
}

// TestGenerateReport_DiffMode exercises the default (diff) mode against a
// real git repo inside the sandbox (testutil.Git commits the fixture's
// current state as HEAD): an uncommitted change to a spec-linked source
// file must surface that atom, an uncommitted change to an atom's own
// .atom.md must surface that atom's id too (the atomFiles path), and no
// uncommitted changes at all must hit the "No atoms found" early return
// (GenerateReport's len(atomIDs) == 0 branch).
func TestGenerateReport_DiffMode(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")
	sb.Git(t)

	t.Run("no_changes", func(t *testing.T) {
		res := sb.Run(func() (string, error) {
			report, err := GenerateReport("", "", "", sb.DocsDir, false, false, nil)
			if err != nil {
				return "", err
			}
			return report.Text, nil
		})
		if res.Err != nil {
			t.Fatalf("GenerateReport(diff, no changes): %v", res.Err)
		}
		if !strings.Contains(res.Output, "No atoms found for the specified scope.") {
			t.Errorf("expected the no-atoms-found message with a clean tree, got:\n%s", res.Output)
		}
	})

	t.Run("changed_code_file", func(t *testing.T) {
		gammaPath := filepath.Join(sb.SrcDir, "gamma.go")
		orig, err := os.ReadFile(gammaPath)
		if err != nil {
			t.Fatalf("reading gamma.go: %v", err)
		}
		if err := os.WriteFile(gammaPath, append(orig, []byte("\n// touched for diff-mode test\n")...), 0644); err != nil {
			t.Fatalf("writing gamma.go: %v", err)
		}

		res := sb.Run(func() (string, error) {
			report, err := GenerateReport("", "", "", sb.DocsDir, false, false, nil)
			if err != nil {
				return "", err
			}
			return report.Text, nil
		})
		if res.Err != nil {
			t.Fatalf("GenerateReport(diff, changed gamma.go): %v", res.Err)
		}
		if !strings.Contains(res.Output, "mode: diff") {
			t.Errorf("expected mode: diff in header, got:\n%s", res.Output)
		}
		rows := parseReportRows(t, res.Output)
		if len(rows) != 1 || rows[0].AtomID != "mech_zzfix_gamma" {
			t.Fatalf("expected exactly one row for mech_zzfix_gamma (the atom spec-linked in gamma.go), got %+v", rows)
		}
	})

	t.Run("changed_atom_file", func(t *testing.T) {
		orphanAtomPath := filepath.Join(sb.DocsDir, "mech_zzfix_orphan.atom.md")
		orig, err := os.ReadFile(orphanAtomPath)
		if err != nil {
			t.Fatalf("reading mech_zzfix_orphan.atom.md: %v", err)
		}
		// Restore the code-file edit from the previous subtest and the
		// previous state of this atom file isn't required here: each Run
		// call re-diffs against the same original HEAD, so both edits are
		// present as uncommitted changes; this subtest only asserts the
		// orphan atom's id is surfaced via the atomFiles path.
		if err := os.WriteFile(orphanAtomPath, append(orig, []byte("\n<!-- touched for diff-mode test -->\n")...), 0644); err != nil {
			t.Fatalf("writing mech_zzfix_orphan.atom.md: %v", err)
		}

		res := sb.Run(func() (string, error) {
			report, err := GenerateReport("", "", "", sb.DocsDir, false, false, nil)
			if err != nil {
				return "", err
			}
			return report.Text, nil
		})
		if res.Err != nil {
			t.Fatalf("GenerateReport(diff, changed mech_zzfix_orphan.atom.md): %v", res.Err)
		}
		rows := parseReportRows(t, res.Output)
		found := false
		for _, r := range rows {
			if r.AtomID == "mech_zzfix_orphan" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected mech_zzfix_orphan to be surfaced via the changed .atom.md file, got rows: %+v", rows)
		}
	})
}

// TestGenerateReport_IDCanonicalization complements scenario S1
// (cmd/atd/cmd/scenario_test.go, driven through runCoverageCheck) at the
// pkg/coverage unit level: calling GenerateReport directly with the
// canonical id, the bare id, and a redundant TYPE_-prefixed id must all
// produce byte-identical reports (Addendum B's workspace-resolver
// strip-and-retry, generalized from pkg/workspace/resolver_test.go's own
// unit tests through to the check engine). This uses a workspace fixture
// (fixture_workspace/zzfix_a) to exercise the workspace.Resolver path
// specifically; pkg/exploration/resolve_test.go's
// TestResolveAtom_StandaloneBareIDAndKnownDefect covers the equivalent
// standalone-project case now that CanonicalAtomID's own strip-and-retry
// (test_atd_07_26.md §8.3 #1) applies there too.
func TestGenerateReport_IDCanonicalization(t *testing.T) {
	t.Parallel()
	ws := testutil.Sandbox(t, "fixture_workspace")
	a := ws.Sub("zzfix_a")

	const canonical = "req_zzfix_ws_alpha"
	const bare = "req_zzfix_ws_alpha"
	const typePrefixed = "requirement_req_zzfix_ws_alpha"

	run := func(id string) (string, error) {
		res := a.Run(func() (string, error) {
			report, err := GenerateReport("atom", id, "", a.DocsDir, false, false, nil)
			if err != nil {
				return "", err
			}
			return report.Text, nil
		})
		return res.Output, res.Err
	}

	canonicalOut, err := run(canonical)
	if err != nil {
		t.Fatalf("canonical id %q: %v", canonical, err)
	}
	bareOut, err := run(bare)
	if err != nil {
		t.Fatalf("bare id %q: %v", bare, err)
	}
	typeOut, err := run(typePrefixed)
	if err != nil {
		t.Fatalf("TYPE_-prefixed id %q: %v", typePrefixed, err)
	}

	if canonicalOut != bareOut {
		t.Errorf("canonical vs bare id gave different reports:\ncanonical:\n%s\nbare:\n%s", canonicalOut, bareOut)
	}
	if canonicalOut != typeOut {
		t.Errorf("canonical vs TYPE_-prefixed id gave different reports:\ncanonical:\n%s\ntype-prefixed:\n%s", canonicalOut, typeOut)
	}

	rows := parseReportRows(t, canonicalOut)
	if len(rows) != 1 || rows[0].Impl != 1 || rows[0].Tests != 1 || rows[0].Status != "OK" {
		t.Fatalf("expected req_zzfix_ws_alpha to show 1 impl / 1 test / OK, got: %+v", rows)
	}

	// The nonsense-id half of "resolve or shout" is pinned at the
	// GenerateReport level too, in the workspace context this time (a
	// Resolver is active, unlike TestGenerateReport_AtomMode_NonsenseID's
	// standalone fixture).
	const nonsense = "req_zzfix_ws_totally_bogus"
	_, err = run(nonsense)
	if err == nil {
		t.Fatal("expected a loud error for a nonsense atom id in a workspace project, got none")
	}
	if !strings.Contains(err.Error(), nonsense) {
		t.Errorf("expected error to name the input id %q, got: %v", nonsense, err)
	}
}
