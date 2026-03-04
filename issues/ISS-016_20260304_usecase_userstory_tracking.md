# Issue: Tracking Use Cases and User Stories for Proper Testing

**ID:** `20260304_usecase_userstory_tracking`
**Ref:** `ISS-016`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Resolved
**Component:** `ATD Framework / Documentation`
**Affects:** Testing Strategy / Requirements Traceability

---

## Summary

Currently, there is a lack of structured tracking for high-level use cases and user stories. Without this traceability, it is difficult to guarantee that the implemented logic (Atoms) directly fulfills user requirements or that appropriate, comprehensive tests cover the intended workflows. We need a systematic way to document and link use cases and user stories to technical specifications and tests.

---

## Technical Description

### Background
The ATD framework defines a robust mechanism for atomic specifications (Atoms) like `RULE`, `MECHANIC`, and `DOMAIN`. However, higher-level context such as use cases and user stories (e.g., "As a user, I want to X so that Y") is not explicitly tracked or linked to the underlying atoms and test cases.

### The Problem Scenario
1. A new feature is requested via a user story.
2. The logic is broken down into various Atoms (e.g., `RULE`, `MECHANIC`).
3. During the testing phase, the QA or development suite verifies individual Atoms but lacks a holistic mechanism to ensure the entire user story workflow is tested and satisfied.
4. If a regression occurs, it's difficult to map it back to the specific user impact without clear traceability from the test to the use case.

### Where This Pattern Exists Today
This gap exists across the ecosystem as there is no standardized `USECASE` or `USER_STORY` ATD type, nor a defined macro-level linking strategy from narrative requirements to execution validation.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium (Features may be marked "complete" atomicly but fail holistically; poor test coverage for real-world scenarios) |
| Detectability | Medium (Found late in the cycle during integration or UAT) |
| Current mitigant | Informal understanding or ad-hoc test plans |

---

## Recommended Fix

**Short term:** Define a convention or a new ATD type (e.g., `REQUIREMENT` sub-type or a dedicated `USECASE` / `USER_STORY` type) that explicitly maps a narrative to technical `parents` or `dependents`. Ensure it's bloat tolerance is high, as it will be used to track high-level requirements and may include many steps.
**Medium term:** Update the testing strategy to ensure that tests can be tagged or associated directly with these high-level use case/user story Atoms, giving visibility into feature complete/incomplete status.
**Long term:** Integrate this tracking into a dashboard or report that maps User Stories -> Atoms -> Test Results.

---

## References

- ATD framework documentation
- Existing testing strategies

---

## Change Log

- **2026-03-04**: Resolved. Added `USECASE` and `USER_STORY` as first-class ATD types.
  - `.atd` config: added `type_overrides` entries with bloat factor `0.1` (auto-pass) for both types.
  - `SKILL.md`: updated type enum and category table with descriptions.
  - Created `usecase_template.atom.md` and `userstory_template.atom.md` in `.agent/skills/atd/` as canonical templates, with `WORKFLOW` and `ACCEPTANCE CRITERIA` sections respectively.
