# Issue: Link WebUI to Project Binaries

**ID:** `ISS-007_20260304_webui_binary_link`
**Ref:** `ISS-007`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `webui`
**Affects:** `webui/backend`, `scripts/`

---

## Summary

Integrate the WebUI with the project's heavy-duty binaries and scripts (e.g., `atd-audit`, cold start scripts). This will allow users to perform audits and other tasks directly from the UI, providing a nicer interface for review and validation.

---

## Technical Description

### Background
Multiple CLI tools exist for ATD lifecycle management (auditing, generation, mapping).

### The Problem Scenario
Users must switch between the WebUI and the terminal to run audits or update the ATD base, creating friction in the workflow.

### Where This Pattern Exists Today
- `scripts/atd-audit/`
- `scripts/atd-full-audit.sh`
- `webui/`

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium |
| Detectability | High |
| Current mitigant | CLI usage |

---

## Recommended Fix

**Short term:** Add button triggers in the WebUI that invoke shell scripts via the backend.
**Medium term:** Stream terminal output from these scripts directly into a WebUI console component.
**Long term:** Implement a job queue for backgrounding long-running tasks like full audits.

---

## References

- [webui directory](../webui/)
- [scripts directory](../scripts/)
