# Issue: MCP Tools Review and Parameter Cleanup

**ID:** `20260324_mcp_tools_refactor_and_cleanup`
**Ref:** `ISS-045`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/mcp_tools.go`
**Affects:** Agent LLM, atd-tools MCP consumers

---

## Summary

The current MCP tools expose internal implementation details (like file paths and debug flags) that should be handled by the `.atd` configuration. Furthermore, tool descriptions and parameter schemas are not optimized for Agent LLMs, leading to potential misuse or confusion. Specific tools like `atd_dissect`, `atd_index`, and `atd_search` require structural simplifications to align with the framework's architectural intent.

---

## Technical Description

### Background
MCP tools should provide a clean, high-level interface for the Agent LLM to interact with the ATD framework. The framework is designed to be self-configuring via `.atd` files, meaning the Agent should not need to specifies paths like `--docs` or `--src` in normal operation.

### The Problem Scenario
The current schema registration in `mcp_tools.go` prints several "leaky" parameters to the Agent:

1. **Leaky Parameters**: Parameters like `docs`, `src`, `threshold`, and `db` appear in many tool schemas. This forces the Agent to "guess" the workspace structure instead of letting the server use its internal knowledge.
2. **Implementation Logic Leak**: `atd_dissect` has an `llm` boolean. The Agent shouldn't care *if* an LLM is used; it should just request the dissection, and the framework should follow the `./.atd` configuration for provider delegation.
3. **Database Mismanagement**: `atd_index` and `atd_search` expose raw SQLite database paths (`db`). The user has clarified that `atd` should maintain exactly two databases (one for docs, one for codebase), and the tools should automatically select the correct one based on `scope`.

```
Currently:
atd_search(query="...", db="/path/to/docs/.atd_index.db", scope="docs")

Desired:
atd_search(query="...", scope="docs") 
// Framework knows where the doc index is
```

### Where This Pattern Exists Today
- `scripts/cmd/atd/cmd/mcp_tools.go`: The `RegisterMCPTools` function (lines 65-460) defines all schemas.
- Specifically:
  - `atd_dissect`: `llm` parameter (line 259)
  - `atd_index`: `dir`, `db`, `mode` parameters (lines 274-278)
  - `atd_search`: `db` parameter (line 300)

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — leads to sub-optimal tool usage and fragile prompts |
| Detectability | High — visible in the Agent's tool list |
| Current mitigant | Some defaults are provided in the Go code, but the parameters are still exposed in the JSON-RPC schema. |

---

## Recommended Fix

**Short term:** 
- Review and refine tool descriptions in `mcp_tools.go` to be more "Agent-centric".
- Remove `docs`, `src`, `threshold`, and `db` from the `InputSchema` of all tools. Hardcode the implementation to use `config.DocsDir()` and `config.ActiveConfig` values.

**Medium term:** 
- Refactor `atd_index` to remove `dir` and `mode`. It should just index both target areas (docs and code) into their respective databases. Add an optional `target` for focused indexing.
- Remove `llm` from `atd_dissect`. Implement provider selection logic in `runDissect` based on `.atd` config.

**Long term:** 
- Implement a "Narrative Schema" system where tool descriptions are dynamically generated to provide better context to the LLM based on the current project's state.

---

## References

- [mcp_tools.go](scripts/cmd/atd/cmd/mcp_tools.go)
- [config/config.go](scripts/cmd/atd/config/config.go)
