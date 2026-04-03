# Issue: Dedicated ATD Tree/Graph View

**ID:** `20260403_webui_atd_tree_view`
**Ref:** `ISS-067`
**Date:** 2026-04-03
**Severity:** High
**Status:** Open
**Component:** `webui/static/js/explorer.js`
**Affects:** `webui/static/js/details.js`

---

## Summary

The current Explorer view provides a waterfall/lane view, but navigating deep hierarchies (ancestors/descendants) can be difficult. A dedicated tree/graph view for a specific ATD is needed for better structural clarity.

---

## Technical Description

### Background

Users need to see exactly what a specific requirement (Customer layer) breaks down into in terms of Architecture and Implementation, including code and tests.

### The Problem Scenario

1. User clicks a "View Tree" button on an ATD card.
2. UI navigates to a new view (e.g., `/explorer/tree/:id`).
3. View shows:
    - Breadcrumb: `Explore > [Current ATD Name]`
    - Parents: Layer, Type, Name, Health, Intent.
    - Current ATD: Full card in center.
    - Dependents: Full graph down to `@spec-link` code and `@test-link` tests.

```
       [Parent 1] [Parent 2]
               \   /
           [Selected ATD]
               /   \
        [Child 1] [Child 2]
           |         |
        [Code]     [Test]
```

### Where This Pattern Exists Today

This is a new navigation state in `webui/static/js/explorer.js`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | High (Navigation and Traceability) |
| Detectability | High |
| Current mitigant | Highlighting in flat explorer |

---

## Recommended Fix

**Short term:** Implement a basic graph layout using D3 or similar, centered on the selected ATD.  
**Medium term:** Add breadcrumb navigation and full integration with the detail side panel.

---

## References

- [explorer.js](file:///home/bastien/work/skill/webui/static/js/explorer.js)
