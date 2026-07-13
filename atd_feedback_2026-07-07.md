# ATD feedback — field report from the battleui → upsilonhub migration

**Date:** 2026-07-07
**Workspace:** `upsilonumbrella` (12 projects), active project `upsilonhub`
**Reporter:** Bastien (findings collected by Claude during migration Phases 1–5)

Every item below was reproduced first-hand on 2026-07-07 against the current
workspace state (post-Phase-5), via the ATD MCP tools. Exact calls are given so
you can replay them. Nothing here is hearsay; tools I did not exercise are
listed at the end as untested.

Context that makes us a good stress test: `upsilonhub` (Go) has **zero local
atoms** — every one of its ~67 linked atoms lives in another project
(`upsilonapi`, `upsilonbattle`, `upsilontypes`, `battleui`, umbrella `shared`),
so cross-project prefixed links (`[[upsilonapi:api_shop_purchase]]`) are our
bread and butter.

---

## What works

### 1. `atd_workspace_list`
Correct: 12 projects, right paths, active project `upsilonhub`. ✔

### 2. `atd_check full:true` — cross-project attribution now works
`atd_check {full: true}` on `upsilonhub` returned **67 atoms, 67 with impl,
52 with tests**, all with correct `project:atom` prefixes across five
projects. I independently counted the unique `[[project:atom]]` links in the
module with grep: also exactly 67. Per-atom impl/test counts spot-checked
(`upsilonapi:api_shop_purchase` 3 impl / 1 test) match the tags in the code.

Note: on 2026-07-04/05 (Phases 1–2, right after the MCP server restart) the
same prefixed links reported `NO_IMPL` in scoped checks, which is why the
migration fell back to mechanical grep validation. Full-scope attribution is
now correct — thank you, that part appears fixed.

### 3. `atd_check file:<path>` — works, with a dedup bug (see below)
`atd_check {file: "internal/gateway/shop.go"}` resolves the file's prefixed
atoms correctly (relative path accepted, resolved against the active project).

### 4. `atd_test_links atom:<prefixed id>`
`atd_test_links {atom: "upsilonapi:api_shop_purchase"}` → correct (3 impl,
1 test). ✔ — but only with the fully-qualified form (see broken §3).

### 5. `atd_query` — the best cross-project citizen right now
`atd_query {field: "id", search: "api_shop_purchase"}` returned the full
frontmatter **plus `linked_codes` pointing at the new Go files**
(`internal/gateway/shop.go:50`, `shop.go:150`,
`internal/platform/economy/pg_inventory.go:21`) — correct attribution across
the project boundary, and it resolved a **bare** id without a prefix. This is
the resolution behaviour we'd like everywhere.

### 6. `atd_trace <atom>` (raw JSON)
Resolves the right atom, ancestry and dependents across projects, includes
full context bodies. Genuinely useful for impact assessment — but its code
links and metrics disagree with `atd_check`/`atd_query` (see broken §5).

### 7. `atd_search grep:`
`atd_search {grep: "MaskBoardState", paths_only: true}` → exactly the four
files that mention it, correctly scoped to the active project. ✔

---

## What doesn't work

### 1. `atd_check` diff mode finds nothing — our per-phase gate is unusable
Every diff-scoped variant returns **"No atoms found for the specified
scope."**, even when the diff is saturated with links:

- `atd_check {base: "HEAD~1"}` — the Phase 5 commit: 48 files, 67 linked atoms → nothing.
- `atd_check {base: "HEAD~1", target: "HEAD"}` — same → nothing.
- `atd_check {}` with a real uncommitted one-line edit to a `@spec-link`-tagged
  file (`internal/gateway/shop.go`) → nothing.

Since full-mode attribution works, the link parser is fine — the diff
computation looks broken (guess: git context, possibly submodule-related; our
projects are git submodules of the umbrella). **This is our highest-priority
ask**: migration acceptance gate 4 is "atd_check for the phase's re-anchored
atoms", and we currently substitute a hand-rolled grep.

