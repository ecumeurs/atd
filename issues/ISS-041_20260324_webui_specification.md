# Issue: WebUI Specification

**ID:** `20260324_webui_specification`
**Ref:** `ISS-041`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Open
**Component:** `webui/`
**Affects:** All WebUI-related issues (ISS-004 through ISS-008)

---

## Summary

The WebUI currently exists only as the kernel of an idea — basic rendering with no formal specification, no connection to the ATD binary tools, and no defined user experience. Further UI development is blocked until a proper specification exists as ATD atoms.

---

## Technical Description

### Background
Several UI-related issues have been filed (ISS-004 navigation, ISS-005 HTML rendering, ISS-006 edit propagation, ISS-007 binary link, ISS-008 LLM integration) but they lack a coherent architectural specification.

### The Problem Scenario
1. A developer wants to improve the WebUI.
2. There are no `MODULE`, `UI`, or `SERVICE` atoms defining the intended architecture, components, or user flows.
3. Development is ad-hoc, with no traceability to requirements.

### Where This Pattern Exists Today
- `webui/` — existing code with no ATD specification.
- ISS-004 through ISS-008 — individual issues without architectural context.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — ad-hoc development, rework risk |
| Detectability | High — absence of spec is obvious |
| Current mitigant | Individual issue descriptions |

---

## Recommended Fix

**Short term:** Create a set of ATD atoms (`MODULE`, `UI`, `SERVICE` types) specifying: graph visualization, atom CRUD operations, audit dashboards, cold-start orchestration, integration with `atd` CLI/MCP tooling.
**Medium term:** Link all existing WebUI issues to the new specification atoms. Use the spec to prioritize and sequence UI development.
**Long term:** The WebUI specification should be the first comprehensive `CUSTOMER` + `ARCHITECTURE` domain test case for ATD itself — dogfooding the framework.

---

## References

- [ATD.md §3.2.10](file:///home/bastien/work/skill/ATD.md)
- [ISS-004](ISS-004_20260304_webui_navigation.md)
- [ISS-005](ISS-005_20260304_webui_html_rendering.md)
- [ISS-006](ISS-006_20260304_webui_edit_propagation.md)
- [ISS-007](ISS-007_20260304_webui_binary_link.md)
- [ISS-008](ISS-008_20260304_webui_llm_integration.md)
