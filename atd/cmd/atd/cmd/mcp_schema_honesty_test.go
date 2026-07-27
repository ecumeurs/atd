package cmd

// Schema-honesty and error-shape tests (test_atd_07_26.md §3.3 #2 and #4).
//
// Before writing paramCases below, every one of the registered tools'
// InputSchema.properties was manually cross-checked against its handler
// closure in cmd/atd/cmd/mcp_tools.go (grepping for mcpArgs.String/Bool/
// Int/Float64/direct args[...] reads of each declared key). The one
// declared-but-never-read parameter this originally found -- atd_check's
// "line", the report's own "known liar" (test_atd_07_26.md §3.3 #2, §1 item
// 3) -- has since been removed from the schema entirely (§8.3 item 4c):
// runCoverageCheck has no line-scoping concept to wire it to, so "simpler,
// preferred" won over inventing one from scratch. Every currently-declared
// property IS read somewhere in its handler. paramCases below is a
// representative spot-check of live params across several tools, to
// demonstrate the with/without method generalizes rather than re-deriving
// the same fact for every one of the registered tools.
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
// declared InputSchema param must measurably change the tool's output
// (asserted here), or be a documented KNOWN DEFECT. As of §8.3's defect
// pass there is no longer a standing declared-but-dead param anywhere in
// the registry -- atd_check's "line" (the report's original "known liar")
// was removed from the schema rather than wired to a nonexistent concept
// in runCoverageCheck (item 4c); every remaining paramCase below asserts a
// live, wired param.
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
				t.Errorf("%s's %q param is declared in InputSchema but produced byte-identical output with and without it -- either wire it into the handler or pin it as a KNOWN DEFECT (set paramCase.knownDefect explaining why)\nwithout: err=%v output=%q\nwith:    err=%v output=%q",
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

	// atd_recon's InputSchema declares `"required": ["atom", "candidate"]`.
	// This used to be purely descriptive: the registry never validated
	// "required" against the schema, and the handler's runMap(candidate,
	// atom, docsDir, false) call would fall through to the DEFAULT
	// (discover) branch on an empty atom instead of refusing the call --
	// the "silent empty/wrong success" shape §7.2 "resolve or shout" warns
	// against, on the write^H^H^Hread side. pkg/mcp.Registry.Call now
	// enforces "required" centrally (test_atd_07_26.md §8.3 item 4a) before
	// any handler runs, so a caller who forgets "atom" is refused loudly,
	// naming both the tool and the missing argument, rather than silently
	// getting a discover-mode recommendation list instead of the confirm-
	// mode verdict they asked for.
	t.Run("missing_required_arg_atom_recon_refused", func(t *testing.T) {
		sb := testutil.Sandbox(t, "fixture_project")
		defer chdirT(t, sb.Root)()
		res := sb.Run(func() (string, error) {
			return r.Call("atd_recon", map[string]any{
				"candidate": filepath.Join(sb.SrcDir, "beta.go"),
			})
		})
		if res.Err == nil {
			t.Fatal("expected atd_recon called without required \"atom\" to be refused, got success")
		}
		if !strings.Contains(res.Err.Error(), "atd_recon") {
			t.Errorf("expected the error to name the tool \"atd_recon\", got: %v", res.Err)
		}
		if !strings.Contains(strings.ToLower(res.Err.Error()), "atom") {
			t.Errorf("expected the error to mention the missing \"atom\" argument, got: %v", res.Err)
		}
	})

	// mcpArgs (mcp_tools.go) used to silently fall back to each accessor's
	// default on a type mismatch instead of erroring -- there was no
	// JSON-arg type validation anywhere on the MCP path. Demonstrated here
	// on atd_check's "full": a string "true" is not a bool, so it is now a
	// loud, named type error instead of being silently treated as false
	// (diff mode) -- the KNOWN DEFECT this pinned (test_atd_07_26.md §8.3
	// item 4b) is fixed.
	t.Run("wrong_arg_type_full_string_instead_of_bool_rejected", func(t *testing.T) {
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
		if stringTrue.Err == nil {
			t.Fatal("expected atd_check(full=\"true\") (string, not bool) to be rejected with a loud type error, got success")
		}
		if !strings.Contains(stringTrue.Err.Error(), "full") {
			t.Errorf("expected the type error to name the offending argument \"full\", got: %v", stringTrue.Err)
		}
	})
}
