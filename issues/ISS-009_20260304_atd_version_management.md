# Issue: ATD Version Management Implementation

**ID:** `ISS-009_20260304_atd_version_management`
**Ref:** `ISS-009`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `atd_management_skill`
**Affects:** `scripts/atd-audit`, `webui`

---

## Summary

The `version` attribute in ATD YAML frontmatter is currently ignored. The system should handle versioning, ensuring that if multiple ATDs respond to the same ID, operations work on the highest version.

---

## Technical Description

### Background
The ATD template includes a `version` field (default 1.0).

### The Problem Scenario
When multiple files exist for the same logical Atom ID (perhaps due to branching or manual duplication), the system does not consistently select the latest version, leading to potential stale data usage.

### Where This Pattern Exists Today
- Parse logic in `scripts/atd-audit/`.
- File discovery in `webui/`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium |
| Detectability | Medium |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** Update ATD parsers to extract and sort by version.
**Medium term:** Implement a "Latest Version" resolver in the core ATD management library.
**Long term:** Automate version incrementing during edits.

---

## References

- [ATD protocol](../PROTOCOL.md)
