# Issue: Missing @spec-link Tags for Implemented Features

**ID:** `20260418_missing_spec_link_tags_documentation_gap`
**Ref:** `ISS-074`
**Date:** 2026-04-18
**Severity:** Medium
**Status:** Open
**Component:** `docs/`, codebase-wide
**Affects:** Documentation Coverage Accuracy, Atom-to-Code Traceability

---

## Summary

Approximately 40 STABLE atoms describe implemented functionality that exists in code but lacks proper @spec-link tags. This is a documentation gap, not a missing feature gap, requiring systematic tag addition to existing code.

---

## Technical Description

### Background
ATD traceability requires @spec-link [[atom_id]] tags in code to establish connections between documentation and implementation. Some features are fully implemented but missing these linking tags.

### The Problem Scenario
1. **Implemented But Untagged**: Core functionality exists in code but corresponding atoms have no @spec-link tags
2. **Discovery Gap**: Makes it difficult to trace which code implements which requirements
3. **Coverage Inaccuracy**: Contributes to false orphan reporting despite features being implemented

### Where This Pattern Exists Today
- **Mechanics Likely Implemented**: Action economy, board generation, character reroll, combat shielding, entity properties, initiative systems, move validation (9 atoms), skill validation (7 atoms)
- **UI/UX Likely Implemented**: Leaderboard components (4 atoms), dashboard components (5 atoms), registration flows (3 atoms), various UI elements
- **Requirements Likely Satisfied**: Admin experience, logging requirements, player experience, security requirements (4 atoms)

### Evidence from Investigation
- **Status**: All are STABLE atoms, indicating implementation completion
- **Code Search**: grep finds relevant functions/classes but no @spec-link tags
- **Impact**: These are the 40 atoms categorized as "Implemented But Not Tagged" in orphan analysis

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Systematic analysis identified ~40 cases) |
| Impact if triggered | Medium (Reduced traceability, harder refactoring impact analysis) |
| Detectability | Medium (Requires manual code-to-atom mapping) |
| Current mitigant | Manual grep searches when investigating specific features |

---

## Recommended Fix

**Short term**: Conduct systematic audit of STABLE atoms to identify implemented code without @spec-link tags. Add tags prioritizing high-impact features: authentication, combat mechanics, matchmaking.

**Medium term**: Implement automated link suggestion tool that analyzes code signatures against atom specifications to recommend @spec-link placements. Create missing-link reports.

**Long term**: Integrate ATD with CI/CD to detect implemented functions/classes without corresponding @spec-link tags and generate warnings.

---

## References

- [Orphan Categorization](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/orphan_categorization.md)
- [Final Summary](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/final_summary.md)
- [ATD Linking Guide](file:///home/bastien/work/skill/ATD.md#surgical-attachment-rules-for-spec-link)