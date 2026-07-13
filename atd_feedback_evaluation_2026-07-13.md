# ATD feedback evaluation — code investigation & fix proposals

**Date:** 2026-07-13
**Evaluator:** Claude (code investigation)
**Source report:** `atd_feedback_2026-07-07.md` (+ 2026-07-12 addendum)
**Codebase:** `atd/` (Go CLI + MCP server), reproduced against `upsilon-hub/` workspace copy
**Binary used:** built from `atd/cmd/atd` at HEAD (`2025556`)

Every finding below was checked against the source. Where I could drive the CLI
directly against `upsilon-hub/upsilonhub` I marked it **REPRODUCED** with the live
output; where the defect only manifests through the MCP server's working-directory
context I marked it **CONFIRMED (by code)** and explain why my standalone run
didn't trip it.

> **Resolution update — 2026-07-13.** **All eight findings + both addenda are now
> FIXED and verified**, plus one deeper bug (§9) surfaced during the work.
> Batch 1: §2 + Add A/B (id canonicalization / prefix stripping / loud failure), §3
> (file-mode dedup), §7 (semantic-search honesty + index moved out of the tree).
> Batch 2: §1 (diff-mode path base), §5 (trace consumes the live link index), §6
> (stats population + inversion), §8 (map path + recon prompt), §4 (trace-summary
> target primacy). §9 (new): workspace "current project" was always mis-detected as
> the root project. Details inline under each finding; roll-up in **"Resolution
> status"** at the end. Changes live on branch `fix/field-report-quickwins`.

---

## Summary

| # | Report item | Verdict | Status | Root cause | Fix effort |
|---|---|---|---|---|---|
| §1 | `atd_check` diff mode finds nothing | **CONFIRMED (by code)** | ✅ **FIXED** | `git diff` paths (repo-root-relative) vs `SpecLink.FilePath` (ProjectRoot-relative) mismatch under MCP | Medium |
| §2 | bare atom id → silent `NO_IMPL` | **REPRODUCED** | ✅ **FIXED** | check/test_links never canonicalize the input id through the resolver | Low |
| §3 | `atd_check file:` duplicates rows | **REPRODUCED** | ✅ **FIXED** | file mode has no per-atom dedup (one row per tag occurrence) | Trivial |
| §4 | `atd_trace summary:true` narrates wrong atom | **CONFIRMED (by code)** | ✅ **FIXED** | prompt gives target no primacy; no guard when target brief is empty | Low |
| §5 | `atd_trace` link data contradicts check/query | **REPRODUCED** | ✅ **FIXED** | three tools derive "code links" from three sources; `Implementations` conflates frontmatter `linked_codes` with live crawl | Medium |
| §6 | `atd_stats` internally inconsistent | **REPRODUCED** | ✅ **FIXED** | project scope counts cross-project atoms dragged in by `ResolveAtom`; `implemented_stable_count` counts *non-orphans*, not *implemented* | Medium |
| §7 | semantic search: silent empty + index artifact in repo | **REPRODUCED** | ✅ **FIXED** | error swallowed by grep fallback; DB path is `docs/.atd_index.db` inside the working tree | Low |
| §8 | `atd_map` path handling + unusable confirm output | **REPRODUCED** | ✅ **FIXED** | `runMap` reads file relative to cwd (not ProjectRoot); confirm returns raw LLM JSON | Low/Medium |
| Add A | bare-id is the highest-yield defect | same as §2 | ✅ **FIXED** | — | Low |
| Add B | `requirement_`-prefixed id → silent not-found | **CONFIRMED (by code)** | ✅ **FIXED** | type/layer prefix not stripped, no loud failure | Low |

**Cross-cutting theme:** four of the eight (§1, §5, §6, §8) are the *same class of
bug* — a tool mixes two coordinate systems (git-root vs ProjectRoot vs cwd) or two
link indexes (live crawl vs persisted `linked_codes`). Fixing the shared substrate
(one path base, one link index) resolves most of the report.

---

## §2 + Addendum A — bare id → silent `NO_IMPL` *(highest priority per reporter)*

