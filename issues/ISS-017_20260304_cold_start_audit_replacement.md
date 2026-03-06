# Issue: Replace Cold Start Mass Generative Step with Audit Loop

**ID:** `20260304_cold_start_audit_replacement`
**Ref:** `ISS-017`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `scripts/atd-cold-start.sh`
**Affects:** `scripts/atd-ollama-generate`, `scripts/atd-full-audit.sh`

---

## Summary

The final step of the cold start pipeline (`atd-cold-start.sh`) instructs the user to dump all generated parsed context into a single Cloud LLM prompt. This limits scalability, is error-prone, and relies heavily on token budgets. The proposal is to replace this step with an automated iterative audit/fix loop, potentially using `atd-ollama-generate` to draft atoms, followed by `atd-audit-fixer` to hone them based on the rules.

---

## Technical Description

### Background
Currently, Iteration 15 of `protocol.md` outlines the cold start generating a series of `domain_*.md.txt` and `dissect_*.json` metadata files. The script pauses there and expects the Cloud Agent to read everything and construct `.atom.md` files. 

### The Problem Scenario
A single manual prompt dumping 20+ files can lead to context collapse. Furthermore, the role and stability of `atd-ollama-generate` as part of this solution currently remains unclear and untested at scale. Using heavy Cloud LLMs sequentially instead of locally bootstrapping Shadow Atoms creates performance issues.

### Where This Pattern Exists Today
`scripts/atd-cold-start.sh` Phase 5 and the manually executed Phase 6.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium |
| Detectability | High — pipeline stops and instructs manual LLM work. |
| Current mitigant | None, step is entirely manual and unconstrained. |

---

## Recommended Fix

**Short term:** Document the capabilities of `atd-ollama-generate`. Verify if it can reliably be used to create the initial draft "Shadow Atoms" in bulk.
**Medium term:** Prototype a bash script that loops `atd-ollama-generate` over the `pipeline_output/dissect_*.json` outputs. Pass the resulting `docs/*.atom.md` proxies through `atd-audit` and `atd-audit-fixer` iteratively to constrain the generation process automatically.
**Long term:** Remove the manual prompt completely. Automate Phase 5 -> Draft Atoms -> Fixer Loop -> `atd-recon` tagging entirely inside `atd-cold-start.sh`.

---

## References

- `scripts/atd-cold-start.sh`
- `scripts/atd-ollama-generate/main.go`
- `protocol.md` (Iteration 11, 15)
