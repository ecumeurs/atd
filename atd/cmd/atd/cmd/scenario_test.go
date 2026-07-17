package cmd

// Scenario tests against the fixture corpus (test_atd_07_26.md §3.2, §4).
// Each Test* function below implements exactly one scenario from the
// catalog and cites its scenario id in its name and doc comment. All run
// against atd/testdata/fixture_project (or fixture_workspace) via
// testutil.Sandbox, which copies the fixture into a fresh t.TempDir() per
// test so t.Parallel() is safe: config.ActiveConfig mutation is confined to
// (*testutil.SB).Run's mutex-protected critical section.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"atd-tools/pkg/atom"
	"atd-tools/pkg/mcp"
	"atd-tools/pkg/testutil"
)

// ── shared helpers ─────────────────────────────────────────────────────

// checkRow is a parsed row from the tabular report coverage.formatCheckReport
// prints ("# ATD Coverage Report ..." followed by a fixed-width table).
// Scenario tests parse this text rather than reaching into the coverage
// package's unexported buildCoverageReport, since the text IS the contract
// both the CLI and MCP surfaces actually return to a caller.
type checkRow struct {
	AtomID string
	Impl   int
	Tests  int
	Status string
}

// normalizeTraceJSON re-marshals trace's raw JSON output with graph_slice's
// code_links/test_links sorted. Those two fields are built from Go maps in
// pkg/exploration/trace.go (codeFiles/testFiles), so their array order is
// not deterministic run to run even for byte-identical input -- a real,
// observed nondeterminism (confirmed via two otherwise-identical trace
// calls returning differently-ordered lists). Comparing trace output for
// equality must sort these first or it flakes on element order that was
// never semantically meaningful. context's map is unaffected: encoding/json
// already sorts map keys when marshaling a Go map into a JSON object.
func normalizeTraceJSON(t *testing.T, s string) string {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("unmarshaling trace output: %v\n%s", err, s)
	}
	if gs, ok := v["graph_slice"].(map[string]any); ok {
		for _, key := range []string{"code_links", "test_links"} {
			if arr, ok := gs[key].([]any); ok {
				sort.Slice(arr, func(i, j int) bool {
					return fmt.Sprint(arr[i]) < fmt.Sprint(arr[j])
				})
				gs[key] = arr
			}
		}
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("re-marshaling trace output: %v", err)
	}
	return string(out)
}

func parseCheckReport(t *testing.T, text string) []checkRow {
	t.Helper()
	var rows []checkRow
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
		rows = append(rows, checkRow{AtomID: atomID, Impl: impl, Tests: tests, Status: status})
	}
	return rows
}

// ── S1: id-form resolution (E1, "resolve or shout") ────────────────────

// TestScenario_S1_IDFormResolution pins test_atd_07_26.md §4 S1: for each of
// check, test_links, trace and query, canonical/bare id forms must return
// identical coverage answers, and a nonsense id must error loudly (naming
// the input) rather than silently returning an empty/zero result.
func TestScenario_S1_IDFormResolution(t *testing.T) {
	t.Run("standalone_project", testScenarioS1Standalone)
	t.Run("workspace_project", testScenarioS1Workspace)
}

