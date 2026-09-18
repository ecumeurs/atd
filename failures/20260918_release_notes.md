# 2026-09-18 — Release notes: KNOWN DEFECT batch fixes

Fixed this batch (pinned in 149a497, fixed in the commits below) and removed
from `failures/` as fully processed:

- `atd update`'s id normalizer no longer double-prefixes an id that already
  starts with an abbreviated type prefix (e.g. `mech_` → was becoming
  `mechanic_mech_...`). (96b521b)
- `atd update --expectation` (and the other body sections) now appends a
  missing `## <SECTION>` header instead of silently no-opping while still
  reporting `Success`. (96b521b)
- `atd audit --atom` now actually scopes the audit to that one atom instead
  of silently ignoring the flag and running a full `--docs` sweep. (e33e0f0)
- `atd search` flags stale/deleted-file hits (`Stale: true`) instead of
  returning them indistinguishable from genuinely live results. (450bb84)
- `atd map` link recommendations are now validated against the atom-ID
  shape and cross-checked against the real registry; hallucinated advice,
  fabricated URLs, and plain sentences are dropped into a "Discarded
  Suggestions" list instead of being rendered as `[[...]]` links. (6b3b52a)
- `atd map`'s JSON-truncation crash is fixed by reordering the LLM response
  schema (`recommendations` before the large free-form `rationale` field)
  and adding salvage logic for partial payloads. (6b3b52a)

Reports removed as fully processed by this batch:
- `20260901_atd_map_hallucinated_code_values.md`
- `20260901_atd_map_malformed_json_from_link_recommender.md`
- `20260901_atd_search_stale_index_returns_deleted_code.md`
- `20260902_atd_map_malformed_json_link_recommender_recurrence.md`
- `20260917_atd_audit_never_returns_docs_and_atom_code_scopes.md`
- `20260917_atd_map_hallucinated_links_and_malformed_json_recurrence.md`
- `20260917_atd_update_double_prefixes_id_that_already_starts_with_type_abbreviation.md`
- `20260917_atd_update_expectation_no_ops_when_section_header_missing.md`

Also fixed this batch, as a follow-up:
- `atd init`'s installed pre-commit hook's orphan check now compares
  against HEAD instead of blocking any commit that merely touches a file
  containing a pre-existing orphaned atom: newly-added orphans and
  last-parent removals still block, a pre-existing orphan untouched by the
  commit's `parents:` now only warns. (a43f553)

Reports removed as fully processed:
- `20260916_orphaned_architecture_atom_blocks_commit.md`

Still open (not touched by this batch, left in `failures/`):
- `20260916_atd_audit_workspace_no_return.md` — separate gap: `atd audit`
  has no bounded per-request LLM timeout or explicit backend-unreachable
  error.
- `20260917_atd_audit_docs_exits_zero_with_no_report.md` — separate gap: a
  clean 0-findings `atd audit` run prints no summary line, indistinguishable
  from a broken run.
- `20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md` —
  deferred pending a policy decision, out of scope for this round.