### 2. `atd_test_links` with a bare atom id silently reports NO_IMPL
`atd_test_links {atom: "api_shop_purchase"}` → `NO_IMPL` (false negative).
`atd_test_links {atom: "upsilonapi:api_shop_purchase"}` → OK.
A bare id should either resolve workspace-wide (like `atd_query` does) or
error with "did you mean upsilonapi:api_shop_purchase" — a silent zero is the
worst outcome, it reads as missing coverage.

### 3. `atd_check file:` duplicates rows
For `internal/gateway/shop.go`, `upsilonapi:api_shop_browse` and
`upsilonapi:api_shop_purchase` each appear **twice** in the table and the
summary counts 6 atoms for 4 distinct ones. Probably one row per tag
occurrence instead of per atom.

### 4. `atd_trace summary:true` narrates the wrong atom
`atd_trace {atom: "upsilonapi:api_shop_purchase", summary: true}` produced a
narrative that literally says "the target atom, `mechanic_shop_inventory_system`"
— the parent, not the requested atom. The bare-id call gave a similar
shop-system-in-general narrative. The raw (non-summary) call resolves the
target correctly, so this is the summarisation step losing the target.

### 5. `atd_trace` link data contradicts `atd_check`/`atd_query`
For the same atom (`api_shop_purchase`), raw trace reports:
- `code_links`: only two battleui Vue components; `test_links`: only the
  battleui Playwright spec — the three Go `@spec-link` sites and the Go
  `@test-link` that `atd_check`/`atd_query` correctly find are absent;
- `total_code_files: 2`, `implementation_rate: 0.5`, and the warning
  "Architecture atom has no Implementation dependents or direct @spec-link" —
  all contradicted by `atd_check` (3 impl / 1 test, status OK).

So trace uses a different (stale or single-project) link index than
check/query. Whichever is canonical, they should agree.

### 6. `atd_stats` numbers are internally inconsistent
- Project scope (`atd_stats {}`, project `upsilonhub`): `total_atoms: 494` —
  for a project with **zero** local atoms. Workspace scope
  (`atd_stats {workspace: true}`): `total_atoms: 302`. A project reporting
  more atoms than its whole workspace can't be right. On-disk ground truth:
  **311** `*.atom.md` files in the workspace (find, node_modules excluded).
- Project scope also reports `implemented_stable_count: 254` >
  `implemented_total_count: 190` — a "stable subset" larger than the total.
We currently trust none of the stats output.

### 7. Semantic search: empty result, no diagnostics, index artifact in the repo
`atd_search {query: "...", limit: 3}` returned an empty "Top 3 Semantic
Matches" block — no error, no "index missing, run atd_index" hint (the tool
description says an index is required; presumably none exists here).
Side effect observed right after: an untracked **`upsilonhub/docs/.atd_index.db`**
appeared — the server created a `docs/` directory (which didn't exist; this
project keeps no local atoms) inside our working tree. Index artifacts inside
the repo working tree will end up committed by accident; please store them
under a cache dir (or at minimum document a .gitignore requirement).

### 8. `atd_map` — inconsistent path handling and unusable confirm output
- `atd_map {file: "internal/gateway/shop.go", atom: ...}` → "failed to read
  file … no such file or directory", while `atd_check` accepts the exact same
  relative path. Path resolution differs between tools; absolute paths work.
- With the absolute path, confirm mode against `upsilonapi:api_shop_purchase`
  on `internal/gateway/shop.go` — a file that carries that very `@spec-link`
  and whose behaviour passes the ported PHP test suite — returned:
  `{"Confidence": 0, "Mismatches": "Yes, there are several mismatches and
  issues with the code that prevent it from logically implementing the Target
  Atom."}`
  Confidence 0 with a one-sentence yes/no and no rationale. (Fairness note:
  the transactional core lives behind a service seam in another file, so a
  strict single-file reading could argue partial implementation — but then the
  mismatches should be listed; the current output is unusable either way, and
  looks like a raw LLM yes/no answer leaking into a field meant for details.)