// testScenarioS1Standalone runs against the standalone fixture_project (no
// workspace). Here "canonical id" and "bare id" are the literal same
// string: without a workspace.Resolver there is no project-qualification
// layer to strip.
//
// FIXED (test_atd_07_26.md §8.3 #1): pkg/workspace/resolver.go's
// strip-known-prefix retry (Addendum B, atd_feedback_evaluation_2026-07-13.md)
// — which lets "requirement_req_x" resolve to "req_x" — used to live only on
// workspace.Resolver. A standalone project (no .atd.workspace found upward)
// never constructs one (pkg/exploration/explorer.go's NewExplorerWithConfig
// leaves Resolver nil), so CanonicalAtomID/ResolveAtom fell back to an exact
// map lookup with no retry: a redundant TYPE_-prefixed id failed EXACTLY
// like a genuinely nonsense id. pkg/exploration/resolve.go's
// CanonicalAtomID/ResolveAtom now apply the same strip-and-retry directly in
// their standalone (no e.Resolver) path, so the TYPE_-prefixed case below now
// resolves just like testScenarioS1Workspace's workspace-resolver case
// already did.
func testScenarioS1Standalone(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	const canonical = "api_zzfix_beta"
	const bare = "api_zzfix_beta" // same string standalone; see comment above
	const typePrefixed = "api_api_zzfix_beta"
	const nonsense = "api_zzfix_does_not_exist_at_all"

	t.Run("check", func(t *testing.T) {
		wantRes := sb.Run(func() (string, error) {
			return runCoverageCheck("atom", canonical, "", sb.DocsDir, false, false, nil)
		})
		if wantRes.Err != nil {
			t.Fatalf("canonical id %q: %v", canonical, wantRes.Err)
		}
		gotRes := sb.Run(func() (string, error) {
			return runCoverageCheck("atom", bare, "", sb.DocsDir, false, false, nil)
		})
		if gotRes.Err != nil {
			t.Fatalf("bare id %q: %v", bare, gotRes.Err)
		}
		if wantRes.Output != gotRes.Output {
			t.Errorf("canonical vs bare id gave different check reports:\ncanonical:\n%s\nbare:\n%s", wantRes.Output, gotRes.Output)
		}

		typeRes := sb.Run(func() (string, error) {
			return runCoverageCheck("atom", typePrefixed, "", sb.DocsDir, false, false, nil)
		})
		if typeRes.Err != nil {
			t.Errorf("expected TYPE_-prefixed id %q to resolve via the standalone strip-and-retry in a standalone project via check, got error: %v", typePrefixed, typeRes.Err)
		}
		if typeRes.Output != wantRes.Output {
			t.Errorf("TYPE_-prefixed id resolved to a different report than canonical:\ntype-prefixed:\n%s\ncanonical:\n%s", typeRes.Output, wantRes.Output)
		}

		nonsenseRes := sb.Run(func() (string, error) {
			return runCoverageCheck("atom", nonsense, "", sb.DocsDir, false, false, nil)
		})
		if nonsenseRes.Err == nil {
			t.Fatal("expected a loud error for a nonsense atom id via check, got none")
		}
		if !strings.Contains(nonsenseRes.Err.Error(), nonsense) {
			t.Errorf("expected check's error to name the input id %q, got: %v", nonsense, nonsenseRes.Err)
		}
	})

	t.Run("test_links", func(t *testing.T) {
		// atd_test_links (MCP) and CLI `check --atom` both delegate to the
		// identical runCoverageCheck("atom", ...) entry point (see
		// mcp_tools.go's atd_test_links handler) -- exercised directly here.
		wantRes := sb.Run(func() (string, error) {
			return runCoverageCheck("atom", canonical, "", sb.DocsDir, false, false, nil)
		})
		gotRes := sb.Run(func() (string, error) {
			return runCoverageCheck("atom", bare, "", sb.DocsDir, false, false, nil)
		})
		if wantRes.Err != nil || gotRes.Err != nil {
			t.Fatalf("canonical/bare id errors: %v / %v", wantRes.Err, gotRes.Err)
		}
		if wantRes.Output != gotRes.Output {
			t.Errorf("canonical vs bare id gave different test_links reports:\ncanonical:\n%s\nbare:\n%s", wantRes.Output, gotRes.Output)
		}

		nonsenseRes := sb.Run(func() (string, error) {
			return runCoverageCheck("atom", nonsense, "", sb.DocsDir, false, false, nil)
		})
		if nonsenseRes.Err == nil {
			t.Fatal("expected a loud error for a nonsense atom id via test_links, got none")
		}
	})

	t.Run("trace", func(t *testing.T) {
		wantRes := sb.Run(func() (string, error) { return runTrace(canonical, sb.DocsDir, sb.Root, false) })
		gotRes := sb.Run(func() (string, error) { return runTrace(bare, sb.DocsDir, sb.Root, false) })
		if wantRes.Err != nil || gotRes.Err != nil {
			t.Fatalf("canonical/bare id errors: %v / %v", wantRes.Err, gotRes.Err)
		}
		if normalizeTraceJSON(t, wantRes.Output) != normalizeTraceJSON(t, gotRes.Output) {
			t.Errorf("canonical vs bare id gave different trace reports:\ncanonical:\n%s\nbare:\n%s", wantRes.Output, gotRes.Output)
		}

		typeRes := sb.Run(func() (string, error) { return runTrace(typePrefixed, sb.DocsDir, sb.Root, false) })
		if typeRes.Err != nil {
			t.Errorf("expected TYPE_-prefixed id %q to resolve via the standalone strip-and-retry in a standalone project via trace, got error: %v", typePrefixed, typeRes.Err)
		}
		if normalizeTraceJSON(t, wantRes.Output) != normalizeTraceJSON(t, typeRes.Output) {
			t.Errorf("TYPE_-prefixed id traced differently than canonical:\ntype-prefixed:\n%s\ncanonical:\n%s", typeRes.Output, wantRes.Output)
		}

		nonsenseRes := sb.Run(func() (string, error) { return runTrace(nonsense, sb.DocsDir, sb.Root, false) })
		if nonsenseRes.Err == nil {
			t.Fatal("expected a loud error for a nonsense atom id via trace, got none")
		}
		if !strings.Contains(nonsenseRes.Err.Error(), nonsense) {
			t.Errorf("expected trace's error to name the input id %q, got: %v", nonsense, nonsenseRes.Err)
		}
	})

	t.Run("query", func(t *testing.T) {
		// FIXED (test_atd_07_26.md §8.3 #2): query used to be architecturally
		// different from check/test_links/trace -- it is a case-insensitive
		// SUBSTRING search over frontmatter fields (pkg/exploration/query.go)
		// that never canonicalizes an id and never errors, so a nonsense
		// search string just returned an empty/null JSON array, not a loud
		// "not found". runQuery (cmd/atd/cmd/query.go) now treats a
		// zero-match field="id" lookup as a resolution, not a general
		// keyword search, and errors loudly naming the input -- matching
		// check/trace/test_links. Other fields are untouched: a substring
		// search with no hits on e.g. field="tags" is still a legitimate
		// empty result (see pkg/exploration/query_test.go).
		wantRes := sb.Run(func() (string, error) { return runQuery("id", canonical, false) })
		gotRes := sb.Run(func() (string, error) { return runQuery("id", bare, false) })
		if wantRes.Err != nil || gotRes.Err != nil {
			t.Fatalf("canonical/bare id errors: %v / %v", wantRes.Err, gotRes.Err)
		}
		if wantRes.Output != gotRes.Output {
			t.Errorf("canonical vs bare id gave different query results:\ncanonical:\n%s\nbare:\n%s", wantRes.Output, gotRes.Output)
		}
		var matches []atom.AtomData
		if err := json.Unmarshal([]byte(wantRes.Output), &matches); err != nil {
			t.Fatalf("unmarshaling query output: %v", err)
		}
		if len(matches) != 1 || matches[0].ID != canonical {
			t.Fatalf("expected query(id=%s) to return exactly one match, got %+v", canonical, matches)
		}

		nonsenseRes := sb.Run(func() (string, error) { return runQuery("id", nonsense, false) })
		if nonsenseRes.Err == nil {
			t.Fatal("expected a loud error for a nonsense atom id via query, got none")
		}
		if !strings.Contains(nonsenseRes.Err.Error(), nonsense) {
			t.Errorf("expected query's error to name the input id %q, got: %v", nonsense, nonsenseRes.Err)
		}
	})
}

