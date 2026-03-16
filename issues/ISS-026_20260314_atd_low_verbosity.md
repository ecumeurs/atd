# Issue: Low ATD Content Verbosity

**ID:** `20260314_atd_low_verbosity`
**Ref:** `ISS-026`
**Date:** 2026-03-14
**Severity:** Low
**Status:** Open
**Component:** `scripts/cmd/atd/update`
**Affects:** `atd` generation pipeline

---

## Summary

Atoms generated during the initial creation phase (Task 03) are often sparse, containing only single-sentence intents and minimal logic. While the tools allow for expansion, the default generation/update process encourages low-quality "oneliner" documentation.

---

## Technical Description

### Background
ATD atoms should be rich enough to serve as a system blueprint. The `SKILL.md` defines several sections (LOGIC, INTERFACE, EXPECTATION) that should be comprehensively populated.

### The Problem Scenario
When creating atoms via `atd update` or `atd generate`, the resulting files often only have the strictly required fields populated, or use very brief summaries that don't capture the full complexity of the underlying logic.

### Where This Pattern Exists Today
- Atoms in `upsilonbattle/docs/` created during Task 03.
- `update.go` logic which facilitates sparse updates.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Low — leads to shallow documentation |
| Detectability | High |
| Current mitigant | Manual review and expansion |

---

## Recommended Fix

**Short term:** Update documentation guidelines to emphasize verbosity.
**Medium term:** Improve the prompt in `atd dissect` and `atd generate` to request more detailed logical breakdowns.
**Long term:** Implement an automated "richness" auditor that flags atoms with insufficient content length or complexity.

---

## References

- [SKILL.md](file:///home/bastien/work/skill/atd_management_skill/.agent/skills/atd/SKILL.md)
- [Tests/03_atd_creation.md](file:///home/bastien/work/skill/Tests/03_atd_creation.md)
