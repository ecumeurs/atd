# Issue: Unified ATD Configuration File

**ID:** `ISS-003_20260304_audit_config_move`
**Ref:** `ISS-003`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `root`
**Affects:** `scripts/atd-audit`, `scripts/atd-ollama-generate`

---

## Summary

Move audit configuration to a general `.atd` configuration file at the root of the project. This file should provide comprehensive configuration info, such as paths to binaries and documentation, and ensure that audit thresholds are generic and accessible to LLM agents.

---

## Technical Description

### Background
Currently, configuration (like thresholds) is often passed via flags or hardcoded in scripts.

### The Problem Scenario
Lack of a central configuration makes it difficult to maintain consistency across different tools (auditor, generator, webui). LLM agents also lack a single source of truth for project-specific ATD constraints.

### Where This Pattern Exists Today
- `scripts/atd-audit/` thresholds.
- Binary paths in various shell scripts.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium |
| Detectability | Low |
| Current mitigant | Environment variables or flags |

---

## Recommended Fix

**Short term:** Define a JSON or YAML schema for `.atd` config.
**Medium term:** Refactor `atd-audit` and `atd-ollama-generate` to read from this file.
**Long term:** Automatically inject this configuration into LLM prompts during ATD operations.

---

## References

- [atd-audit logic](../scripts/atd-audit/)