// testScenarioS1Workspace exercises the same TYPE_-prefixed-id case against
// fixture_workspace's zzfix_a project, where a workspace.Resolver IS active
// (an .atd.workspace exists above it) -- proving Addendum B's strip-and-
// retry mechanism generalizes to check/test_links/trace (not just
// pkg/workspace/resolver_test.go's own unit tests), the same way
// testScenarioS1Standalone above now proves it for the standalone path too.
func testScenarioS1Workspace(t *testing.T) {
	t.Parallel()
	ws := testutil.Sandbox(t, "fixture_workspace")
	a := ws.Sub("zzfix_a")

	const canonical = "req_zzfix_ws_alpha"
	const bare = "req_zzfix_ws_alpha"
	const typePrefixed = "requirement_req_zzfix_ws_alpha"
	const nonsense = "req_zzfix_ws_totally_bogus"

	t.Run("check", func(t *testing.T) {
		wantRes := a.Run(func() (string, error) {
			return runCoverageCheck("atom", canonical, "", a.DocsDir, false, false, nil)
		})
		if wantRes.Err != nil {
			t.Fatalf("canonical id: %v", wantRes.Err)
		}
		bareRes := a.Run(func() (string, error) {
			return runCoverageCheck("atom", bare, "", a.DocsDir, false, false, nil)
		})
		if bareRes.Err != nil || bareRes.Output != wantRes.Output {
			t.Errorf("bare id gave a different/erroring result: err=%v output=%q want=%q", bareRes.Err, bareRes.Output, wantRes.Output)
		}
		typeRes := a.Run(func() (string, error) {
			return runCoverageCheck("atom", typePrefixed, "", a.DocsDir, false, false, nil)
		})
		if typeRes.Err != nil {
			t.Fatalf("expected TYPE_-prefixed id %q to resolve via the workspace resolver's strip-and-retry, got error: %v", typePrefixed, typeRes.Err)
		}
		if typeRes.Output != wantRes.Output {
			t.Errorf("TYPE_-prefixed id resolved to a different report than canonical:\ntype-prefixed:\n%s\ncanonical:\n%s", typeRes.Output, wantRes.Output)
		}
		rows := parseCheckReport(t, wantRes.Output)
		if len(rows) != 1 || rows[0].Impl != 1 || rows[0].Tests != 1 {
			t.Fatalf("expected req_zzfix_ws_alpha to show 1 impl / 1 test, got: %+v", rows)
		}

		nonsenseRes := a.Run(func() (string, error) {
			return runCoverageCheck("atom", nonsense, "", a.DocsDir, false, false, nil)
		})
		if nonsenseRes.Err == nil {
			t.Fatal("expected a loud error for a nonsense atom id, got none")
		}
	})

	t.Run("trace", func(t *testing.T) {
		wantRes := a.Run(func() (string, error) { return runTrace(canonical, a.DocsDir, a.Root, false) })
		typeRes := a.Run(func() (string, error) { return runTrace(typePrefixed, a.DocsDir, a.Root, false) })
		if wantRes.Err != nil || typeRes.Err != nil {
			t.Fatalf("canonical/type-prefixed errors: %v / %v", wantRes.Err, typeRes.Err)
		}
		if normalizeTraceJSON(t, wantRes.Output) != normalizeTraceJSON(t, typeRes.Output) {
			t.Errorf("TYPE_-prefixed id traced differently than canonical:\ntype-prefixed:\n%s\ncanonical:\n%s", typeRes.Output, wantRes.Output)
		}

		nonsenseRes := a.Run(func() (string, error) { return runTrace(nonsense, a.DocsDir, a.Root, false) })
		if nonsenseRes.Err == nil {
			t.Fatal("expected a loud error for a nonsense atom id via trace, got none")
		}
	})
}

