---
id: service_atd_audit
human_name: "ATD Audit"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, audit, integrity]
parents:
  - [[module_atd_cli]]
dependents:
  - [[mechanic_atd_compare]]
  - [[mechanic_atd_fix]]
layer: IMPLEMENTATION
---

# ATD Audit

## INTENT
To systematically assess the structural integrity of the ATD ecosystem — detecting bloated atoms, semantic collisions, and missing abstractions — and optionally verifying code compliance against ATD rules.

## THE RULE / LOGIC
Operates in two modes focusing on internal ATD health:

**Mode 1 (Bloat Detection):**
Feeds each atom's INTENT and LOGIC sections to the LLM for a binary YES/NO bloat assessment based on the atom type's strictness. Files are processed through a bounded worker pool: at most `concurrency` files (CLI `--concurrency`, MCP `concurrency`; default 0 falls back to a package constant of 4) are audited in parallel, so only that many outbound bloat-check LLM calls to the local Ollama provider are ever in flight at once, keeping the sweep from overwhelming the provider with simultaneous requests. Per-file results are still collected and merged back into original file order once the pool drains, so the report text and its counters are identical regardless of the concurrency level used.

**Mode 2 (Collision Detection):**
Computes Nomic embeddings for all atoms and builds a pairwise cosine similarity matrix. For pairs exceeding the threshold, performs an ancestry BFS walk to determine if they represent a COLLISION (missing abstraction) or are structurally related.

Audit has no atom-vs-code compliance-checking capability — it only ever assesses the docs graph's own internal health (bloat + collisions), never a specific code file against a specific atom. `--atom <path>` (CLI) / `atom` (MCP) alone scopes both modes to a single atom file instead of sweeping the whole docs directory. Passing `--code`/`code` together with `--atom`/`atom` is a hard error — the command refuses to run and reports: "audit does not compare an atom against code; use \"atd map --atom <atom> --file <code>\" for a single-atom-vs-single-file compliance check instead." `--code`/`code` given alone (without `--atom`/`atom`) is not an error but has no effect either — audit still runs a full-directory sweep and the value is never read. Single-atom-vs-single-file compliance checking belongs to `atd map --atom <atom_id> --file <code_path>` (MCP: `atd_recon`) instead; broader diff-driven, multi-file compliance verification belongs to `atd verify`.

A bloat-check or embed call that errors (including an LLM request timing out once the tiered LLM provider's `llm.generate_timeout_ms` elapses) is never silently treated as a pass or silently dropped from collision detection — it is logged as an explicit `[ERROR]` line naming the atom and cause. Every run, clean or not, ends with a guaranteed `Summary: N atom(s) scanned, M bloated, K collision(s), E LLM error(s)` line, so a run can never exit having produced no readable signal.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd audit [--docs <dir>] [--workspace] [--threshold <float>] [--atom <atom_path>] [--code <code_path>] [--concurrency <int>]`
- **MCP Tool:** `atd_audit` with `docs`, `threshold`, `atom`, `code`, `concurrency` params (no `workspace` param — MCP always scopes to the single active project's docs directory)
- **`--atom`/`atom`:** scopes the bloat/collision sweep to one atom file instead of sweeping `--docs`/`docs`.
- **`--code`/`code`:** not a compliance-check input for this command. Combined with `--atom`/`atom`, it is rejected with a hard error directing the caller to `atd map`/`atd_recon`; given alone it is accepted but silently unused.
- **`--concurrency`/`concurrency`:** bounds how many atom files Phase 1 (bloat detection) audits in parallel against the LLM provider at once; `<= 0` (including unset/0) falls back to the package default of 4.
- **LLM Tasks:** `audit_bloat` (bloat pass/fail per atom), `embed` (collision similarity)
- **Function:** `pkg/audit.RunFullAudit(docsDir, threshold, workspace, concurrency)` / `pkg/audit.RunScopedAudit(atomPath, threshold, concurrency)`
- **Code Tag:** `@spec-link [[service_atd_audit]]`

## EXPECTATION
A run always terminates with a `Summary: N atom(s) scanned, M bloated, K collision(s), E LLM error(s)` line, never a silent empty result, and report text/counters are identical for the same input regardless of `--concurrency`/`concurrency` value used. `--atom <path>` without `--code` narrows both bloat and collision detection to that one atom file. `--atom <path>` combined with `--code <path>` (or MCP `atom`+`code` together) always fails fast with an error naming `atd map --atom <atom> --file <code>` as the correct tool, and performs no audit pass at all in that case — it never falls back to running the bloat/collision sweep anyway.
