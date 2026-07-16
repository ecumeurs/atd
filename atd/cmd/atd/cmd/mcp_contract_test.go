package cmd

// MCP contract tests (test_atd_07_26.md §3.3, S11, WP-4). pkg/mcp.Registry
// is the right seam per §3.3: RegisterMCPTools(r) then r.Call(name, args) --
// no transport needed, and no LLM required in the default path. The fixture
// project (testdata/fixture_project/.atd) ships with zero configured LLM
// providers, so every LLM-backed tool below deterministically takes the
// "IDE fallback" path (pkg/ollama's ResolveProvider falls through to
// Resolution{IsIDE: true} when llmConfig.Providers is empty) rather than
// touching a network -- confirmed by reading pkg/ollama/provider.go,
// pkg/audit/audit.go, pkg/indexer/index.go and cmd/atd/cmd/{dissect,map}.go
// before writing these cases, not assumed.
//
// This file covers §3.3 #1 (wiring sweep) and #2 (schema honesty) plus the
// error-shape cases from #4. Tool-set <-> docs equality (§3.3 #3) lives in
// mcp_docs_equality_test.go since it reads the REAL repo docs/, not the
// fixture, and deserves its own file-level doc comment about that.
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/mcp"
	"atd-tools/pkg/testutil"
)

// ── #1: wiring sweep ─────────────────────────────────────────────────────

// wiringCase is one entry in the table every registered tool must appear
// in exactly once (test_atd_07_26.md §3.3 #1). Args returns the minimal
// valid arguments to exercise the tool once against a fresh, git-initialized
// sandbox; llmBacked relaxes the assertion from "must not error" to "must
// not panic and must not be an empty success" (a graceful provider-
// unavailable error is fine); workspace switches the fixture to
// fixture_workspace for the three workspace_* tools, which error immediately
// against a workspace-less project.
type wiringCase struct {
	llmBacked bool
	workspace bool
	args      func(sb *testutil.SB) map[string]any
}

