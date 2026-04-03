# Issue: Force Weaving from ATD Detail Panel

**ID:** `20260403_webui_atd_force_weave`
**Ref:** `ISS-065`
**Date:** 2026-04-03
**Severity:** Medium
**Status:** Open
**Component:** `webui/static/js/details.js`
**Affects:** `webui/static/js/explorer.js`

---

## Summary

The ATD Detail side panel currently lacks a direct way to trigger a "force weave" operation on the selected atom. Weaving is essential for synchronizing `parents` and `dependents` links across the atom graph.

---

## Technical Description

### Background

ATD atoms use bidirectional linking. If a user modifies the `parents` field in one atom, the corresponding `dependents` fields in those parent atoms must be updated. This is handled by the `atd_weave` tool.

### The Problem Scenario

1. User opens the WebUI Explorer.
2. User selects an ATD atom and edits its content or frontmatter (e.g., adding a parent).
3. The graph may become inconsistent if weaving is not executed.
4. Currently, the user must wait for a background process or run the CLI manually.

```
[WebUI Detail Panel] --> [Action: Force Weave] 
                           |
                           v
                     [atd_weave]
                           |
                           v
                 [Update dependant links]
```

### Where This Pattern Exists Today

The detail side panel logic is primarily in `webui/static/js/details.js`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium (Graph inconsistency) |
| Detectability | Medium (Links don't show up in Explorer) |
| Current mitigant | Manual CLI usage |

---

## Recommended Fix

**Short term:** Add a "Force Weave" button in the ATD detail panel that calls a new backend endpoint, which internally will communicate with the weaving function of the CLI which should be accessible, as its in the same binary. 

---

## References

- [details.js](file:///home/bastien/work/skill/webui/static/js/details.js)
