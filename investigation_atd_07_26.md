# ATD Toolkit — Usability Investigation for LLM Agents

**Date:** 2026-07-13
**Scope:** ATD MCP toolset + CLI + orchestration scripts + agent-facing documentation (ATD.md, `.agent/rules/ATD.md`, CLAUDE.md, README.md). WebUI explicitly out of scope.
**Question:** How usable is ATD for an LLM agent today, and would opencode-style **skills** materially improve correct usage? What else would bring the system closer to its own stated philosophy?
**Method:** Read the documentation surface end-to-end, then verified every claim against the Go source (`atd/cmd/atd/cmd/*.go`, esp. `mcp_tools.go`), the orchestration scripts, the `.atd` configs, the issues backlog, and the two prior field reports (`atd_feedback_2026-07-07.md`, `atd_feedback_evaluation_2026-07-13.md`). All file:line citations are verifiable.

---

## 0. TL;DR

1. **ATD's deterministic core is genuinely good** — fast, token-free, well-conceived tools (`query`, `crawl`, `weave`, `update`, `check`, `trace`, `heatmap*`, workspace ops). An agent that sticks to these gets real value.
2. **The documentation is actively misleading in several places**, and the single most-loaded rule (`.agent/rules/ATD.md`, `always_on`) tells the agent to call **`atd_discover`, a tool that does not exist**. The real tool is `atd_map`. This is the #1 usability defect and it is self-inflicted.
3. **The LLM-backed and coverage tools silently return wrong/empty results** in common cases (bare atom id → clean zero; missing index → empty; type-prefixed id → not-found; `trace` disagrees with `check`). For an agent, a *confident wrong answer is worse than an error*. This is the #1 *tooling* defect class and it is the one most corrosive to agent trust.
4. **Skills will help — but they are not the fix.** Skills are the right *container* for the heavy procedural workflows (create / alter / pre-code-change), and they let you slim the always-on context. But a skill that wraps tools which silently lie cannot guarantee correctness. **Sequence: fix the substrate, then write the skills on top of it.**
5. ATD also **does not dogfood itself reliably** (§4): its own `STABLE` atoms describe tools under names that no longer exist, 10 of 25 tools have *no* atom, there is no `CONTRACT`/`VISION`, and non-canonical types/layers are in active use across its own corpus. Fixing that through the ATD workflow is both the cheapest win and the most convincing proof the philosophy works.

---

## 0.5 Status Update — 2026-07-14 (agent work session)

A first pass acted on the highest-leverage items. **Nothing is committed yet** (changes live in the working tree). Verification was done against an isolated git worktree at clean `HEAD`, because the nested `cmd/atd` module was concurrently mid-refactor by a separate code-quality effort and did not build during this session.

### First correction to this report

**TL;DR #3 (silent wrong answers) is already fixed** and was *not* re-done. The 2026-07-13 field-report commits (`3a94502`, `5af356a`) resolved all four §3.1 paths and the §3.2 trace/check disagreement in the substrate — verified against current source (`check_coverage.go:70-77` canonicalizes via `explorer.CanonicalAtomID`/`SuggestAtomID`; `atd_test_links` and `atd_trace` route through the same resolver; `trace` derives links from the live crawl). **§3.1 and §3.2 of this document are stale** and describe the pre-07-13 state. The genuinely-live defect was TL;DR **#2** (phantom `atd_discover`), addressed below.

### Done this session (verified in isolation, not committed)

| Report item | Action taken | Status |
|---|---|---|
| §8 #1 / §7.1 / D2 — phantom `atd_discover` | `.agent/rules/ATD.md`: all 3 refs → `atd_map` (correct 3-mode syntax). Renamed+rewrote the 2 phantom atoms → `api_atd_serve_map` / `api_atd_serve_check` with accurate schemas; updated inbound `[[…]]` refs; fixed the 2 stale `@spec-link` comments in `mcp_tools.go`. | ✅ done |
| §4 #1/#3 (D2) — `discover→map` / `verify→check` renames | Propagated end-to-end (atoms + source tags + rules). Fills 2 of the 10 missing tool-atom gaps. | ✅ done |
| §7.5 / ISS-010 — STABLE+BUSINESS update guard | `atom.Update` + `atd update` now refuse to modify an existing STABLE+BUSINESS atom without `--force` (CLI) / `Force:true`. MCP path enforces by default (see deferred item below). | ✅ done |
| §4 #6 — non-canonical `CUSTOMER` layer (22 atoms) | Migrated by type per §1.5 (`DOMAIN`/`REQUIREMENT`→BUSINESS, `UI`/`SPECIFICATION`→ARCHITECTURE, `USAGE`→IMPLEMENTATION). Blanket→BUSINESS would have created new §1.5 violations, so remap is type-aware. | ✅ done |
| §4 #4 / D3 — no CONTRACT/VISION | Created `contract_atd` + `vision_atd` as **DRAFT** (agent-proposed per §1.4 bootstrapping; await human ratification). | ⚠️ needs ratification |
| §4 #5 / D5 — lint doesn't flag taxonomy | `atd lint` now validates `type` against the ATD.md-sanctioned union (§1.3+§1.5, so `SERVICE`/`USAGE` are accepted, not false-flagged) and enforces CONTRACT/VISION presence (gated on BUSINESS atoms) + uniqueness. Note: the invalid-*layer* enum check already existed — that part of §4 #6 was stale. | ✅ done |
| **New bug (found via dogfooding)** — `atom.Parse` H3-swallow | Section parser treated `###` subheadings as section terminators, silently dropping `## TECHNICAL INTERFACE`/`## EXPECTATION` content — which also starves `check --semantic`, `trace --summary`, and `assemble`. Fixed to H2-only boundary + regression test. Cleared all 17 `api_atd_serve_*` false "missing section" lint errors. | ✅ done |

