package cmd

// Schema-honesty and error-shape tests (test_atd_07_26.md §3.3 #2 and #4).
//
// Before writing paramCases below, every one of the 25 registered tools'
// InputSchema.properties was manually cross-checked against its handler
// closure in cmd/atd/cmd/mcp_tools.go (grepping for argString/argBool/
// argInt/direct args[...] reads of each declared key). Exactly one
// declared-but-never-read parameter was found across the entire registry:
// atd_check's "line" -- the report's own "known liar" (test_atd_07_26.md
// §3.3 #2, §1 item 3). Every other declared property IS read somewhere in
// its handler. paramCases below pins that one true defect plus a
// representative spot-check of live params across several other tools, to
// demonstrate the with/without method generalizes rather than re-deriving
// the same fact 60 times.
import (
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/mcp"
	"atd-tools/pkg/testutil"
)

// paramCase exercises one (tool, param) pair by calling the handler once
// with `without` args and once with `with` args, then comparing the two
// Results. If knownDefect is empty, the two calls must differ (proving the
// param has an effect); if knownDefect is set, they must be IDENTICAL
// (proving the param is dead), and knownDefect records why -- mirroring the
// "KNOWN DEFECT" house style already established in scenario_test.go's S1/S4.
type paramCase struct {
	tool        string
	param       string
	workspace   bool
	without     func(sb *testutil.SB) map[string]any
	with        func(sb *testutil.SB) map[string]any
	knownDefect string
}

var paramCases = []paramCase{
	{
		// The report's headline example: atd_check declares "line" (requires
		// "file") but runCoverageCheck's signature has no line parameter at
		// all -- the MCP handler never extracts args["line"], so any value
		// is silently discarded.
		tool:  "atd_check",
		param: "line",
		without: func(sb *testutil.SB) map[string]any {
			return map[string]any{"file": filepath.Join(sb.SrcDir, "beta.go")}
		},
		with: func(sb *testutil.SB) map[string]any {
			return map[string]any{"file": filepath.Join(sb.SrcDir, "beta.go"), "line": 999999}
		},
		knownDefect: "KNOWN DEFECT (test_atd_07_26.md §3.3 #2, §1 item 3): atd_check's \"line\" schema param is never read by the handler (mcp_tools.go's atd_check closure extracts base/target/full/file/semantic only) or passed into runCoverageCheck, which has no line parameter in its signature at all. An out-of-range line (999999) produces byte-identical output to omitting it entirely.",
	},
	{
		tool:  "atd_query",
		param: "paths_only",
		without: func(sb *testutil.SB) map[string]any {
			return map[string]any{"search": "zzfix"}
		},
		with: func(sb *testutil.SB) map[string]any {
			return map[string]any{"search": "zzfix", "paths_only": true}
		},
	},
	{
		tool:    "atd_crawl",
		param:   "gaps",
		without: func(sb *testutil.SB) map[string]any { return map[string]any{} },
		with:    func(sb *testutil.SB) map[string]any { return map[string]any{"gaps": true} },
	},
	{
		tool:  "atd_trace",
		param: "summary",
		without: func(sb *testutil.SB) map[string]any {
			return map[string]any{"atom": "req_zzfix_alpha"}
		},
		with: func(sb *testutil.SB) map[string]any {
			return map[string]any{"atom": "req_zzfix_alpha", "summary": true}
		},
	},
	{
		tool:  "atd_search",
		param: "paths_only",
		without: func(sb *testutil.SB) map[string]any {
			return map[string]any{"grep": "zzfix"}
		},
		with: func(sb *testutil.SB) map[string]any {
			return map[string]any{"grep": "zzfix", "paths_only": true}
		},
	},
	{
		tool:  "atd_heatmap_project",
		param: "layer",
		without: func(sb *testutil.SB) map[string]any {
			return map[string]any{}
		},
		with: func(sb *testutil.SB) map[string]any {
			return map[string]any{"layer": "dependency"}
		},
	},
	{
		tool:  "atd_config",
		param: "list",
		without: func(sb *testutil.SB) map[string]any {
			return map[string]any{}
		},
		with: func(sb *testutil.SB) map[string]any {
			return map[string]any{"list": true}
		},
	},
}

