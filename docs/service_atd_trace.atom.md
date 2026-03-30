---
id: service_atd_trace
status: REVIEW
type: SERVICE
layer: IMPLEMENTATION
priority: 3
parents: [[api_atd_mcp_ops]]
version: 1.0
dependents:
  - [[atd_health_snapshot_schema]]
human_name: Trace Service
---

# New Atom

## INTENT
Orchestrate the cross-referencing of an atom's graph position with its real-world implementation and test coverage to produce a machine-readable health snapshot.

## THE RULE / LOGIC
- Crawl all atoms to build the standard dependency graph.
- Extract all @spec-link and @test-link tags from the codebase.
- For the target atom:
  - Walk up parents to ensure a path to CUSTOMER layer (ancestry compliance).
  - Walk down dependents to identify all IMPLEMENTATION layer leaves.
  - Intersection of dependents and @spec-link sources defines Implementation Rate.
  - Intersection of implemented dependents and @test-link files defines Test Coverage Rate.
- Emit structured JSON matching TraceSnapshot schema.

## TECHNICAL INTERFACE
- **Source File**: `scripts/cmd/atd/cmd/trace.go`
- **Spec Link**: `@spec-link [[service_atd_trace]]`

## EXPECTATION
When triggered from the API, correctly walks the ancestry (parents) and descendants (dependents) of any valid atom. Correctly detects layer violations (e.g., ARCHITECTURE atom with no CUSTOMER parent). Produces accurate implementation and test coverage ratios based on @spec-link and @test-link indices.