> ✅ **FIXED (2026-07-13).** `runCoverageCheck`'s `--atom` branch and `runTrace` now
> canonicalize the id through the resolver before lookup, via new
> `Explorer.CanonicalAtomID` (+ `SuggestAtomID` for "did you mean"). On failure they
> return a **loud** `atom '<id>' not found in workspace (did you mean '<x>'?)` instead
> of a clean zero row. `atd_test_links` and the MCP `atd_trace`/`atd_check` handlers
> route through the same functions, so they inherit the fix.
> Verified: `check --atom api_shop_purchase` → `upsilonapi:api_shop_purchase 3/2/OK`;
> `check --atom totally_bogus_xyz` → `Error: … not found in workspace` (exit 1).
> Files: `pkg/exploration/exploration.go` (new `CanonicalAtomID`/`SuggestAtomID`,
> `Trace` entry), `cmd/atd/cmd/check_coverage.go` (atom branch).

**REPRODUCED.** From `upsilon-hub/upsilonhub`:

```
$ atd check --atom api_shop_purchase
api_shop_purchase              0   0   -   NO_IMPL      ← false negative
$ atd check --atom upsilonapi:api_shop_purchase
upsilonapi:api_shop_purchase   3   2   -   OK
```

**Root cause.** `atd_test_links` and `atd_check --atom` both call
`runCoverageCheck("atom", atomID, …)` (`cmd/atd/cmd/mcp_tools.go:319`,
`check_coverage.go:68`). In atom mode the id is used verbatim:

```go
// check_coverage.go:68
case atomID != "":
    atomIDs = []string{atomID}          // never resolved
```

Coverage is then keyed off `SpecLink.AtomID` (`check_coverage.go:126`), which the
link crawler stores in **canonical prefixed** form (`upsilonapi:api_shop_purchase`)
— see `exploration.go:132-134`. A bare key therefore matches nothing and reads as
zero coverage. The workspace `Resolver` already resolves bare ids workspace-wide
(`resolver.go:113-139`) — `atd_query` uses it, which is why query "just works" — but
check/test_links skip it.

**Fix.** Canonicalize the id before lookup, in the `atomID != ""` branch:

```go
case atomID != "":
    mode = "atom"
    if node, err := explorer.ResolveAtom(atomID); err == nil {
        atomID = node.ID              // ResolveAtom returns canonical prefixed id
    }
    atomIDs = []string{atomID}
```

`ResolveAtom` (`exploration.go:687`) already returns the canonical id and errors on
unknown project. If resolution fails, return a **loud** message
(`atom '%s' not found; did you mean upsilonapi:%s?`) instead of a clean zero row —
the reporter is explicit that a silent zero is the worst outcome and is
self-propagating (Addendum A: ~15 of 55 scenarios). Low effort, high value.

Apply the same one-liner wherever an atom id enters from the user: `atd_trace`
(`trace.go:47` → `explorer.Trace`), which is the phantom-stub case in Addendum A.

---

## Addendum B — `requirement_`-prefixed id resolves to nothing, silently

> ✅ **FIXED (2026-07-13).** `workspace.Resolver.Resolve` now retries once with a
> stripped leading type/layer token (`requirement_`, `rule_`, `api_`, `mech(anic)_`,
> `ui_`, `entity_`, `module_`, `domain_`, `us_`/`usecase_`/`user_story_`, `vision_`,
> `contract_`, `workflow_`) when the bare id misses; still returns loud
> `ErrAtomNotFound` if nothing matches after stripping. Unit tests added in
> `pkg/workspace/resolver_test.go`. Verified:
> `trace requirement_req_ui_session_timeout` → strips `requirement` and resolves to
> the real `upsilonbattleui:req_ui_session_timeout`. (Note: that turned out to be a
> genuine separate atom whose own parent is `shared:req_security_token_ttl`; the
> mechanical strip-and-retry correctly lands on the real atom.)
> File: `pkg/workspace/resolver.go`.

**CONFIRMED (by code).** `atd_trace {atom: "requirement_req_ui_session_timeout"}`
goes to `Explorer.Trace` → `e.Graph.Atoms[targetID]` (`exploration.go:414`). The
string contains no `:` so `ParseReference` treats the whole thing as a bare atom id
(`resolver.go:60-65`); no atom has that literal id, so it's unresolved and Trace
returns a bare `atom not found` — but the MCP layer surfaces it as an empty result.

