# Issue: WebUI ATD Editing and ID Propagation

**ID:** `ISS-006_20260304_webui_edit_propagation`
**Ref:** `ISS-006`
**Date:** 2026-03-04
**Severity:** High
**Status:** Resolved
**Component:** `webui`
**Affects:** `webui/backend`, `webui/app.js`, `atd_management_skill`

---

## Summary

The WebUI needs to allow altering ATDs (all fields). Crucially, changing an Atom ID must trigger a filename change and update all dependent ATDs (parents/dependents lists) to maintain graph integrity.

---

## Technical Description

### Background
ATDs are linked via IDs in their YAML frontmatter. Filenames are typically `[ID].atom.md`.

### The Problem Scenario
Currently, editing is either not supported or does not handle the fallout of an ID change, leading to broken links, orphaned files, and inconsistent metadata.

### Where This Pattern Exists Today
- `webui/` backend for saving ATDs.
- `atd_management_skill` for manual editing.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High |
| Detectability | Medium |
| Current mitigant | Manual ID/Filename sync |

---

## Recommended Fix

**Short term:** Implement basic field editing without ID modification.
**Medium term:** Add ID change logic that renames the file and performs a workspace-wide search-and-replace for the old ID.
**Long term:** Use the `atd_management_skill` or a dedicated library to handle atomic updates and link propagation.

---

## References

- [atd_management_skill](../atd_management_skill/)
