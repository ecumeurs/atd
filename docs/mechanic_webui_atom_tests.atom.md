---
id: mechanic_webui_atom_tests
status: STABLE
layer: IMPLEMENTATION
priority: 3
version: 1.0
dependents: []
human_name: WebUI Atom Tests Mechanic
type: MECHANIC
parents:
  - [[api_webui_atom_tests]]
tags: webui,mechanic,tests
---

# WebUI Atom Tests Mechanic

## INTENT
Implement the /atd/:id/tests endpoint.

## THE RULE / LOGIC
Lookup atom by `id` parameter, return `HasTests` and `IsGreen` fields from memory map.

## TECHNICAL INTERFACE

## EXPECTATION
Responses reflect memory cache data.
