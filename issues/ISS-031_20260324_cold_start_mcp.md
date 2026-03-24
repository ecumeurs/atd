# Issue: Cold-Start and Full Audit via MCP

**ID:** `20260324_cold_start_mcp`
**Ref:** `ISS-031`
**Date:** 2026-03-24
**Severity:** High
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/serve.go`
**Affects:** MCP consumers, IDE agents

---

## Summary

The cold-start pipeline (`roadmap` → `index` → `dissect` → `weave` → `discover` → `recon` → `audit`) is only available as a shell script (`atd-cold-start.sh`). It cannot be triggered from MCP, forcing MCP-only consumers (IDE agents) to manually orchestrate 7+ sequential tool calls. Similarly, a full audit (bloat + collision + gap analysis) requires multiple separate MCP tool invocations with no unified entry point.

---

## Technical Description

### Background
The `atd-cold-start.sh` script orchestrates the entire bootstrapping pipeline in sequence. MCP exposes each step as an individual tool, but provides no composite operation.

### The Problem Scenario
1. An IDE agent needs to bootstrap ATD on a new project via MCP.
2. The agent must call `atd_roadmap`, then `atd_index`, then `atd_dissect` (for each file), then `atd_weave`, then `atd_discover`, then `atd_recon`, then `atd_audit` — in exact order with error handling.
3. This is error-prone, token-expensive (many round-trips), and duplicates orchestration logic that already exists in the shell script.

### Where This Pattern Exists Today
- `scripts/atd-cold-start.sh`: Shell pipeline not exposed to MCP.
- `scripts/cmd/atd/cmd/mcp_tools.go`: No composite tool registered.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High — blocks MCP-only bootstrapping |
| Detectability | High — users will attempt cold-start via MCP and fail |
| Current mitigant | Manual shell script execution outside MCP |

---

## Recommended Fix

**Short term:** Expose `atd_cold_start` as a new MCP tool that wraps the pipeline. Accept `dir` (project root) and `src` (source directory) as parameters. Return progress updates and final results.
**Medium term:** Similarly, expose `atd_full_audit` as a composite MCP tool that runs bloat check + collision detection + gap analysis + test-link validation in a single call.
**Long term:** Support streaming progress events via MCP notifications so the IDE agent can display real-time pipeline status.

---

## References

- [atd-cold-start.sh](file:///home/bastien/work/skill/scripts/atd-cold-start.sh)
- [mcp_tools.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/mcp_tools.go)
- [ATD.md §1.6](file:///home/bastien/work/skill/ATD.md)
