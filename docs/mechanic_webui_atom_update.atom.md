---
id: mechanic_webui_atom_update
status: STABLE
dependents: []
type: MECHANIC
layer: IMPLEMENTATION
tags: webui,mechanic,update
version: 1.0
human_name: WebUI Atom Update Mechanic
priority: 3
parents: [[api_webui_atom_update]]
---

# New Atom

## INTENT
Implement the mutate operations via the atd update CLI sub-tool.

## THE RULE / LOGIC
Proxy incoming changes to `atd update --set xyz` command and manual content rewrites. Refresh map upon success.

## TECHNICAL INTERFACE

## EXPECTATION
Executes CLI correctly and updates file bytes for content.