// ── S2: check --file dedup ──────────────────────────────────────────────

// TestScenario_S2_CheckFileDedup pins test_atd_07_26.md §4 S2: `check --file`
// must list each spec-linked atom exactly once, even when that file tags
// the same atom more than once (src/beta.go tags api_zzfix_beta twice).
func TestScenario_S2_CheckFileDedup(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	res := sb.Run(func() (string, error) {
		return runCoverageCheck("file", "", "src/beta.go", sb.DocsDir, false, false, nil)
	})
	if res.Err != nil {
		t.Fatalf("check --file: %v", res.Err)
	}

	rows := parseCheckReport(t, res.Output)
	if len(rows) != 1 {
		t.Fatalf("expected exactly 1 row for src/beta.go (api_zzfix_beta tagged twice), got %d: %+v", len(rows), rows)
	}
	if rows[0].AtomID != "api_zzfix_beta" {
		t.Errorf("expected the row to be api_zzfix_beta, got %q", rows[0].AtomID)
	}
	if rows[0].Impl != 2 {
		t.Errorf("expected 2 impl links (Beta + BetaHelper both tag api_zzfix_beta), got %d", rows[0].Impl)
	}
	if rows[0].Tests != 1 {
		t.Errorf("expected 1 test link, got %d", rows[0].Tests)
	}
}

// ── S3: consistency oracle ───────────────────────────────────────────────

// TestScenario_S3_ConsistencyOracle pins test_atd_07_26.md §4 S3: for every
// atom in the fixture, check, trace, and query must agree on impl-link and
// test-link presence -- the literal E2 defect shape ("trace/check/query
// disagreement").
//
// check and query draw from identical underlying data
// (Explorer.SpecLinksForAtom/TestLinksForAtom and
// AtomData.Implementations/HasTests respectively) and are compared for
// exact agreement on every atom, with no exceptions.
//
// trace is compared too, but only for atoms with no dependents: trace's
// CodeLinks/TestLinks aggregate the target's own links WITH every
// dependent's (pkg/exploration/trace.go's processNode loop over
// snap.GraphSlice.Dependents) as a deliberate rollup for BUSINESS/
// ARCHITECTURE ancestry health -- a wider, non-comparable metric for any
// atom that itself has dependents. Comparing only leaf atoms keeps the
// check honest rather than asserting a false equality across differently
// scoped metrics.
func TestScenario_S3_ConsistencyOracle(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	atomFiles, err := filepath.Glob(filepath.Join(sb.DocsDir, "*.atom.md"))
	if err != nil || len(atomFiles) == 0 {
		t.Fatalf("failed to glob fixture atoms: %v (found %d)", err, len(atomFiles))
	}

	var ids []string
	hasDependents := map[string]bool{}
	for _, f := range atomFiles {
		a, err := atom.Parse(f)
		if err != nil {
			t.Fatalf("parsing fixture atom %s: %v", f, err)
		}
		ids = append(ids, a.ID)
		if len(a.Dependents) > 0 {
			hasDependents[a.ID] = true
		}
	}
	sort.Strings(ids)

	for _, id := range ids {
		id := id
		t.Run(id, func(t *testing.T) {
			checkRes := sb.Run(func() (string, error) {
				return runCoverageCheck("atom", id, "", sb.DocsDir, false, false, nil)
			})
			if checkRes.Err != nil {
				t.Fatalf("check --atom %s: %v", id, checkRes.Err)
			}
			rows := parseCheckReport(t, checkRes.Output)
			if len(rows) != 1 {
				t.Fatalf("expected exactly 1 check row for %s, got %d: %+v", id, len(rows), rows)
			}
			checkImpl, checkTests := rows[0].Impl, rows[0].Tests

			queryRes := sb.Run(func() (string, error) { return runQuery("id", id, false) })
			if queryRes.Err != nil {
				t.Fatalf("query id=%s: %v", id, queryRes.Err)
			}
			var atoms []atom.AtomData
			if err := json.Unmarshal([]byte(queryRes.Output), &atoms); err != nil {
				t.Fatalf("unmarshaling query output for %s: %v\n%s", id, err, queryRes.Output)
			}
			var match *atom.AtomData
			for i := range atoms {
				if atoms[i].ID == id {
					match = &atoms[i]
					break
				}
			}
			if match == nil {
				t.Fatalf("query id=%s returned no exact match among %d result(s)", id, len(atoms))
			}

			if checkImpl != len(match.Implementations) {
				t.Errorf("check/query disagree on impl-link count for %s: check=%d query=%d", id, checkImpl, len(match.Implementations))
			}
			if (checkTests > 0) != match.HasTests {
				t.Errorf("check/query disagree on has-test for %s: check tests=%d query has_tests=%v", id, checkTests, match.HasTests)
			}

			traceRes := sb.Run(func() (string, error) { return runTrace(id, sb.DocsDir, sb.Root, false) })
			if traceRes.Err != nil {
				t.Fatalf("trace %s: %v", id, traceRes.Err)
			}
			if hasDependents[id] {
				// Wider rollup metric -- resolves fine, not compared 1:1.
				return
			}

			var snap struct {
				GraphSlice struct {
					CodeLinks []string `json:"code_links"`
					TestLinks []string `json:"test_links"`
				} `json:"graph_slice"`
			}
			if err := json.Unmarshal([]byte(traceRes.Output), &snap); err != nil {
				t.Fatalf("unmarshaling trace output for %s: %v", id, err)
			}
			traceHasImpl := len(snap.GraphSlice.CodeLinks) > 0
			traceHasTests := len(snap.GraphSlice.TestLinks) > 0

			if traceHasImpl != (checkImpl > 0) {
				t.Errorf("check/trace disagree on has-impl for leaf atom %s: check impl=%d trace code_links=%v", id, checkImpl, snap.GraphSlice.CodeLinks)
			}
			if traceHasTests != (checkTests > 0) {
				t.Errorf("check/trace disagree on has-test for leaf atom %s: check tests=%d trace test_links=%v", id, checkTests, snap.GraphSlice.TestLinks)
			}
		})
	}
}

