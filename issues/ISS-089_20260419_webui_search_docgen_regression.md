# Issue: WebUI Search and Document Generation Regression

**ID:** `20260419_webui_search_docgen_regression`
**Ref:** `ISS-089`
**Date:** 2026-04-19
**Severity:** High
**Status:** Open
**Component:** `webui/`
**Affects:** `Ctrl + K` (Command Palette), Document Generation Tool

---

## Summary

Critical functionality in the WebUI has regressed:
1. The **Command Palette (Ctrl + K)** is broken and fails to find ATD atoms.
2. **Document Generation** is also broken, preventing users from creating narratives from their ATD graph.

---

## Technical Description

### Background
The Command Palette and Document Generation are the two primary "AI-aided" entry points for interacting with the ATD graph. Search uses a hybrid of local caching and backend semantic search, while Document Generation uses an LLM-backed orchestration to thread atoms together.

### The Problem Scenario
- **Search (Ctrl + K):** The user triggers search and types a query. No results are returned, or the UI fails to react.
- **Document Generation:** The tool fails to initiate or produces errors during the assembly phase.
This seems to be a regression from previous stable states (possibly related to recent changes in the search handlers or frontend connectivity).

### Where This Pattern Exists Today
- `scripts/pkg/webui/static/js/search.js`
- `scripts/pkg/webui/handlers.go`
- `atd/pkg/exploration/` (if the underlying search logic moved)

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Reported as "broken") |
| Impact if triggered | High — core interactive features are offline |
| Detectability | High |
| Current mitigant | Using the CLI directly for searching and stats. |

---

## Recommended Fix

**Short term:**
- Investigate the network console and server logs to identify the exact point of failure for `Ctrl + K`.
- Verify if the `/api/search` and `/api/assemble` endpoints are still responding correctly.
- Fix immediate regressions in event listeners or API call paths.

**Medium term:**
- Implement a suite of integration tests for the WebUI that cover search and doc-gen.
- Ensure the state management in the frontend is robust against partial index failures.

---

## References

- [ISS-064: Search Regression (Older)](file:///home/bastien/work/skill/issues/ISS-064_20260403_webui_search_regression_and_perf.md)
- [ISS-061: Doc Gen Bugs (Older)](file:///home/bastien/work/skill/issues/ISS-061_20260402_webui_document_generation_bugs.md)
- [scripts/pkg/webui/static/js/search.js](file:///home/bastien/work/skill/scripts/pkg/webui/static/js/search.js)
