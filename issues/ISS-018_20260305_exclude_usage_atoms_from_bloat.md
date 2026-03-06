# Issue: Exclude User Stories and Use Cases from Bloat Checks

**ID:** `20260305_exclude_usage_atoms_from_bloat`
**Ref:** `ISS-018`
**Date:** 2026-03-05
**Severity:** Medium
**Status:** Open
**Component:** `scripts/atd-audit`, `scripts/atd-audit-fixer`
**Affects:** All ATD Atoms of type `USAGE`, `UI` (User Stories, Use Cases)

---

## Summary

Atoms that represent User Stories and Use Cases (typically typed as `USAGE` or `UI`) are inherently comprehensive. By their very nature, they require extensive bullet points, steps, and robust narrative descriptions to capture full user flows. However, the current ATD structural linter (`atd-audit`) and its auto-fixer (`atd-audit-fixer`) aggressively check for "bloating" and the "Minimum Atomic Scale," mistakenly flagging these necessary narrative structures as bloated and attempting to split them.

---

## Technical Description

### Background
The "Architectural Linter" (Iteration 17 via `atd-audit`) evaluates ATD contents using a local LLaMA syntactic validator. It explicitly checks if `## INTENT` or `## THE RULE / LOGIC` contains compound rules, flagging atoms as `[BLOATED]` if they violate the Minimum Atomic Scale. `atd-audit-fixer` then tries to refactor these.

### The Problem Scenario
When an Architect defines a User Story (e.g., `login-user-flow` atom typed as `USAGE`), the logic block will correctly contain a 10-step sequence. The `atd-audit` tool will parse this list, incorrectly identify it as compound/bloated mechanics, and flag it. The fixer might then erroneously attempt to split a single cohesive User Story into 10 disjointed atoms.

### Where This Pattern Exists Today
- `scripts/atd-audit` (The syntactic evaluation phase)
- `scripts/atd-audit-fixer` (The resolution phase)
- `scripts/atd-full-audit.sh` (The orchestration pipeline)

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (guaranteed when evaluating USAGE/UI atoms) |
| Impact if triggered | High (destroys cohesive user stories by aggressively splitting them) |
| Detectability | High (flagged visibly in the audit report) |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** Update the bash and Go files (`atd-audit` and `atd-audit-fixer`) to inspect the YAML frontmatter `type:` before executing the bloat check constraint. 
**Medium term:** If `type` equals `USAGE`, `UI`, or another narrative-heavy type, immediately bypass the Syntactic LLaMA Bloat Check and mark the bloat status as `PASS`. Ensure they are still subject to the Semantic Collision check so they are integrated into the graph.

---

## References

- `scripts/atd-audit`
- `scripts/atd-audit-fixer`
- ATD Structure definitions (`SKILL.md` / `template.atom.md`)
