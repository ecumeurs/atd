# Issue: Enhance WebUI ATD Health Indicators

**ID:** `20260403_webui_atd_health_indicators`
**Ref:** `ISS-068`
**Date:** 2026-04-03
**Severity:** Medium
**Status:** Open
**Component:** `scripts/pkg/webui`
**Affects:** `scripts/pkg/webui/static/js/explorer.js`

---

## Summary

This issue track the enhancement of the ATD cards in the WebUI Explorer view with advanced health indicators. These indicators represent an atom's traceability (ancestry and dependencies) based on its layer, alongside existing implementation and test coverage.

---

## Technical Description

### Background

Currently, each ATD card shows two squares representing implementation and test coverage.

### The Problem Scenario

The objective is to add another series of health indicators left to the two original one. The specific indicators depend on the ATD's layer:
1.  **IMPLEMENTATION layer**: [Customer Ancestor identified] [Architecture Ancestor identified]
2.  **ARCHITECTURE layer**: [Customer Ancestor identified] [Implementation dependent identified]
3.  **CUSTOMER layer**: [Architecture dependent identified] [Implementation dependent identified]

Rules of display:
- Both series (Ancestry/Dependency and Impl/Test) should be separated by `|`.
- If both indicators in a series are satisfied, they should be replaced by a green tick `✓`.
- Each indicator must have a clear tooltip.

Wait, this feature is blocked because the backend current crawling and caching logic isn't optimized for graph-wide health calculation (Ref: ISS-060).

### Where This Pattern Exists Today

- `scripts/pkg/webui/static/js/explorer.js:createAtomCard` (Line 144)
- `scripts/pkg/webui/static/styles.css`

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High — requested UX improvement |
| Impact if triggered | Medium — improves traceability awareness |
| Detectability | High — visual indicators in Explorer |
| Current mitigant | Existing basic coverage indicators |

---

## Recommended Fix

**Short term:** Implement the UI logic in `explorer.js` once the backend provides the necessary health fields.

**Medium term:** Update `atom.AtomData` and `exploration.DependencyGraph` to compute these flags after crawling.

**Long term:** Ensure unified exploration and caching (ISS-060) to avoid redundant crawls.

---

## References

- [ISS-060](file:///home/bastien/work/skill/issues/ISS-060_20260402_refactor_atd_exploration_commands.md)
- [explorer.js](file:///home/bastien/work/skill/scripts/pkg/webui/static/js/explorer.js)
- [styles.css](file:///home/bastien/work/skill/scripts/pkg/webui/static/styles.css)
