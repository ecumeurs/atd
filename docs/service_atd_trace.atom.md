---
id: service_atd_trace
status: DRAFT
type: SERVICE
layer: IMPLEMENTATION
priority: 3
parents: [[api_atd_mcp_ops]]
version: 1.0
dependents: []
human_name: ATD Trace Command
---

# New Atom

## INTENT
Provide an automated, snapshot view of an atom's health constraints, tracking upward ancestry layer compliance and downward test/implementation coverage.

## THE RULE / LOGIC
- Perform an upward graph traversal of `parents` to identify missing `CUSTOMER` or `ARCHITECTURE` origin layers, and assert that all ancestors are `STABLE`.
- Perform a downward traversal of `dependents` to identify missing `IMPLEMENTATION` layers.
- Cross-reference downstream dependent leaf nodes against `@spec-link` source indices to compute the `implementation_rate`.
- Calculate `test_coverage_rate` by identifying which of those implemented downstream nodes also appear alongside `@test-link` annotations.
- Warn if an atom breaks layer constraints (e.g. IMPLEMENTATION atom with zero source/test implementations).

## TECHNICAL INTERFACE

## EXPECTATION
Given a real atom ID, output matches the `HealthSnapshot` JSON schema. Handles invalid IDs and missing directories gracefully. Warns correctly about missing layer origins and implementations.
