# Issue: Formalize Test Handling and Traceability in ATD System

**ID:** `20260311_test_handling_shortcoming`
**Ref:** `ISS-022`
**Date:** 2026-03-11
**Severity:** Medium
**Status:** Resolved
**Component:** `atd_management_skill`
**Affects:** `tests`, `tooling`

---

## Summary

Currently, test files and test functions are not handled in a specific way by the ATD system. They lack a dedicated tagging convention (other than the general `@spec-link`), and there is no structured mechanism to ensure tests are correctly linked to their corresponding ATDs or to perform impact analysis based on ATD changes.

---

## Technical Description

### Background
The ATD system uses `@spec-link [[ATOM_ID]]` to link code to specifications. However, tests often cover multiple requirements, specifications, or mechanics, and should be treated as first-class citizens in terms of traceability.

### The Problem Scenario
1. A developer writes a test for a feature.
2. The test is either not linked at all or uses a generic `@spec-link`.
3. If an ATD (requirement) changes, it's difficult to quickly identify which tests are affected because they aren't explicitly tagged as tests linked to that ATD (and its dependent/inherited ATDs).
4. There is no automated way to discover all tests associated with an ATD hierarchy.

### Where This Pattern Exists Today
- All test files in the workspace (e.g., in `scripts/`, `upsilon/`, etc.) lack formalized ATD test-links.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | Medium |
| Current mitigant | Manual review of tests and generic spec links. |

---

## Recommended Fix

**Short term:** 
- Introduce `@test-link [[ATOM_ID]]` tag.
- Update skill instructions to require `@test-link` for all new/updated tests.
- Ensure tests reference all applicable ATDs, including hereditary ones.

**Medium term:** 
- Implement a toolkit to discover all tests linked to an ATD (recursively following dependencies).
- Update the ATD indexer to track test links.

**Long term:** 
- Automated impact analysis that flags specific tests for re-run or review when an ATD changes.

---

## References

- [atd_management_skill/SKILL.md](file:///home/bastien/work/skill/atd_management_skill/SKILL.md)
- [protocol.md](file:///home/bastien/work/skill/protocol.md)
