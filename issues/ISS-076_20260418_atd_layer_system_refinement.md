# Issue: ATD Layer System Overload and Ambiguity

**ID:** `20260418_atd_layer_system_refinement`
**Ref:** `ISS-076`
**Date:** 2026-04-18
**Severity:** Medium
**Status:** Open
**Component:** `ATD.md`, `scripts/pkg/atom/layers.go`, documentation system
**Affects:** Atom Placement Guidance, Layer Compliance Checking, Documentation Structure

---

## Summary

Current ATD layer system (CUSTOMER, ARCHITECTURE, IMPLEMENTATION) has overloaded ARCHITECTURE layer with both high-level design and implementation contracts, and underutilized IMPLEMENTATION layer, causing ambiguity in atom placement.

---

## Technical Description

### Background
ATD layers organize atoms by their relationship to change and human oversight, guiding their placement and expected behavior.

### The Problem Scenario
1. **ARCHITECTURE Layer Overload**: Contains both high-level system design and implementation contracts (APIs, UI, data models)
2. **IMPLEMENTATION Layer Underutilized**: Many implementation details live in ARCHITECTURE layer instead
3. **Gray Areas**: Some atoms don't clearly fit one layer, leading to inconsistent placement
4. **82% of @spec-link Tags**: Point to ARCHITECTURE layer atoms, suggesting layer misuse

### Where This Pattern Exists Today
- **Current Layers**: CUSTOMER, ARCHITECTURE, IMPLEMENTATION (3 layers)
- **Layer Definitions**: `ATD.md §1.4 Document Hierarchy & Layers`
- **Layer Assignment**: Atom creation and validation logic

### Evidence from Investigation
- **ARCHITECTURE Layer**: Contains API contracts, UI components, data models, business rules
- **IMPLEMENTATION Layer**: Contains mostly low-level mechanics, underutilized
- **Link Distribution**: 82% of @spec-link tags point to ARCHITECTURE layer atoms
- **Recommendation**: Rename ARCHITECTURE → DESIGN for accuracy, narrow IMPLEMENTATION scope

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Confirmed via analysis of layer usage and link distribution) |
| Impact if triggered | Medium (Ambiguous atom placement, inconsistent documentation structure) |
| Detectability | High (Visible in layer distribution and atom classification) |
| Current mitigant | Manual layer selection guidance in ATD.md and CLAUDE.md |

---

## Recommended Fix

**Short term**: Rename ARCHITECTURE → DESIGN for accuracy. Clarify layer scope: DESIGN for "what", IMPLEMENTATION for "how" (complex algorithms only). Update ATD.md layer definitions.

**Medium term**: Implement IMPLEMENTATION layer criteria: algorithm complexity > 10 lines pseudo-code, mathematical calculations, performance optimizations, cross-cutting concerns. Move simple mechanics from IMPLEMENTATION to DESIGN layer.

**Long term**: Add layer compliance checking to `atd_lint` to verify atoms are in appropriate layers. Implement layer-specific orphan detection rules and coverage expectations.

---

## References

- [ATD System Analysis](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/atd_system_analysis.md)
- [ATD.md Document Hierarchy](file:///home/bastien/work/skill/ATD.md#14-document-hierarchy--layers)
- [Layer Implementation](file:///home/bastien/work/skill/scripts/pkg/atom/layers.go)