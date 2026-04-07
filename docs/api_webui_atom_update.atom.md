---
id: api_webui_atom_update
status: STABLE
type: API
priority: 3
tags: webui,api,update
dependents:
  - [[mechanic_webui_atom_update]]
human_name: WebUI Atom Update API
layer: ARCHITECTURE
parents: [[api_webui_atd_router]]
version: 1.0
---

# New Atom

## INTENT
Allow frontend clients to mutate atom fields and content.

## THE RULE / LOGIC
Endpoint processes a partial JSON atom representation and applies it to disk.

## TECHNICAL INTERFACE

## EXPECTATION
A mutation request successfully alters the matching atom or errors appropriately.