**Fix.** In the resolver, if a bare id fails and its leading token is a known
type/layer word (`requirement`, `rule`, `api`, `mech`/`mechanic`, `ui`, `entity`,
`module`, `us`/`user_story`, …), retry once with the prefix stripped and, on a hit,
return it with a `resolution` note (`interpreted 'requirement_x' as 'x'`). Otherwise
fail loudly. Same remedy family as §2: **resolve or shout, never return silent
empty.**

---

## §3 — `atd_check file:` duplicates rows

> ✅ **FIXED (2026-07-13).** File mode now uses the same per-atom `seen` dedup as diff
> mode. Verified: `check --file internal/gateway/shop.go` → 4 distinct rows, summary
> "4 atoms | 4 with impl | 4 with tests" (was 6). File:
> `cmd/atd/cmd/check_coverage.go` (file branch).

**REPRODUCED.** `internal/gateway/shop.go` tags `api_shop_browse` and
`api_shop_purchase` twice each (lines 28/149 and 50/150):

```
$ atd check --file internal/gateway/shop.go
upsilonapi:api_shop_browse       4  1  -  OK
upsilonapi:api_shop_purchase     3  2  -  OK
upsilonapi:mechanic_shop_...     2  1  -  OK
upsilonapi:api_inventory_list    5  1  -  OK
upsilonapi:api_shop_browse       4  1  -  OK   ← dup
upsilonapi:api_shop_purchase     3  2  -  OK   ← dup
Summary: 6 atoms | 6 with impl | 6 with tests   ← 4 distinct
```

**Root cause.** File mode appends one id per matching tag with no dedup set —
unlike diff mode, which already uses a `seen` map (`check_coverage.go:86-97`):

```go
// check_coverage.go:73-78  (file mode — no dedup)
for _, sl := range explorer.SpecLinks {
    if sl.FilePath == filePath {
        atomIDs = append(atomIDs, sl.AtomID)
    }
}
```

**Fix (trivial).** Reuse the same dedup pattern:

```go
seen := map[string]bool{}
for _, sl := range explorer.SpecLinks {
    if sl.FilePath == filePath && !seen[sl.AtomID] {
        atomIDs = append(atomIDs, sl.AtomID)
        seen[sl.AtomID] = true
    }
}
```

(The per-atom impl/test *counts* are already correct — 3/2 — because
`buildCoverageReport` counts all links for the id; only the *rows* are duplicated.)

---

## §1 — `atd_check` diff mode finds nothing

**CONFIRMED (by code).** Standalone from inside `upsilon-hub/upsilonhub`, diff mode
*works* (an uncommitted edit to `shop.go` correctly surfaced its atoms). It breaks
in the reporter's setup because of a **coordinate-system mismatch that only appears
when the process cwd ≠ the active project root** — i.e. the MCP server driving a
workspace project.

**Root cause.** `diffChangedFiles` shells out to git with no working directory set:

```go
// check_coverage.go:244
cmd := exec.Command("git", args...)     // runs in the server's cwd
```

`git diff --name-only` returns paths relative to **the git repo that owns cwd**. The
diff loop then matches those against `SpecLink.FilePath` (`check_coverage.go:92-93`),
which the crawler stores **relative to `config.ProjectRoot()`** (the active project)
via `filepath.Rel(e.ProjectRoot, path)` (`exploration.go:384`, `152-156`).

- Standalone CLI in the project dir: cwd == git root == ProjectRoot → strings match → works.
- MCP server (cwd = umbrella root, active project = `upsilonhub` submodule):
  git runs against the **umbrella** repo, where a submodule shows up as a single
  gitlink entry (`upsilonhub`), never its internal `internal/gateway/shop.go`. Even
  a committed change surfaces no code-file paths that match the ProjectRoot-relative
  `SpecLink` keys → **"No atoms found."**

**Fix.** Run git in the project and normalize both sides to the same base:

```go
func diffChangedFiles(gitArgs []string) (code, atoms []string, err error) {
    root := config.ProjectRoot()
    args := append([]string{"-C", root, "diff", "--name-only"}, gitArgs...)
    cmd := exec.Command("git", args...)
    // …
    // git -C <root> yields paths relative to the repo that contains <root>;
    // if that repo root ≠ ProjectRoot, rebase each path onto ProjectRoot before
    // comparing to SpecLink.FilePath.
}
```

