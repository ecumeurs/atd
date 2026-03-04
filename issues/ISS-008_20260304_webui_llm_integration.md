# Issue: WebUI LLM Integration for Content Iteration

**ID:** `ISS-008_20260304_webui_llm_integration`
**Ref:** `ISS-008`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `webui`
**Affects:** `webui/backend`, `scripts/atd-ollama-generate`

---

## Summary

Link the WebUI to a "true" LLM (via Ollama or an API) to allow for content iteration and spec drafting directly within the interface.

---

## Technical Description

### Background
ATD creation is currently either manual or via CLI-based generation scripts.

### The Problem Scenario
Users cannot easily iterate on ATD content using AI assistance while viewing the ATD graph in the WebUI.

### Where This Pattern Exists Today
- `webui/`
- `scripts/atd-ollama-generate`

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium |
| Detectability | High |
| Current mitigant | Manual drafting |

---

## Recommended Fix

**Short term:** Provide an "AI Improve" button for selected fields that calls the local Ollama instance.
**Medium term:** Implement a chat-like interface in a sidebar for interactive spec drafting.
**Long term:** Integrate with the full `atd-dissect` pipeline to allow uploading documents for AI-assisted ATD deconstruction via the WebUI.

---

## References

- [atd-ollama-generate](../scripts/atd-ollama-generate)
