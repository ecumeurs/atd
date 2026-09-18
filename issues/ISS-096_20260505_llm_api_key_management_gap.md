# Issue: Secure LLM API Key Management Gap for CLI/Backend

**ID:** `20260505_llm_api_key_management_gap`
**Ref:** `ISS-096`
**Date:** 2026-05-05
**Severity:** Medium
**Status:** Open
**Component:** `atd/config`
**Affects:** `atd` (CLI), `atd_check`, `atd_audit`

---

## Summary

The current implementation of OpenAI/Minimax support relies on browser local storage for API keys. While this works for the WebUI, it leaves the CLI and other backend-driven ATD tools (like `atd_check --semantic`) without a secure mechanism to retrieve these keys. This prevents the full utilization of cost-effective LLMs across the entire toolchain.

---

## Technical Description

### Background
The WebUI was recently updated to support OpenAI-compatible providers. To avoid leaking secrets into the `.atd` manifest (which is often committed to git), keys are stored in the browser's `localStorage`.

### The Problem Scenario
1. A user configures Minimax/OpenAI in the WebUI.
2. The user runs `atd_check --semantic` or `atd_audit` from the terminal.
3. These tools read the `.atd` config, see the "openai" provider type, but have no way to access the API key stored in the browser.
4. The tools fail or revert to local Ollama instances, leading to inconsistent results between the UI and CLI.

### Where This Pattern Exists Today
- `atd/pkg/webui/static/js/spec-builder.js` (Key storage in UI)
- `atd/pkg/chat/openai.go` (Expects a key but lacks a generic retrieval mechanism for CLI)
- `.atd` configuration schema (Lacks a placeholder or environment variable mapping for keys)

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium (Feature fragmentation) |
| Detectability | High — CLI tools will report auth errors or missing keys |
| Current mitigant | None for CLI; only WebUI is functional with these providers |

---

## Recommended Fix

**Short term:** Implement support for environment variables (e.g., `ATD_OPENAI_API_KEY`) that the Go backend can pick up when running in CLI mode.
**Medium term:** Standardize a `.atd.keys` or similar file (auto-added to `.gitignore`) to store local credentials for non-browser environments.
**Long term:** Implement a unified credential manager or vault integration for enterprise/shared environments.

---

## References

- [config.go](file:///home/bastien/work/atd/atd/config/config.go)
- [openai.go](file:///home/bastien/work/atd/atd/pkg/chat/openai.go)
- [handlers_llm.go](file:///home/bastien/work/atd/atd/pkg/webui/handlers_llm.go)