Concretely: resolve `git -C <root> rev-parse --show-toplevel`, and for each returned
path compute its path *relative to ProjectRoot* so it matches the crawler's keys.
This also makes diff mode submodule-correct (git inside the submodule sees the real
file paths). This is the reporter's #1 ask (phase acceptance gate); medium effort,
mostly path bookkeeping + a submodule test.

---

## §5 — `atd_trace` link data contradicts check/query

**REPRODUCED.** For the same atom, trace and check disagree:

```
$ atd check --atom upsilonapi:api_shop_purchase
… 3 impl / 2 tests / OK
$ atd trace upsilonapi:api_shop_purchase
"implementation_rate": 0.5, "total_code_files": 2, "total_tests": 2
```

Check counts **3 @spec-link sites**; trace reports **2 code files** and an
`implementation_rate` of 0.5 with the "Architecture atom has no Implementation
dependents…" warning. In the reporter's live server the divergence was starker
(trace showed *battleui* Vue/Playwright files that check never saw) — that is the
same bug one step worse, and the mechanism explains both.

**Root cause — two independent link derivations:**

1. **check** iterates the **live crawl** result `explorer.SpecLinks` /
   `explorer.TestLinks` (`check_coverage.go:126-131`) — authoritative, cross-project.
2. **trace** reads `node.Implementations` (`exploration.go:487`), a field whose JSON
   tag is **`linked_codes`** (`atom/parse.go:40`): it is populated from the atom's
   **frontmatter** at parse time and *then* has live crawl hits appended
   (`exploration.go:151`). So it is a **mix of persisted (possibly stale) frontmatter
   and live data**, deduped by *file* (loses the per-site count → "2 files" vs "3
   links"). When the frontmatter still lists old battleui paths and the current
   project only crawls Go, trace shows the battleui ghosts and misses/undercounts
   the Go reality.
3. **query** returns the raw `linked_codes` frontmatter, which is why it looked
   correct for the reporter (that project's frontmatter had been rewritten to the Go
   files) yet trace, keyed on the same field but computed in a different project
   context, disagreed.

`implementation_rate: 0.5` compounds it: for an ARCHITECTURE atom the health
denominator counts the atom + implementation dependents (`exploration.go:590-618`),
a different quantity than check's "does this atom have ≥1 @spec-link", so the two
numbers are not comparable even when both are "right".

**Fix.** Make the **live crawl the single source of truth** and have trace consume
it, exactly as check does:

- In `Explorer.Trace`, derive code/test links from `e.SpecLinks` / `e.TestLinks`
  filtered by atom id (and its descendants), not from `node.Implementations`.
- Stop persisting/reading coverage from frontmatter `linked_codes` for
  *computation*; if you keep the field, treat it as a cached render regenerated from
  the crawl, never as an input to trace/query metrics.
- Report the same primitive check reports (distinct impl links, distinct test
  links) so trace's `total_code_files`/rate line up with `atd check --atom`.

This is the reporter's #2 priority (trace drives Phase-6 impact assessment before
PHP deletion). Medium effort; the payoff is check/query/trace finally agreeing.

---

## §6 — `atd_stats` numbers internally inconsistent

**REPRODUCED** (numbers differ from the report because ProjectRoot differs, but the
two defects are both present):

```
$ atd stats               # project scope, in upsilonhub (0 local atoms)
"total_atoms": 75         ← should be ~0; these are OTHER projects' atoms
"implemented_stable_count": 66, "implemented_total_count": 75
$ atd stats --workspace
"total_atoms": 302, "implemented_stable_count": 210, "implemented_total_count": 293
$ find . -name '*.atom.md' | wc -l   # under upsilonhub
0
```

**Two distinct root causes:**

**(a) Project scope counts cross-project atoms.** `runStats` (no `--workspace`) calls
`explorer.Load(false)`, whose link crawl calls `ResolveAtom` for every `@spec-link`,
and `ResolveAtom` **inserts the resolved cross-project atom into the graph**
(`exploration.go:699-711`). So a project with zero local atoms accumulates one graph
node per distinct atom its code references (75 here; the report's 494 is the same
mechanism from a broader ProjectRoot). `total_atoms` is then neither "local atoms"
nor "workspace atoms" — hence a project reporting *more* than its workspace.

