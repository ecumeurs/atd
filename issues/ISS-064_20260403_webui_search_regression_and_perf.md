# Issue: WebUI Search Regression and Performance

**ID:** `20260403_webui_search_regression_and_perf`
**Ref:** `ISS-064`
**Date:** 2026-04-03
**Severity:** High
**Status:** Open
**Component:** `scripts/pkg/webui`
**Affects:** `scripts/pkg/webui/static/js/search.js`, `scripts/pkg/webui/handlers.go`, `scripts/pkg/webui/static/js/api.js`

---

## Summary

The "Ctrl+K" command palette in the WebUI is currently failing to reliably find atoms by their ID or Name. Furthermore, the search process is perceived as slow because it primarily relies on semantic (embedding-based) search without providing a fast-path for exact string matches or adequate intermediate feedback to the user.

---

## Technical Description

### Background
The WebUI provides a search overlay (Ctrl+K) to quickly locate ATD atoms. This feature is critical for navigating the growing graph of documentation.

### The Problem Scenario
1. **Broken Lookup:** Searching for a known Atom ID (e.g., `ui_issue_tab`) or a specific human name often returns no results or irrelevant semantic matches, instead of prioritizing the exact match.
2. **High Latency:** Semantic search via Nomic embeddings is computationally expensive compared to literal string matching. 
3. **Lack of Feedback:** While a "Searching..." hint exists, the UI remains static and unresponsive during the potentially multi-second embedding/search phase.
4. **As-you-type Latency:** Multiple search requests are triggered sequentially as the user types, each incurring the full embedding cost:
   ```text
   [GIN] 2026/04/03 - 09:40:12 | 200 |    2.33s |             ::1 | GET      "/api/search?q=how"
   [LLM] Task=embed Model=nomic-embed-text:latest Provider=local
   [GIN] 2026/04/03 - 09:40:13 | 200 |    2.31s |             ::1 | GET      "/api/search?q=how%20are"
   [LLM] Task=embed Model=nomic-embed-text:latest Provider=local
   [GIN] 2026/04/03 - 09:40:15 | 200 |    2.31s |             ::1 | GET      "/api/search?q=how%20are%20audit%20handled"
   [LLM] Task=embed Model=nomic-embed-text:latest Provider=local
   ```
5. **Missing Fast-Path:** The backend `handleSearch` implementation does not appear to prioritize name/ID matches before falling back to the vector database.

### Where This Pattern Exists Today
- Frontend: `scripts/pkg/webui/static/js/search.js` (UI logic) and `scripts/pkg/webui/static/js/api.js` (API call).
- Backend: `scripts/pkg/webui/handlers.go` (`handleSearch` using `exploration.Search`).

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High — breaks core navigation and usability |
| Detectability | High — users immediately notice search failure |
| Current mitigant | Browsing the tree manually or using the CLI `atd query` |

---

## Recommended Fix

**Short term:** 
- Implement a "Fast-Path" in the frontend (`search.js`) that first filters the locally cached `state.atoms` by ID and Name to provide instantaneous results.
- Ensure the "Searching..." status is highly visible (e.g., a spinner or progress bar).

**Medium term:** 
- Update the backend `/api/search` handler to perform a hybrid search: combine exact matches for ID/Name with semantic results.
- Implement a "Search as you type" behavior that populates local matches first, then overlays semantic results asynchronously.

**Long term:** 
- Optimize the embedding indexing and search performance in `atd-tools`.

---

## References

- [search.js](file:///home/bastien/work/skill/scripts/pkg/webui/static/js/search.js)
- [handlers.go](file:///home/bastien/work/skill/scripts/pkg/webui/handlers.go)
- [api.js](file:///home/bastien/work/skill/scripts/pkg/webui/static/js/api.js)
