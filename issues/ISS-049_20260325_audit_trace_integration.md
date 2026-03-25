# Issue: Audit Should Request Trace for Coverage Detection

**ID:** `20260325_audit_trace_integration`
**Ref:** `ISS-049`
**Date:** 2026-03-25
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/audit.go`
**Affects:** `atd audit`, `atd trace`, CI/CD pipelines

---

## Summary

When running an audit (full or scoped), the system currently focuses on bloat and semantic collisions. It should proactively suggest or trigger a `trace` for the audited atoms to detect whether parts of the documentation, implementation, or tests are missing. This integration would provide a more holistic view of an atom's health during the verification phase.

---

## Technical Description

### Background
`atd audit` checks for internal document quality (bloat) and cross-document overlap (collisions). `atd trace` provides a health snapshot inclusive of implementation and test coverage.

### The Problem Scenario
1. A developer runs `atd audit` on a set of atoms.
2. The audit passes (no bloat, no collisions).
3. However, some of these atoms might have 0% implementation or test coverage, which `audit` does not currently surface.
4. The developer assumes the documentation is "green" when it is actually orphaned or unimplemented.

### Where This Pattern Exists Today
- `scripts/cmd/atd/cmd/audit.go`: Does not invoke or recommend `trace`.
- `scripts/cmd/atd/cmd/mcp_tools.go`: `atd_audit` and `atd_trace` are separate tools with no combined logic.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — leads to "false sense of security" regarding doc completeness |
| Detectability | Low — requires manual `atd trace` or `atd crawl --gaps` to find coverage issues |
| Current mitigant | Manually running `atd trace` or checking `atd stats` |

---

## Recommended Fix

**Short term:** Add a recommendation to the `atd audit` output suggesting the user run `atd trace [[ID]]` for any atoms with low or unknown health.
**Medium term:** Integrate a lightweight "coverage check" (trace-lite) into the audit command that flags atoms with no `@spec-link` or `@test-link` detections.
**Long term:** A unified `atd verify` command that runs both structural audit and traceability trace.

---

## References

- [audit.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/audit.go)
- [trace.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/trace.go)
- [ATD.md §2.2](file:///home/bastien/work/skill/ATD.md)