*Fix:* count only atoms whose `FilePath` is under the project's own docs dir (or
that carry no cross-project prefix), or compute `total_atoms` from a docs-only walk
separate from the link-crawl graph.

**(b) `implemented_stable_count` counts non-orphans, not implemented atoms** — this
is the `stable > total` inversion the report saw. In `stats.go`:

```go
// stats.go:99-106
if node.Status == "STABLE" {
    stableCount++
    if explorer.IsOrphan(node) {
        report.OrphanCount++
    } else {
        report.ImplementedStableCount++   // ← "not an orphan" ≠ "implemented"
    }
}
```

`IsOrphan` returns false for many *unimplemented* atoms — BUSINESS-layer atoms
(`BusinessLayerException`, `exploration.go:353`), excluded types, and parents
"covered by proxy" (`exploration.go:358-369`). A STABLE business atom with **zero**
code is therefore counted as `implemented_stable`, while `implemented_total`
(`stats.go:94-97`) only counts `len(Implementations) > 0`. Enough STABLE-but-uncoded
business atoms and `implemented_stable_count` exceeds `implemented_total_count` — a
mathematical impossibility if the label were accurate.

*Fix:* gate on actual implementation, and rename for honesty:

```go
if node.Status == "STABLE" {
    stableCount++
    if explorer.IsOrphan(node) { report.OrphanCount++ }
    if len(node.Implementations) > 0 { report.ImplementedStableCount++ }
}
```

Reporter's #3 priority (Phase-6 cutover report). Medium effort.

---

## §7 — semantic search: silent empty + index artifact in the working tree

> ✅ **FIXED (2026-07-13).**
> - **Honest failures:** `pkg/exploration/search.go` now distinguishes a *missing/empty
>   index* (new `ErrIndexMissing` — db file absent, `no such table`, or 0 rows) from a
>   *provider offline* failure. Missing index → hard, loud error `no semantic index
>   found for this project — run 'atd index' first` (no silent grep fallback);
>   provider offline → still falls back to grep but **prints a notice** first. Genuine
>   empty result sets print `(0 results)` instead of a blank block. The MCP
>   `atd_search` handler surfaces these messages too.
> - **Index out of the tree:** new `config.IndexDBPath(docsDir)` returns
>   `$XDG_CACHE_HOME/atd/<sha256(docsDir)[:16]>/index.db` (fallback `~/.cache/atd/…`),
>   now used by every writer/reader — `atd index`, `search`, `map`, workspace
>   per-project search, the two MCP tools, **and** the webui handlers (the last two
>   webui refs were updated in a follow-up so reader and writer agree). `.gitignore`
>   also gained explicit `.atd_index.db` / `**/.atd_index.db` entries. A pre-`sql.Open`
>   `os.Stat` guard prevents the 0-byte-artifact-on-failure.
> Verified: with no index, `search --query "shop purchase"` → the loud error above and
> **no** `docs/.atd_index.db` created; after `atd index`, the db lands at
> `~/.cache/atd/<hash>/index.db` and search returns real matches.
> Files: `config/config.go` (new `IndexDBPath`), `pkg/exploration/search.go`,
> `cmd/atd/cmd/{search,index,map,mcp_tools}.go`, `pkg/webui/handlers.go`, `.gitignore`.

**REPRODUCED.** From `upsilon-hub/upsilonhub` (embedding provider was actually
reachable), a fresh search returned nothing *and* dropped a db file into the repo:

```
$ atd search --query "shop purchase" --limit 3
--- Top 3 Semantic Matches ---
                              ← empty, no diagnostic
$ ls -la docs/.atd_index.db
-rw-r--r-- … 0  docs/.atd_index.db   ← 0-byte artifact created inside the repo
```

**Two root causes:**

**(a) The "index missing" error is swallowed.** `Search` falls back to grep on *any*
semantic failure (`exploration/search.go:46-51`), and `SemanticSearch` opens the db
lazily — the missing `atom_index` table errors at query time
(`search.go:155-158`), the error is discarded, grep runs for the literal string
`"shop purchase"`, finds nothing, and the tool prints an empty match block. The CLI
`search` has a provider pre-check (`search.go:28-36`) but the **MCP `atd_search`
path calls `runSemanticSearch` directly** (`mcp_tools.go:393`) with no pre-check and
no "did you run atd_index?" hint.

