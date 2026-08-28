---
id: mechanic_webui_atom_code
status: STABLE
type: MECHANIC
layer: IMPLEMENTATION
parents:
  - [[api_webui_atom_code]]
version: 1.0
dependents: []
human_name: WebUI Atom Code Mechanic
priority: 3
tags: webui,mechanic,code
---

# WebUI Atom Code Mechanic

## INTENT
Implement the /atd/:id/code endpoint.

## THE RULE / LOGIC
Extract `id` param, lookup `Atoms` map, and retrieve `LinkedCodes` field.

## TECHNICAL INTERFACE

## EXPECTATION
Correctly returns the linked code lines.
