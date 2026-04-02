# Issue: WebUI Document Generation Failures and UI Regressions

**ID:** `ISS-061_20260402_webui_document_generation_bugs`
**Ref:** `ISS-061`
**Date:** 2026-04-02
**Severity:** High
**Status:** Open
**Component:** `webui/`
**Affects:** `webui/main.go`, `webui/static/`

---

## Summary

The WebUI document generation flow is currently broken and suffers from several UI regressions. Key issues include contextual search failures for known ATDs, non-functional "add specific atom" popups, layout inconsistencies (inline form instead of modal), and the absence of expected UI components and atoms.

---

## Technical Description

### Background

The document generation feature should allow users to search for ATDs, select them, and generate a narrative. This process usually involves a modal-based UI (consistent with Ctrl+K) and a functional search that matches the global ATD index.

### The Problem Scenario

1. **Contextual Search Failure:**
   - User hits Ctrl+K and searches "spec builder" -> ATD found.
   - User clicks "Generate Document" with the same query -> "No contextual ATDs found. Please add manually."

2. **Broken "Add Specific Atom" Popup:**
   - User clicks "Add specific atom".
   - Popup accepts input, but nothing happens (no atom added to the selection list).

3. **Layout Regression:**
   - "Generate Document" button triggers an inline form between the header and the view instead of opening a modal.

4. **UI Inconsistency:**
   - The "Add specific atom" dialog does not match the spec builder context's "+" button dialog.

5. **Missing Documentation:**
   - `ui_webui_document_viewer.atom.md` does not exist in the `docs/` directory despite being referenced/expected.

### Where This Pattern Exists Today

- Search logic in document generation handlers.
- Popup/Modal implementation in the frontend (likely `webui/static/`).
- Missing atom in `docs/`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High — Core documentation feature is unusable |
| Detectability | High — Obvious UI failure |
| Current mitigant | None |

---

## Recommended Fix

**Short term:**
- Fix the search mapping to ensure "Generate Document" ATD search provide results (even when provided with a natural language query). 
- Restore the modal behavior for document generation.
- Fix the event handling for the "Add specific atom" popup to correctly update the state.

**Medium term:**
- Unify the "Add atom" dialog components across different contexts (Spec Builder vs. Document Generator).
- Create the missing `ui_webui_document_viewer.atom.md` file.

**Long term:**
- Implement a shared frontend controller for all ATD-searching modals to prevent logic divergence.

---

## References

- [docs/mechanic_webui_document_generation.atom.md](file:///home/bastien/work/skill/docs/mechanic_webui_document_generation.atom.md)
- [webui/main.go](file:///home/bastien/work/skill/webui/main.go)
- [webui/README.md](file:///home/bastien/work/skill/webui/README.md)
