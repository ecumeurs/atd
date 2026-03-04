# Issue: Lack of Project Documentation and ATD

**ID:** `ISS-013_20260304_lack_of_atd_documentation`
**Ref:** `ISS-013`
**Date:** 2026-03-04
**Severity:** High
**Status:** Open
**Component:** `root`
**Affects:** `all`

---

## Summary

This project lacks comprehensive documentation, including the Atomic Technical Documentation (ATD) it is designed to manage. There is no high-level overview or atomic breakdown of the system itself.

---

## Technical Description

### Background
The system is built to manage specifications and rules via Atoms (ATD), yet it does not follow its own dogfooding principle.

### The Problem Scenario
A new developer or agent arriving at the project has no structured documentation to understand the project's own architecture, components (scripts, skills, webui), or rules. This leads to friction and errors when attempting to extend or use the tools.

### Where This Pattern Exists Today
- The root `README.md` is minimal.
- There is no `docs/` folder or ATD base for the project's own logic.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High |
| Detectability | High |
| Current mitigant | Code exploration |

---

## Recommended Fix

**Short term:** Write a high-level `ARCHITECTURE.md` and expand the root `README.md`.
**Medium term:** Create an initial set of ATD Atoms for the core components (Audit, Dissect, Generate, WebUI).
**Long term:** Use the tool's own `dissect` and `generate` workflows to fully document the project's codebase as a comprehensive ATD base.

---

## References

- [ATD protocol](../PROTOCOL.md)
