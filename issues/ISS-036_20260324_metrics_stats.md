# Issue: ATD Metrics and Stats Dashboard

**ID:** `20260324_metrics_stats`
**Ref:** `ISS-036`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/cmd/atd`
**Affects:** CI pipelines, WebUI, project health reporting

---

## Summary

There is no way to get a quantitative view of documentation health: coverage percentage, orphan ratio, bloat score distribution, atom count by type/status/domain. This makes it impossible to enforce documentation standards in CI or present health dashboards.

---

## Technical Description

### Background
Individual tools provide fragments of health data (e.g., `crawl --gaps` for orphans, `audit` for bloat), but there is no unified summary.

### The Problem Scenario
1. A project manager asks "What percentage of our `STABLE` atoms are implemented?"
2. Answering requires running multiple commands and manually aggregating.

### Where This Pattern Exists Today
- No unified metrics tool exists.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — no CI-enforced documentation standards |
| Detectability | High — absence of dashboard is obvious |
| Current mitigant | Manual tool invocation and counting |

---

## Recommended Fix

**Short term:** Add `atd stats` subcommand (deterministic, no LLM) producing JSON: total atoms, atoms by type, atoms by status, atoms by domain, coverage ratio, orphan count.
**Medium term:** Expose as MCP tool `atd_stats`. Output a CI‐friendly exit code (fail if coverage below threshold). Generate CI badge data.

---

## References

- [ATD.md §3.2.4](file:///home/bastien/work/skill/ATD.md)
