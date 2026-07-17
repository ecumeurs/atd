# ATD Toolkit — Test Strategy Report

**Date:** 2026-07-26
**Scope:** The ATD toolkit itself — `atd/` (Go: `pkg/`, `config/`, `cmd/atd/`), the CLI surface, and the MCP surface. **WebUI and the VSCode extension are explicitly out of scope.**
**Companion reports:** `code_quality_07_26.md` (refactor execution) and `investigation_atd_07_26.md` (agent-usability defects). Both generated large pending changes; the friction of verifying those changes is itself primary evidence for this report.
**Method:** Full inventory of existing `*_test.go` files, per-package `go test -cover` runs on both modules, and a post-mortem of the defects that shipped *despite* a green test suite.

---

## 0. TL;DR

> ✅ **Status update (2026-07-17): WP-0 through WP-7 have been executed and merged to `main`** — see the execution record in §8. The warning below is preserved for historical context; the sandbox contract landed first (commit `ca641eb`) and incident I-1 is closed.

> ⚠️ **Warning (historical, resolved by WP-0) — the suite was not sandboxed and had already damaged the repo.** A full `go test ./...` run in `cmd/atd` could **rewrite its own test source files** (reproduced; root cause in §2.1: leaked global config + cwd fallback + `UpdateLinks`' repo-walk), and one such self-mutation was committed to `main` unnoticed. Until the sandbox contract (§3.6) landed, any unexplained working-tree diff after a test run had to be treated as suite-inflicted.

1. **The suite is green but shallow.** ~2 450 test LOC against ~13 000 production LOC, and the tests that exist are mostly flag-existence checks, string-classification micro-tests, and happy-path parses. Packages at the heart of ATD's value proposition are the least tested: `pkg/exploration` (the graph engine) sits at **10.7 %** coverage, `pkg/coverage` (the `check` engine) at **23 %**, and `pkg/audit`, `pkg/indexer`, `pkg/llmservice`, `pkg/mcp`, `pkg/chat` have **zero** test files.
2. **The bug classes that actually hurt were all integration-level** — bare-id silent zeros, `trace`/`check` disagreement, the phantom-tool drift, the H3 parser swallow, the MCP handlers referencing deleted functions. Not one of them was catchable by the current unit tests, and not one *would be* caught by adding more of the same kind of unit test. The missing layer is **scenario tests against a fixture corpus** and **contract tests on the MCP surface**.
3. **The two-module layout is a standing trap.** `go build ./...` at `atd/` does **not** compile `cmd/atd/` (separate `go.mod`). This session, HEAD was "verified green" while the nested module had five compile errors. Every verification step — local, hook, future CI — must run **both** modules, always.
4. **ATD's own docs corpus is the best integration test it will ever have, and it is not wired in.** Running `atd lint` / `atd check --full` against `docs/` as a regression gate ("dogfood gate") would have caught the phantom atoms, the layer drift, and the parser bug years before a human did.
5. **The LLM boundary is already mockable — the codebase just doesn't use the seam.** `ollama.ListModels` is a swappable package var and `chat.Provider` is an interface; a deterministic fake provider unlocks testing of `audit`, `dissect`, `search`, `map`, `recon` logic without a model. Live-model tests should exist but be opt-in (env-gated), never in the default path.

**Recommended order:** sandbox the suite (WP-0 — non-negotiable first step) → two-module `verify` gate → fixture corpus + CLI scenario harness → MCP contract tests + dogfood gate in the pre-commit hook/CI → fake LLM provider → targeted unit back-fill on `exploration`/`coverage`. Orchestration-ready work packages with acceptance criteria in §6.

---

## 1. Where Testing Stands Today

### 1.1 Inventory (both modules, 2026-07-26)

| Area | Package | Coverage | Test character |
|---|---|---|---|
| Graph engine | `pkg/exploration` | **10.7 %** | 2 tests (workspace ref resolution, workspace search/assemble). Nothing on trace, crawl, orphan detection, spec-link extraction. |
| Check engine | `pkg/coverage` | **23.2 %** | 3 tests: status enum, report formatting, git-diff path rebasing. Nothing on the resolver path, semantic mode, full mode. |
| Atom I/O | `pkg/atom` | 30.2 % | Parse happy path + H3 regression + update basics. No BuildContent round-trip, no malformed-frontmatter cases. |
| Config / workspace | `config`, `pkg/workspace` | 39–41 % | Decent: multi-project detection, malformed config, resolver. |
| Storage | `pkg/store` | **77 %** | Good — the refactor shipped with real tests. The model to copy. |
| Provider resolution | `pkg/ollama` | 44.3 % | Good pattern: swaps the `ListModels` seam, no network. |
| Pipeline / prompt | `pkg/pipeline`, `pkg/prompt` | 84.5 % / 41 % | Task-list parsing well covered; prompt schemas smoke-checked. |
| **Zero tests** | `pkg/audit`, `pkg/indexer`, `pkg/llmservice`, `pkg/mcp`, `pkg/chat` | **0 %** | The entire LLM-backed half of the product plus the MCP transport. |
| CLI layer | `cmd/atd/cmd` (nested module) | 23.2 % | ~16 files, mostly "flag exists" + input-validation checks. One real integration test (`workspace_integration_test.go`) and one real regression test (trace traversal). |

Totals: **~12 980 prod LOC / ~2 454 test LOC (≈19 %)**. `go test -race` passes on the packages that matter (store, exploration) — the race flag is free to adopt.

### 1.2 What the current tests are good at

- Pure functions with obvious tables (cosine, pipeline, status enums).
- Config edge cases (the workspace tests are genuinely defensive).
- Post-incident regressions — the best tests in the repo (`TestDiffChangedFilesRebasesOntoProjectRoot`, `TestParseSectionWithSubheadings`, `TestTraceRecursiveTraversal`) each encode one shipped bug. That instinct is right; it just fires *after* the field report instead of before.

### 1.3 What they are structurally blind to

- **Cross-tool consistency** (`trace` vs `check` vs `query` on the same atom) — no test compares two tools' answers.
- **The CLI as users invoke it** — nothing executes the built binary against a real project layout end-to-end (except one workspace test).
- **The MCP surface** — zero tests. Schema/handler drift (`atd_check`'s `line` param is declared but never read), handler wiring (this week: three handlers calling deleted functions), and JSON-arg coercion are all unchecked.
- **The docs corpus itself** — lint/check are never run against `docs/` in any automated path.
- **LLM-path *logic*** — prompt assembly, response parsing, fallback routing: all reachable deterministically, none tested.

---

## 2. Why Bugs Escape — Six Observed Escape Paths

Each of these is a real, dated incident. They define the requirements for §3.

| # | Incident | Why the suite was blind | Test layer that would have caught it |
|---|---|---|---|
| E1 | **Bare-id silent zeros** (`atd_check`/`test_links`/`trace` returned confident 0 for unprefixed ids; ~15/55 scenarios in the 07-07 field report) | No test called two entry points with the same real-world input and compared; no fixture corpus with realistic ids | Fixture-corpus scenario tests (§3.2) + resolver contract test |
| E2 | **`trace`/`check`/`query` disagreement** on one atom's coverage | Same — consistency across tools was never asserted anywhere | Cross-tool consistency suite (§4, S3) |
| E3 | **Phantom tools** (`atd_discover` in rules/atoms, STABLE atoms with wrong schemas, stale `@spec-link` tags) | Nothing compares the registered tool set against the docs corpus | Dogfood gate (§3.4) + MCP registry contract test (§3.3) |
| E4 | **`atom.Parse` H3-swallow** (`###` subheadings silently dropped section content, starving `check --semantic`, `trace --summary`, `assemble`) | Parse tests used only minimal synthetic atoms; no test parsed a *real* atom from `docs/` | Corpus round-trip test: parse every atom in `docs/`, assert non-empty sections where the raw file has content (§4, S6) |
| E5 | **Nested module compile break** (refactor left `mcp_tools.go` calling `runCoverageCheck`/`runIndex`/`runFullAudit` after they were deleted; `ResolveProviderEx` renamed under `config.go`) | "Verified: `go build ./...` passes" was run at `atd/` only — the nested `cmd/atd` module is invisible from there | Two-module build matrix as a hard rule in every gate (§5.1) |
| E6 | **The suite rewrites its own source code** (`TestUpdateLinks` broken repeatedly, one broken fixture even committed to `main`) — full incident below | Tests run against **live global config** with a cwd fallback; no sandbox | The sandbox contract (§3.6) — the single most important item in this report |

The pattern: **every escape happened between components or between doc and code** — exactly the seams ATD exists to govern. The test strategy must mirror the product's own philosophy: the corpus and the code are one system; test them as one system.

### 2.1 ⚠️ Incident I-1 — the self-rewriting test suite (root-caused 2026-07-26)

For a day, `atd/cmd/atd/cmd/update_test.go` appeared to be edited by a "ghost": its fixture strings changed three times across sessions (`[[old_id]]` → `[[module_my_new_module]]` → `[[new_id]]`), each time leaving the test incoherent and failing, with no IDE or human touching the file. One mutated version was even **committed to `main`** unnoticed (in `45abc0f`). The ghost is the test suite itself. Reproduced deterministically; the chain has four links, each a lesson:

1. **A test leaks global config.** `trace_test.go:15` calls `config.LoadFromDir(t.TempDir())` and never restores `config.ActiveConfig` (several `map_test.go` paths leak too — probe-bisected: `TestMap` and `TestTrace` both arm the bug; `TestLint` does not).
2. **The config loader falls back to cwd.** The temp dir contains no `.atd`, so `LoadFromDirLegacy` climbs to `/`, finds nothing, and silently sets `loadedFromDir = os.Getwd()` (`config/config.go:286-289`). For a `go test` binary, **cwd is the package source directory** — `atd/cmd/atd/cmd/`.
3. **A product feature trusts that global.** `atom.UpdateLinks` (`pkg/atom/update.go:395-408`) performs its second walk — spec-link rename propagation across *source files* — over `config.ProjectRoot()`, i.e. over the test suite's own source directory. When `TestUpdateNamingConvention`/`TestUpdateLinks` later trigger renames, the walk rewrites every bracketed `[[old-id]]` literal inside `update_test.go` itself. (Only `[[…]]`-wrapped occurrences are rewritten — which is exactly why the *call arguments* survived while the *fixture strings* churned, producing incoherent tests.)
4. **The failure is displaced in time.** The already-compiled test binary is unaffected, so the mutating run is often green; the **next** compilation picks up the rewritten fixture and fails. Cause and symptom land in different sessions — which is why it read as a phantom concurrent editor.

Same disease, second symptom: `TestMapNewFlagProducesSkeleton` writes `pipeline_output/task_list.md` into the package source dir via the same cwd-anchored config, and a stale run's artifact is **committed** at `atd/cmd/atd/cmd/pipeline_output/` (dated April, `/tmp/Test…` paths inside).

**Why this matters beyond one flaky test:** a test suite that can silently modify — and has already silently committed modifications to — the repository it lives in has *negative* trust value: a red test no longer means the product is broken, and a green run no longer means nothing was damaged. Every recommendation in §3 assumes tests are deterministic functions of the checked-out tree; that assumption is currently false. **The sandbox contract (§3.6) is therefore not hygiene — it is the precondition for the rest of this report**, and the product-side scope guard (§3.6, P-1) also closes a real end-user hazard: any `atd update` rename executed with a mis-anchored cwd-fallback config will rewrite `[[refs]]` across whatever directory it happens to be anchored to.

---

## 3. Target Test Architecture — Four Layers + the LLM Boundary

### 3.1 Layer 1 — Unit tests (exists; back-fill selectively)

Keep the current style but aim it where the complexity is. Priority back-fill, in order:

1. **`pkg/exploration`** (10.7 %): `extractLinks` on every comment style (`//`, `#`, `<!-- -->`, block), `ResolveAtom` (bare id, `project:` prefix, type-prefixed, ambiguous, miss), crawl over a temp tree, orphan detection, trace depth/cycles.
2. **`pkg/coverage`**: `buildCoverageReport` per mode (diff/atom/file/full) against an in-memory explorer; dedup in file mode; the canonicalization path (the E1 fix must keep a test pinning it).
3. **`pkg/atom`**: `BuildContent` ⇄ `Parse` round-trip property (`Parse(BuildContent(a)) == a` for arbitrary field values); malformed frontmatter (no closing `---`, duplicate keys, CRLF); the STABLE+BUSINESS force guard (accept, refuse, `Force: true`).
4. **`pkg/audit` / `pkg/indexer` / `pkg/llmservice`**: unit-test the deterministic parts (chunking, cache keys, fallback selection, prompt assembly) once the fake provider (§3.5) exists.

*Not* a coverage-percentage crusade: `cmd/` flag tests and formatting tests have low value; don't add more of them.

### 3.2 Layer 2 — Fixture corpus + CLI scenario harness (the big missing piece)

Create `atd/testdata/fixture_project/` — a small but *complete* ATD project checked into the repo:

```
testdata/fixture_project/
├── .atd                      # real config
├── docs/
│   ├── contract_zzfix.atom.md    # 1 CONTRACT, 1 VISION (governance present)
│   ├── vision_zzfix.atom.md
│   ├── req_zzfix_alpha.atom.md       # BUSINESS
│   ├── api_zzfix_beta.atom.md        # ARCHITECTURE, parents: req
│   ├── mech_zzfix_gamma.atom.md      # IMPLEMENTATION, parents: api
│   ├── mech_zzfix_orphan.atom.md     # deliberate orphan (no parents)
│   └── rule_zzfix_untested.atom.md   # has @spec-link, no @test-link
└── src/
    ├── beta.go               # // @spec-link [[api_zzfix_beta]]
    ├── gamma.go              # @spec-link with ### -style sections upstream
    └── beta_test.go          # // @test-link [[api_zzfix_beta]]
```

Design rules for the fixture:

- **Every id carries the `zzfix` marker** so repo-wide rename/propagation tooling never collides with it (lesson E6/I-1).
- It deliberately contains **one instance of every state the tools must classify**: covered atom, impl-without-test, orphan, cross-layer chain, an atom using `###` subheadings, a `[[project:atom]]` cross-ref (add a second mini-project for workspace tests).
- A scenario harness (`cmd/atd/cmd/scenario_test.go` or a separate `atd/e2e/` package) copies the fixture to `t.TempDir()`, then drives the **CLI entry points** (the `run*`/pkg functions the commands delegate to — same functions the MCP handlers call, so one harness covers both surfaces) and asserts on structured output.
- **Golden files** for report-shaped output (`check --full`, `lint`, `stats`): store expected output in `testdata/golden/`, compare normalized (strip timestamps/paths), regenerate with an `-update` flag. This makes output regressions visible in diffs instead of invisible.

Harness API sketch (lives in a new `atd-tools/pkg/testutil`, importable from both modules):

```go
// Sandbox copies the named fixture into t.TempDir(), loads config from the
// copy, and registers cleanups that (a) restore the prior config.ActiveConfig
// and (b) run the tripwire (§3.6). Everything a test touches lives under Root.
func Sandbox(t *testing.T, fixture string) *SB   // SB{Root, DocsDir, ProjectDir string}

func (s *SB) Run(fn func() (string, error)) Result      // capture output+err of a run* entry point
func (s *SB) Golden(t *testing.T, name, got string)      // normalized golden-file compare, -update aware
func (s *SB) Git(t *testing.T)                           // git-init the sandbox for diff-mode scenarios
```

Scenario tests then read as: `s := testutil.Sandbox(t, "fixture_project")` → call the entry point → assert. No test constructs atoms inline unless the atom's *content* is the thing under test; everything else comes from the fixture, so realism accrues in one place.

This layer directly buys: E1 (bare id vs prefixed id both resolve to the same answer), E2 (see S3 below), E4 (real-shaped atoms), and a regression bed for every future field-report defect.

### 3.3 Layer 3 — MCP contract tests

`pkg/mcp.Registry` is already the right seam: `RegisterMCPTools(r)` then `r.Call(name, args)` — no transport needed. Add `cmd/atd/cmd/mcp_contract_test.go`:

1. **Wiring test:** register everything, iterate `r.List()`, call each tool with minimal valid args against the fixture project. Assert: no "unknown tool", no panic, and for deterministic tools a non-error result. *This single test catches E5-class breaks (handler referencing deleted code compiles fine only when it exists — but arg-shape drift and runtime nil-derefs it does catch) and any future rename that forgets a handler.*
2. **Schema honesty test:** for each tool, walk `InputSchema.properties` and assert every declared parameter is actually read by the handler (greppable via a table the test maintains, or by calling the handler twice — with and without the param — and asserting the output differs where it should). Start with the known liar: `atd_check`'s `line`. This makes E3-class "declared but dead" params fail loudly.
3. **Tool-set ↔ docs test (drift killer):** assert the set of registered tool names equals the set of `api_atd_serve_*` atoms in `docs/` (once D1's 8 missing atoms are authored, this pins graph completeness forever), and that each atom's documented param names ⊆ the registered schema's properties. This is the automated version of §4 of the investigation report — the test ATD's philosophy says should exist.
4. **Error-shape test:** unknown atom id, missing required arg, wrong type → assert the error message is loud and actionable (contains the id, suggests candidates), never an empty success. This pins the "resolve or shout" invariant (investigation §7.2).

### 3.4 Layer 4 — The dogfood gate

Run ATD on ATD, automatically:

- **Pre-commit hook extension:** the existing hook only checks staged atoms for missing parents. Extend it (or add a `make verify`) to run `atd lint` over `docs/` and fail on *new* errors relative to a checked-in baseline count (ratchet pattern — the ~90 pre-existing empty-section debts don't block, but regressions do).
- **CI job (once CI exists — none today):** `atd lint`, `atd check --full`, and the Layer-3 tool-set↔docs test against the real corpus on every push. This is simultaneously an integration test of the binary and the living proof-of-philosophy the investigation report called for.
- The 07-13/07-14 sessions found the H3 parser bug and the phantom atoms *by manually doing exactly this*. Automate what already proved itself.

### 3.5 The LLM boundary — deterministic fakes, opt-in live tests

The rule: **no test in the default path may require a model, network, or GPU.** Two mechanisms, both already half-built:

1. **`chat.Provider` fake** (`pkg/chat/chat.go:64` is an interface): a `fakeProvider` returning canned, schema-valid JSON per prompt-type. Inject via `llmservice` (post-refactor it accepts config/providers). This unlocks *logic* testing of `audit` (bloat/collision paths), `dissect` (boundary parsing), `map` (confirm/propose modes), `recon`, and `search` (embedding mode with a fake embedder returning fixed vectors — cosine math is already 100 % covered).
   - Test the **parsers hard**: feed the fake malformed LLM output (markdown-fenced JSON, trailing prose, wrong field types, the real-world `{"Confidence": 0, "Mismatches": "Yes, there are several…"}` case from the field report) and assert graceful, loud handling. LLM-response parsing is the single most failure-prone code in the product and currently has zero tests.
2. **`ollama.ListModels` var seam** (already used by `provider_test.go`): keep for resolution-order tests.
3. **Opt-in live suite:** `//go:build live` or `if os.Getenv("ATD_LIVE_LLM") == ""` { t.Skip } — a handful of smoke tests (index 3 files, semantic-search them, one `map` confirm) for humans with a running provider. Never in CI's default lane.
4. **Prompt snapshot tests:** golden-file the assembled prompts per task type (`pkg/prompt` + `llmservice/prompts.go`). Prompt drift silently changes product behavior; a diff in a golden file makes it a reviewed decision.

### 3.6 The sandbox contract (prerequisite — see incident §2.1)

Incident I-1 proved the suite can rewrite its own source. The fix has a test-side contract, a product-side guard, and a tripwire that keeps both honest.

**T-1 — Config isolation (test side).** `config.ActiveConfig` is a mutable global; every test that loads or mutates it MUST snapshot and restore it. Concretely:

- Add `config.Snapshot() Config` / `config.Restore(Config)` (or just document the `saved := config.ActiveConfig; defer func(){ config.ActiveConfig = saved }()` idiom `map_test.go:85` already uses) and apply it to **every** offender. Known offenders today: `trace_test.go:15` (naked `LoadFromDir`, no restore) and at least one `map_test.go` path (probe-bisected). Audit all `*_test.go` for `LoadFromDir|config.Load|ActiveConfig =` and fix each.
- `testutil.Sandbox` (§3.2) does this automatically via `t.Cleanup`, so the rule for new tests is simply: **touch atoms or config → use Sandbox.**

**T-2 — No cwd-anchored writes (test side).** Tests must never let product code resolve paths against the process cwd. The `Sandbox` helper always produces a sandbox root that *contains* a `.atd`, so the cwd fallback can never fire inside a sandboxed test. Delete the committed test debris at `atd/cmd/atd/cmd/pipeline_output/` and add it to `.gitignore` as a tombstone.

**P-1 — Scope guard in `UpdateLinks` (product side).** The second walk in `atom.UpdateLinks` (`update.go:395`) must not trust `ProjectRoot()` blindly:

- If `docsPath` is **not inside** `ProjectRoot()`, skip the source-file walk (a rename in project X must never rewrite files outside X). This alone makes the incident impossible regardless of test hygiene, because the tests pass temp-dir docs paths.
- Additionally, `LoadFromDirLegacy`'s cwd fallback should mark the config as fallback-anchored (e.g. `loadedFromFallback bool`), and destructive repo-wide operations (rename propagation, `fix`, future bulk edits) should refuse or warn when anchored by fallback rather than a real `.atd`. A loud refusal here is the same "resolve or shout" principle the investigation report demands for reads — applied to writes, where it matters more.

**P-2 — Skip test sources in the propagation walk (product side, cheap belt-and-braces).** The source-file walk should skip `_test.go` files and `testdata/` directories: fixture literals are not spec-links, and no legitimate `@spec-link` propagation target lives in a test fixture string.

**W-1 — The tripwire (keeps everyone honest).** After every test run in the gate (`make verify`, CI): `git status --porcelain` must be empty. Implement twice:

- In `testutil.Sandbox`'s cleanup for the local loop (cheap best-effort: fail the test if files under the *package source dir* changed mtime during the test), and
- As a hard CI step after the test job: `test -z "$(git status --porcelain)" || (git diff; exit 1)`. This converts any future I-1-class bug from "ghost editor haunting sessions for a day" into a red build with the diff printed.

Ordering: **W-1 and P-1 first** (they make the failure loud and impossible respectively), then T-1/T-2 cleanup, then P-2. All four are small; together they are WP-0 in §6.

---

## 4. Scenario Catalog (concrete, ordered by yield)

Each scenario is phrased as: setup → action → assertion. All run against the §3.2 fixture unless noted.

| ID | Scenario | Pins |
|---|---|---|
| S1 | For each of `check`, `test_links`, `trace`, `query`: call with (a) canonical id, (b) bare id, (c) `TYPE_`-prefixed id, (d) nonsense id. (a)–(c) return identical coverage; (d) errors loudly naming the input and suggesting near-matches. | E1, §7.2 "resolve or shout" |
| S2 | `check --file src/beta.go` lists each spec-link site exactly once (dedup). | field-report §3 |
| S3 | **Consistency oracle:** for every atom in the fixture, `check`, `trace` (raw), and `query` agree on impl-link count and test-link count. One loop, entire corpus. | E2 |
| S4 | STABLE+BUSINESS guard: `update` on `req_zzfix_alpha` (STABLE) → refused with actionable message; with `--force`/`Force:true` → succeeds; DRAFT atom → no friction. MCP path: same via `r.Call("atd_update", …)`. | ISS-010, this week's guard |
| S5 | Governance lint: corpus with 0 CONTRACT → error; 2 CONTRACTs → error; non-canonical type `FOO` → error naming the file; `SERVICE`/`USAGE` → **no** error (sanctioned union). | this week's lint rules |
| S6 | **Corpus parse audit** (runs on the *real* `docs/`, not the fixture): for every atom file, `Parse` yields non-empty `Intent`; and if the raw file contains a `## TECHNICAL INTERFACE` H2 with non-blank body lines, `Interface` is non-empty. | E4 — would have caught the H3 bug |
| S7 | Rename propagation: `update set id=…` on an atom with 2 inbound `[[refs]]` (one in frontmatter parents, one in body) → file renamed per convention, both refs rewritten, `lint` clean afterwards. | `UpdateLinks`, E6's underlying feature |
| S8 | Orphan/gaps: `crawl gaps=true` reports exactly `mech_zzfix_orphan` and `rule_zzfix_untested`'s missing test-link; no false positives on the covered chain. | investigation §5 |
| S9 | Workspace: two fixture projects; `workspace_use` A then query B's atom via `[[projB:atom]]` → resolves with correct project attribution; bare cross-project id → loud, disambiguating error. | cross-project model |
| S10 | Missing semantic index: `search` (semantic mode) with no index file → explicit "index missing, run atd_index" message, never a silent empty result; and the index db is created **outside** the docs tree. | field-report §7 |
| S11 | MCP wiring sweep + schema honesty + tool-set↔docs equality (§3.3 #1–#3). | E3, E5 |
| S12 | LLM-response robustness: fake provider returns fenced/truncated/prose-wrapped JSON to `audit`/`dissect`/`map` parsers → structured error or salvage, never a zero-value success. | §3.3 field-report class |
| S13 | Diff mode: fixture in a temp git repo; modify `src/beta.go` uncommitted → `check` (diff mode) flags exactly the touched atom; modify an atom file → flags the code side. | the bidirectional promise itself |

S1–S6 are the highest yield: they encode every defect class that has actually shipped.

---

## 5. Harness & Infrastructure

### 5.1 Kill the two-module trap (do this first, it's an hour)

- Add `atd/Makefile` (or a `verify.sh`) with the canonical gate, and make every human, agent, hook, and future CI use *only* it:

```make
verify:            ## the only definition of "green"
	cd $(ROOT) && go build ./... && go vet ./... && go test ./...
	cd $(ROOT)/cmd/atd && go build ./... && go vet ./... && go test ./...

verify-race:
	… same with -race …
```

- Consider `go.work` at `atd/` so `go build ./...` from the root sees both modules — that removes the trap at the language level. (Check the release process tolerates a `go.work` first; otherwise the Makefile is the guard.)
- Document in `CLAUDE.md`/agent rules: **"green" means `make verify`, nothing less.** Three of this week's incidents trace to partial verification.

### 5.2 CI (currently: none)

Single GitHub-Actions-style workflow, three jobs: `verify` (both modules, `-race`), `dogfood` (`atd lint` + `check --full` + tool-set↔docs on the real corpus, ratcheted), `fixture-e2e` (Layer 2 scenarios). No LLM job in the default lane; a manual/`live`-tagged job for the opt-in suite.

### 5.3 Coverage policy

No blanket threshold (it breeds flag-tests). Instead, per-package floors on the engines only — `exploration` ≥ 60 %, `coverage` ≥ 60 %, `atom` ≥ 70 % — enforced in CI once the §3.1 back-fill lands, ratcheted upward, never downward.

### 5.4 Fixture hygiene (lesson E6 / incident I-1)

- All fixture atom ids/strings carry a `zzfix` (or similar) marker no real tooling will ever rename.
- Prefer `testdata/` files over string literals inside `_test.go` for anything an ATD tool might rewrite — rename-propagation and find-replace passes skip `testdata/` by convention (and add that convention to the agent rules).
- When a test encodes a shipped bug, name it and comment it after the incident (`TestParseSectionWithSubheadings` is the house style — keep it).

### 5.5 Cheap extras worth taking

- `go vet` already in the gate; add `staticcheck` if appetite exists (the codebase is clean enough post-refactor that adoption is cheap now).
- **Fuzz** `atom.Parse` and `extractLinks` (`go test -fuzz`, native tooling): both consume arbitrary user files; the H3 bug is exactly the class fuzzing plus the S6 invariant would surface.
- `t.Parallel()` on the scenario harness — fixtures are copied per-test, so it's free speed.

---

## 6. Rollout — Orchestration-Ready Work Packages

Written to be executed by a coordinating agent delegating to subagents. Each WP lists scope, deliverables, acceptance criteria (AC), and dependencies. WPs marked ∥ can run in parallel once their dependencies are met. **Every WP finishes with `make verify` (WP-1) green and `git status --porcelain` empty** — no exceptions, that's the tripwire working.

### WP-0 — Sandbox the suite (FIRST; blocks everything) — §2.1, §3.6
**Scope:** `pkg/atom/update.go`, `config/config.go`, `cmd/atd/cmd/{trace,map}_test.go` (+ audit all `*_test.go`), new `pkg/testutil`, delete `cmd/atd/cmd/pipeline_output/`.
**Deliverables:** P-1 scope guard (`docsPath` outside `ProjectRoot` → skip source walk) + `loadedFromFallback` refusal for destructive ops; P-2 skip `_test.go`/`testdata/` in the propagation walk; T-1 config snapshot/restore in every offending test; T-2 debris deletion + `.gitignore`; W-1 tripwire (in `testutil` cleanup and as a `verify` step).
**AC:** (1) `go test -count=5 ./...` in `cmd/atd` leaves the tree byte-identical — run it in a loop, diff each time; (2) a deliberately re-introduced naked `LoadFromDir(t.TempDir())` no longer causes any repo write (proves P-1 alone suffices); (3) new regression test pins the scope guard.
**Size:** small (≤ ½ day). **Not parallelizable internally** — one agent, sequential.

### WP-1 — The `verify` gate — §5.1
**Scope:** `atd/Makefile` (or `verify.sh`), optional `go.work`, a note in `CLAUDE.md`/`.agent/rules`.
**Deliverables:** `make verify` = build+vet+test **both modules**; `make verify-race`; tripwire step from WP-0.
**AC:** running `make verify` from a tree with a compile error in *either* module fails; docs updated so "green" is defined as this target.
**Depends:** nothing (can land with WP-0). **Size:** ~1 h. ∥

### WP-2 — Fixture project + scenario harness — §3.2
**Scope:** `atd/testdata/fixture_project/` (+ a sibling mini-project for workspace scenarios), `pkg/testutil` (Sandbox/Run/Golden/Git), scenario tests S1–S5, S7, S8, S13.
**Deliverables:** the fixture corpus (every classifiable state, all ids `zzfix_`-marked), the harness API from §3.2, golden files under `testdata/golden/` with `-update` flag.
**AC:** each scenario S1–S5, S7, S8, S13 is one named test citing its scenario id; S1 asserts identical output across all four id forms *per tool*; S3 loops the whole fixture corpus; goldens regenerate deterministically (two consecutive `-update` runs produce no diff).
**Depends:** WP-0 (Sandbox), WP-1. **Size:** the big one, 1–2 days. Internally parallelizable: fixture+harness first, then scenarios can be split across agents by scenario id.

### WP-3 — Dogfood gate — §3.4, S6
**Scope:** pre-commit hook extension, S6 corpus parse audit as a Go test (build-tagged `dogfood` so it reads the real `docs/`), lint-error baseline ratchet file.
**Deliverables:** `make dogfood` = `atd lint` + `check --full` + S6 against the real corpus, failing on regressions vs the checked-in baseline; hook calls it on atom-touching commits.
**AC:** introducing a fresh lint error (e.g. a non-canonical type in a scratch atom) fails the gate; the ~90 pre-existing authoring debts do not.
**Depends:** WP-1. ∥ with WP-2. **Size:** ~½ day.

### WP-4 — MCP contract tests — §3.3, S11
**Scope:** new `cmd/atd/cmd/mcp_contract_test.go`.
**Deliverables:** wiring sweep (every registered tool callable against the fixture), schema-honesty test (start with `atd_check.line`), tool-set ↔ `api_atd_serve_*` atom equality, error-shape test (S11 + the loud-error checks from S1's (d) case).
**AC:** removing any tool registration, adding an undocumented schema param, or deleting a tool atom each turn the suite red. Known current gap: the 8 missing tool atoms — either author them in this WP (pairs with investigation follow-up D1) or start the equality test with an explicit, shrinking allowlist.
**Depends:** WP-0, WP-2 (fixture). ∥ with WP-3, WP-5. **Size:** ~1 day.

### WP-5 — CI — §5.2
**Scope:** one workflow file; jobs `verify` (-race), `dogfood`, `fixture-e2e`; manual `live` lane stub.
**AC:** all three jobs green on main; W-1 tripwire step present in each job; no job requires network/LLM.
**Depends:** WP-1, WP-3 (and picks up WP-2/WP-4 suites as they land). **Size:** ~½ day. ∥

### WP-6 — Fake LLM provider + parser robustness — §3.5, S12
**Scope:** `pkg/testutil/fakeprovider` implementing `chat.Provider` (+ fake embedder), tests for `audit`/`dissect`/`map`/`recon` response parsing, prompt goldens for `pkg/prompt` + `llmservice/prompts.go`.
**Deliverables:** canned-response fake with per-prompt-type fixtures including malformed variants (fenced JSON, trailing prose, wrong types, the real `{"Confidence": 0, "Mismatches": "Yes…"}` case); S12 tests; prompt snapshot goldens.
**AC:** all LLM-path logic tests run with no network (verify by running with `http_proxy=127.0.0.1:1` set); each malformed-input case yields a structured error or salvage, never a zero-value success.
**Depends:** WP-0, WP-2 (testutil). ∥ with WP-4. **Size:** 1–2 days.

### WP-7 — Unit back-fill + coverage floors — §3.1
**Scope:** `pkg/exploration` (extractLinks, ResolveAtom, crawl, orphans, trace), `pkg/coverage` (per-mode reports, dedup, canonicalization), `pkg/atom` (BuildContent⇄Parse round-trip, malformed frontmatter, force guard); then CI floors (exploration ≥ 60 %, coverage ≥ 60 %, atom ≥ 70 %).
**AC:** floors met and enforced in CI; round-trip property test present.
**Depends:** WP-0 (idioms), WP-5 (floor enforcement). Fully parallelizable by package. **Size:** ongoing; first pass ~1–2 days.

### WP-8 — Long tail (opportunistic)
Fuzz targets for `atom.Parse`/`extractLinks`; opt-in live-LLM smoke suite (`ATD_LIVE_LLM=1`); `staticcheck` adoption. **Depends:** WP-2. No AC beyond "exists and is documented".

**Critical path:** WP-0 → WP-2 → {WP-4, WP-6} with {WP-1, WP-3, WP-5} slotting in alongside. WP-0 is deliberately not parallel and not delegated lightly: it changes product behavior (`UpdateLinks`) and must be reviewed against S7's rename-propagation expectations so the guard doesn't break the legitimate feature.

---

## 7. Conclusion

ATD's failure history is unusually legible: every serious defect lived **on a seam** — between two tools' answers, between the declared schema and the handler, between the docs corpus and the registered reality, between two Go modules. The current suite tests *inside* components, so it was green through all of it. The fix is not "more coverage" in the abstract; it is a fixture corpus that exercises the CLI/MCP surface the way agents actually hit it, contract tests that make schema-vs-handler and docs-vs-registry drift build-breaking, and a dogfood gate that makes ATD's own corpus the permanent integration test. That last piece is also the point: a governance tool whose CI proves it governs itself is its own best demo.

And incident I-1 sets the order of operations beyond argument: before the suite can be *extended*, it must be *contained*. A test run that can rewrite its own fixtures — and once committed the damage to `main` — fails the most basic requirement of a test system: that observing the code does not change it. WP-0 is a half-day of work; nothing else in this report is trustworthy until it lands.

---

## 8. Execution Record — orchestrated pass of 2026-07-16/17

All of §6 was executed by a coordinating agent delegating to 7 subagents (WP-0 solo and first per the critical path; WP-2∥WP-3, then WP-4∥WP-6, then WP-7 split three ways by package, each in an isolated git worktree merged back to `main`). Every WP finished with `make -C atd verify` green and `git status --porcelain` empty. WP-8's cheap parts were folded in (FuzzParse seed corpus via WP-7, `TestLive*` opt-in suite via WP-6). Range: `ca641eb`..`3b7fdd7`, ~6 800 lines of tests/infrastructure.

### 8.1 Per-WP results

| WP | Commit(s) | Outcome |
|---|---|---|
| WP-0 sandbox | `ca641eb` | P-1 scope guard (`pathInside` + `LoadedFromFallback()` refusal, loud stderr warning), P-2 `_test.go`/`testdata/` skip, T-1 snapshot/restore in **all** offenders (audit found more than the report listed: also `lint`, `workspace_integration`, `config/workspace_test.go`, `exploration`, `ollama`, `coverage` tests, plus the cobra `--project` persistent-flag leak), T-2 debris deleted + tombstoned, W-1 tripwire in `pkg/testutil`. AC2 proven: re-introduced naked `LoadFromDir(t.TempDir())` causes zero repo writes. |
| WP-1 verify gate | `c3954ca` | `make verify`/`verify-race` over both modules + tripwire; CLAUDE.md updated. Finding: a tracked `go.work` **already existed** and does *not* close the two-module gap (`go list ./...` from `atd/` still never sees the nested module) — the Makefile is the guard, as §5.1's fallback anticipated. |
| WP-2 fixture + harness | `44efa65`,`1e4ee15`,`2693176` | `testdata/fixture_project/` + `fixture_workspace/`, full `testutil.Sandbox/Run/Golden/Git` API, scenarios S1–S5/S7/S8/S13 all green under `t.Parallel()`/`-race`, goldens deterministic. |
| WP-3 dogfood gate | `4ee2196` | `make dogfood`/`dogfood-quick`/`dogfood-update-baseline`; baseline = 176 fingerprints (47 missing EXPECTATION, 43 missing TECHNICAL INTERFACE, 31 impl-without-test, 51 unresolved links, 4 misc); S6 corpus audit (`dogfood` build tag) — zero H3-swallow on the real corpus; hook template extended in `init.go`. |
| WP-4 MCP contracts | `1996365` | Wiring sweep over all 25 registered tools, schema-honesty + exhaustive per-tool param accounting, tool-set↔docs equality (pins exactly the 8 missing `api_atd_serve_*` atoms: env, config, workspace_list/use/stats, heatmap ×3), error-shape tests. All three red-flip ACs demonstrated. |
| WP-5 CI | `f9d729f`, `3b7fdd7` | `ci.yml`: verify-race + coverage-floors, dogfood, isolated `TestScenario_|TestGolden_` suite — every job tripwired; `live.yml` is `workflow_dispatch`-only. All job command sequences verified locally (no remote/CI runner exists yet). |
| WP-6 LLM boundary | `47da6e1` | `testutil/fakeprovider` covering **both** seams (`chat.Provider`, and new `ollama.Generate`/`Embed` package vars mirroring the `ListModels` seam — the only product seam change). S12 malformed-input matrix, deterministic audit/dissect/map/recon/search-embedding logic tests, 28 prompt goldens, `TestLive*` stubs. Whole suite proven green with `http_proxy` pointed at a dead port. |
| WP-7 back-fill | `9152f05`,`c5e07e7`,`6d5ee70`, `3b7fdd7` | `pkg/exploration` 10.7 % → **73.0 %**, `pkg/coverage` 23.2 % → **64.6 %**, `pkg/atom` ~40 % → **79.2 %**; floors (60/60/70) enforced via `make coverage-floors` in CI, ratchet-up-only. |

### 8.2 Product bugs found and FIXED during the pass

- **`BuildContent` parent-link corruption** (found by the WP-7 round-trip property): the first `parents:` entry was emitted on the key line, so `Parse` returned `"- <id>"` instead of `"<id>"` — silently corrupting every child atom produced by `atd fix`'s split path. Fixed in `pkg/atom/parse.go`.
- **Five zero-value-success parser paths** (WP-6, surgical): `dissect.go` and `generate.go` returned an **empty string as success** on any malformed LLM response (discarded `MarshalIndent` error); `map.go`'s propose/discover/recommendation paths discarded `Unmarshal` errors and proceeded on blank intents. All five now error loudly or salvage, each pinned by a regression test.

### 8.3 Defect backlog — pinned as `KNOWN DEFECT`, not fixed (grep `KNOWN DEFECT` under `atd/`)

1. Strip-`TYPE_`-prefix resolver retry exists only on the workspace resolver; standalone projects fail prefixed ids like nonsense, no suggestions (E1-adjacent).
2. `query` never errors — nonsense id → silent empty array; the one tool of the four violating "resolve or shout".
3. MCP `atd_update` has no `force` param — MCP clients can never override the STABLE+BUSINESS guard (CLI `--force` works).
4. MCP registry: `required` is unenforced (`atd_recon` without `atom` silently enters discover mode); `argBool/argString/argInt` silently coerce wrong-typed args to zero values; `atd_check.line` declared-but-never-read; 6 tools have docs-vs-schema param drift (allowlisted in `mcp_docs_equality_test.go`).
5. Nondeterministic output ordering: `coverage.GenerateReport` full mode and trace `CodeLinks`/`TestLinks` (unordered map iteration).
6. Trace `ImplementationRate` double-counts dependents (fully-implemented 3-chain reports 0.5); `config.MaxDepth` is dead code for trace.
7. cwd-anchored path resolution: `runStats` hardcodes `"."`; `atd_roadmap`/`atd_crawl`/`atd_trace` MCP handlers resolve against process cwd (I-1-shaped). **Needs a human decision:** `atd_config`'s `task`+`model` branch calls a bare `config.Load()` and can *write* to whatever real `.atd` sits above the process cwd.
8. `BuildContent`⇄`Parse` round-trip broken for `Tags`/`Dependents`/`Interface`/`Metadata` (hardcoded/omitted in `BuildContent` — needs a serialization-format decision).
9. Remaining zero-value-success paths: audit-bloat JSON missing `is_bloated` → silent PASS; reconcile response without `diffs` → prints `null`, exits 0.
10. `check --full` walks the entire project root regardless of `--docs`, picking up nested test fixtures (`tests/trace/docs/`, `test-workspace/`) — why the dogfood gate ratchets on `lint` and uses `check --full` only as a smoke test.
11. Ambiguous bare id in a workspace resolves first-registered-project-wins with no ambiguity signal; `ListFiles` dotdir filtering misses top-level `.git/`.

### 8.4 Spec deviations worth knowing

- The report's line numbers held (~395, ~286-289), but its offender list for T-1 was incomplete — see WP-0 row above; the audit-everything instruction was what caught the rest.
- `fixture-e2e` in CI was rescoped from placeholder to the real isolated scenario suite once WP-2 settled on the `TestScenario_` naming convention.
- The dogfood gate does not ratchet on `check --full` NO_IMPL/NO_TESTS counts (would fight docs-first authoring — a fresh unimplemented atom is *supposed* to show NO_IMPL); it ratchets on `lint` + S6 only.
