# Issue: ATD Type System Redundancy and Confusion

**ID:** `20260418_atd_type_system_simplification`
**Ref:** `ISS-075`
**Date:** 2026-04-18
**Severity:** Medium
**Status:** Open
**Component:** `ATD.md`, `scripts/pkg/atom/types.go`, documentation system
**Affects:** Agent Decision Making, Documentation Clarity, Atom Creation Guidance

---

## Summary

Current ATD type system contains 13 types with significant overlap and redundancy (USECASE vs USER_STORY, SERVICE vs MODULE, etc.), causing confusion for agents and developers when choosing appropriate atom types.

---

## Technical Description

### Background
ATD atoms are categorized into types to guide their purpose and granularity. The type system should be clear and mutually exclusive.

### The Problem Scenario
1. **Redundant Types**: USECASE and USER_STORY describe similar content with unclear distinction
2. **Overloaded Types**: SERVICE and MODULE both describe architectural groupings
3. **Unclear Scope**: BUILD and DATA types have unclear boundaries with MECHANIC and ENTITY
4. **Agent Confusion**: Agents struggle to choose correct type, leading to inconsistent categorization

### Where This Pattern Exists Today
- **Current Types**: API, BUILD, DATA, DOMAIN, ENTITY, MECHANIC, MODULE, REQUIREMENT, RULE, SERVICE, SPECIFICATION, UI, USECASE, USER_STORY (13 total)
- **Type Definitions**: `ATD.md §1.3 Document Types` and type validation logic
- **Agent Guidance**: ATD.md provides type guidance but distinctions remain unclear

### Evidence from Investigation
- **USECASE vs USER_STORY**: Both describe user-facing workflows, leading to inconsistent usage
- **SERVICE vs MODULE**: SERVICE rarely used, unclear when to apply vs MODULE
- **SPECIFICATION**: Too broad, overlaps with multiple other types
- **BUILD/DATA**: Overly specific, could be subsumed under other types

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Confirmed via analysis of existing atom usage patterns) |
| Impact if triggered | Medium (Inconsistent categorization, agent decision errors) |
| Detectability | High (Visible in atom type distribution and creation patterns) |
| Current mitigant | Manual type selection guidance in ATD.md and CLAUDE.md |

---

## Recommended Fix

**Short term**: Consolidate redundant types: USECASE + USER_STORY → USER_STORY, SERVICE → MODULE, BUILD → MECHANIC, DATA → ENTITY. Update ATD.md type guidance.

**Medium term**: Deprecate SPECIFICATION type (use REQUIREMENT with appropriate scope). Clear type definitions with use-case examples for each remaining type. Update atom validation to enforce proper type usage.

**Long term**: Implement type inference suggestions when creating atoms. Add type consistency checking across related atoms in dependency graph.

---

## References

- [ATD System Analysis](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/atd_system_analysis.md)
- [ATD.md Document Types](file:///home/bastien/work/skill/ATD.md#13-document-types)
- [Type Implementation](file:///home/bastien/work/skill/scripts/pkg/atom/types.go)