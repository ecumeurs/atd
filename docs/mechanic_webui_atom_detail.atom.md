---
id: mechanic_webui_atom_detail
status: STABLE
human_name: WebUI Atom Detail Mechanic
priority: 3
parents: [[api_webui_atom_detail]]
tags: webui,mechanic,detail
version: 1.0
dependents: []
type: MECHANIC
layer: IMPLEMENTATION
---

# New Atom

## INTENT
Implement the /atd/:id endpoint by looking up the ID in the Atoms map.

## THE RULE / LOGIC
Extract `id` param, perform map lookup on `Atoms`. Return 404 if not found.

## TECHNICAL INTERFACE

## EXPECTATION
Fetches correctly or errors gracefully.
