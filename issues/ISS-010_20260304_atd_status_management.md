# Issue: ATD Status Management and Workflow

**ID:** `ISS-010_20260304_atd_status_management`
**Ref:** `ISS-010`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `atd_management_skill`
**Affects:** `scripts/atd-audit`, `webui`

---

## Summary

The `status` attribute is currently ignored. Implementing status-based logic is required: `DRAFT` documents should allow swifter merging, while `REVIEW` and `STABLE` should have stricter requirements. `DISCARDED` or `OUTDATED` statuses should cause the ATD to be ignored entirely.

---

## Technical Description

### Background
The ATD template includes a `status` field (DRAFT, REVIEW, STABLE).

### The Problem Scenario
All ATDs are treated equally regardless of their maturity level. This prevents implementing a gated workflow where stable specifications are protected and draft ones are easily iterable.

### Where This Pattern Exists Today
- `scripts/atd-audit/` validation rules.
- `atd_management_skill` merge logic.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | Medium |
| Current mitigant | Manual review |

---

## Recommended Fix

**Short term:** Implement status-based filtering in `atd-audit` (e.g., skip complex checks for DRAFT).
**Medium term:** Enforce strict validation in `atd-audit` for ATDs marked as STABLE.
**Long term:** Implement a state machine for ATD status transitions within the WebUI.

---

## References

- [ATD protocol](../PROTOCOL.md)
