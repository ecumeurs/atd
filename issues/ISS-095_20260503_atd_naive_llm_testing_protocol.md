# Issue: Naive LLM Semantic Testing Protocol

**ID:** `20260503_atd_naive_llm_testing_protocol`
**Ref:** `ISS-095`
**Date:** 2026-05-03
**Severity:** High
**Status:** In Progress (Protocol Deployed)
**Component:** `atd/pkg/prompt`, `atd/pkg/pipeline`
**Affects:** `atd_check`, documentation quality, architectural governance

---

## Summary

The current `atd_verify --semantic` flag uses a generic "ATD Auditor" prompt to validate code against documentation. This approach is too lenient and lacks layer-awareness. It often allows technical jargon to leak into Business atoms and fails to verify specific architectural guarantees.

This issue proposes implementing the "Naive LLM" testing protocol using two specialized personas: the **Synthetic PM** (for Business layer) and the **Synthetic Tech Lead** (for Architecture layer).

---

## Technical Description

### Background
The "Naive LLM" protocol relies on models that are instructed **not to infer** missing context. If the information isn't in the atom, the test fails. This forces documentation to be explicit and high-quality.

### The Problem Scenario
A `USER_STORY` (Business) atom that describes a database table instead of a user value is technically "correct" to an auditor but fails the purpose of the Business layer. Conversely, an `API` (Architecture) atom that doesn't state its data ownership is a "black box" that creates integration risk.

### Implementation Requirements

#### 1. The BUSINESS Layer Test (The Synthetic PM)
- **Target Types**: `PERSONA`, `WORKFLOW`, `REQUIREMENT`, `USER_STORY`.
- **Persona**: A non-technical Product Manager.
- **Rules**: Must identify "Who", "Value", and "One Strict Rule".
- **Rejection Criteria**: If the model replies "MISSING" or if it finds technical jargon (API ports, DB names, Go structs).

#### 2. The ARCHITECTURE Layer Test (The Synthetic Tech Lead)
- **Target Types**: `CONTRACT`, `MODULE`, `API`, `ENTITY`.
- **Persona**: An Integration Engineer.
- **Rules**: Must identify "External Systems", "Data Ownership", and "Primary Guarantee".
- **Rejection Criteria**: If the model replies "MISSING" or if the "Guarantee/Contract" is absent.

#### 3. The IMPLEMENTATION Layer Test (Structural)
- **Target Types**: `MECHANIC`.
- **Rule**: No LLM required; validation is purely structural (Upward parent link + Downward `@spec-link` in code).

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High — ensures documentation remains a "Source of Truth" rather than a code-mirror |
| Detectability | High — `atd_verify` will show PASS/FAIL per layer |
| Current mitigant | Manual peer review of ATD files |

---

## Recommended Fix

**Short term:** Add the "Synthetic PM" and "Synthetic Tech Lead" prompts to `atd/pkg/prompt`.  
**Medium term:** Update `runCoverageCheck` in `check_coverage.go` to select the correct prompt based on the atom's layer.  
**Long term:** Integrate with local Naive models (Ollama 7B/8B) to ensure zero-cost, high-rigor validation.

## Current Progress (2026-05-03)

### Completed
- [x] **Synthetic Persona Implementation**: Added `Synthetic PM` and `Synthetic Tech Lead` prompt builders in `atd/pkg/prompt`.
- [x] **Pipeline Integration**: Refactored `atd check --semantic` to select personas based on atom layer.
- [x] **Data Curation**: Implemented `CuratedAuditAtom` to strip technical bloat and focus LLM on semantic intent/logic.
- [x] **Report Enhancements**: Added `ResolutionMessage` capture and reporting for semantic failures.
- [x] **Tooling Synchronization**: Renamed `atd_verify` to `atd_check` across all docs, CLI, and MCP tools.

### Observations
- **Rigor Increase**: Initial tests show that `llama3.2` is significantly less likely to "hallucinate" compliance when given a curated, persona-driven prompt.
- **Failures Detected**: Several atoms (e.g., `rule_dto_strict_typing`) are now failing semantic audits that were previously passing, indicating that the new personas are catching documentation-code mismatches.
- **Divergence**: Discovered that some `BUSINESS` layer atoms still contain implementation details that the `Synthetic PM` persona correctly flags as "technical jargon".

### Next Steps
- [ ] **Unpack Failures**: Analyze specific failure modes for `rule_dto_strict_typing` and other failing atoms.
- [ ] **Protocol Refinement**: Adjust prompt temperatures or few-shot examples if personas are *too* strict.
- [ ] **Bulk Verification**: Run `atd check --full --semantic` on `upsilon-hub` to establish a new baseline.

---

## References

- [ATD.md](file:///home/bastien/work/skill/ATD.md)
- [atd/pkg/prompt/audit_code.go](file:///home/bastien/work/skill/atd/pkg/prompt/audit_code.go)
- [atd/cmd/atd/cmd/check_coverage.go](file:///home/bastien/work/skill/atd/cmd/atd/cmd/check_coverage.go)
