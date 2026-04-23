# Issue: E2E Test @test-link Overcrowding

**ID:** `20260422_e2e_test_link_overcrowding`
**Ref:** `ISS-090`
**Date:** 2026-04-22
**Severity:** Medium
**Status:** Open
**Component:** `atd/pkg/exploration`
**Affects:** `E2E test files and ATD health metrics`

---

## Summary

The current ATD model requires individual `@test-link` tags for each atom implementation. In end-to-end (E2E) tests which cover multiple features and atoms, this leads to significant tag overcrowding, maintenance burden, and reduced readability of test files.

---

## Technical Description

### Background
Currently, the ATD verification engine (`atd_verify`, `atd_trace`) expects a direct 1:1 or 1:N mapping via `@test-link [[atom_id]]` tags in test files to determine test coverage for an atom.

### The Problem Scenario
An E2E test file often validates multiple customer requirements and their corresponding architectural/implementation layers simultaneously. Under the current model, the developer must manually list every single atom ID involved:
- `uc_player_login`
- `api_auth_login`
- `api_auth_logout`
- `mechanic_mech_cli_sensitive_data_masking`

This becomes unmanageable as the project grows, leading to developers omitting tags (false negatives in health stats) or files becoming cluttered with metadata.

### Where This Pattern Exists Today
This is a systemic design constraint in the ATD core logic, affecting any large-scale test suite.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | High — manifests as excessive tagging or missing coverage stats |
| Current mitigant | None, developers must manually tag everything |

---

## Recommended Fix

**Short term:** Update documentation to acknowledge the limitation and suggest grouping atoms under broader architectural parents.
**Medium term:** Implement a **Test Propagation Model** in `atd/pkg/exploration`. If a Customer layer atom is linked to a test, allow the "tested" status to propagate downward to its descendant ARCHITECTURE and IMPLEMENTATION atoms.
**Long term:** Introduce a `TEST_SUITE` atom type that can explicitly define a scope of atoms it covers, decoupling the mapping from source code tags where appropriate.

---

## References

- [atd/pkg/exploration/exploration.go](file:///home/bastien/work/skill/atd/pkg/exploration/exploration.go)
- [atd/pkg/exploration/weave.go](file:///home/bastien/work/skill/atd/pkg/exploration/weave.go)