// TestMCPContract_SchemaHonesty pins test_atd_07_26.md §3.3 #2: every
// declared InputSchema param either measurably changes the tool's output
// (asserted here) or is a documented KNOWN DEFECT (atd_check's "line").
// Choice recorded for the AC report: "line" is pinned rather than trivially
// deleted from the schema, matching WP-2's established house style
// (scenario_test.go's S4 pins atd_update's missing "force" the same way)
// and keeping this WP's deliverable confined to the new test file rather
// than reaching into mcp_tools.go.
func TestMCPContract_SchemaHonesty(t *testing.T) {
	r := mcp.NewRegistry()
	RegisterMCPTools(r)

	for _, pc := range paramCases {
		pc := pc
		t.Run(pc.tool+"/"+pc.param, func(t *testing.T) {
			fixture := "fixture_project"
			if pc.workspace {
				fixture = "fixture_workspace"
			}
			sb := testutil.Sandbox(t, fixture)
			sb.Git(t)
			// See TestMCPContract_WiringSweep's chdirT comment: some
			// handlers (e.g. atd_trace hardcodes srcPath=".") resolve
			// relative to the process cwd, not config.ProjectRoot().
			defer chdirT(t, sb.Root)()

			without := sb.Run(func() (string, error) { return r.Call(pc.tool, pc.without(sb)) })
			with := sb.Run(func() (string, error) { return r.Call(pc.tool, pc.with(sb)) })

			differs := without.Output != with.Output || (without.Err == nil) != (with.Err == nil) ||
				(without.Err != nil && with.Err != nil && without.Err.Error() != with.Err.Error())

			if pc.knownDefect != "" {
				if differs {
					t.Errorf("%s\n\nexpectation changed: %s/%s now DOES differ with/without the param -- if the handler was fixed, remove this pin and move the case to a live-param assertion", pc.knownDefect, pc.tool, pc.param)
				}
				return
			}

			if !differs {
				t.Errorf("%s's %q param is declared in InputSchema but produced byte-identical output with and without it -- either wire it into the handler or pin it as a KNOWN DEFECT (see atd_check/line above for the pattern)\nwithout: err=%v output=%q\nwith:    err=%v output=%q",
					pc.tool, pc.param, without.Err, without.Output, with.Err, with.Output)
			}
		})
	}
}

// ── #4: error-shape tests ────────────────────────────────────────────────