*Fix:* distinguish "provider offline" (fall back to grep, say so) from "index
missing/empty" (return `no semantic index found — run atd_index` — do **not** silently
grep). Surface a one-line notice whenever a fallback happens, and print `(0 results —
index empty?)` rather than a blank block.

**(b) Index path is inside the working tree.** The db path is
`config.DocsDir() + "/.atd_index.db"` (`mcp_tools.go:387`, `search.go:52`,
`map.go:168`). `sql.Open`+query creates the file (and `docs/` if the driver can) in
the repo, so it lands in the working tree and is a `git add .` accident waiting to
happen.

*Fix:* store the index under a cache dir outside the tree — e.g.
`$XDG_CACHE_HOME/atd/<project-hash>/index.db` (or `.atd/cache/`), created on demand.
At minimum, add `**/.atd_index.db` to the shipped `.gitignore` and document it. Note
the driver also creates an empty file even on failure — guard with an existence
check before `sql.Open`, and clean up 0-byte artifacts.

---

## §8 — `atd_map` path handling + unusable confirm output

**REPRODUCED (both halves).**

**(a) Path resolution differs from check.** `runMap` reads the target with
`os.ReadFile(filePath)` (`map.go:32`) — **relative to the process cwd**. Check, by
contrast, matches against ProjectRoot-relative crawl keys. Standalone in the project
dir the relative path worked for me; under the MCP server (cwd = umbrella root) the
same relative path `internal/gateway/shop.go` is not found, exactly as reported,
while `atd_check` accepts it. Same cwd-vs-ProjectRoot split as §1/§8.

*Fix:* resolve `filePath` against `config.ProjectRoot()` when it is relative, before
`os.ReadFile`, so map and check accept identical inputs:

```go
if !filepath.IsAbs(filePath) {
    filePath = filepath.Join(config.ProjectRoot(), filePath)
}
```

**(b) Confirm output is a raw LLM yes/no.** `runMapConfirm` unmarshals the model
response and re-emits it verbatim (`map.go:84-89`); with the current `ReconFormat`
the model can return `{"Confidence": 0, "Mismatches": "Yes, there are several
mismatches…"}` — a boolean-ish sentence in a field meant for an itemized rationale.
Reproduced live: `"Confidence": 0` with a one-line `Mismatches`.

*Fix:* two parts. (1) Tighten `ReconFormat`/`ReconBuild` (`pkg/prompt/recon.go`) to
require a **list** of concrete mismatches (`"mismatches": [{"aspect","expected",
"found"}]`) and a calibrated confidence, and reject a bare yes/no. (2) In
`runMapConfirm`, validate the parsed result and, when `Confidence == 0` with no
itemized mismatches, return a "model returned no usable rationale — retry / widen to
the service seam" message rather than passing the empty verdict through. Also feed
confirm the **caller-declared** cross-file context (the report notes the
transactional core lives behind a service seam in another file) so single-file
readings stop producing false zeros.

---

## Suggested sequencing

1. ✅ **§2 / Addendum A+B — id canonicalization** (Low, highest reporter value) —
   **DONE 2026-07-13.** One resolver call at each user-id entry point (check-atom,
   test_links, trace) + loud failure. Kills the self-propagating silent-zero class.
2. ✅ **§3 — file-mode dedup** (Trivial) — **DONE 2026-07-13.** Copied the diff-mode
   `seen` map.
3. ✅ **§7 — index location + honest empty** (Low) — **DONE 2026-07-13.** DB moved to
   a cache dir outside the tree; missing-index error now loud, not swallowed.
4. **§1 — diff mode path base** (Medium). `git -C ProjectRoot` + rebase paths;
   add a submodule test. Unblocks the phase acceptance gate.
5. **§5 — one link index** (Medium). Trace consumes the live crawl, not
   `linked_codes`. Makes check/query/trace agree.
6. **§6 — stats semantics** (Medium). Docs-only atom count + implementation-gated
   stable count.
7. **§8 — map path + recon prompt** (Low/Medium). Path join + structured mismatches.

Items 1–3 are same-day, low-risk wins that clear the two defects the reporter called
"worst outcome"/self-propagating. Items 4–6 are the shared-substrate fixes (one path
base, one link index) that collapse the remaining four findings into two coherent
changes.

---

## Reproduction notes

