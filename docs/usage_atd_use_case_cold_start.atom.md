---
id: usage_atd_use_case_cold_start
human_name: "Use Case: Cold Start with ATD"
type: USAGE
version: 1.0
status: STABLE
priority: 5
tags: [atd, usecase, coldstart, legacy]
parents:
  - [[domain_atd_usage_protocol]]
dependents: []
layer: IMPLEMENTATION
---

# Use Case: Cold Start with ATD

## INTENT
To illustrate how to bootstrap ATD in an existing, undocumented codebase.

## THE RULE / LOGIC
1. **Initialize**: Run `atd init` to create the `.atd` config and the `docs/` directory.
2. **Prioritize**: Run `atd roadmap` to rank key source files by density, then author atoms for the highest-priority files in `docs/`.
3. **Refine**: Review and edit the authored atoms in `docs/` or through the WebUI.
4. **Link**: Manually or automatically (via `atd recon`) place `@spec-link` tags in the code.
5. **Freeze**: Once stable, run `atd index` to build the semantic search database.

## TECHNICAL INTERFACE
- **Command:** `atd init`, `atd roadmap`, `atd index`.

## EXPECTATION
- A `docs/` directory populated with atoms.
- Source files containing `@spec-link` tags matching atom IDs.
- A functional semantic search via `atd search`.
