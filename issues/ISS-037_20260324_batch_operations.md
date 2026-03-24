# Issue: Batch Operations for atd update

**ID:** `20260324_batch_operations`
**Ref:** `ISS-037`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/cmd/atd/cmd/update.go`
**Affects:** MCP efficiency, bulk workflows

---

## Summary

When promoting multiple atoms from `DRAFT` to `REVIEW`, or bulk-tagging a directory, each atom requires a separate `atd update` call. This is especially painful over MCP, where the IDE agent must make many sequential tool calls.

---

## Technical Description

### Background
`atd update` operates on a single atom file at a time. There is no way to filter and update multiple atoms in one invocation.

### The Problem Scenario
1. An architect reviews 15 `DRAFT` atoms and wants to promote them all to `REVIEW`.
2. This requires 15 individual `atd update --file X --set status=REVIEW` calls.
3. Over MCP, this means 15 tool invocations and round-trips.

### Where This Pattern Exists Today
- `scripts/cmd/atd/cmd/update.go` — single-file operation only.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — productivity loss, token waste |
| Detectability | High — immediately apparent during bulk workflows |
| Current mitigant | Shell loops or manual repetition |

---

## Recommended Fix

**Short term:** Add `--filter` flag to `atd update` (e.g., `--filter "type=RULE,status=DRAFT" --set status=REVIEW`) that applies the update to all matching atoms.
**Medium term:** Expose batch mode through MCP with a filter parameter.
**Long term:** Support transactional batch updates with rollback on failure.

---

## References

- [ATD.md §3.2.5](file:///home/bastien/work/skill/ATD.md)

## Change Log
- **2026-03-24**: Added `--filter` flag to `updateCmd` and `filter` property to `atd_update` MCP tool, calling a unified `runBatchUpdate` method that applies the updates to multiple matching `.atom.md` files locally.
