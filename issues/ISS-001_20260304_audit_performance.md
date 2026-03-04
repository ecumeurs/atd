# Issue: Audit Performance Optimization

**ID:** `ISS-001_20260304_audit_performance`
**Ref:** `ISS-001`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `scripts/atd-audit`
**Affects:** `scripts/atd-full-audit.sh`

---

## Summary

The current auditing process is too slow. It requires access to a more performant LLM (better Ollama model) or general performance improvements to be viable for large-scale ATD bases.

---

## Technical Description

### Background
The `atd-audit` tool uses local LLMs via Ollama to perform validation and bloat checks on ATDs.

### The Problem Scenario
When running audits on a directory with many ATDs, the processing time per ATD is high, leading to extremely long wait times for a full project audit. This discourages frequent use of the audit tool.

### Where This Pattern Exists Today
- `scripts/atd-audit/`
- `scripts/atd-full-audit.sh`

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | High |
| Current mitigant | Caching for bloat checks |

---

## Recommended Fix

**Short term:** Test with more performant models (e.g., Llama 3 8B or Phi-3) and optimize prompt size.
**Medium term:** Implement parallel processing for ATD auditing.
**Long term:** Offload heavy auditing tasks to a dedicated high-performance LLM endpoint.

---

## References

- [atd-full-audit.sh](../scripts/atd-full-audit.sh)
