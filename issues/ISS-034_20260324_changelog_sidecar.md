# Issue: Change History Sidecar per Atom

**ID:** `20260324_changelog_sidecar`
**Ref:** `ISS-034`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/update.go`
**Affects:** Governance, audit trail, `STABLE` atom traceability

---

## Summary

Atoms have a `version` field but no change log. When an atom is modified, there is no record of what changed, when, or why. This gap is critical for `STABLE` atoms in the `CUSTOMER` domain where spec change traceability is a governance requirement.

---

## Technical Description

### Background
The `atd update` command modifies atom files in-place via surgical edits. It currently does not record any history of those edits.

### The Problem Scenario
1. An architect modifies a `STABLE` `REQUIREMENT` atom's logic.
2. A month later, a developer questions why the implementation diverges from the original spec.
3. There is no record of the change — only the current state of the atom.

### Where This Pattern Exists Today
- `scripts/cmd/atd/cmd/update.go` — writes changes but does not log them.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — loss of traceability for spec changes |
| Detectability | Low — absence of data is invisible |
| Current mitigant | Git history (coarse-grained, not atom-specific) |

---

## Recommended Fix

**Short term:** Implement a sidecar `<atom_id>.changelog.json` file co-located next to the `.atom.md`. `atd update` should auto-append an entry with: timestamp, field changed, old → new value, user/agent identifier.
**Medium term:** `atd query` and MCP tools should be able to retrieve changelog data. WebUI should display change history per atom.
**Long term:** Integrate with version incrementing (ISS-009) so each changelog entry maps to a version bump.

---

## References

- [ATD.md §3.2.2](file:///home/bastien/work/skill/ATD.md)
- [ISS-009](ISS-009_20260304_atd_version_management.md)
