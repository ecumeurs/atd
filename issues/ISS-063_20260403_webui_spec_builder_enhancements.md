# Issue: WebUI Spec Builder Enhancements and Bug Fixes

**ID:** `20260403_webui_spec_builder_enhancements`
**Ref:** `ISS-063`
**Date:** 2026-04-03
**Severity:** High
**Status:** Open
**Component:** `scripts/pkg/webui`
**Affects:** `scripts/pkg/webui/static/js/app.js`, `scripts/pkg/webui/handlers_gemini.go`, `scripts/pkg/webui/static/styles.css`, `scripts/pkg/webui/static/spec-builder.css`

---

## Summary

The Gemini Spec Builder tab in the WebUI requires several enhancements to improve usability and fix logical inconsistencies. These include UI fixes for card overflow and panel resizing, new features for editing proposals and requesting elaboration, and a fix for a grouping bug affecting proposal updates.

---

## Technical Description

### Background
The Spec Builder tab provides a conversational interface for generating and refining ATD atoms. Proposed atoms are displayed as cards in a side panel for review, acceptance, or rejection.

### The Problem Scenario
1. **Overflow Issues:** When the list of proposals grows long, the current layout does not handle overflow gracefully, preventing cards from expanding to show their full content.
2. **Editing Shortcoming:** Users cannot edit the content of proposed cards directly. Edits are essential for fine-tuning before acceptance and must be communicated back to the LLM to maintain context.
3. **Panel Constraint:** The side panel is currently fixed at approximately 1/6 to 1/5 of the screen width, which is insufficient for comfortable reading and editing of complex atom specifications.
4. **Missing Elaboration Flow:** There is no dedicated way to ask the LLM to "flesh out" a specific proposal without manual prompting, which may dilute the LLM's focus.
5. **Grouping Bug:** Proposals with the same ID that are neither validated nor rejected are sometimes duplicated in subsequent chat iterations instead of being updated/grouped as expected.

### Where This Pattern Exists Today
- UI Layout: `scripts/pkg/webui/static/styles.css`, `scripts/pkg/webui/static/spec-builder.css` and `scripts/pkg/webui/static/js/app.js` (Spec Builder tab logic).
- Backend Logic: `scripts/pkg/webui/handlers_gemini.go` (Proposal handling and LLM interaction).

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | High — visible UI glitches and logic duplication |
| Current mitigant | Manual prompting and browser inspector hacks |

---

## Recommended Fix

**Short term:** 
- Fix CSS for card overflow and implement basic flex/grid adjustments for the side panel.
- Add an "Edit" mode to proposal cards and a simple "Elaborate" button that sends a specialized prompt.

**Medium term:** 
- Implement a draggable divider or toggle to allow the side panel to grow up to 1/2 screen width.
- Refactor the frontend proposal state management to ensure consistent grouping by ID across iterations.
- Update the LLM manifesto/system prompt to handle elaboration requests more effectively.

**Long term:** 
- Implement a full bidirectional synchronization between the UI state and the LLM's internal representation of the "proposed graph".

---

## References

- [app.js](file:///home/bastien/work/skill/scripts/pkg/webui/static/js/app.js)
- [handlers_gemini.go](file:///home/bastien/work/skill/scripts/pkg/webui/handlers_gemini.go)
- [styles.css](file:///home/bastien/work/skill/scripts/pkg/webui/static/styles.css)
- [spec-builder.css](file:///home/bastien/work/skill/scripts/pkg/webui/static/spec-builder.css)
