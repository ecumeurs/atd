# Issue: `atd trace --summary` Narrative Mode

**ID:** `20260503_atd_trace_narrative_summary`
**Ref:** `ISS-094`
**Date:** 2026-05-03
**Severity:** Medium
**Status:** Open
**Component:** `atd/pkg/exploration`, `atd/cmd/atd/cmd`
**Affects:** Agent workflow, WebUI, CLI

---

## Summary

Currently, `atd trace` provides a purely structural and quantitative snapshot (JSON) of an atom's lineage. While useful for automation, it lacks the narrative context required for an agent or human to quickly understand the "Story" of an atom: why it exists (Ancestry), what it specifically does (Objective), and what it forces upon the system (Downward Impact). 

This issue proposes adding a `--summary` flag to `atd trace` that uses an LLM to generate a ~500-word contextual analysis based on the vertical slice of the graph.

---

## Technical Description

### Background
`atd_trace` currently generates a `TraceSnapshot` which identifies all parents and dependents. The WebUI has a "Summary" tab that performs intent-aggregation, but this is not exposed to the CLI or MCP tools in a narrative format.

### The Problem Scenario
An agent entering a codebase to fix a bug in a `MECHANIC` atom needs to know if changing that mechanic violates a `BUSINESS` rule. Currently, the agent must manually read multiple files in the ancestry chain. A narrative summary would provide this context in a single token-efficient pass.

### Implementation Requirements
1. **CLI/MCP Update**: Add a `summary` boolean flag to `atd_trace`.
2. **Context Aggregation**: When `summary` is true, the tool must not only find the IDs of parents/dependents but also extract their:
   - `human_name`
   - `type`
   - `layer`
   - `intent`
   - `logic` (The Rule)
3. **The "Contextual Auditor" Prompt**: A new prompt in `atd/pkg/prompt` that takes the aggregated context and generates a structured narrative:
   - **Upper Context**: The "Why" from the parents.
   - **Current Objective**: The "What" of the target atom.
   - **Downward Impact**: The "How" it affects descendants.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — helps prevent scope creep and logic violations |
| Detectability | High — summary will be visible in agent logs |
| Current mitigant | Manual reading of parent ATDs |

---

## Recommended Fix

**Short term:** Implement the `atd trace --summary` flag and the basic intent-aggregation logic.  
**Medium term:** Integrate the "Contextual Auditor" prompt with the LLM provider (Ollama/Nomic).  
**Long term:** Make this summary a mandatory pre-requisite for any Agent "Edit" action via `CLAUDE.md` behavioral triggers.

---

## References

- [ATD.md](file:///home/bastien/work/skill/ATD.md)
- [atd/cmd/atd/cmd/trace.go](file:///home/bastien/work/skill/atd/cmd/atd/cmd/trace.go)
- [atd/pkg/exploration/exploration.go](file:///home/bastien/work/skill/atd/pkg/exploration/exploration.go)
