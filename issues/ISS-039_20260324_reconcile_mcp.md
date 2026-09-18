# Issue: Reconcile Tool via MCP

**ID:** `20260324_reconcile_mcp`
**Ref:** `ISS-039`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/serve.go`
**Affects:** Onboarding workflow, conflict resolution

---

## Summary

When new external requirements arrive that semantically overlap with existing atoms, there is no structured merge/conflict resolution protocol via MCP. The `atd reconcile` CLI concept exists but is not exposed as an MCP tool.

---

## Technical Description

### Background
The Reconciler sub-mode is described in the skill documentation as a way to match inbound spec requirements against the populated library and classify overlap as MERGE, UPDATE, or NEW.

### The Problem Scenario
1. A customer sends updated requirements.
2. The IDE agent needs to determine which existing atoms overlap, which need updating, and which are genuinely new.
3. Without an MCP reconcile tool, this requires manual semantic comparison using `atd search` + human judgment.

### Where This Pattern Exists Today
- `scripts/cmd/atd/cmd/reconcile.go` — CLI exists but not MCP-registered.
- `atd_management_skill/.agent/rules/ATD.md` — describes "The Reconciler" sub-mode.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium — redundant atoms, missed updates |
| Detectability | Medium — semantic overlap is subtle |
| Current mitigant | Manual `atd search` + human comparison |

---

## Recommended Fix

**Short term:** Register `atd_reconcile` as an MCP tool wrapping the existing CLI command.
**Medium term:** Enhance to return structured JSON with MERGE/UPDATE/NEW classifications and confidence scores.
**Long term:** Integrate with changelog sidecar (ISS-034) to record reconciliation decisions.

---

## References

- [ATD.md §3.2.8](file:///home/bastien/work/atd/ATD.md)
- [atd_reconcile.atom.md](file:///home/bastien/work/atd/docs/atd_reconcile.atom.md)
