# Issue: ATD Content Reformulation Modal

**ID:** `20260403_webui_atd_reformulate_modal`
**Ref:** `ISS-066`
**Date:** 2026-04-03
**Severity:** High
**Status:** Open
**Component:** `webui/static/js/details.js`
**Affects:** `webui/static/js/explorer.js`, `scripts/cmd/atd`

---

## Summary

When an ATD has defined parents and dependents, its logic, technical interface, and expectations should align with its intent and its neighbors' content. There is currently no UI feature to "reformulate" these sections automatically.

---

## Technical Description

### Background

The `atd_assemble` tool can stitch atoms together. This can be used to provide context for an LLM to "rewrite" or "reformulate" an atom's section based on its position in the graph.

### The Problem Scenario

1. User selects an ATD that is a "Mechanic" in the implementation layer.
2. The user wants to ensure the "Technical Interface" correctly reflects the requirements from its "Rule" parent.
3. The user wants to see a side-by-side comparison before applying changes.

```
[Select ATD] -> [Click Reformulate]
   |
   v
[Open Modal]
   [Left: Current Content] <-> [Right: Proposed Content (LLM/Assemble)]
   [Sections: Logic, Interface, Expectations]
   |
   +--> [Regenerate Section (Local/Assemble)]
   +--> [Regenerate Section (External/LLM)]
   |
   v
[Apply Changes]
```

### Where This Pattern Exists Today

This requires new UI in `webui/static/js/details.js` and a new modal component.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | High (Productivity boost) |
| Detectability | High |
| Current mitigant | Manual editing |

---

## Recommended Fix

**Short term:** Define the API for atom reformulation on the backend using `atd_assemble`.  
**Medium term:** Implement the side-by-side modal in the WebUI with regeneration options for each section.

---

## References

- [details.js](file:///home/bastien/work/atd/scripts/pkg/webui/static/js/details.js)
- [README.md](file:///home/bastien/work/atd/scripts/pkg/webui/README.md)
