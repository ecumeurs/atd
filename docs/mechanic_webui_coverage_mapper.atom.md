---
id: mechanic_webui_coverage_mapper
status: DRAFT
human_name: Coverage Mapper Mechanic
layer: IMPLEMENTATION
version: 1.0
priority: 3
parents:
  - [[api_webui_health_stats]]
dependents: []
type: MECHANIC
tags: [webui, mechanic, coverage]
---

# Coverage Mapper Mechanic

## INTENT
Map source code `@spec-link` and `@test-link` tags to the atom graph for status calculation.

## THE RULE / LOGIC
Extend `parser.FindLinkedCode` to produce separate `implementation_rate` and `test_coverage_rate` metrics.
1. Scan source files for `@spec-link [[atom_id]]`.
2. Scan test files (e.g., `_test.go`) for `@test-link [[atom_id]]`.
3. For each atom, calculate binary status:
   - **Implemented**: True if linked in source.
   - **Tested**: True if linked in test.
   - **Green**: True if both Implemented and Tested.

## TECHNICAL INTERFACE (The Bridge)
- **Go Function**: `parser.CalculateCoverage(atom)`
- **Code Tag**: `@spec-link [[mechanic_webui_coverage_mapper]]`
- **Test Names**: `TestCoverageMapper`

## EXPECTATION (For Testing)
Status calculation correctly differentiates between "Implemented" and "Tested".