- All CLI runs were from `upsilon-hub/upsilonhub/` using a binary built from
  `atd/cmd/atd` at HEAD.
- §1 and §8(a) did **not** fail in that standalone run because cwd coincided with the
  project root; both are cwd-vs-ProjectRoot bugs that surface through the MCP server's
  working directory. The code paths (`check_coverage.go:244`, `map.go:32`) are cited
  above; a faithful repro needs the MCP server with an active workspace project whose
  root differs from the server cwd.
- §4 was not driven end-to-end (needs the summary LLM); the root cause is read
  directly from `trace.go:52-95` + `pkg/prompt/trace_summary.go` and is lower
  confidence than the reproduced items — flagged as such.
- No source files were modified; the one uncommitted edit made to `shop.go` for the
  §1 probe was reverted, and the `docs/.atd_index.db` artifact from the §7 probe was
  removed.

---

## Resolution status (2026-07-13)

All eight findings + both addenda are fixed, in two batches, on branch
`fix/field-report-quickwins`. Batch 1 also merged as commit `3a94502`.

### Batch 1 — quick wins (§2 + Add A/B, §3, §7)

**Verified together, from `upsilon-hub/upsilonhub` with the rebuilt binary:**

```
check --atom api_shop_purchase   → upsilonapi:api_shop_purchase  3  2  -  OK   (was 0/0/NO_IMPL)
check --atom totally_bogus_xyz   → Error: atom '…' not found in workspace       (was silent 0-row)
check --file internal/gateway/shop.go → 4 distinct rows | "4 atoms"             (was 6)
trace requirement_req_ui_session_timeout → resolves to upsilonbattleui:req_ui_session_timeout
search --query "shop purchase"   → Error: no semantic index found — run 'atd index'  (was blank block)
  └─ docs/.atd_index.db NOT created inside the repo; index now at ~/.cache/atd/<hash>/index.db
```

**Build/test:** `go build ./...` and `go vet ./...` clean in both the `atd-tools`
root module and the `cmd/atd` module. `go test ./...` passes except two failures
confirmed **pre-existing** (present on stash/HEAD, untouched by this work):
`pkg/prompt` build error in `dissect_test.go` (`AuditCodeBuild` arity) and
`cmd/atd/cmd` `TestUpdateLinks` (unrelated `atd_update` link-rewrite). New
`pkg/workspace/resolver_test.go` cases pass.

**Files changed by the fixes:**
- `pkg/workspace/resolver.go` (+ `resolver_test.go`) — bare-id strip-and-retry.
- `pkg/exploration/exploration.go` — `CanonicalAtomID`, `SuggestAtomID`, resolved `Trace` entry.
- `cmd/atd/cmd/check_coverage.go` — atom-mode canonicalization + loud failure; file-mode dedup.
- `config/config.go` — `IndexDBPath` (cache-dir index location).
- `pkg/exploration/search.go` — `ErrIndexMissing`, honest fallback vs hard error.
- `cmd/atd/cmd/{search,index,map,mcp_tools}.go`, `pkg/webui/handlers.go` — use `IndexDBPath`; `(0 results)`.
- `.gitignore` — explicit `.atd_index.db` entries.

**Caveats worth noting for the maintainer:**
- The addendum narrative said the real atom for `requirement_req_ui_session_timeout`
  was `req_security_token_ttl`; the mechanical strip-and-retry instead lands on the
  genuinely-existing `upsilonbattleui:req_ui_session_timeout` (whose parent *is*
  `shared:req_security_token_ttl`). This is the correct behavior — the aside in the
  report referred to a separate hand-correction, not to what the tool should resolve.
- Moving the index location means any **already-built** `docs/.atd_index.db` files are
  now stale/ignored; a one-time `atd index` rebuild per project repopulates the cache
  dir. The large pre-existing `.atd_index.db` files under `upsilon-hub/*/docs` (dated
  ~09:22) can be deleted at leisure — they are no longer read.

### Batch 2 — shared-substrate + remaining (§1, §5, §6, §8, §4) + new §9

**Verified together, from the `upsilon-hub` workspace with the rebuilt binary:**

