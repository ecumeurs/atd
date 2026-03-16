# Issue: Insufficient Dissection Granularity for Complex Files

**ID:** `20260314_atd_dissect_quality`
**Ref:** `ISS-027`
**Date:** 2026-03-14
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/dissect`
**Affects:** `atd` cold-start and manual dissection pipelines

---

## Summary

The `atd dissect` tool fails to identify a sufficient number of atomic boundaries in complex source files. For example, `ruler.go` (a ~400 line file with state management, notification handling, and rule implementation) was dissected into only one or two atoms, when it should have yielded at least five distinct logical fragments (Service, Call Methods, Notifications, Entity State, Mechanics).

---

## Technical Description

### Background
`atd dissect` uses an LLM to identify "logical boundaries" in a file. It is expected to adhere to the "One Rule" Rule defined in `SKILL.md`, where each atom represents a single state-changing rule or focused objective.

### The Problem Scenario
When run on `battlearena/ruler/ruler.go`, the tool produced a very coarse dissection. This results in "bloated" atoms that combine too many responsibilities, violating the core principle of atomic documentation.

### Where This Pattern Exists Today
- `scripts/cmd/atd/cmd/dissect.go`
- Observed during Task 03 dissection of `ruler.go`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium — results in bloated, non-atomic documentation |
| Detectability | High — evident in the number of atoms proposed |
| Current mitigant | Manual splitting by the IDE Agent |

---

## Recommended Fix

**Short term:** Update the dissection prompt to explicitly request higher granularity and list examples of expected atom types (Service, Method, Notification, etc.).
**Medium term:** Implement a multi-pass dissection that first identifies major modules and then recurses into each module to find smaller rules.
**Long term:** Use AST-based boundary detection to provide the LLM with structural hints (functions, types, switch cases) to guide the dissection.

---

## References

- [SKILL.md](file:///home/bastien/work/skill/atd_management_skill/.agent/skills/atd/SKILL.md) (Granularity Control)
- [scripts/cmd/atd/cmd/dissect.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/dissect.go)
- [battlearena/ruler/ruler.go](file:///home/bastien/work/skill/upsilonbattle/battlearena/ruler/ruler.go)
