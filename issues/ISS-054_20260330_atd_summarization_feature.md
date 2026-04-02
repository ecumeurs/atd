# Issue: ATD Summarization Feature for Atoms

**ID:** `20260330_atd_summarization_feature`
**Ref:** `ISS-054`
**Date:** 2026-03-30
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/cmd/atd`, `scripts/cmd/atd-serve`
**Affects:** CLI Users, MCP Clients

---

## Summary

Add a summarization capability avoiding redundant commands by extending `atd assemble`. This feature allows users to get a concise, context-aware summary of any specific atom ID. The summary must incorporate information from the atom's entire ancestry (parents) and descendants to provide a holistic view of its role in the system.

---

## Technical Description

### Background
Currently, users can use `atd trace` to see health metrics and graph relationships, or `atd assemble` to stitch atoms into a narrative. However, there is no direct way to get a focused summary of a single atom that automatically pulls in relevant context from its neighbors in the dependency graph.

### The Problem Scenario
When a developer or stakeholder wants to understand a specific `MECHANIC` or `RULE`, they often need to manually look up its `parents` (to understand "Why?") and its `dependents` (to understand "What uses this?"). A "Summary" feature would automate this synthesis.

### Functional Requirements
1.  **Contextual Awareness**: The summary must traverse the graph to include: Upper layers, Middle layers, Lower layers.
2.  **Structured Output**: Group atoms by layer when using `--structured`.
3.  **Configurable Length**:
    *   `short`: ~100 words.
    *   `default`: ~300 words.
    *   `extended`: ~500 words.
    *   `long`: ~1000 words.
4.  **Metadata**: Append a list of all involved `atom_id`s with their file paths, layers, and types (not counted in word limits).
5.  **LLM Support**:
    *   Use configured LLM models (e.g., Ollama, see `.atd config`) via guided prompting. Make a multi-pass request grouping by layer if `--structured` is active.
    *   **LLM Light Option**: A fallback mode that aggregates `intent` sections sequentially.
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

**Resolution (2026-04-01):** We decided NOT to implement a redundant `atd summary` command. Instead, `atd assemble` was heavily refactored:
- Prompts were rewritten to support intent-based processing and length constraints.
- A `--structured` flag was added to do multi-pass LLM summaries by Layer.
- A `--json` flag was added to output `{"customer_layer": "...", "content": "...", "metadata": [...]}` perfectly formatted for IDE and WebUI ingestion.
- The `mcp_tools.go` schema for `atd_assemble` was updated to expose these properties.

---

## References

- [ATD.md](file:///home/bastien/work/skill/.agent/rules/ATD.md)
- Existing tools in `scripts/cmd/` (trace, assemble, etc.)
