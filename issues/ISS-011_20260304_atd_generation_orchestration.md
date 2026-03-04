# Issue: ATD Generation Orchestration and Local Dissection

**ID:** `ISS-011_20260304_atd_generation_orchestration`
**Ref:** `ISS-011`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `atd_management_skill`
**Affects:** `scripts/atd-ollama-generate`, `scripts/atd-dissect`

---

## Summary

There is a lack of orchestration between the IDE agent and the local ATD generation tools (`atd-dissect` and `atd-ollama-generate`). The agent does not appear to prioritize these tools for ATD generation. Additionally, the dissection process relies on the IDE agent's cloud LLM, which could be replaced by a local solution (e.g., Llama/Nomic) to reduce token costs.

---

## Technical Description

### Background
`atd-dissect` handles document deconstruction, and `atd-ollama-generate` handles Atom creation using local models. These should be the primary paths for ATD work.

### The Problem Scenario
1. When asked to generate ATDs, the IDE agent often creates them manually or via internal logic instead of calling the specialized `atd-ollama-generate` script or the `dissect` workflow.
2. The `SKILL.md` for `atd_management` might be too vague, failing to force the agent to use these specific binaries.
3. Dissection prompts are currently handled by high-cost cloud models instead of leveraging the local Ollama infrastructure already in place.

### Where This Pattern Exists Today
- `atd_management_skill/SKILL.md` instructions.
- `scripts/atd-dissect/` prompt logic.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | High |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** Update `atd_management_skill/SKILL.md` with explicit instructions to use `atd-ollama-generate` for all new Atom creation.
**Medium term:** Investigate running `atd-dissect` prompts against local Ollama models (Llama 3, Nomic-Embed-Text).
**Long term:** Create a unified `atd-gen` wrapper that handles the entire pipeline (Dissect -> Generate -> Audit) autonomously via local models.

---

## References

- [atd_management_skill/SKILL.md](../atd_management_skill/.agent/skills/atd/SKILL.md)
- [atd-ollama-generate](../scripts/atd-ollama-generate/)
