# ATD Toolkit — Test Strategy Report

**Date:** 2026-07-26
**Scope:** The ATD toolkit itself — `atd/` (Go: `pkg/`, `config/`, `cmd/atd/`), the CLI surface, and the MCP surface. **WebUI and the VSCode extension are explicitly out of scope.**
**Companion reports:** `code_quality_07_26.md` (refactor execution) and `investigation_atd_07_26.md` (agent-usability defects). Both generated large pending changes; the friction of verifying those changes is itself primary evidence for this report.
**Method:** Full inventory of existing `*_test.go` files, per-package `go test -cover` runs on both modules, and a post-mortem of the defects that shipped *despite* a green test suite.

---

## 0. TL;DR

1. **The suite is green but shallow.** ~2 450 test LOC against ~13 000 production LOC, and the tests that exist are mostly flag-existence checks, string-classification micro-tests, and happy-path parses. Packages at the heart of ATD's value proposition are the least tested: `pkg/exploration` (the graph engine) sits at **10.7 %** coverage, `pkg/coverage` (the `check` engine) at **23 %**, and `pkg/audit`, `pkg/indexer`, `pkg/llmservice`, `pkg/mcp`, `pkg/chat` have **zero** test files.
2. **The bug classes that actually hurt were all integration-level** — bare-id silent zeros, `trace`/`check` disagreement, the phantom-tool drift, the H3 parser swallow, the MCP handlers referencing deleted functions. Not one of them was catchable by the current unit tests, and not one *would be* caught by adding more of the same kind of unit test. The missing layer is **scenario tests against a fixture corpus** and **contract tests on the MCP surface**.
3. **The two-module layout is a standing trap.** `go build ./...` at `atd/` does **not** compile `cmd/atd/` (separate `go.mod`). This session, HEAD was "verified green" while the nested module had five compile errors. Every verification step — local, hook, future CI — must run **both** modules, always.
4. **ATD's own docs corpus is the best integration test it will ever have, and it is not wired in.** Running `atd lint` / `atd check --full` against `docs/` as a regression gate ("dogfood gate") would have caught the phantom atoms, the layer drift, and the parser bug years before a human did.
5. **The LLM boundary is already mockable — the codebase just doesn't use the seam.** `ollama.ListModels` is a swappable package var and `chat.Provider` is an interface; a deterministic fake provider unlocks testing of `audit`, `dissect`, `search`, `map`, `recon` logic without a model. Live-model tests should exist but be opt-in (env-gated), never in the default path.

**Recommended order:** fixture corpus + CLI scenario harness → MCP contract tests → dogfood gate in the pre-commit hook/CI → fake LLM provider → targeted unit back-fill on `exploration`/`coverage`. Details in §6.

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
| E6 | **Test-fixture churn under concurrent editing** (`TestUpdateLinks` broken three times in one day by uncoordinated find-replace edits across sessions) | Fixture literals (`old_id`, `module_my_new_module`) collide with strings other tools rewrite | Quarantine fixture strings (unique prefixes like `zz_fixture_*`), or move fixtures to `testdata/` files that rename-propagation ignores (§5.4) |

The pattern: **every escape happened between components or between doc and code** — exactly the seams ATD exists to govern. The test strategy must mirror the product's own philosophy: the corpus and the code are one system; test them as one system.

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

- **Every id carries the `zzfix` marker** so repo-wide rename/propagation tooling never collides with it (lesson E6).
- It deliberately contains **one instance of every state the tools must classify**: covered atom, impl-without-test, orphan, cross-layer chain, an atom using `###` subheadings, a `[[project:atom]]` cross-ref (add a second mini-project for workspace tests).
- A scenario harness (`cmd/atd/cmd/scenario_test.go` or a separate `atd/e2e/` package) copies the fixture to `t.TempDir()`, then drives the **CLI entry points** (the `run*`/pkg functions the commands delegate to — same functions the MCP handlers call, so one harness covers both surfaces) and asserts on structured output.
- **Golden files** for report-shaped output (`check --full`, `lint`, `stats`): store expected output in `testdata/golden/`, compare normalized (strip timestamps/paths), regenerate with an `-update` flag. This makes output regressions visible in diffs instead of invisible.

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

### 5.4 Fixture hygiene (lesson E6)

- All fixture atom ids/strings carry a `zzfix` (or similar) marker no real tooling will ever rename.
- Prefer `testdata/` files over string literals inside `_test.go` for anything an ATD tool might rewrite — rename-propagation and find-replace passes skip `testdata/` by convention (and add that convention to the agent rules).
- When a test encodes a shipped bug, name it and comment it after the incident (`TestParseSectionWithSubheadings` is the house style — keep it).

### 5.5 Cheap extras worth taking

- `go vet` already in the gate; add `staticcheck` if appetite exists (the codebase is clean enough post-refactor that adoption is cheap now).
- **Fuzz** `atom.Parse` and `extractLinks` (`go test -fuzz`, native tooling): both consume arbitrary user files; the H3 bug is exactly the class fuzzing plus the S6 invariant would surface.
- `t.Parallel()` on the scenario harness — fixtures are copied per-test, so it's free speed.

---

## 6. Prioritized Rollout

| # | Work item | Effort | Buys |
|---|---|---|---|
| 1 | `make verify` two-module gate + agent-rules note (+ optional `go.work`) | ~1 h | Kills E5 permanently; prerequisite for everything else |
| 2 | Fixture project (`testdata/fixture_project/`) + scenario harness with S1–S5, S7, S8 | 1–2 days | The integration layer; regression bed for all field-report classes |
| 3 | S6 corpus parse audit + dogfood ratchet in pre-commit / CI | ~½ day | ATD tests itself on every commit; catches doc-code drift at the source |
| 4 | MCP contract tests (S11: wiring sweep, schema honesty, tool-set↔docs) | ~1 day | Pins the MCP surface; makes the D1 atom back-fill verifiable and permanent |
| 5 | CI workflow wiring (§5.2) | ~½ day | Makes 1–4 unskippable |
| 6 | Fake `chat.Provider` + S12 parser robustness + prompt goldens | 1–2 days | First-ever tests on the LLM half; de-risks the least reliable code |
| 7 | Unit back-fill: `exploration` → `coverage` → `atom` round-trip (§3.1) + coverage floors | ongoing | Depth under the engines |
| 8 | Fuzz targets for `Parse`/`extractLinks`; opt-in live-LLM smoke suite | opportunistic | Long-tail robustness |

Items 1–3 are the "stop the bleeding" set and are independent of any pending product work. Item 4 pairs naturally with authoring the 8 missing `api_atd_serve_*` atoms (investigation follow-up D1) — write the atoms and the test that locks them in the same change.

---

## 7. Conclusion

ATD's failure history is unusually legible: every serious defect lived **on a seam** — between two tools' answers, between the declared schema and the handler, between the docs corpus and the registered reality, between two Go modules. The current suite tests *inside* components, so it was green through all of it. The fix is not "more coverage" in the abstract; it is a fixture corpus that exercises the CLI/MCP surface the way agents actually hit it, contract tests that make schema-vs-handler and docs-vs-registry drift build-breaking, and a dogfood gate that makes ATD's own corpus the permanent integration test. That last piece is also the point: a governance tool whose CI proves it governs itself is its own best demo.