// ── S4: STABLE+BUSINESS governance guard ────────────────────────────────

// TestScenario_S4_StableBusinessGuard pins test_atd_07_26.md §4 S4: `update`
// on a STABLE BUSINESS atom is refused with an actionable message; --force
// (CLI) overrides it; a DRAFT BUSINESS atom sees no friction at all.
func TestScenario_S4_StableBusinessGuard(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")
	reqAlphaPath := filepath.Join(sb.DocsDir, "req_zzfix_alpha.atom.md")
	reqDraftPath := filepath.Join(sb.DocsDir, "req_zzfix_draft.atom.md")

	t.Run("stable_business_refused_without_force", func(t *testing.T) {
		res := sb.Run(func() (string, error) {
			return runUpdate(reqAlphaPath, []string{"priority=1"}, "", "", "", "", "", "", false)
		})
		if res.Err == nil {
			t.Fatal("expected refusal updating a STABLE BUSINESS atom without --force, got success")
		}
		if !strings.Contains(res.Err.Error(), "req_zzfix_alpha") {
			t.Errorf("expected the refusal to name the atom id, got: %v", res.Err)
		}
		if !strings.Contains(strings.ToLower(res.Err.Error()), "force") {
			t.Errorf("expected the refusal to mention --force as the escape hatch, got: %v", res.Err)
		}
	})

	t.Run("stable_business_succeeds_with_force", func(t *testing.T) {
		res := sb.Run(func() (string, error) {
			return runUpdate(reqAlphaPath, []string{"priority=1"}, "", "", "", "", "", "", true)
		})
		if res.Err != nil {
			t.Fatalf("expected --force to override the guard, got error: %v", res.Err)
		}
		updated, err := atom.Parse(reqAlphaPath)
		if err != nil {
			t.Fatalf("re-parsing updated atom: %v", err)
		}
		if updated.Priority != "1" {
			t.Errorf("expected priority updated to 1, got %q", updated.Priority)
		}
	})

	t.Run("draft_atom_no_friction", func(t *testing.T) {
		res := sb.Run(func() (string, error) {
			return runUpdate(reqDraftPath, []string{"priority=2"}, "", "", "", "", "", "", false)
		})
		if res.Err != nil {
			t.Fatalf("expected a DRAFT atom update to proceed without --force, got error: %v", res.Err)
		}
	})

	// KNOWN DEFECT: the MCP atd_update tool (cmd/atd/cmd/mcp_tools.go) has
	// no `force` parameter in its InputSchema, and its handler never reads
	// one -- unlike the CLI's --force flag. test_atd_07_26.md's S4
	// explicitly expects an MCP path "same via r.Call(atd_update, ...)";
	// today an MCP client has NO way to override the STABLE+BUSINESS guard
	// at all. This pins the current (always-refuses, force-silently-
	// ignored) behavior rather than silently loosening the assertion.
	t.Run("mcp_path_force_param_not_wired", func(t *testing.T) {
		sb2 := testutil.Sandbox(t, "fixture_project")
		path := filepath.Join(sb2.DocsDir, "req_zzfix_alpha.atom.md")

		r := mcp.NewRegistry()
		RegisterMCPTools(r)

		res := sb2.Run(func() (string, error) {
			return r.Call("atd_update", map[string]any{
				"file":  path,
				"set":   []any{"priority=9"},
				"force": true, // accepted into the args map, but never read by the handler
			})
		})
		if res.Err == nil {
			t.Fatal("KNOWN DEFECT expectation changed: atd_update via MCP with force:true unexpectedly succeeded on a STABLE BUSINESS atom -- if force is now wired into the MCP handler, update/remove this pin")
		}
	})
}