**Verification:** root `atd-tools` module builds clean; all `pkg/atom` tests pass incl. the new regression; `lint.go` builds + all 3 lint unit tests pass in the isolated worktree; corpus dogfood shows my target lint categories at **0** (non-canonical types, invalid layers, project-governance, contract/vision self-consistency).

### Follow-up recommendations (deferred — for a later session)

1. **Ratify CONTRACT/VISION.** Review `contract_atd` (6 invariants) + `vision_atd` (scope boundary) and promote DRAFT→STABLE if they match intent. Once STABLE, editing them will itself require `--force` (the new guard).
2. **Wire the MCP `force` override.** Deliberately *not* added to the `atd_update` MCP handler this session because `mcp_tools.go` was mid-refactor. Current effect: MCP agents cannot override the STABLE+BUSINESS guard at all (defensible — such edits arguably belong on a human CLI). Add a `force` param to the handler + thread to `runUpdate` if MCP override is wanted.
3. **Re-run full `cmd/atd` build + `atd lint` on the whole corpus** once the concurrent code-quality refactor lands — the cmd-layer changes here (`lint.go`, 2 `mcp_tools.go` comments) are verified only in isolation.
4. **Commit strategy.** These changes are entangled in the working tree with a separate code-quality refactor (`audit.go`, `fix.go`, `index.go`, `search.go`, `extension.js`, deletions). Separate into coherent commits (governance/dogfooding vs. refactor) before landing.
5. **Remaining lint debt is pre-existing authoring gaps, not addressed here:** ~43 genuinely-empty `## TECHNICAL INTERFACE` sections, ~47 empty `## EXPECTATION`, ~38 impl-without-`@test-link` traceability gaps, and cross-project unresolved `[[project:atom]]` links (some may be single-project-lint-context false positives). Candidate for a bulk `atd_update` authoring pass.
6. **Still open from §8 (untouched this session):** #2 tool-count reconciliation (README 19→25) + issues/README drift (§3.6/§7.9); #7 generate ATD.md tool reference from `mcp_tools.go` (§7.3); #10 unify link index (mostly landed via §5 fix, confirm); #13 `atd_doctor` orientation probe (§7.6); #14 expose cold-start/full-audit/reconcile via MCP (§7.7); D1 the remaining 8 missing tool atoms; D4 the `SERVICE`/`USAGE`/`SPECIFICATION` spec-vs-config reconciliation (§4 #5/#7 — a spec decision, deliberately left to a human).
7. **Skills (§6).** Substrate is now in better shape (silent-zeros fixed pre-session; governance guard + phantom tool fixed this session). Skills A (`create-atom`) and C (`pre-code-change`) remain the recommended next build; Skill B (`alter-atom`) dependencies are closer but still want the MCP `force` wiring (#2) to make its "require confirmation" step tool-backed rather than prose.

---

## 1. The Real Tool Surface (ground truth from source)

### 1.1 MCP tools — actually 25, registered in `atd/cmd/atd/cmd/mcp_tools.go`

| Deterministic (no LLM) | LLM-backed | Config / workspace |
|---|---|---|
| `atd_query`, `atd_crawl`, `atd_weave`, `atd_update`, `atd_roadmap`, `atd_stats`, `atd_check`, `atd_assemble`*, `atd_trace`†, `atd_test_links`, `atd_lint` | `atd_dissect`, `atd_index`, `atd_search`, `atd_audit`, `atd_recon`, `atd_map` | `atd_env`, `atd_config`, `atd_workspace_list`, `atd_workspace_use`, `atd_workspace_stats`, `atd_heatmap`, `atd_heatmap_code`, `atd_heatmap_project` |

\* `atd_assemble` has an LLM *summarization* path; structurally deterministic otherwise.
† `atd_trace(summary=true)` is LLM-backed; raw JSON is deterministic.

### 1.2 CLI-only commands (NOT exposed via MCP)

`init`, `workspace init|add`, `compare`, `congruence`, `continue`, `fix`, `generate`, `reconcile`, `verify`, `version`, `webui`. Plus the two orchestration shell scripts `atd-cold-start.sh` and `atd-full-audit.sh`.

**Implication:** an MCP-only agent (e.g. Claude Desktop with no shell) **cannot** bootstrap a project (`init`), run the cold-start pipeline, run a full audit, or reconcile overlapping requirements. Only shell-capable agents (opencode, Claude Code) can. This is tracked as ISS-031 / ISS-050 / ISS-039 but unfixed.

### 1.3 Documentation says three different things about the tool count

| Source | Claimed tool count | Reality |
|---|---|---|
| `README.md:16, 63` | **19** | wrong |
| `ATD.md:357` | **25** | correct |
| `.agent/rules/ATD.md` | (lists ~12 by name, incl. a phantom) | partially wrong |

---

## 2. Documentation Layers & Their Drift

There are **four** overlapping sources of agent guidance. They disagree with each other and with the code.

| Doc | Audience | Load model | Size | State |
|---|---|---|---|---|
| `ATD.md` (root) | reference manual | on-demand | 941 lines | accurate on concepts, **wrong on several tool signatures** |
| `.agent/rules/ATD.md` | always-on ruleset | `trigger: always_on` | 181 lines | references **phantom `atd_discover`**; otherwise the best short guide |
| `CLAUDE.md` | UpsilonBattle dev guide | injected into system prompt | ~heavy | project-specific; partly stale |
| `README.md` | onboarding | on-demand | 223 lines | **wrong tool count (19)**; references 52 issues, only 11 on disk |

### 2.1 Concrete drift between `ATD.md` and the real MCP schemas

| Tool | `ATD.md` says | Code actually does | Evidence |
|---|---|---|---|
| `atd_assemble` | params `starts, purpose, snapshot, theme` | params `starts, intent, length, structured, json, only_parents, only_dependents` | `mcp_tools.go:261-282` vs `ATD.md:507-513` |
| `atd_check` | supports `line` (int) | `line` is in the JSON schema but **never read** by the handler | `mcp_tools.go:234` (schema) vs `:238-253` (handler ignores it) |
| `atd_query` / `atd_search` / `atd_crawl` | (undocumented extra params) | all accept `paths_only`; `crawl` accepts `workspace` | `mcp_tools.go:83, 104, 374` |
| `atd_discover` | (described at length in rules + the atom) | **No such tool.** Renamed to `atd_map` (3 modes) | `mcp_tools.go:438-462` |
| `atd_verify` | (issues + a STABLE atom use this name) | MCP name is `atd_check`; `verify` is CLI-only | `mcp_tools.go:219-253` |

### 2.2 The phantom `atd_discover` — the single worst usability bug

The always-on ruleset instructs the agent to use `atd_discover` in **three** places:

- `.agent/rules/ATD.md:73-76` (section 4, "Delegate LLM tasks")
- `.agent/rules/ATD.md:91` (section 6, "Discovery")
- `.agent/rules/ATD.md:177-180` (Quick Reference, four example calls)

`atd_discover` is **not registered**. The real tool is `atd_map` (default / confirm / propose modes). An agent following its own ruleset will emit a tool call that the MCP server rejects, then either improvise or give up. Worse, the project's *own* atom `docs/api_atd_serve_discover.atom.md` is:

- `status: STABLE`, `priority: 5`
- titled `"MCP Tool: atd_discover"`
- documents a **different schema** (`{file, docs}`) than the real `atd_map` (`{file, atom, new}`)

So ATD's source of truth contains a STABLE atom that is wrong about its own product. The code even leaves a fingerprint of the rename: `mcp_tools.go:438` carries the comment `// @spec-link [[api_atd_serve_discover]]` directly above the `atd_map` registration, and `mcp_tools.go:219` tags the `atd_check` tool with `@spec-link [[api_atd_serve_verify]]`. Both renames (discover→map, verify→check) were never propagated.

> **This is the strongest argument the philosophy isn't yet self-sustaining:** the bidirectional-traceability promise is exactly what should have caught these renames, and didn't.

---

## 3. Usability Assessment for an LLM Agent

Organized by *theme*, because most defects are instances of two or three cross-cutting root causes (already identified in `atd_feedback_evaluation_2026-07-13.md`).

### 3.1 Silent wrong answers (worst class for agents)

An agent can recover from an explicit error. It cannot recover from a confident zero. ATD currently produces confident zeros in at least four situations — all reproduced first-hand in the field reports:

| Trigger | Symptom | Root cause | Evidence |
|---|---|---|---|
| Bare atom id (`api_shop_purchase`) to `atd_check` / `atd_test_links` / `atd_trace` | clean `NO_IMPL` / 0 coverage | input id never canonicalized through the resolver | `check_coverage.go:68`; resolver at `exploration.go:687` already does this, the tools just skip it |
| Type/layer-prefixed id (`requirement_req_ui_session_timeout`) | silent not-found | prefix not stripped, no loud failure | `resolver.go:60-65` |
| Semantic `atd_search` with no/empty index | empty result block, no diagnostic | missing-index error swallowed by grep fallback | `search.go:46-51, 155-158` |
| `atd_search` side-effect | writes `docs/.atd_index.db` into the working tree (a `git add .` accident) | db path under `DocsDir()` | `mcp_tools.go:387`; `search.go:52` |

The reporter (`atd_feedback_2026-07-07.md` Addendum A) calls the bare-id case **the single highest-yield defect**: it recurred in ~15 of 55 audited scenarios and is *self-propagating*, because a clean zero looks like "no coverage," so authors copy-paste the bare form and nothing flags it.

**Agent impact:** catastrophic. The agent builds a false world model ("this atom is unimplemented") and proceeds to write duplicate code or file spurious gaps.

### 3.2 Cross-tool disagreement (the agent can't form a consistent model)

For the *same* atom, `atd_check`, `atd_trace`, and `atd_query` report different coverage numbers. Reproduced and root-caused in the evaluation report:

- `check` iterates the **live crawl** (`explorer.SpecLinks`) — authoritative, cross-project.
- `trace` reads `node.Implementations` (`linked_codes`), a field that is **frontmatter (possibly stale) + live crawl**, deduped by file (loses per-site counts) → `implementation_rate: 0.5` while `check` says 3 impl / OK.
- `query` returns raw frontmatter `linked_codes`.

Evidence: `exploration.go:151, 487`; `atom/parse.go:40`; `check_coverage.go:126-131`. Root theme: **two coordinate systems** (cwd vs `ProjectRoot` vs git-root) and **two link indexes** (live crawl vs persisted `linked_codes`) are mixed across tools (`atd_feedback_evaluation_2026-07-13.md` §1, §5, §6, §8).

**Agent impact:** an agent running `trace` before a code change (the ruleset's MANDATORY step) may be told the atom is unhealthy when `check` says it's fine — leading to either false alarms or learned distrust of the tools.

### 3.3 LLM-backed tools produce unstructured / wrong output

- `atd_trace(summary=true)` narrated the **parent** atom instead of the requested target (`atd_feedback_evaluation_2026-07-13.md` §4; root cause in `trace.go:52-95` + `pkg/prompt/trace_summary.go`). This is the ruleset's mandated "investigation" step.
- `atd_map` confirm mode returned `{"Confidence": 0, "Mismatches": "Yes, there are several mismatches…"}` — a boolean-ish sentence in a field meant for an itemized rationale, with no way to act on it (§8).

**Agent impact:** the highest-value semantic tools are the least reliable, so a careful agent learns to avoid them and falls back to grep — defeating the point of having LLM-backed tooling.

### 3.4 Governance is advisory, not enforced

- The `status` field (`DRAFT`/`REVIEW`/`STABLE`) is **ignored by all tooling** (ISS-010, open). Nothing actually prevents a `atd_update` from rewriting a `STABLE` BUSINESS atom. The "heavy human sign-off" the docs promise is purely the agent's own discipline.
- `CONTRACT` / `VISION` gating (ATD.md §1.4) is described as mandatory ("MUST be read", "MUST propose") but there is no tool that enforces or even *checks* it. It's prose the agent is supposed to obey.
- Layer ancestry ("every IMPLEMENTATION atom must have a BUSINESS ancestor") is enforced only softly via `atd_trace`'s `has_customer_origin` flag, which the ruleset correctly says is *the one* warning to treat as a blocker (`.agent/rules/ATD.md:98-103`). Good rule — but it lives in the ruleset, not the tool.

**Agent impact:** the agent is the entire enforcement layer. A less-careful agent (or a smaller model) will quietly violate governance because nothing pushes back.

### 3.5 Naming instability erodes trust

Three renames were half-done: `discover → map`, `verify → check`, and the `atd_cli → atd` binary rename (still referenced in `.claude/settings.local.json:8-10` permissions, which now match nothing). The issues backlog still talks about `atd_verify` (ISS-051) and `atd_discover`. An agent reading issues + rules + atoms simultaneously gets three vocabularies.

### 3.6 Issues backlog drift

`README.md` references **52** issues; only **11** exist on disk in `issues/`. The tracker's own README is out of sync with reality. An agent told to "check the issues" gets an unreliable index.

### 3.7 Discovery / onboarding friction

- No single "ground truth" probe. To orient, an agent must call `atd_workspace_list`, `atd_stats`, `atd_env`, `atd_config(list)`, and guess an `atd_query` — five calls, several of which may return inconsistent numbers (§3.2).
- The `docs/.atd_index.db` artifact and the missing-index silent failure (§3.1) mean semantic search — the natural "find related atoms" entry point — either works invisibly or silently doesn't.

---

## 4. ATD Does Not Dogfood Its Own Framework

ATD's central promise is bidirectional traceability and doc-code co-evolution: the graph should catch drift, and `STABLE` means "approved and enforced, code must comply." The project's own `docs/` corpus (178 atoms, including 18 that document the MCP tools) is the natural proof of concept. It currently fails that proof in at least ten concrete ways — each one a defect class that ATD's own tooling is supposed to prevent.

| # | Failure | Evidence | ATD principle violated |
|---|---|---|---|
| 1 | **2 STABLE atoms document tools that don't exist** (`atd_discover`, `atd_verify`), with wrong schemas | `docs/api_atd_serve_discover.atom.md` (`status: STABLE`, schema `{file, docs}`); `docs/api_atd_serve_verify.atom.md` (`status: STABLE`, schema `{}`, "Required: None") — real tools are `atd_map` (`{file, atom, new}`) and `atd_check` (6 params) | "STABLE = enforced, code must comply"; bidirectional traceability |
| 2 | **10 of 25 tools have no atom at all** — 40% of the surface is absent from the graph | No `api_atd_serve_*` for: `atd_check`, `atd_map`, `atd_env`, `atd_config`, `atd_workspace_{list,use,stats}`, `atd_heatmap{,_code,_project}` | "Every code module links back to its governing atom" |
| 3 | **Source still tags renamed tools with the old atom ids** | `mcp_tools.go:219` `// @spec-link [[api_atd_serve_verify]]` above the `atd_check` registration; `mcp_tools.go:438` `// @spec-link [[api_atd_serve_discover]]` above `atd_map` | Doc-code co-evolution (renames not propagated) |
| 4 | **No `CONTRACT` or `VISION` atom exists**, despite §1.4 mandating "exactly one of each per project" | `grep 'type: CONTRACT\|type: VISION' docs/` → empty | Project governance gates (§1.4) — not instantiated |
| 5 | **Non-canonical types in active use**: `SERVICE` (17 atoms, incl. core ancestor `service_atd_serve`), `USAGE` (4), `SPECIFICATION` (3) — none in the canonical 11-type list (§1.3) | type distribution across `docs/*.atom.md`; ISS-075 confirms | Type-system atomicity (§1.3) |
| 6 | **Non-canonical layer `CUSTOMER`** used by 22 atoms, though §1.5 defines only `BUSINESS`/`ARCHITECTURE`/`IMPLEMENTATION` | layer distribution across `docs/`; ISS-076 confirms | Layer hierarchy (§1.5) |
| 7 | **`.atd` config references non-canonical types** in `bloating_factor.type_overrides` (`USECASE`, `SPECIFICATION`) | root `.atd` | Config ↔ type-system agreement |
| 8 | **`status` is demonstrably meaningless**: 104 STABLE atoms, yet that set includes the 2 wrong atoms from #1; nothing enforces maturity | ISS-010 (open); status distribution | Lifecycle / status gating (§1.6) |
| 9 | **An atom carries the stale `linked_codes` frontmatter field** — the exact field that breaks `trace`/`check` agreement (field report §5) | `grep -l linked_codes docs/*.atom.md` → 1 hit | "Live crawl is the single source of truth" |
| 10 | **Same-family layering is inconsistent**: `api_atd_serve_lint` is `layer: IMPLEMENTATION` while its 16 MCP siblings are `layer: ARCHITECTURE` (and it carries an extra parent `mechanic_atd_lint`) | `docs/api_atd_serve_lint.atom.md` vs siblings | Atomicity / consistency within a family |

> **One thing it does get right:** the MCP-tools ancestry *does* reach the BUSINESS layer (`api_atd_serve_*` → `api_atd_mcp_ops` → `service_atd_serve` → `module_atd_cli` → `domain_atd_philosophy` [DOMAIN/BUSINESS]). So the `has_customer_origin` invariant — the one warning the ruleset treats as a hard blocker — is satisfied for this subtree. That is the correct dogfoading outcome; it is also the *only* structural invariant in the list above that currently holds.

### What this means for an agent

An agent told to "trust the ATD graph" and then asked "what does the `atd_check` tool do?" will, in sequence: (a) find no atom for `atd_check`; (b) possibly surface `api_atd_serve_verify` via a fuzzy/grep search; (c) read a STABLE atom asserting the schema is `{}` with no parameters; (d) call `atd_check` with the wrong mental model. Every layer of the promise — the atom exists → it's STABLE so it's authoritative → its schema is accurate — fails in turn, and the agent cannot recover without cross-checking the Go source. That is precisely the cross-check the graph is supposed to make unnecessary.

### The deeper point

ATD's own backlog shows these are *known* (ISS-010, ISS-075, ISS-076) — yet the tooling does not flag them on its own corpus. `atd_lint` should catch non-canonical types/layers and the missing CONTRACT/VISION; `atd_audit` should catch atoms describing nonexistent tools; `atd_crawl(gaps=true)` should surface the 10 missing tool atoms; `atd_check` should report the `verify`/`discover` source tags as pointing at atoms whose schemas don't match the call sites. Either these tools aren't being run on ATD's own docs, or they don't detect these classes. Either way, a prospective user's first question — "does it work for you?" — currently has an uncomfortable answer, and that answer undermines the whole value proposition more than any single bug in §3 does.

### Cheapest credible fix

Resolve the `discover→map` and `verify→check` renames **through the ATD workflow itself** (Skill B: `atd_update` the two atoms *and* their source `@spec-link` tags, re-`weave`, re-`lint`), create the 10 missing tool atoms, and instantiate CONTRACT/VISION. Doing this *with the tooling* — not by hand-editing Markdown — is the single most convincing demonstration that ATD delivers on its philosophy, and it doubles as the integration test the project currently lacks.

---

## 5. What Works Well (keep / amplify)

- **Deterministic core** (`query`, `crawl`, `weave`, `update`, `check`, `lint`, `trace` raw JSON, `heatmap*`, workspace ops) — fast, token-free, well-designed. The embedded MCP descriptions in `mcp_tools.go` are *more accurate than ATD.md*; they are a hidden asset.
- **`atd_query`** is the gold standard: it resolves a **bare id workspace-wide** and returns frontmatter + `linked_codes` with correct cross-project attribution (`atd_feedback_2026-07-07.md` §5). Every other id-accepting tool should behave like `query`.
- **`atd_check full:true`** cross-project attribution now works correctly (67/67 atoms, correct `project:atom` prefixes).
- **`atd_search grep:`** is reliable and correctly scoped.
- **Workspace tools** (`list`/`use`/`stats`) work and the cross-project `[[project:atom]]` model is sound.
- **The day-to-day lifecycle** (Plan → Specify → Implement → Verify → Evolve) is a genuinely good mental model for an agent; it just needs tooling that matches it.
- **`has_customer_origin` as the one hard blocker** (`.agent/rules/ATD.md:98-103`) is the right pragmatic call: top-down incompleteness is tolerated, code without a business root is not.

---

## 6. Skills Evaluation

### 6.1 Are skills the right tool? — Yes, for *procedure*; no, for *invariants*.

The user's instinct is correct: the heavy procedural guidance currently stuffed into `ATD.md` (941 lines) and the always-on rules (181 lines) is a poor fit for "always loaded." opencode skills are loaded **on demand** when the task matches the skill's `description`. That maps cleanly onto ATD's own lifecycle stages.

**The splitting principle:**
- **Always-on rules** should carry only *invariants*: the atom blueprint, the "no parent, no code" rule, surgical tag placement, the 3–4 tool names that matter, and the "resolve-or-shout" expectation. Target: ≤60 lines.
- **Skills** should carry *procedures*: multi-step workflows with decision points, checks, and rollback. Loaded only when the agent is doing that specific thing.

This is strictly better than today's "181-line always-on ruleset that references a phantom tool."

### 6.2 Proposed skills (3 core + 2 optional)

Each skill below: **trigger** (when opencode loads it), **scope** (what it covers), **value**, **dependency** (what must be true for the skill's guarantees to hold).

#### Skill A — `atd-create-atom` (core)
- **Trigger:** "create a new atom / new spec / new requirement / add documentation for a feature."
- **Scope:**
  1. `atd_search` (grep + semantic) for existing/overlapping atoms → avoid duplicates.
  2. Verify a parent BUSINESS or ARCHITECTURE atom exists (the "no parent, no code" rule). If not, **stop** and propose the missing upstream atom to the user.
  3. Read `CONTRACT`/`VISION` if creating a BUSINESS-layer atom (governance gate).
  4. Pick `type`/`layer`; `atd_config(bloating_factor=<type>)` to calibrate granularity; enforce the "no 'and' in INTENT" rule.
  5. Author the four H2 sections; pick the type-specific template (USECASE/USER_STORY/API — see ISS-038).
  6. Create via `atd_update` (never hand-write the file); set `parents`; `status=DRAFT`.
  7. `atd_weave`; `atd_audit` for collisions; `atd_check` on touched files.
- **Value:** high. This is the most common agent action and the one most often done wrong (orphan atoms, missing ancestry, bloat).
- **Dependency:** `atd_audit`/`atd_search` must not silently empty-out (§3.1).

#### Skill B — `atd-alter-atom` (core, higher impact than A)
- **Trigger:** "change / update / refine / rename / reparent an existing atom," especially anything `STABLE` or BUSINESS/ARCHITECTURE layer.
- **Scope:**
  1. `atd_crawl` (blast radius) + `atd_trace(atom, summary=true)` (vertical context) **before** editing.
  2. If `status=STABLE` or layer=BUSINESS → require explicit user confirmation; surface the crawl impact in the prompt (the ruleset already demands this at `.agent/rules/ATD.md:94`).
  3. Edit **surgically** via `atd_update` (set/intent/logic/…); never rewrite the file.
  4. If `id`/`type` changes → confirm reference propagation happened (`mcp_tools.go:444` claims it does; verify with `atd_lint`).
  5. Re-`weave`; re-`check`; re-`lint`; if the change alters intent, flag downstream `@spec-link`/`@test-link` sites for review (use `atd_crawl` + `atd_test_links`).
- **Value:** very high — this is where governance actually matters.
- **Dependency:** `trace`/`crawl`/`check` must agree (§3.2); `trace summary=true` must narrate the *right* atom (§3.3); and ideally `status` is enforced by the tool (ISS-010) so the skill's "require confirmation" step is backed, not just prose.

#### Skill C — `atd-pre-code-change` (core — directly answers the user's "ensuring alteration doesn't break intent")
- **Trigger:** "edit / refactor / fix code" that carries (or should carry) `@spec-link`.
- **Scope:**
  1. Identify `@spec-link` atoms in the file(s) being touched (`atd_heatmap_code` or grep).
  2. For each: `atd_trace(atom, summary=true)` → load the spec's INTENT and EXPECTATION into context.
  3. State explicitly whether the planned change still satisfies INTENT/EXPECTATION; if it *changes* behavior, route through Skill B (alter the atom first) before touching code.
  4. After the edit: `atd_check` (structural, fast) on the file; add/move `@spec-link`/`@test-link` following surgical placement rules; if behavior shifted, update EXPECTATION.
- **Value:** high — this is the bridge between "doc-code co-evolution" (Principle 3) and daily reality. Today nothing enforces this loop; the agent just edits.
- **Dependency:** `atd_check file:` dedup (§3 of field report — trivial fix), path-base consistency (§1), and a reliable `trace summary`.

#### Skill D — `atd-verify` (optional; folds into A/B/C as a closing step)
- **Trigger:** "before commit / pre-commit / verify my changes / CI gate."
- **Scope:** `atd_lint` → `atd_check` → `atd_test_links` → `atd_crawl(gaps=true)`. Optionally `atd_check(semantic=true)` for STABLE promotions (token cost — make it opt-in).
- **Value:** medium as a standalone skill; high as the *closing block* of A/B/C.

#### Skill E — `atd-cold-start` (optional, niche, currently blocked)
- **Trigger:** "bootstrap / onboard ATD onto a new or legacy codebase."
- **Scope:** the `roadmap → index → dissect → weave → discover → recon → audit` pipeline.
- **Value:** medium (rare but high-complexity).
- **Blocker:** the pipeline is shell-only today (`atd-cold-start.sh`). An MCP-only agent cannot run it. Either expose it via MCP (ISS-031/050) or restrict this skill to shell-capable agents and have it shell out explicitly.

### 6.3 What skills deliberately CANNOT fix

Skills are instructions to an agent. They cannot fix:

- Silent zeros (§3.1) — the skill says "call `atd_trace`," the tool returns a phantom, the skill's guarantee is hollow.
- Cross-tool disagreement (§3.2) — a skill that says "trust `trace`" and a skill that says "trust `check`" will disagree exactly when the tools do.
- LLM-output quality (§3.3) — Skill C *depends* on `trace summary=true` narrating the correct atom.
- Unenforced status (ISS-010) — Skill B's "require confirmation" is a request, not a guarantee, until the tool refuses.

> **Recommendation:** ship Skills A and C first (highest value, fewest hard dependencies), but treat the §3.1/§3.2 fixes as **prerequisites for Skill B** being trustworthy. Do not advertise Skill B's governance guarantees until `trace`/`check`/`query` agree and bare ids resolve loudly.

### 6.4 Existing skill pattern to follow

`.agent/skills/issue_management/SKILL.md` is already a clean, working example in this repo: frontmatter (`name`, `description`), "When to Activate," numbered steps, templates, quick-reference checklist. Use it as the structural template for the ATD skills. (Note: it currently references `/workspace/issues/` and an `issues` script — confirm path conventions before copy-pasting.)

---

## 7. Other Recommendations (beyond skills)

Ordered by **value-to-effort**, and framed as preliminary-research options since ATD semantics changes are deferred.

### 7.1 Same-day wins (cheap, unblock trust)

1. **Delete `atd_discover` from the always-on rules.** Replace with `atd_map` (and the 3 modes). One file, ~10 edits. This alone removes the single most damaging piece of misinformation an agent receives every turn. (`.agent/rules/ATD.md:73-76, 91, 177-180`.)
2. **Reconcile the tool count.** README says 19, ATD.md says 25. Pick one source of truth (see 6.3).
3. **Trivial code fixes already specified** in `atd_feedback_evaluation_2026-07-13.md`: file-mode dedup in `atd_check` (§3, "copy the diff-mode `seen` map"), index-db relocation out of the working tree + honest "index missing" message (§7). These are same-day, low-risk, and kill two of the worst silent-failure modes.

### 7.2 The cross-cutting principle: "resolve or shout, never silent empty"

Make this a hard invariant for **every** tool that accepts an atom id or returns coverage: either return the canonical, cross-project-correct answer, or return a loud, actionable error ("atom 'X' not found; did you mean `upsilonapi:X`?"). The reporter is explicit that a silent zero is the worst outcome and is self-propagating. Concretely: route every user-supplied id through `explorer.ResolveAtom` (which already does the right thing — that's why `atd_query` works) at *every* entry point (`check`, `test_links`, `trace`, `map`, `audit`). Low effort, highest agent-trust payoff.

### 7.3 Make the source the single source of truth (dogfood the philosophy)

The embedded MCP descriptions in `mcp_tools.go` are accurate and current; `ATD.md`'s tool reference is not. **Generate the ATD.md tool reference from the source** (the registration calls already carry `Description` + `InputSchema`). This makes drift structurally impossible and is a perfect demonstration of ATD's own "doc-code co-evolution" principle applied to ATD itself.

### 7.4 Resolve the half-done renames (a dogfooding exercise)

`discover → map`, `verify → check`, `atd_cli → atd`. Update the STABLE atoms (`api_atd_serve_discover`, `api_atd_serve_verify`), the `@spec-link` tags in `mcp_tools.go:219,438`, the issues backlog, and `.claude/settings.local.json`. Doing this *through the ATD workflow* (Skill B, with `atd_update` + `weave` + reference propagation) would be the most credible possible proof that the system works.

### 7.5 Minimal status enforcement (ISS-010, scoped)

Full status semantics can wait (deferred per scope). But a *minimal* version unlocks Skill B's guarantees cheaply: have `atd_update` emit a warning (and require an explicit `--force`/confirmation arg) when the target is `status=STABLE` and `layer=BUSINESS`. That single guard turns "the agent is supposed to ask" into "the tooling requires the agent to ask."

### 7.6 One cheap orientation tool: `atd_doctor` (or extend `atd_env`)

A single MCP call an agent can run at session start that returns: active project, docs atom count vs files on disk, orphan count, last index mtime, provider reachability, and **the canonical tool list with signatures**. Today the agent spends 5 calls to build a model that may still be wrong (§3.7). One deterministic probe would make every subsequent decision better-grounded.

### 7.7 Expose the pipelines via MCP (ISS-031 / 050 / 039)

Cold-start, full-audit, and reconcile are shell-only. If MCP-only agents are a target audience, these need MCP wrappers — otherwise Skill E is shell-only by necessity and the "agent governs everything" vision has a hole where governance *begins*.

### 7.8 Type-specific templates (ISS-038)

`API` atoms want request/response schemas; `USECASE` wants `## WORKFLOW`; `USER_STORY` wants `## ACCEPTANCE CRITERIA`. Push these into `atd_update` (so every agent benefits), and mirror them in Skill A's templates.

### 7.9 Fix the issues tracker drift

11 files on disk, 52 referenced in README. Either restore the missing files or prune the README. An agent told "check the issues for known risks" currently gets a misleading index — which is exactly the class of problem ATD exists to prevent.

---

## 8. Prioritized Action List

| # | Action | Type | Effort | Unblocks |
|---|---|---|---|---|
| 1 | Remove phantom `atd_discover` from `.agent/rules/ATD.md`; correct tool names | doc fix | trivial | every agent turn |
| 2 | Reconcile tool count (README 19 → 25); prune issues/README drift | doc fix | trivial | onboarding trust |
| 3 | `atd_check` file-mode dedup + index-db relocation + honest empty | code | low | silent-failure class |
| 4 | Route all id-entry points through `ResolveAtom`; loud failure on miss | code | low | §3.1 (worst class) |
| 5 | Slim always-on rules to invariants only (≤60 lines) | doc refactor | low | context budget |
| 6 | **Ship Skill A (`create-atom`) + Skill C (`pre-code-change`)** | skill | medium | daily workflows |
| 7 | Generate ATD.md tool reference from `mcp_tools.go` | tooling | medium | drift-proofing |
| 8 | Resolve `discover→map` / `verify→check` renames end-to-end via ATD | dogfood | medium | trust + credibility |
| 9 | Minimal `status=STABLE + BUSINESS` update guard (ISS-010 scoped) | code | medium | Skill B guarantees |
| 10 | Unify link index: `trace` consumes live crawl, not `linked_codes` | code | medium | §3.2 cross-tool agreement |
| 11 | Fix `trace summary=true` target primacy + `map` confirm schema | code | low-med | §3.3 LLM-tool quality |
| 12 | **Ship Skill B (`alter-atom`)** once #4/#9/#10/#11 land | skill | medium | governance workflow |
| 13 | `atd_doctor` orientation probe (or extend `atd_env`) | code | medium | session bootstrap |
| 14 | Expose cold-start / full-audit / reconcile via MCP | code | high | MCP-only agents, Skill E |

### Dogfooding cleanup (§4) — do this *through* the ATD workflow, not by hand

| # | Action | Type | Effort | Unblocks |
|---|---|---|---|---|
| D1 | Create the 10 missing `api_atd_serve_*` atoms (check, map, env, config, 3×workspace, 3×heatmap) from the live `mcp_tools.go` registrations | atom authoring | low | graph completeness (§4 #2) |
| D2 | Retire/rewrite the 2 phantom atoms: merge `api_atd_serve_discover`→ document `atd_map`; `api_atd_serve_verify`→ document `atd_check`; update the `@spec-link` tags at `mcp_tools.go:219,438` in the same pass | atom + code | low | naming trust (§4 #1, #3) |
| D3 | Instantiate `CONTRACT` + `VISION` for the ATD project; have `atd_lint` enforce their presence | atom + lint | low-med | governance gates (§4 #4) |
| D4 | Reconcile non-canonical types (`SERVICE`/`USAGE`/`SPECIFICATION`) and the `CUSTOMER` layer against §1.3/§1.5, or formally adopt them and update the spec + `.atd` config | atom + spec | medium | type/layer integrity (§4 #5–#7) |
| D5 | Add lint/audit rules that catch the §4 classes: atom documents a tool not in the registered set; tool registered with no atom; STABLE atom whose schema ≠ the code's schema | lint/audit | medium | makes §4 self-preventing |

Items 1–5 are preliminary-research-safe (docs + trivial code). Items 6 and 12 are the skills. Items 7–11 are the substrate fixes that make Skill B's guarantees real. Item 14 is the larger investment. **D1–D3 are low-effort and double as the integration test ATD currently lacks** — they should run alongside items 1–2 as the first visible proof the philosophy works.

---

## 9. Conclusion

ATD's *philosophy* is sound and its *deterministic tooling* is genuinely good. What hurts agent usability today is **not the absence of skills** — it is that (a) the documentation the agent is forced to trust is partly wrong (phantom tools, wrong counts, drifted signatures), (b) several tools return **confidently wrong answers** in common cases, which is the one failure mode an LLM agent cannot self-correct from, and (c) as §4 documents, **ATD's own corpus fails ten different self-governance rules** — including two `STABLE` atoms that describe tools which don't exist. If ATD cannot keep its own house in order, no agent can be expected to trust the graph it produces.

**Skills are the right container** for the procedural workflows the user identified (create / alter / pre-code-change), and adopting them lets the always-on ruleset shrink to just its invariants. That is a clear win regardless. But skills are a **layer on top of the substrate**; they cannot guarantee correctness while the substrate silently lies or the governing corpus is itself unreliable. The highest-leverage preliminary work is therefore:

> Fix the four silent-zero paths, make `trace`/`check`/`query` agree, delete the phantom tool from the ruleset, generate the tool reference from source, and run the §4 dogfooding cleanup (D1–D3) **through** the ATD workflow. *Then* build Skills A and C, and Skill B once its dependencies hold.

This sequencing gives ATD something it currently lacks proof of: a demonstration that its own doc-code co-evolution catches and fixes drift in its own product. Resolving the `discover`/`verify` renames *via `atd_update` + `weave` + `lint`* — not by editing Markdown — would be that proof, and would do more for adoption than any feature in the backlog.

---

## Appendix A — Evidence Index

All claims above trace to one of:

- **Source:** `atd/cmd/atd/cmd/mcp_tools.go` (tool registrations + the `discover`/`verify` spec-link leftovers at `:219, :438`), `check_coverage.go`, `pkg/exploration/*`, `pkg/prompt/*`.
- **Always-on rules:** `.agent/rules/ATD.md` (phantom `atd_discover` at `:73-76, :91, :177-180`; the good `has_customer_origin` rule at `:98-103`).
- **Reference manual:** `ATD.md` (tool-count claim `:357`; `assemble`/`check` signature drift `:507-513`, `:491-495`).
- **README:** tool-count `:16, :63`; 52 issues referenced vs 11 on disk.
- **Field reports:** `atd_feedback_2026-07-07.md` (incl. 2026-07-12 addendum on bare-id prevalence) and `atd_feedback_evaluation_2026-07-13.md` (code-level root causes + proposed fixes for all 8 defects).
- **Issues backlog:** ISS-010 (status ignored), ISS-031/050 (cold-start/audit not in MCP), ISS-038 (type templates), ISS-039 (reconcile not in MCP).
- **Dogfooding evidence (§4):** `docs/api_atd_serve_discover.atom.md` + `docs/api_atd_serve_verify.atom.md` (STABLE atoms documenting nonexistent tools with wrong schemas); 10 missing `api_atd_serve_*` atoms (no atom for `atd_check`, `atd_map`, `atd_env`, `atd_config`, `atd_workspace_*`, `atd_heatmap*`); `mcp_tools.go:219,438` (stale `@spec-link` tags); empty CONTRACT/VISION grep; type/layer distribution showing non-canonical `SERVICE`/`USAGE`/`SPECIFICATION` types and `CUSTOMER` layer (ISS-075, ISS-076); `.atd` `bloating_factor.type_overrides` referencing `USECASE`/`SPECIFICATION`; 1 atom carrying the stale `linked_codes` frontmatter field.

## Appendix B — Real MCP tool list (canonical, 25)

```
Deterministic: atd_query  atd_crawl  atd_weave  atd_update  atd_roadmap
               atd_stats  atd_check  atd_assemble  atd_trace  atd_test_links
               atd_lint
LLM-backed:    atd_dissect  atd_index  atd_search  atd_audit  atd_recon  atd_map
Config/diag:   atd_env  atd_config
Workspace:     atd_workspace_list  atd_workspace_use  atd_workspace_stats
Heatmap:       atd_heatmap  atd_heatmap_code  atd_heatmap_project
```

CLI-only (not MCP): `init`, `workspace init|add`, `compare`, `congruence`, `continue`, `fix`, `generate`, `reconcile`, `verify`, `version`, `webui`, plus `atd-cold-start.sh` and `atd-full-audit.sh`.