// TestMCPContract_ErrorShapes pins test_atd_07_26.md §3.3 #4 / S11: unknown
// atom id, missing required arg, and wrong arg type must each be loud and
// actionable (never an empty/silent success), exercised through the MCP
// path (r.Call), not the underlying run* functions directly (those are
// already covered per-tool by scenario_test.go's S1).
func TestMCPContract_ErrorShapes(t *testing.T) {
	r := mcp.NewRegistry()
	RegisterMCPTools(r)

	t.Run("unknown_tool_name", func(t *testing.T) {
		sb := testutil.Sandbox(t, "fixture_project")
		res := sb.Run(func() (string, error) { return r.Call("atd_this_tool_does_not_exist", map[string]any{}) })
		if res.Err == nil {
			t.Fatal("expected an error calling an unregistered tool name, got success")
		}
		if !strings.Contains(res.Err.Error(), "atd_this_tool_does_not_exist") {
			t.Errorf("expected the error to name the unknown tool, got: %v", res.Err)
		}
	})

	t.Run("unknown_atom_id_via_trace", func(t *testing.T) {
		sb := testutil.Sandbox(t, "fixture_project")
		defer chdirT(t, sb.Root)()
		const nonsense = "zzfix_totally_bogus_atom_id_xyz"
		res := sb.Run(func() (string, error) { return r.Call("atd_trace", map[string]any{"atom": nonsense}) })
		if res.Err == nil {
			t.Fatal("expected a loud error for a nonsense atom id via atd_trace, got none")
		}
		if !strings.Contains(res.Err.Error(), nonsense) {
			t.Errorf("expected atd_trace's error to name the input id %q, got: %v", nonsense, res.Err)
		}
	})

	t.Run("unknown_atom_id_via_test_links", func(t *testing.T) {
		sb := testutil.Sandbox(t, "fixture_project")
		const nonsense = "zzfix_totally_bogus_atom_id_xyz"
		res := sb.Run(func() (string, error) { return r.Call("atd_test_links", map[string]any{"atom": nonsense}) })
		if res.Err == nil {
			t.Fatal("expected a loud error for a nonsense atom id via atd_test_links, got none")
		}
		if !strings.Contains(res.Err.Error(), nonsense) {
			t.Errorf("expected atd_test_links' error to name the input id %q, got: %v", nonsense, res.Err)
		}
	})

	t.Run("missing_required_arg_atom_trace", func(t *testing.T) {
		sb := testutil.Sandbox(t, "fixture_project")
		defer chdirT(t, sb.Root)()
		res := sb.Run(func() (string, error) { return r.Call("atd_trace", map[string]any{}) })
		if res.Err == nil {
			t.Fatal("expected atd_trace called with no \"atom\" to error, got success")
		}
		if !strings.Contains(strings.ToLower(res.Err.Error()), "atom") {
			t.Errorf("expected the error to mention the missing \"atom\" argument, got: %v", res.Err)
		}
	})

	// KNOWN DEFECT: atd_recon's InputSchema declares `"required": ["atom",
	// "candidate"]`, but the registry (pkg/mcp.Registry.Call) never
	// validates "required" against the schema -- it is purely descriptive.
	// The atd_recon handler (mcp_tools.go) calls runMap(candidate, atom,
	// docsDir, false); when atom=="" and isNew==false, runMap's switch falls
	// through to its DEFAULT branch (runMapDiscover) instead of the confirm
	// branch, silently changing MODE rather than refusing the call. This is
	// the "silent empty/wrong success" shape §7.2 "resolve or shout" warns
	// against, on the write^H^H^Hread side: a caller who forgot "atom"
	// wanted a confirm-mode verdict and got a discover-mode recommendation
	// list instead, with no error at all.
	t.Run("missing_required_arg_atom_recon_silently_changes_mode", func(t *testing.T) {
		sb := testutil.Sandbox(t, "fixture_project")
		defer chdirT(t, sb.Root)()
		res := sb.Run(func() (string, error) {
			return r.Call("atd_recon", map[string]any{
				"candidate": filepath.Join(sb.SrcDir, "beta.go"),
			})
		})
		if res.Err != nil {
			t.Errorf("KNOWN DEFECT expectation changed: atd_recon called without required \"atom\" now errors (%v) -- if the registry or handler now validates required args, this pin should be relaxed to assert an error, and the wiring case's minimal-args assumption should be revisited", res.Err)
		}
	})

	// KNOWN DEFECT: the argBool/argString/argInt helpers in mcp_tools.go
	// silently fall back to their default on a type mismatch instead of
	// erroring -- there is no JSON-arg type validation anywhere on the MCP
	// path. Demonstrated on atd_check's "full": a string "true" is not a
	// bool, so argBool's `v.(bool)` type assertion fails and `full` is
	// silently treated as false (diff mode) rather than rejected or
	// coerced -- producing a materially different report than the intended
	// full-project audit, with no error surfaced anywhere.
	t.Run("wrong_arg_type_full_string_instead_of_bool", func(t *testing.T) {
		sb := testutil.Sandbox(t, "fixture_project")
		sb.Git(t)
		defer chdirT(t, sb.Root)()

		boolTrue := sb.Run(func() (string, error) {
			return r.Call("atd_check", map[string]any{"full": true})
		})
		stringTrue := sb.Run(func() (string, error) {
			return r.Call("atd_check", map[string]any{"full": "true"})
		})
		if boolTrue.Err != nil {
			t.Fatalf("atd_check(full=true) errored: %v", boolTrue.Err)
		}
		if stringTrue.Err != nil {
			t.Fatalf("KNOWN DEFECT expectation changed: atd_check(full=\"true\") now errors (%v) -- if wrong-type args are now rejected, this pin should assert that error instead", stringTrue.Err)
		}
		if boolTrue.Output == stringTrue.Output {
			t.Errorf("KNOWN DEFECT expectation changed: atd_check(full=\"true\") (string) now produces the same output as atd_check(full=true) (bool) -- if argBool now coerces string booleans, update this pin")
		}
	})
}