// ── S5: governance lint ──────────────────────────────────────────────────

// TestScenario_S5_GovernanceLint pins test_atd_07_26.md §4 S5: a corpus
// with 0 CONTRACT atoms errors; 2 CONTRACT atoms errors; a non-canonical
// type (FOO) errors naming the offending atom; the sanctioned-union types
// SERVICE/USAGE do not error.
//
// These use ad-hoc temp corpora (not fixture_project) since the atom
// content/count IS the thing under test here, per the harness's own
// guidance ("no test constructs atoms inline unless the atom's content is
// the thing under test").
func TestScenario_S5_GovernanceLint(t *testing.T) {
	t.Run("zero_contracts_errors", func(t *testing.T) {
		t.Parallel()
		root, docsDir := newGovFixture(t)
		writeGovAtom(t, docsDir, "req_zzfix_gov_a", "REQUIREMENT", "BUSINESS", "STABLE", nil)

		res := runLintInDir(root, docsDir)
		if res.Err == nil {
			t.Fatal("expected a lint error for a corpus with 0 CONTRACT atoms and a BUSINESS atom present")
		}
		if !strings.Contains(res.Output, "CONTRACT") {
			t.Errorf("expected lint output to mention the missing CONTRACT, got:\n%s", res.Output)
		}
	})

	t.Run("two_contracts_errors", func(t *testing.T) {
		t.Parallel()
		root, docsDir := newGovFixture(t)
		writeGovAtom(t, docsDir, "contract_zzfix_gov_one", "CONTRACT", "BUSINESS", "STABLE", nil)
		writeGovAtom(t, docsDir, "contract_zzfix_gov_two", "CONTRACT", "BUSINESS", "STABLE", nil)
		writeGovAtom(t, docsDir, "vision_zzfix_gov", "VISION", "BUSINESS", "STABLE", nil)

		res := runLintInDir(root, docsDir)
		if res.Err == nil {
			t.Fatal("expected a lint error for a corpus with 2 CONTRACT atoms")
		}
		if !strings.Contains(res.Output, "Multiple CONTRACT atoms") {
			t.Errorf("expected lint output to name the multiple-CONTRACT violation, got:\n%s", res.Output)
		}
	})

	t.Run("non_canonical_type_errors_naming_the_atom", func(t *testing.T) {
		t.Parallel()
		root, docsDir := newGovFixture(t)
		writeGovAtom(t, docsDir, "zzfix_gov_bad_type", "FOO", "ARCHITECTURE", "STABLE", nil)

		res := runLintInDir(root, docsDir)
		if res.Err == nil {
			t.Fatal("expected a lint error for a non-canonical type FOO")
		}
		if !strings.Contains(res.Output, "zzfix_gov_bad_type") {
			t.Errorf("expected lint output to name the offending atom, got:\n%s", res.Output)
		}
		if !strings.Contains(res.Output, "Non-canonical type: FOO") {
			t.Errorf("expected lint output to name the bad type FOO, got:\n%s", res.Output)
		}
	})

	t.Run("service_and_usage_are_sanctioned_no_error", func(t *testing.T) {
		t.Parallel()
		root, docsDir := newGovFixture(t)
		writeGovAtom(t, docsDir, "zzfix_gov_service", "SERVICE", "ARCHITECTURE", "STABLE", nil)
		writeGovAtom(t, docsDir, "zzfix_gov_usage", "USAGE", "ARCHITECTURE", "STABLE", nil)

		res := runLintInDir(root, docsDir)
		if res.Err != nil {
			t.Errorf("expected SERVICE/USAGE types to lint clean, got error: %v\n%s", res.Err, res.Output)
		}
	})
}

