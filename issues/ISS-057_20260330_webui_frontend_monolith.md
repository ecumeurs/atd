# Issue: Frontend Monoliths (app.js, spec-builder.js) and Inappropriate ATD Granularity

**ID:** `20260330_webui_frontend_monolith`
**Ref:** `ISS-057`
**Date:** 2026-03-30
**Severity:** High
**Status:** Open
**Component:** `webui/static/`
**Affects:** `webui/static/app.js`, `webui/static/spec-builder.js`, ATD index

---

## Summary

The frontend codebase for the WebUI is distributed across two major files: `app.js` (560+ lines) and `spec-builder.js` (1380+ lines). Both files have become monolithic, mixing DOM manipulation, state management, complex API orchestration, and logic that should be separated into modules. Additionally, many functionalities within these files are linked to the same ATD atoms (e.g., `ui_webui_spec_builder`, `rule_webui_context_history_management`), indicating that the ATD granularity is too low to reflect the diversity of the implementation.

---

## Technical Description

### Background
The WebUI frontend is a single-page application (SPA) style interface. `app.js` handles the main explorer and atom details, while `spec-builder.js` manages the conversational AI interface for spec creation.

### The Problem Scenario
1.  **Monolithic Files:**
    *   `app.js` handles everything from treemap rendering (D3) to bulk updates and atom editing.
    *   `spec-builder.js` is extremely large (1380 lines) and manages chat history, proposal versioning, model selection, and context injection in a single IIFE.
2.  **Inappropriate Granularity:**
    *   `ui_webui_spec_builder` is linked to at least 9 different locations in `spec-builder.js`, covering initialization, message rendering, proposal handling, and more.
    *   `rule_webui_context_history_management` is linked to 4 locations in `spec-builder.js`.
    *   `ui_webui_traceability_explorer` is used for both treemap and treeview rendering in `app.js`.

### Where This Pattern Exists Today
- `webui/static/app.js` lines:
    - 244, 296, 384: `@spec-link [[ui_webui_traceability_explorer]]`
- `webui/static/spec-builder.js` lines:
    - 54, 227, 358, 379, 534, 554, 732, 1240, 1301: `@spec-link [[ui_webui_spec_builder]]`
    - 864, 918, 967, 1137: `@spec-link [[rule_webui_context_history_management]]`

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High |
| Detectability | High — evident via file sizes and ATD trace alerts |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** 
1.  Decompose `app.js` into smaller modules (e.g., `explorer.js`, `details.js`, `api.js`).
2.  Decompose `spec-builder.js` by extracting proposal handling and chat rendering into separate files.

**Medium term:** 
1.  Increase ATD granularity:
    *   Split `ui_webui_spec_builder` into `ui_webui_chat_input`, `ui_webui_message_list`, `ui_webui_proposal_card`, etc.
    *   Split `ui_webui_traceability_explorer` into `ui_webui_treemap` and `ui_webui_treeview`.
2.  Repopulate `@spec-link` tags to point to the new, more granular atoms.

**Long term:** 
1.  Consider a modern component-based architecture or simple ES modules to manage state and rendering more effectively.
2.  Ensure each atom describes exactly ONE state-changing rule or focused UI component.

---

## References

- [webui/static/app.js](file:///home/bastien/work/skill/webui/static/app.js)
- [webui/static/spec-builder.js](file:///home/bastien/work/skill/webui/static/spec-builder.js)
- [ui_webui_spec_builder.atom.md](file:///home/bastien/work/skill/docs/ui_webui_spec_builder.atom.md)
- [ui_webui_traceability_explorer.atom.md](file:///home/bastien/work/skill/docs/ui_webui_traceability_explorer.atom.md)
