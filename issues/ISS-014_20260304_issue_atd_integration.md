# Issue: Integrate Issue Management with ATD Management

**ID:** `ISS-014_20260304_issue_atd_integration`
**Ref:** `ISS-014`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `.agent/skills/atd`
**Affects:** `.agent/skills/issue_management`, `webui`

---

## Summary

Integrate the `issue_management` skill as a side-skill for `atd_management`. This allows issues to be bidirectionally linked with ATDs, providing better traceability between technical debt/bugs and the specifications they impact. This integration should be reflected in the WebUI.

---

## Technical Description

### Background

Currently, `issue_management` and `atd_management` operate independently. Issues are tracked in `/workspace/issues/` and ATDs in `.atd/`. There is no formal way to link an issue to the specific Atom it relates to, other than manual mentions in text.

### The Problem Scenario

When an issue is discovered, it often relates to a specific business rule or technical mechanic defined in an ATD. Conversely, when viewing an ATD, it's useful to know if there are any open issues or technical debt associated with it.

1.  User identifies a bug in a logic block.
2.  User creates an issue.
3.  The issue has no machine-readable link to the ATD defining that logic.
4.  Developer viewing the ATD is unaware of the active bug/risk.

### Where This Pattern Exists Today

- `.agent/skills/issue_management/templates/issue.md` lacks an explicit ATD reference field.
- `ATD.md` template (and existing atoms) lacks a structured `Issues` list in metadata or interface.
- WebUI rendering logic does not cross-reference these two domains.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — leads to fragmented documentation and missed context. |
| Detectability | Medium — requires manual searching to find related issues/ATDs. |
| Current mitigant | Manual mentions in `## References` or `Related Issue` in ATDs. |

---

## Recommended Fix

**Short term:** Add a `Relates to ATDs:` field in the issue template and a `Related Issues:` field in the ATD template.

**Medium term:** Update `list_issues.py` and `atd` scripts to support cross-referencing and validation of这些 links.

**Long term:** Enhance the WebUI to automatically show linked issues on ATD pages and vice versa, with deep links between them.

---

## References

- [ATD.md](file:///home/bastien/work/skill/ATD.md)
- [.agent/skills/issue_management/SKILL.md](file:///home/bastien/work/skill/.agent/skills/issue_management/SKILL.md)
- [.agent/skills/atd/SKILL.md](file:///home/bastien/work/skill/atd_management_skill/.agent/skills/atd/SKILL.md)
