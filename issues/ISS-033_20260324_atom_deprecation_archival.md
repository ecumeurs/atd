# Issue: Atom Deprecation and Archival Statuses

**ID:** `20260324_atom_deprecation_archival`
**Ref:** `ISS-033`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd`
**Affects:** `atd audit`, `atd crawl`, `atd query`, all MCP tools

---

## Summary

There is no `DEPRECATED` or `ARCHIVED` status for atoms. When a feature is removed, the corresponding atom lingers as `STABLE` with no formal lifecycle end, leading to confusion and stale references.

---

## Technical Description

### Background
The current status enum is `DRAFT` → `REVIEW` → `STABLE`. There is no terminal state beyond `STABLE`.

### The Problem Scenario
1. A feature is removed from the codebase.
2. The corresponding `STABLE` atom has no status to reflect deprecation.
3. `atd crawl --gaps` flags it as an orphan (no code links), but cannot distinguish "removed on purpose" from "not yet implemented".
4. Other atoms still listing it as a parent create broken graph edges.

### Where This Pattern Exists Today
- Status enum in YAML parsing logic.
- All atom-processing tools.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium — stale atoms pollute the graph |
| Detectability | Medium — orphan gap reports will surface them |
| Current mitigant | Manual deletion of atom files |

---

## Recommended Fix

**Short term:** Add `DEPRECATED` and `ARCHIVED` to the status enum in parsers and `atd update`.
**Medium term:** `atd audit` should warn if code still links (`@spec-link`) to a `DEPRECATED` atom. `ARCHIVED` atoms should be excluded from default tool outputs unless explicitly queried.
**Long term:** `atd update --set status=DEPRECATED` should auto-warn about downstream dependents.

---

## References

- [ATD.md §3.2.1](file:///home/bastien/work/skill/ATD.md)
- [ATD.md §1.6 Status Transitions](file:///home/bastien/work/skill/ATD.md)
