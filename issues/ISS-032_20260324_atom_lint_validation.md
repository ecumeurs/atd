# Issue: Atom Structural Lint / Validation Tool

**ID:** `20260324_atom_lint_validation`
**Ref:** `ISS-032`
**Date:** 2026-03-24
**Severity:** High
**Status:** Resolved
**Component:** `scripts/cmd/atd`
**Affects:** `atd audit`, CI pipelines, MCP consumers

---

## Summary

There is no lightweight, deterministic structural validation tool for ATD atoms. `atd audit` handles bloat and collision detection (LLM-heavy), but cannot cheaply catch missing mandatory fields, malformed `[[id]]` references, broken parent/dependent links (pointing to non-existent atoms), or empty sections.

---

## Technical Description

### Background
Atom files have strict structural requirements: all frontmatter fields must be present, all `[[id]]` references must resolve, and all four mandatory H2 sections must have content.

### The Problem Scenario
1. An agent creates an atom via `atd update` but omits `domain` or `priority`.
2. The atom passes `atd audit` (which focuses on semantic quality) but is structurally invalid.
3. Downstream tools (crawl, weave, search) encounter unexpected missing data.

### Where This Pattern Exists Today
- All atom files in `docs/` — structural validation is purely manual.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — broken graph, inconsistent data |
| Detectability | Low — structural errors are silent until another tool fails |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** Add `atd lint` subcommand (deterministic, no LLM) that validates: mandatory fields present, `[[id]]` references resolve, sections non-empty, `domain` is valid enum, `priority` is 1-5.
**Medium term:** Expose as MCP tool `atd_lint`. Run automatically in CI on every commit.
**Long term:** Integrate lint warnings into `atd update` so they are caught at write time.

---

## References

- [ATD.md §3.2.6](file:///home/bastien/work/skill/ATD.md)
- [atd_structure.atom.md](file:///home/bastien/work/skill/docs/atd_structure.atom.md)
