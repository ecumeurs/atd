# Issue: Research efficient API logic tracking for ATD atoms

**ID:** `20260306_api_logic_tracking_research`
**Ref:** `ISS-020`
**Date:** 2026-03-06
**Severity:** Medium
**Status:** Open
**Component:** `atd_management_skill`
**Affects:** Efficiency of ATD generation and LLM context usage

---

## Summary

Current API-typed Atoms use free-form text or simplified summaries that often lose technical precision (payload nesting, side-effects, etc.). This research issue aims to identify structured tools or formats (e.g., TypeSpec, Smithy, OpenAPI) that can efficiently track API logic in a way that remains parsable by LLMs while maintaining high fidelity for ATD documentation.

---

## Technical Description

### Background
The ATD system relies on Markdown files to store "Atoms". When an Atom describes an API, it needs to convey enough technical detail for a developer (or an agent) to implement or call it correctly. Plain text descriptions are prone to loss of detail during LLM-assisted generation.

### The Problem Scenario
A user provides a complex JSON payload and several behavioral rules for an endpoint. The current LLM-based extraction might condense this into: "Endpoint X takes a JSON payload and starts a game." This loses the specific field names, types, and the exact clock rules mentioned in the source material.

### Research Goals
- Identify schema-first languages that are "LLM-friendly" (low token count, high signal).
    - **Found:** TypeSpec (TypeScript-like, highly modular) and Smithy (IDL-based, protocol agnostic).
- Evaluate if embedding these schemas directly into ATD Markdown files is feasible.
- Explore tools that can "re-hydrate" a brief ATD summary into a full technical contract by interrogating source code or external specs.
    - **Found:** Model Context Protocol (MCP) is the emerging standard for this re-hydration/interrogation flow.

### Findings Summary
- **TypeSpec (Microsoft):** Excellent for LLMs due to modularity and official MCP emitters. It provides strong typing "guiderails" for agents.
- **Smithy (AWS):** Features specific `smithy.ai#prompts` traits to embed LLM guidance directly in the IDL. High abstraction helps LLMs ignore transport details.
- **OpenAPI:** Industry standard but often too verbose for context windows. Best used as a secondary artifact generated from TypeSpec/Smithy.
- **MCP (Model Context Protocol):** Critical for efficiency. Instead of putting full specs in the ATD, the ATD can point to an MCP-capable endpoint that provides the spec on-demand.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | N/A (Research) |
| Impact if triggered | High (Potential for significantly better documentation) |
| Detectability | Easy to compare structured formats vs plain text |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** Conduct literature/tooling review for API description languages tailored for LLMs.
**Medium term:** Prototype an ATD extension that supports a specific structured format (e.g., TypeSpec).
**Long term:** Standardize all `type: API` Atoms to use this structured format.

---

## References

- [ISS-019](ISS-019_20260306_api_atd_payload_capture_shortcoming.md) (Related issue regarding capture shortcomings)
- [atd_management_skill](file:///home/bastien/work/skill/atd_management_skill/)
