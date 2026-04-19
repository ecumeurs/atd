# Issue: Granular Instance-Based Verification with JSON Prompting

**ID:** `20260419_granular_verify_json_recap`
**Ref:** `ISS-083`
**Date:** 2026-04-19
**Severity:** Medium
**Status:** Open
**Component:** `atd/cmd/atd/cmd/verify.go`
**Affects:** CI/CD compliance audits, IDE Agent verification workflow

---

## Summary

The current `atd verify` command produces a single, large "Document Bundle" prompt that contains all linked atoms and all modified source code. This lack of granularity forces the LLM to process a massive context window and manually map atoms to specific code sections, increasing the risk of hallucinations and missing compliance details. Furthermore, testing is currently directory-based rather than graph-based, leading to missing proof for cross-file tests. This issue tracks the transition to a per-tag ("Atomic Instance") audit model with surgical snippet capture, ancestry context, @test-link discovery, and structured JSON prompting.

---

## Technical Description

### Background
Currently, `runVerify` collects all `@spec-link` IDs and all file contents, then concatenates them into one large prompt. The LLM is then asked to "Audit all changes."

### The Problem Scenario
1. Developer modifies 5 files touching 10 different atoms.
2. `atd verify` generates one prompt with 10 atoms and 5 large files.
3. The LLM must reason across all 10 rules simultaneously.
4. If one specific rule is violated in a subtle way, it may be lost in the "noise" of the other 9 rules being compliant.
5. There is no machine-readable recap showing which specific tags passed or failed.

### Where This Pattern Exists Today
`atd/cmd/atd/cmd/verify.go`: The `runVerify` function lacks iteration over individual `@spec-link` instances.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | Medium — manifests as "false pass" audits from the LLM |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** 
- Refactor `verify.go` to iterate through every `@spec-link` found.
- For each link, extract a surgical code snippet (heuristic-based).
- For each link, perform a global graph search for `@test-link [[atom_id]]` to include relevant verification proof cross-file.
- Include the atom's ancestry chain (parents) for context.
- Format the instance prompt to request a JSON response from the LLM.

**Medium term:** 
- Implement a final recap assembly that prints a table of valid/invalid tags with file:line locations.
- Include a "Coverage Warning" in the recap for atoms with implementations but 0 verified tests.
- Add "Fix Recommendations" in the recap: suggest using `atd discover` or `atd trace` when compliance or coverage gaps are found.
- Add an option to target a specific tag at a specific file/line: `atd verify --file path/to/code.go --line 123`.

**Long term:** 
- Integrate with an AST parser for perfect code section extraction instead of heuristics.

---

## References

- [verify.go](file:///home/bastien/work/skill/atd/cmd/atd/cmd/verify.go)
- [exploration.go](file:///home/bastien/work/skill/atd/pkg/exploration/exploration.go)
