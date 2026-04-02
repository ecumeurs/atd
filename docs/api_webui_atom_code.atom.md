---
id: api_webui_atom_code
status: STABLE
human_name: WebUI Atom Code API
version: 1.0
type: API
layer: ARCHITECTURE
priority: 3
parents: [[api_webui_atd_router]]
tags: webui,api,code
dependents: [[[mechanic_webui_atom_code]]]
---

# New Atom

## INTENT
Provide mapped source code blocks linked to a specific atom.

## THE RULE / LOGIC
Endpoint retrieves the linked codes for a given atom ID.

## TECHNICAL INTERFACE

## EXPECTATION
/atd/:id/code returns related snippet list or 404.
