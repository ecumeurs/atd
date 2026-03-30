# Issue: ATD Summarization Feature for Atoms

**ID:** `20260330_atd_summarization_feature`
**Ref:** `ISS-054`
**Date:** 2026-03-30
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd`, `scripts/cmd/atd-serve`
**Affects:** CLI Users, MCP Clients

---

## Summary

Add a new `summary` command to both the ATD CLI and MCP server. This feature allows users to get a concise, context-aware summary of any specific atom ID. The summary must incorporate information from the atom's entire ancestry (parents) and descendants to provide a holistic view of its role in the system.

---

## Technical Description

### Background
Currently, users can use `atd trace` to see health metrics and graph relationships, or `atd assemble` to stitch atoms into a narrative. However, there is no direct way to get a focused summary of a single atom that automatically pulls in relevant context from its neighbors in the dependency graph.

### The Problem Scenario
When a developer or stakeholder wants to understand a specific `MECHANIC` or `RULE`, they often need to manually look up its `parents` (to understand "Why?") and its `dependents` (to understand "What uses this?"). A "Summary" feature would automate this synthesis.

### Functional Requirements
1.  **Contextual Awareness**: The summary must traverse the graph to include:
    *   Upper layers (Customer Requirements) to explain the business value.
    *   Middle layers (Architectural Decisions) to explain the design.
    *   Lower layers (Implementations) to show how it's realized.
2.  **Structured Output**:
    *   First: Customer Requirements.
    *   Second: Architectural Decisions.
    *   Third: Implementations.
    *   The "Meat": The target atom's logic and intent should be the primary focus.
3.  **Configurable Length**:
    *   `short`: ~100 words.
    *   `default`: ~300 words.
    *   `extended`: ~500 words.
    *   `long`: ~1000 words.
4.  **Metadata**: Append a list of all involved `atom_id`s with their file paths, layers, and types (not counted in word limits).
5.  **LLM Support**:
    *   Use configured LLM models (e.g., Ollama, see `.atd config`) via guided prompting.
    *   **LLM Light Option**: A fallback mode that aggregates `intent` sections by layer and summarizes them simply, avoiding complex cross-atom reasoning if resources are limited.
6.  **Traceability**: Use `[[atom_id]]` syntax within the summary for quick linking.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Feature is requested) |
| Impact if triggered | Medium (Improves developer onboarding and visibility) |
| Detectability | High — command will either return a summary or fail |
| Current mitigant | Manual lookup of related atoms via `atd trace` or `grep` |

---

## Recommended Fix

**Short term:** Implement a basic version that aggregates `intent` sections of parents and descendants and presents them in the requested order.
**Medium term:** Integrate with the LLM provider to generate more natural and synthesized summaries based on the full content of the atoms.
**Long term:** Add this to the WebUI for interactive exploration.

---

## References

- [ATD.md](file:///home/bastien/work/skill/.agent/rules/ATD.md)
- Existing tools in `scripts/cmd/` (trace, assemble, etc.)