var wiringCases = map[string]wiringCase{
	"atd_query":  {args: func(sb *testutil.SB) map[string]any { return map[string]any{"search": "zzfix"} }},
	"atd_crawl":  {args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_weave":  {args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_update": {args: func(sb *testutil.SB) map[string]any {
		// req_zzfix_draft is DRAFT, not STABLE -- avoids the S4 governance
		// guard so this exercises pure wiring, not the (separately tested)
		// business rule.
		return map[string]any{
			"file": filepath.Join(sb.DocsDir, "req_zzfix_draft.atom.md"),
			"set":  []any{"priority=3"},
		}
	}},
	"atd_roadmap": {args: func(sb *testutil.SB) map[string]any {
		// "dir" is the only required param; deliberately NOT passing "out"
		// exercises the true MCP-caller default (relative "roadmap.json"),
		// which is why the wiring sweep chdirs into sb.Root for every case
		// (see TestMCPContract_WiringSweep) -- otherwise this would write a
		// stray roadmap.json into this package's own source directory,
		// exactly the class of debris incident I-1 (test_atd_07_26.md §2.1)
		// warns about.
		return map[string]any{"dir": sb.SrcDir}
	}},
	"atd_stats": {args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_check": {args: func(sb *testutil.SB) map[string]any { return map[string]any{"full": true} }},
	"atd_assemble": {args: func(sb *testutil.SB) map[string]any {
		return map[string]any{"starts": "req_zzfix_alpha"}
	}},
	"atd_trace":      {args: func(sb *testutil.SB) map[string]any { return map[string]any{"atom": "req_zzfix_alpha"} }},
	"atd_test_links": {args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_dissect": {llmBacked: true, args: func(sb *testutil.SB) map[string]any {
		return map[string]any{"file": filepath.Join(sb.SrcDir, "beta.go")}
	}},
	"atd_index":  {llmBacked: true, args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_search": {args: func(sb *testutil.SB) map[string]any { return map[string]any{"grep": "zzfix"} }},
	"atd_audit":  {llmBacked: true, args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_recon": {llmBacked: true, args: func(sb *testutil.SB) map[string]any {
		return map[string]any{"atom": "api_zzfix_beta", "candidate": filepath.Join(sb.SrcDir, "beta.go")}
	}},
	"atd_map": {llmBacked: true, args: func(sb *testutil.SB) map[string]any {
		return map[string]any{"file": filepath.Join(sb.SrcDir, "beta.go")}
	}},
	"atd_env":            {args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_config":         {args: func(sb *testutil.SB) map[string]any { return map[string]any{"list": true} }},
	"atd_lint":           {args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_workspace_list": {workspace: true, args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_workspace_use": {workspace: true, args: func(sb *testutil.SB) map[string]any {
		return map[string]any{"project": "zzfix_a"}
	}},
	"atd_workspace_stats": {workspace: true, args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
	"atd_heatmap":         {args: func(sb *testutil.SB) map[string]any { return map[string]any{"atom": "req_zzfix_alpha"} }},
	"atd_heatmap_code": {args: func(sb *testutil.SB) map[string]any {
		return map[string]any{"file": filepath.Join(sb.SrcDir, "beta.go")}
	}},
	"atd_heatmap_project": {args: func(sb *testutil.SB) map[string]any { return map[string]any{} }},
}

// TestMCPContract_WiringSweep pins test_atd_07_26.md §3.3 #1 / S11: every
// tool RegisterMCPTools registers must be callable with minimal valid args
// against a real (if disposable) project -- no "unknown tool", no panic,
// and a non-error result for deterministic tools. It also enforces the two
// halves of the drift-catching AC in §6 WP-4: removing a Register call
// makes the reverse-direction loop below fail (a wiringCases entry with no
// matching registered tool), and adding a new Register call without adding
// a wiringCases entry fails immediately (the forward loop has nothing to
// call it with).
func TestMCPContract_WiringSweep(t *testing.T) {
	r := mcp.NewRegistry()
	RegisterMCPTools(r)

	seen := make(map[string]bool, len(wiringCases))

	for _, tool := range r.List() {
		name := tool.Name
		tc, ok := wiringCases[name]
		if !ok {
			t.Errorf("registered tool %q has no entry in wiringCases -- add one (this is exactly the drift E5/S11 exists to catch: a tool registered but never exercised)", name)
			continue
		}
		seen[name] = true

		t.Run(name, func(t *testing.T) {
			fixture := "fixture_project"
			if tc.workspace {
				fixture = "fixture_workspace"
			}
			sb := testutil.Sandbox(t, fixture)
			sb.Git(t)

			args := tc.args(sb)

			// Several handlers (roadmap's default "out", crawl's hardcoded
			// "." src) resolve relative paths against the process cwd, not
			// against config.ProjectRoot() -- exactly the cwd-anchoring
			// hazard incident I-1 (test_atd_07_26.md §2.1) documents for
			// product code. A real MCP server's cwd is the project root, so
			// chdir into the sandbox for the duration of the call: it makes
			// the test both safe (no relative writes escape into this
			// package's own source directory) and realistic.
			restoreCwd := chdirT(t, sb.Root)
			defer restoreCwd()

			var res testutil.Result
			func() {
				defer func() {
					if p := recover(); p != nil {
						t.Fatalf("tool %q panicked on args %v: %v", name, args, p)
					}
				}()
				res = sb.Run(func() (string, error) { return r.Call(name, args) })
			}()

			if res.Err != nil && strings.Contains(res.Err.Error(), "unknown tool") {
				t.Fatalf("registry reports %q as unknown despite being listed by r.List(): %v", name, res.Err)
			}

			if tc.llmBacked {
				// Graceful provider-unavailable error OR a non-empty
				// delegated/salvaged success -- panic is already excluded
				// above; the one shape that must never happen is a
				// *silent, empty* success.
				if res.Err == nil && strings.TrimSpace(res.Output) == "" {
					t.Errorf("LLM-backed tool %q returned an empty success (no error, no output)", name)
				}
				return
			}

			if res.Err != nil {
				t.Errorf("deterministic tool %q errored on minimal valid args %v: %v", name, args, res.Err)
			}
		})
	}

	for name := range wiringCases {
		if !seen[name] {
			t.Errorf("wiringCases has an entry for %q but RegisterMCPTools no longer registers a tool with that name -- tool removed? update wiringCases", name)
		}
	}
}

// chdirT temporarily changes the process working directory to dir and
// returns a func restoring the original one. Not t.Parallel()-safe (cwd is
// process-global) -- TestMCPContract_WiringSweep deliberately does not call
// t.Parallel() on its subtests for this reason.
func chdirT(t *testing.T, dir string) func() {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("chdirT: os.Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdirT: os.Chdir(%s): %v", dir, err)
	}
	return func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("chdirT: restoring cwd to %s: %v", prev, err)
		}
	}
}
