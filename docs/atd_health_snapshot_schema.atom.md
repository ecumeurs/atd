---
id: atd_health_snapshot_schema
human_name: "ATD Health Snapshot Schema"
type: ENTITY
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 3
tags: [schema, json, trace]
parents:
  - [[service_atd_trace]]
dependents: []
---

# ATD Health Snapshot Schema

## INTENT
Define the formal JSON structure of the health snapshot produced by the `atd trace` command.

## THE RULE / LOGIC
The output must be a JSON object with the following structure:
- `target_id`: String (ID of the traced atom)
- `layer`: Enum (CUSTOMER, ARCHITECTURE, IMPLEMENTATION, etc.)
- `health_summary`:
  - `ancestry_complete`: Boolean (True if all parents lead to a CUSTOMER origin)
  - `has_customer_origin`: Boolean
  - `implementation_rate`: Number (0.0 to 1.0)
  - `test_coverage_rate`: Number (0.0 to 1.0)
- `metrics`:
  - `total_dependents`: Integer
  - `implemented_dependents`: Integer
  - `total_code_files`: Integer
  - `total_tests`: Integer
- `graph_slice`:
  - `parents`: List of Strings
  - `dependents`: List of Strings
  - `code_links`: List of Objects (source file paths)
  - `test_links`: List of Objects (test file paths)
- `warnings`: List of Strings

## TECHNICAL INTERFACE
- **Source Module**: `internal/trace`
- **Output Format**: JSON-RPC compatible

## EXPECTATION
The `atd trace` command produces valid JSON that adheres strictly to this schema for any target atom.
