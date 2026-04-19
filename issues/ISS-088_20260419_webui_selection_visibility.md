# Issue: WebUI Explorer - Improved Visibility for Related Atoms on Selection

**ID:** `20260419_webui_selection_visibility`
**Ref:** `ISS-088`
**Date:** 2026-04-19
**Severity:** Medium
**Status:** Open
**Component:** `webui/static/js/explorer.js` (or similar)
**Affects:** WebUI Explorer View

---

## Summary

When an ATD is clicked in the Explorer view, it, its ancestors, and its descendants are highlighted. However, the user currently has to manually sift through the columns to find these highlighted atoms. The UI should automatically prioritize or filter the view to make these related atoms immediately visible.

---

## Technical Description

### Background
The WebUI Explorer highlights the dependency path of a selected atom. This is a crucial feature for "Blast Radius" analysis and understanding intent propagation.

### The Problem Scenario
A user clicks on an "Implementation" atom. The system highlights 3 Architecture parents and 1 Customer parent. However, these parents are scattered across their respective columns, some potentially off-screen or buried among dozens of other atoms. The user has to scroll and scan to find the "green" highlights.

### Where This Pattern Exists Today
The highlight logic in the frontend (CSS class toggling on click).

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium (Reduced usability, friction in navigation) |
| Detectability | High |
| Current mitigant | Manual scrolling and hunting for highlights. |

---

## Recommended Fix

**Short term:**
- When an atom is selected, ensure that all related atoms (those with the highlight class) are scrolled into view or moved to the top of their respective columns.
- Alternatively, apply a temporary "filter" mode where the columns only show the selected atom and its lineage.

**Medium term:**
- Implement a "Focus Mode" toggle that hides all non-related atoms when one is selected.
- Add connector lines (SVG) between the selected atom and its parents/dependents to visually guide the eye.

**Long term:**
- Integrate the graph layout more tightly so that selection naturally reshapes the flow to highlight the active path.

---

## References

- [docs/ui_webui_explorer_treeview.atom.md](file:///home/bastien/work/skill/docs/ui_webui_explorer_treeview.atom.md)
- [scripts/pkg/webui/static/js/main.js](file:///home/bastien/work/skill/scripts/pkg/webui/static/js/main.js) (assuming this manages selection)
