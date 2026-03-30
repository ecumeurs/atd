# Issue: `atd stats` coverage and ancestry reporting

**ID:** `20260325_atd_stats_coverage_ancestry`
**Ref:** `ISS-052`
**Date:** 2026-03-25
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/stats.go`
**Affects:** ATD health metrics, documentation quality

---

## Summary
`atd stats` currently lacks detailed reporting on implementation and test coverage at the implementation layer, as well as ancestry validation to ensure all atoms (except customer layer) have a path upward to the customer layer.

---

## Technical Description

### Background
`atd stats` provides high-level metrics about the ATD graph. ATDs in the implementation layer should ideally be linked to code via `@spec-link` and to tests via `@test-link`. Furthermore, the ATD methodology requires a top-down connection where implementation atoms are rooted in architectural and customer requirements.

### The Problem Scenario
1.  **Missing Coverage Metrics**: There is no easy way to see what percentage of the implementation layer is actually implemented and tested.
    - Ratio should be: (ATDs linked to code/tests) / (Total leaf ATDs in the implementation layer).
    - Non-leaf ATDs in the implementation layer that have implementation/tests should also be included in both numerator and denominator.
2.  **Disconnected Implementation**: Implementation atoms might be created "bottom-up" without being linked to any higher-level architectural or customer requirements, violating the "Implementations Require Roots" principle.
    - An implementation layer ATD should have at least one architecture layer and one customer layer ancestor.

### Where This Pattern Exists Today
The existing `atd stats` implementation in `scripts/cmd/atd/cmd/stats.go` needs to be expanded to traverse the graph and calculate these specific ratios and ancestry paths.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | Low — requires manual inspection without this tool |
| Current mitigant | `atd trace` can check individual atoms, but there's no aggregate report. |

---

## Recommended Fix

**Short term:** Implement implementation and test coverage ratios in `atd stats` for the implementation layer.
**Medium term:** Add an "orphan" or "unrooted" check that flags implementation/architecture atoms lacking a path to the customer layer.
All such atoms should be listed with their full path, and correctly tagged (e.g. with `orphan` or `unrooted`) in the output of `atd stats` for quick identification and fixing.
**Long term:** Integrate these metrics into CI/CD pass/fail criteria for documentation health.

---

## References
- [ATD.md](file:///home/bastien/work/skill/ATD.md)
- [atd trace output](file:///home/bastien/work/skill/docs/service_atd_trace.atom.md) (Related logic)
