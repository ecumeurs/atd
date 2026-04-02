# Issue: WebUI Monolith and Inappropriate ATD Granularity

**ID:** `20260330_webui_monolith_architecture`
**Ref:** `ISS-056`
**Date:** 2026-03-30
**Severity:** High
**Status:** Resolved
**Component:** `webui/`
**Affects:** `webui/main.go`, ATD index

---

## Summary

The `webui` component is currently implemented as a single, bloated `main.go` file (750+ lines). This monolithic structure makes it difficult to maintain, test, and evolve. Furthermore, the ATD mapping is not granular enough: multiple distinct functionalities (Gemini proxy, atom search, atom details, chat) are all linked to the same ATD `mechanic_webui_gemini_proxy`, violating the "single responsibility" principle of ATD.

---

## Technical Description

### Background
The `webui` was intended to provide a conversational interface for ATD management. It uses Gin for the web server and interacts with the Gemini API and the local ATD toolkit.

### The Problem Scenario
1.  **Monolith:** All logic, from configuration loading to complex Gemini chat handling and ATD updates via CLI execution, lives in `main.go`. This includes a massive string constant `atdManifesto`.
2.  **Granularity Issue:**
    *   `@spec-link [[mechanic_webui_gemini_proxy]]` is applied to:
        *   `GET /api/gemini/models`
        *   `GET /api/gemini/atoms`
        *   `GET /api/gemini/atom/:id`
        *   `func handleGeminiChat`
    *   These are logically distinct operations (listing models, querying the local index, fetching atom details, and orchestrating a LLM chat).

### Where This Pattern Exists Today
- `webui/main.go` lines:
    - 351: `@spec-link [[mechanic_webui_gemini_proxy]]`
    - 397: `@spec-link [[mechanic_webui_gemini_proxy]]`
    - 416: `@spec-link [[mechanic_webui_gemini_proxy]]`
    - 593: `@spec-link [[mechanic_webui_gemini_proxy]]`

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High |
| Detectability | High — evident via code bloat and ATD trace alerts |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** 
1.  Extract `atdManifesto` and configuration types into separate files or packages.
2.  Split `main.go` into multiple files within the `webui` package (e.g., `routes.go`, `handlers_gemini.go`, `handlers_atd.go`, `config.go`).

**Medium term:** 
1.  Decompose `mechanic_webui_gemini_proxy` into more granular mechanics:
    *   `mechanic_webui_gemini_model_list`
    *   `mechanic_webui_atom_search`
    *   `mechanic_webui_atom_detail_fetch`
    *   `mechanic_webui_gemini_chat_orchestration`
2.  Update `@spec-link` tags in the code to reflect these new atoms.

**Long term:** 
1.  Refactor `webui` to use a cleaner architecture (e.g., separating transport/delivery from business logic).
2.  Ensure each atom describes exactly ONE state-changing rule or focused mechanic as per ATD mandate.

---

## References

- [webui/main.go](file:///home/bastien/work/skill/webui/main.go)
- [mechanic_webui_gemini_proxy.atom.md](file:///home/bastien/work/skill/docs/mechanic_webui_gemini_proxy.atom.md) (presumed path)
