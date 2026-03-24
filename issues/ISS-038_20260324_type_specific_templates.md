# Issue: Type-Specific Atom Templates

**ID:** `20260324_type_specific_templates`
**Ref:** `ISS-038`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/update.go`
**Affects:** Atom quality, creation workflow

---

## Summary

All atoms use the same generic template regardless of type. `API` atoms would benefit from built-in request/response schema sections; `USECASE` atoms need a `## WORKFLOW` section; `USER_STORY` atoms need `## ACCEPTANCE CRITERIA`. Templates for `USECASE` and `USER_STORY` already exist in the skill but are not integrated into the CLI/MCP tooling.

---

## Technical Description

### Background
The generic template has four sections: INTENT, LOGIC, INTERFACE, EXPECTATION. Some types have additional required sections (e.g., USECASE needs WORKFLOW).

### The Problem Scenario
1. An agent creates a `USECASE` atom via `atd update`.
2. The atom gets the generic template, missing the `## WORKFLOW` section.
3. Manual editing is required to add the type-specific sections.

### Where This Pattern Exists Today
- `atd_management_skill/.agent/skills/atd/`: Has `usecase_template.atom.md` and `userstory_template.atom.md` but these are not used by the CLI.
- `scripts/cmd/atd/cmd/update.go` — uses a single template for all types.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Low — missing sections can be added manually |
| Detectability | Medium — only noticeable if you know the correct template |
| Current mitigant | Skill templates for agent reference |

---

## Recommended Fix

**Short term:** Integrate type-specific templates into `atd update` — when creating a new atom, select the template based on `type`.
**Medium term:** Allow custom templates per project via `.atd` config (e.g., `templates.USECASE: path/to/template.md`).
**Long term:** `atd lint` (ISS-032) should validate type-specific section requirements.

---

## References

- [ATD.md §3.2.7](file:///home/bastien/work/skill/ATD.md)
- [usecase_template.atom.md](file:///home/bastien/work/skill/atd_management_skill/.agent/skills/atd/usecase_template.atom.md)