func newGovFixture(t *testing.T) (root, docsDir string) {
	t.Helper()
	testutil.SnapshotConfigLocked(t)
	root = t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	docsDir = filepath.Join(root, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}
	return root, docsDir
}

func writeGovAtom(t *testing.T, docsDir, id, atomType, layer, status string, parents []string) {
	t.Helper()
	parentsYAML := "[]"
	if len(parents) > 0 {
		var b strings.Builder
		for _, p := range parents {
			b.WriteString(fmt.Sprintf("\n  - [[%s]]", p))
		}
		parentsYAML = b.String()
	}
	content := fmt.Sprintf(`---
id: %s
human_name: "%s"
type: %s
layer: %s
version: 1.0
status: %s
priority: 3
tags: [zzfix]
parents: %s
dependents: []
---

# %s

## INTENT
Intent for %s.

## THE RULE / LOGIC
Logic for %s.

## TECHNICAL INTERFACE
- **Code Tag:** `+"`"+`@spec-link [[%s]]`+"`"+`

## EXPECTATION
Expectation for %s.
`, id, id, atomType, layer, status, parentsYAML, id, id, id, id, id)
	if err := os.WriteFile(filepath.Join(docsDir, id+".atom.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// runLintInDir runs runLint(docsDir) with config.ActiveConfig loaded from
// root, under the same mutex (*testutil.SB).Run uses -- constructed
// directly (not via Sandbox) since these are ad-hoc temp dirs, not a
// checked-in fixture copy. SB.Run only touches its own exported fields, so
// this is safe.
func runLintInDir(root, docsDir string) testutil.Result {
	sb := &testutil.SB{Root: root, DocsDir: docsDir}
	return sb.Run(func() (string, error) { return runLint(docsDir) })
}

// ── S7: rename propagation ────────────────────────────────────────────

// TestScenario_S7_RenamePropagation pins test_atd_07_26.md §4 S7: `update
// set id=...` on an atom with two inbound [[refs]] (one in another atom's
// frontmatter parents:, one in a third atom's body prose) renames the file
// per convention, rewrites both refs, and leaves the corpus lint-clean.
func TestScenario_S7_RenamePropagation(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	const oldID = "req_zzfix_alpha"
	// Deliberately already carries the REQUIREMENT type's full-word naming
	// convention atom.Update enforces (see pkg/atom/update.go's prefix
	// check), so the rename lands exactly on this id with no incidental
	// re-mangling unrelated to what this test is pinning.
	const newID = "requirement_zzfix_alpha_renamed"

	oldPath := filepath.Join(sb.DocsDir, oldID+".atom.md")
	newPath := filepath.Join(sb.DocsDir, newID+".atom.md")

	// req_zzfix_alpha is STABLE+BUSINESS: force is required (S4 covers the
	// refusal itself; this test is about propagation, not the guard).
	res := sb.Run(func() (string, error) {
		return runUpdate(oldPath, []string{"id=" + newID}, "", "", "", "", "", "", true)
	})
	if res.Err != nil {
		t.Fatalf("rename update: %v", res.Err)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("expected old file %s to be gone (stat err: %v)", oldPath, err)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("expected renamed file %s to exist: %v", newPath, err)
	}

	betaDoc, err := os.ReadFile(filepath.Join(sb.DocsDir, "api_zzfix_beta.atom.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(betaDoc), "[["+oldID+"]]") {
		t.Errorf("api_zzfix_beta.atom.md (frontmatter parents: ref) still references old id [[%s]]", oldID)
	}
	if !strings.Contains(string(betaDoc), "[["+newID+"]]") {
		t.Errorf("api_zzfix_beta.atom.md was not rewritten to reference [[%s]]", newID)
	}

	ruleDoc, err := os.ReadFile(filepath.Join(sb.DocsDir, "rule_zzfix_untested.atom.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(ruleDoc), "[["+oldID+"]]") {
		t.Errorf("rule_zzfix_untested.atom.md (body prose ref) still references old id [[%s]]", oldID)
	}
	if !strings.Contains(string(ruleDoc), "[["+newID+"]]") {
		t.Errorf("rule_zzfix_untested.atom.md was not rewritten to reference [[%s]]", newID)
	}

	lintRes := sb.Run(func() (string, error) { return runLint(sb.DocsDir) })
	if lintRes.Err != nil {
		t.Errorf("expected lint clean after rename propagation, got error: %v\n%s", lintRes.Err, lintRes.Output)
	}
}

// ── S8: orphan / gaps ────────────────────────────────────────────────────

// TestScenario_S8_OrphanGaps pins test_atd_07_26.md §4 S8: `crawl
// gaps=true` reports exactly the orphan mechanic, with no false positives
// on the covered chain; the untested rule's missing test-link is a
// distinct state check/lint surface (crawl's gaps detect missing
// *implementation*, not missing *tests* -- see comment inline).
func TestScenario_S8_OrphanGaps(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	gapsRes := sb.Run(func() (string, error) { return runCrawl(sb.Root, sb.DocsDir, true, false) })
	if gapsRes.Err != nil {
		t.Fatalf("crawl --gaps: %v", gapsRes.Err)
	}
	var report GapReport
	if err := json.Unmarshal([]byte(gapsRes.Output), &report); err != nil {
		t.Fatalf("unmarshaling gap report: %v\n%s", err, gapsRes.Output)
	}
	if len(report.OrphanedAtoms) != 1 || report.OrphanedAtoms[0] != "mech_zzfix_orphan" {
		t.Errorf("expected exactly [mech_zzfix_orphan], got %v", report.OrphanedAtoms)
	}
	for _, id := range []string{"req_zzfix_alpha", "req_zzfix_draft", "api_zzfix_beta", "mech_zzfix_gamma", "rule_zzfix_untested"} {
		for _, orphan := range report.OrphanedAtoms {
			if orphan == id {
				t.Errorf("false positive: %s (part of the covered chain) reported as orphan", id)
			}
		}
	}

	// crawl's gaps=true detects missing IMPLEMENTATION only (IsOrphan,
	// pkg/exploration/orphan.go); a missing TEST-link is a distinct state
	// that check (and lint's IMPLEMENTATION-layer traceability-gap rule)
	// surface as NO_TESTS. Together these are the two facets of "gaps"
	// S8 asks for.
	checkRes := sb.Run(func() (string, error) {
		return runCoverageCheck("atom", "rule_zzfix_untested", "", sb.DocsDir, false, false, nil)
	})
	if checkRes.Err != nil {
		t.Fatalf("check --atom rule_zzfix_untested: %v", checkRes.Err)
	}
	rows := parseCheckReport(t, checkRes.Output)
	if len(rows) != 1 {
		t.Fatalf("expected exactly 1 row, got %d: %+v", len(rows), rows)
	}
	if rows[0].Impl == 0 || rows[0].Tests != 0 || rows[0].Status != "NO_TESTS" {
		t.Errorf("expected rule_zzfix_untested to show impl>0, tests=0, status NO_TESTS, got: %+v", rows[0])
	}
}

// ── S13: diff mode ────────────────────────────────────────────────────

// TestScenario_S13_DiffMode pins test_atd_07_26.md §4 S13: with the
// sandbox under git, an uncommitted source-file edit makes `check` (diff
// mode) flag exactly the touched atom; an uncommitted atom-file edit
// flags that atom's code-side coverage -- the bidirectional promise
// itself.
func TestScenario_S13_DiffMode(t *testing.T) {
	t.Run("code_change_flags_touched_atom", func(t *testing.T) {
		t.Parallel()
		sb := testutil.Sandbox(t, "fixture_project")
		sb.Git(t)

		betaPath := filepath.Join(sb.SrcDir, "beta.go")
		content, err := os.ReadFile(betaPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(betaPath, append(content, []byte("\n// uncommitted change for S13\n")...), 0644); err != nil {
			t.Fatal(err)
		}

		res := sb.Run(func() (string, error) {
			return runCoverageCheck("diff", "", "", sb.DocsDir, false, false, nil)
		})
		if res.Err != nil {
			t.Fatalf("check (diff mode): %v", res.Err)
		}
		rows := parseCheckReport(t, res.Output)
		if len(rows) != 1 || rows[0].AtomID != "api_zzfix_beta" {
			t.Errorf("expected diff-mode check to flag exactly api_zzfix_beta, got: %+v", rows)
		}
	})

	t.Run("atom_change_flags_the_code_side", func(t *testing.T) {
		t.Parallel()
		sb := testutil.Sandbox(t, "fixture_project")
		sb.Git(t)

		atomPath := filepath.Join(sb.DocsDir, "api_zzfix_beta.atom.md")
		content, err := os.ReadFile(atomPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(atomPath, append(content, []byte("\n<!-- uncommitted doc edit for S13 -->\n")...), 0644); err != nil {
			t.Fatal(err)
		}

		res := sb.Run(func() (string, error) {
			return runCoverageCheck("diff", "", "", sb.DocsDir, false, false, nil)
		})
		if res.Err != nil {
			t.Fatalf("check (diff mode): %v", res.Err)
		}
		rows := parseCheckReport(t, res.Output)
		if len(rows) != 1 || rows[0].AtomID != "api_zzfix_beta" {
			t.Fatalf("expected diff-mode check to flag api_zzfix_beta, got: %+v", rows)
		}
		if rows[0].Impl == 0 {
			t.Errorf("expected api_zzfix_beta's code-side coverage (impl>0) to be shown, got: %+v", rows[0])
		}
	})
}
