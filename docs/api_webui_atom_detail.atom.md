---
id: api_webui_atom_detail
status: STABLE
layer: ARCHITECTURE
priority: 3
parents: [[api_webui_atd_router]]
version: 1.0
dependents:
  - [[mechanic_webui_atom_detail]]
human_name: WebUI Atom Detail API
type: API
tags: webui,api,detail
---

# New Atom

## INTENT
Provide detailed data for a specific ATD atom.

## THE RULE / LOGIC
Endpoint looks up an atom by ID and returns it.

## TECHNICAL INTERFACE

## EXPECTATION
/atd/:id returns the atom or 404.
