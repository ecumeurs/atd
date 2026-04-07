---
id: api_webui_atom_tests
status: STABLE
version: 1.0
layer: ARCHITECTURE
parents: [[api_webui_atd_router]]
tags: webui,api,tests
dependents:
  - [[mechanic_webui_atom_tests]]
human_name: WebUI Atom Tests API
type: API
priority: 3
---

# New Atom

## INTENT
Provide test coverage metrics for a specific atom.

## THE RULE / LOGIC
Endpoint returns boolean indicators of whether an atom has tests and if they are passing.

## TECHNICAL INTERFACE

## EXPECTATION
/atd/:id/tests returns corresponding test coverage data.
