# Issue: ATD Linter Aggressive Noise and Parsing Bug

**ID:** `20260415_linter_noise_and_parser_bug`
**Ref:** `ISS-070`
**Date:** 2026-04-15
**Severity:** Medium
**Status:** Open
**Component:** `scripts/pkg/atom/`, `scripts/cmd/atd/cmd/`
**Affects:** Developer workflow (CI/Linting), Documentation Quality Assurance

---

## Summary

The `atd lint` tool currently produces a high volume of false positives and incorrectly identifies sections as missing. This is caused by a fundamental flaw in how the parser handles Markdown headers and a lack of type-awareness in the linter's validation logic.

---

## Technical Description

### Background
The ATD linter is designed to ensure that all `.atom.md` files contain mandatory sections (`## INTENT`, `## THE RULE / LOGIC`, `## TECHNICAL INTERFACE`, and `## EXPECTATION`).

### The Problem Scenario
1. **Parser Reset Bug**: In `scripts/pkg/atom/parse.go`, the section parser resets its state whenever it encounters a line starting with `##`. This includes H3+ sub-headers (e.g., `### Color Palette`). If an atom uses sub-headers immediately after a mandatory section header, the content buffer for that section remains empty, leading to a "Missing mandatory section" error.
2. **Type Over-Enforcement**: The linter requires `## EXPECTATION` for all atoms. However, architectural grouping atoms like `MODULE` or `DOMAIN` often do not have direct testable expectations, as they only aggregate lower-level rules.

### Where This Pattern Exists Today
- `scripts/pkg/atom/parse.go:L175-194`: The header detection logic is too broad.
- `scripts/cmd/atd/cmd/lint.go:L99-110`: The validation logic is type-agnostic.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Happens in any project with sub-headers or MODULE types) |
| Impact if triggered | Medium (Developer fatigue, lint results ignored due to noise) |
| Detectability | High (Visible in `atd lint` output) |
| Current mitigant | Manually ignoring lint errors or adding empty placeholder sections. |

---

## Recommended Fix

**Short term**: Fix the parser in `parse.go` to only switch states on exact H2 headers (`## `) followed by mandatory keywords, ignoring `###` and other sub-headers.

**Medium term**: Update the linter to make `EXPECTATION`, `THE RULE / LOGIC`, and `TECHNICAL INTERFACE` optional for specific types:
- `MODULE`, `DOMAIN`, `requirement`, `SPECIFICATION`, `USER_STORY`.

**Long term**: Implement a proper Markdown parser instead of the line-by-line scanner to handle nested structures more robustly.

---

## References

- [parse.go](file:///home/bastien/work/atd/scripts/pkg/atom/parse.go)
- [lint.go](file:///home/bastien/work/atd/scripts/cmd/atd/cmd/lint.go)
- [Implementation Plan](file:///home/bastien/.gemini/antigravity/brain/ba71aad7-ea04-442b-b082-bd6aa329d02c/implementation_plan.md)
