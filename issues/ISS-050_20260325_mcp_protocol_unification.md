# Issue: Unified Cold-Start and Audit Protocol as MCP Tools

**ID:** `20260325_mcp_protocol_unification`
**Ref:** `ISS-050`
**Date:** 2026-03-25
**Severity:** High
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/serve.go`
**Affects:** MCP consumers, IDE agents

---

## Summary

The full cold-start and auditing protocols (multi-step pipelines) should be exposed as first-class MCP tools. While individual steps (roadmap, index, dissect, etc.) exist, the full *protocol* that orchestrates them into a cohesive workflow is currently restricted to shell scripts or requires 7+ sequential round-trips from the IDE agent. This hinders the agent's ability to efficiently bootstrap or verify a project in restricted environments.

---

## Technical Description

### Background
This issue expands upon and consolidates the requirements mentioned in **ISS-031**. The user specifically requested that the "auditing and cold start protocol" be available as MCP tools.

### The Problem Scenario
1. An IDE agent is tasked with "bootstrapping this project" or "running a full audit".
2. The agent must navigate the complexity of ordering `roadmap`, `index`, `dissect`, `weave`, etc.
3. If any step fails or requires specific flags, the agent must handle this logic, which is already perfected in the `atd-cold-start.sh` and `atd-audit.sh` (or internal logic).

### Where This Pattern Exists Today
- `scripts/atd-cold-start.sh`: Orchestrates the bootstrap.
- `scripts/cmd/atd/cmd/mcp_tools.go`: Missing a `atd_protocol_cold_start` or similar composite tool.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High — high barrier to entry for new project ATD adoption |
| Detectability | High — evident when trying to bootstrap via MCP |
| Current mitigant | Manual orchestration by the agent (unreliable) |

---

## Recommended Fix

**Short term:** Create a new MCP tool `atd_protocol_cold_start` that executes the full pipeline sequentially and returns a summary.
**Medium term:** Create `atd_protocol_full_audit` that aggregates bloat, collision, gap, and trace results.
**Long term:** Support a "workflow" or "pipeline" primitive in the ATD Go core that can be invoked via CLI, WebUI, or MCP identically.

---

## References

- [ISS-031](file:///home/bastien/work/skill/issues/ISS-031_20260324_cold_start_mcp.md)
- [atd-cold-start.sh](file:///home/bastien/work/skill/scripts/atd-cold-start.sh)
- [mcp_tools.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/mcp_tools.go)