```
§1  diff-mode check, run with cwd ≠ ProjectRoot (umbrella root, active project
    upsilonhub) after touching a tagged file → now lists that file's atoms
    (was "No atoms found"). git now runs `-C ProjectRoot` and rebases paths.
§5  trace upsilonapi:api_shop_purchase → code/test links now sourced from the live
    crawl (same as check): 2 code files / 2 tests, and the bogus "Architecture atom
    has no Implementation dependents or direct @spec-link" warning is gone.
§6  stats (project upsilonhub, 0 local atoms) → total_atoms 0 (was 75);
    stats --workspace → implemented_stable 205 ≤ implemented_total 293 (inversion gone).
§8  map --file <relative> --atom … from cwd ≠ ProjectRoot → file now found
    (was "no such file"); confirm mode now returns itemized {aspect,expected,found}
    mismatches with a calibrated confidence, not a raw yes/no.
§4  trace --summary <bare id> → narrative now names and centres the target atom
    (was narrating a dependent/parent); empty-target guard added.
§9  stats/resolution for a project WITH local bare-tagged atoms (e.g. upsilonapi) →
    implemented_total 10 (was 0): the "current project" is now detected correctly.
```

**Batch 2 files changed:**
- `cmd/atd/cmd/check_coverage.go` (+ `check_coverage_test.go`) — §1: `git -C ProjectRoot` + path rebasing onto ProjectRoot; regression test for the nested/submodule case.
- `cmd/atd/cmd/map.go` (+ `map_test.go`), `pkg/prompt/recon.go` — §8: resolve relative path against ProjectRoot; structured `Mismatches` schema + degenerate-verdict guard.
- `pkg/exploration/exploration.go` — §5: `SpecLinksForAtom`/`TestLinksForAtom`; `Trace` derives code/test links + warnings from the live crawl, not frontmatter `linked_codes` (frontmatter seeding left intact for other consumers).
- `cmd/atd/cmd/trace.go`, `pkg/prompt/trace_summary.go` — §4: look up the target brief by canonical id with a rebuild fallback; prompt names the target as the subject and labels ancestry/dependents as context only.
- `cmd/atd/cmd/stats.go` — §6: `isLocalAtom` filter (exclude `project:`-prefixed and absolute-path atoms) applied to every aggregation in project scope; `implemented_stable_count` gated on `len(Implementations) > 0`.
- `pkg/workspace/loader.go` — §9 (new, see below).

### §9 (new) — workspace "current project" always mis-detected as the root project

Surfaced while fixing §6. `Workspace.FindProjectByCWD` (`pkg/workspace/loader.go:120`)
returned the **first** project whose path prefixes the cwd. Because the `shared`
project has path `.` (= workspace root) and is listed first, it prefixes *every* cwd,
so the "current project" resolved to `shared` no matter where you were. That silently
mis-attributed every **bare** `@spec-link` tag to the wrong project — e.g. `upsilonapi`
project-scope stats reported `implemented_total_count: 0` despite 45 real tags in its
own code. Fix: pick the **longest (most specific)** matching project with a
boundary-aware prefix check (`absCWD == projPath` or under `projPath + separator`), so
a root-level project no longer shadows the real one. `implemented_total_count` for
`upsilonapi` went 0 → 10 after the fix.

*Residual (out of scope, noted for the maintainer):* the stats JSON `"project"` label
still prints `shared` in some project-scoped runs — that value comes from
`config.ActiveConfig.ActiveProject` (config load), a separate path from
`FindProjectByCWD`, and is cosmetic (the atom numbers are now correct). Worth a
follow-up to make the config's active-project detection use the same longest-match
logic.

### Build/test (both batches)

`go build ./...` and `go vet ./...` clean in both modules. `go test ./...` green
except two failures confirmed **pre-existing** (identical on clean HEAD via `git
stash`): `pkg/prompt` build error in `dissect_test.go` (`AuditCodeBuild` arity) and
`cmd/atd/cmd` `TestUpdateLinks`. New tests added: resolver strip-and-retry
(`pkg/workspace/resolver_test.go`), diff-path rebasing + recon-degenerate + map-path
(`check_coverage_test.go`, `map_test.go`). `upsilon-hub` left clean throughout.

**Nothing from the original report remains open.** Recommended follow-ups (maintainer's
call, not in the report): the §9 residual config-label mismatch above; and the two
pre-existing unrelated test failures (`dissect_test.go` arity, `TestUpdateLinks`).
