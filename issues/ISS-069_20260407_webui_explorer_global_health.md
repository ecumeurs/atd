# Issue: WebUI Explorer Global Health Dashboard

**ID:** `20260407_webui_explorer_global_health`
**Ref:** `ISS-069`
**Date:** 2026-04-07
**Severity:** Medium
**Status:** Open
**Component:** `scripts/pkg/webui`
**Affects:** `static/js/explorer.js`, `static/js/app.js`, `handlers.go`

---

## Summary

The current Explorer view provides a "Waterfall of Intent" but lacks a high-level, aggregate summary of the project's documentation health. This issue tracks the implementation of a global health check component that surfaces key metrics (coverage, orphans, missing links) and provides one-click filtering to address these gaps.

---

## Technical Description

### Background
The WebUI currently visualizes ATDs in three columns (Customer, Architecture, Implementation). While individual atoms show health indicators, there is no way to see the total "Documentation Debt" at a glance.

### The Problem Scenario
A lead architect wants to know:
1. What percentage of the project is documented?
2. How many "Implementation" atoms have no "Architecture" parent (orphans)?
3. How many "Architecture" atoms are missing "Customer" requirements?
4. Which atoms are "Done" but missing actual code or test links?

Currently, these require manual inspection or running CLI commands like `atd stats` or `atd crawl --gaps`.

### Proposed Solution
Add a "Global Health" header or sidebar to the Explorer tab that displays:
- **Documentation Coverage:** Ratio of linked code to total codebase (using `atd_stats` logic).
- **Orphaned Atoms:** Count of atoms with no parents (excluding top-level Customer atoms).
- **Missing Customer Link:** Count of Architecture/Implementation chains that don't terminate in a Customer requirement.
- **Implementation Gaps:** Implementation atoms with 0% code coverage or 0% test coverage.

Each metric should act as a **Filter**: clicking it should instantly update the Explorer view to show only the affected atoms.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Users need this for project management) |
| Impact if triggered | Medium (Better visibility into doc debt) |
| Detectability | N/A (Feature request) |
| Current mitigant | CLI tools `atd stats` and `atd trace` |

---

## Recommended Fix

**Short term:** Add a simple horizontal bar at the top of the Explorer view with these 4 metrics.  
**Medium term:** Implement the "Filter" logic in `static/js/explorer.js` to handle these dynamic health-based queries.  
**Long term:** Integrate this into the "Waterfall" as a specialized "Health" mode toggle.

---

## References

- [explorer.js](file:///home/bastien/work/skill/scripts/pkg/webui/static/js/explorer.js)
- [handlers.go](file:///home/bastien/work/skill/scripts/pkg/webui/handlers.go) (for the backend data source)
- [atd_stats](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/stats.go)
- [atd_crawl](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/crawl.go)
