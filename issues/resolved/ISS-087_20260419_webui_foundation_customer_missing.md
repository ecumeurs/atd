# Issue: WebUI Foundation Zone - Implemented Customer Atoms Missing from Column

**ID:** `20260419_webui_foundation_customer_missing`
**Ref:** `ISS-087`
**Date:** 2026-04-19
**Severity:** Medium
**Status:** Resolved
**Component:** `webui/static/js/ui.js` (or similar rendering logic)
**Affects:** WebUI Explorer View, Waterfall of Intent

---

## Summary

In the WebUI Explorer, when a Customer Layer ATD is fully implemented, it correctly appears in the "Foundation" zone (bottom area). However, it seems to be removed from its original position in the "Customer" column. This makes it harder for users to see the complete set of customer requirements in one place.

---

## Technical Description

### Background
The Waterfall of Intent separates atoms into columns (Customer, Architecture, Implementation) and zones based on health/implementation status. The Foundation zone is reserved for atoms that are fully implemented and stable.

### The Problem Scenario
A user looks at the Customer column to see all requirements. An atom that was recently fully implemented moves to the Foundation zone at the bottom of the screen but is no longer rendered within the vertical flow of the Customer column. The user loses the visual context of where that requirement sits in the hierarchy relative to other requirements.

### Where This Pattern Exists Today
The rendering logic (likely in `webui/static/js/` or the templating engine) that assigns atoms to lanes and zones.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Consistent behavior) |
| Impact if triggered | Low/Medium (Visual confusion, perceived data loss) |
| Detectability | High |
| Current mitigant | Scrolling to the bottom to find the atom. |

---

## Recommended Fix

**Short term:**
- Modify the rendering logic to ensure Customer ATDs remain visible in the Customer column even when they meet the criteria for the Foundation zone.
- Apply a specific visual style (e.g., a textured green background) to these atoms so they are easily recognizable as "fully implemented and stable" while remaining in their natural layer column.

**Medium term:**
- Review the "Foundation Zone" logic: should it be a separate zone, or just a status-based highlighting within the columns?

---

## References

- [docs/ui_webui_waterfall_explorer.atom.md](file:///home/bastien/work/skill/docs/ui_webui_waterfall_explorer.atom.md)
- [scripts/pkg/webui/static/index.html](file:///home/bastien/work/skill/scripts/pkg/webui/static/index.html)