---

## Not tested (no claims made)

`atd_index` (didn't want to kick off an embedding build mid-migration),
semantic search post-index, `atd_check semantic:true` (token cost),
`atd_lint`, `atd_audit`, `atd_weave`, `atd_assemble`, `atd_dissect`,
`atd_crawl`, `atd_recon`, `atd_roadmap`, `atd_heatmap*`, `atd_env`,
`atd_config`, `atd_update`, `atd_workspace_use`.

---

## Priority from the migration's point of view

1. **Diff-mode `atd_check`** (§1) — it's the phase acceptance gate; we're
   grepping by hand instead.
2. **One link index for check/query/trace** (§5) — trace drives impact
   assessment before we delete PHP at Phase 6; wrong counts there are risky.
3. **Stats correctness** (§6) — wanted for the Phase 6 cutover report.
4. Bare-id resolution or loud failure (§2), file-mode dedup (§3),
   summary-target fix (§4), index artifact location + missing-index message
   (§7), map path/output (§8).

Happy to re-run any of these against a patched server — every repro above is a
single MCP call in this workspace.

---

## Addendum — 2026-07-12, from the ISS-107 edge-case audit

**Reporter:** Bastien (findings collected by Claude during the ISS-107 audit of
the 55 `upsiloncli/tests/scenarios/edge_*.js` CI scenarios). Same workspace,
five months of scenario/link drift later. Two items, both reinforcing §2.

### A. §2 (bare-id → silent NO_IMPL) is the single highest-yield defect we hit — please prioritise it above its current #4 slot
Across the 55-scenario audit, "author wrote the **bare** id, it resolved to a
different, zero-coverage phantom stub than the module-qualified form that
carries the real links" was the most common single defect class — recurred in
~15 of 55 scenarios. It is actively self-propagating: because the bare form
returns a clean zero (not an error), authors copy-paste it between files and
nothing flags it, while the qualified form sits right there in the same atom's
`test_links`. Fresh first-hand repros in this workspace:
- `atd_trace {atom: "rule_password_policy"}` → 0-coverage phantom;
  `atd_trace {atom: "upsilonapi:rule_password_policy"}` → 100% impl/test, and
  its own `test_links` already lists files that tag it with the *bare* id. The
  bare form appears in ~15 files repo-wide (incl. the atom's own doc).
- Same shape confirmed for `req_ui_session_timeout`
  (real: `upsilonbattleui:req_ui_session_timeout`),
  `mech_character_reroll` (real: `upsilonbattle:mech_character_reroll`),
  `api_leaderboard`, `api_websocket`, and ~a dozen others.

The §2 ask stands and we'd raise its priority: bare ids should resolve
workspace-wide (like `atd_query` already does) **or** fail loudly with a "did
you mean `upsilonapi:rule_password_policy`?" — a silent zero is worst-case,
it reads as genuinely missing coverage and trains authors to leave it wrong.

### B. New: a `requirement_`-prefixed id resolves to *nothing* (not-found), also silently
`atd_trace {atom: "requirement_req_ui_session_timeout"}` → not found at all
(distinct from §2's phantom-with-zero-coverage: here there is no object).
This looks like an author pasting the atom's **layer/type** (`requirement`)
in front of its id and the tool neither stripping the type prefix nor erroring
helpfully. Found in the CR-16 mapping (`CI.md`, `tests/ci_report.sh`,
`e2e_session_timeout.js` — since corrected in-repo to the real backend atom
`req_security_token_ttl`). Same remedy as §2: resolve or fail loudly, never
return a silent empty.

Both are single-MCP-call repros in this workspace; happy to re-run against a
patched server.
