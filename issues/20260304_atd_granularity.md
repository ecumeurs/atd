# Issue: ATD Dissection Granularity Enforcement

**ID:** `20260304_atd_granularity`
**Ref:** `ISS-002`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `scripts/atd-ollama-generate`
**Affects:** `atd_management_skill`

---

## Summary

Ensure that the dissection of documents and general ATD creation strictly follows the granularity guidelines defined in the "Minimum Atomic Scale" of the ATD protocol.

---

## Technical Description

### Background
The ATD protocol specifies a "Minimum Atomic Scale" to prevent overly broad definitions (e.g., the "One Rule" Rule).

### The Problem Scenario
Current dissection tools and manual creation often group multiple independent rules or intents into a single Atom, violating the principle of atomicity and making dependencies harder to track.

### Where This Pattern Exists Today
- Automated dissection logic in `scripts/atd-ollama-generate`.
- Manual ATD creation via `atd_management_skill`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | High |
| Detectability | Medium |
| Current mitigant | Manual review |

---

## Recommended Fix

**Short term:** Add validation logic to `atd-audit` to detect "and" or "also" in intent strings.
**Medium term:** Refactor dissection prompts to explicitly mention and enforce the "One Rule" Rule.
**Long term:** Implement automated splitting suggestions during the ATD creation workflow.

---

## References

- [ATD protocol](../PROTOCOL.md)
