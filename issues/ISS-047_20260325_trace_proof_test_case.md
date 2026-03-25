# Issue: Proof Test Trace - Comprehensive Test Case

**ID:** `20260325_trace_proof_test_case`
**Ref:** `ISS-047`
**Date:** 2026-03-25
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/cmd/atd/cmd/trace.go`
**Affects:** `atd trace` validation

---

## Summary

The `atd trace` command lacks a comprehensive proof test case that validates all documented health check warnings and structural rules. To ensure regression stability and correct health scoring, we need a test environment with a complex graph (24+ atoms) including specific edge cases like orphans, missing Customer ancestry, and layer jumping.

---

## Technical Description

### Background

The `atd trace` tool is responsible for traversing the atom dependency graph (upwards to CUSTOMER and downwards to IMPLEMENTATION) and calculating health metrics based on development status, linked code (`@spec-link`), and linked tests (`@test-link`).

### The Problem Scenario

Current tests for `trace` might be too simple to catch subtle logic errors in:
1. Multi-level dependency traversal.
2. Ancestry checks (e.g., verifying if a CUSTOMER atom exists at the root).
3. Layer-specific warnings (CUSTOMER -> ARCHITECTURE -> IMPLEMENTATION).

The following edge cases must be explicitly tested:
- **Orphan:** An atom with no parents and no dependents.
- **No Customer Ancestry:** A graph that stops at the ARCHITECTURE layer without reaching CUSTOMER.
- **Missing Architecture Layer:** A direct link from CUSTOMER to IMPLEMENTATION.
- **Missing implementation:** An IMPLEMENTATION atom with no `@spec-link`.
- **Missing tests:** An IMPLEMENTATION atom with implementation but no `@test-link`.

### Where This Pattern Exists Today

The implementation in `scripts/cmd/atd/cmd/trace.go` makes several assumptions about the graph structure that need rigorous verification against a non-trivial dataset.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium |
| Detectability | High — manifests as incorrect health reports or missing warnings |
| Current mitigant | Basic unit tests |

---

## Recommended Fix

**Short term:** Create a dedicated test environment in `tests/trace/` with 24+ `.atom.md` files and mock source files.

**Medium term:** Integrate this test case into the CI pipeline to run automated assertions against the `atd trace` JSON output.

---

## References

- [trace.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/trace.go)
- [task.md](file:///home/bastien/.gemini/antigravity/brain/13e2e7dc-e125-4baa-9b76-86069d19a956/task.md)
