# Issue: Cross-Project Atom Sharing with Destination Selection

**ID:** `20260324_cross_project_sharing`
**Ref:** `ISS-035`
**Date:** 2026-03-24
**Severity:** Low
**Status:** Open
**Component:** `.atd` config, `scripts/cmd/atd`
**Affects:** Multi-repository projects

---

## Summary

ATD is currently single-project. There is no mechanism for sharing atoms between repositories (e.g., a shared `DOMAIN` atom for company-wide business rules). Multi-repository projects would benefit from cross-project references and destination selection.

---

## Technical Description

### Background
The `.atd` config file specifies a single `docs_path`. All tools operate within this single project boundary. Many real-world projects span multiple repositories with shared business rules.

### The Problem Scenario
1. A company has a shared auth service used by three projects.
2. The auth domain rules are documented as atoms in one repo.
3. Other repos need to link to these atoms via `parents` but cannot resolve `[[auth_rule_xyz]]` across repository boundaries.
4. When creating new atoms, there is no way to specify which repository the atom belongs to.

### Where This Pattern Exists Today
- `.atd` config schema — single `docs_path`.
- All atom resolution logic — local filesystem only.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Low — workaround is manual documentation |
| Detectability | High — cross-repo references fail to resolve |
| Current mitigant | Copy atoms manually or use conventions |

---

## Recommended Fix

**Short term:** Document conventions for cross-repo atom referencing (e.g., `[[repo:atom_id]]` format).
**Medium term:** Add `refs` or `imports` section in `.atd` config pointing to external atom repositories (git URLs or local paths). `atd crawl` and `atd search` should resolve cross-project links. Upgrade tooling to handle destination selection when creating or linking atoms.
**Long term:** Implement a federated atom registry for organization-wide governance.

---

## References

- [ATD.md §3.2.3](file:///home/bastien/work/skill/ATD.md)
