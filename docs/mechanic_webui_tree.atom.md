---
id: mechanic_webui_tree
status: STABLE
type: MECHANIC
priority: 3
version: 1.0
human_name: WebUI Tree Mechanic
layer: IMPLEMENTATION
parents:
  - [[api_webui_tree]]
tags: webui,mechanic,tree
dependents: []
---

# New Atom

## INTENT
Implement the /tree endpoint by transforming the Atoms map to a slice.

## THE RULE / LOGIC
Iterate over the `Atoms` map and append to a slice.

## TECHNICAL INTERFACE

## EXPECTATION
Successfully converts map to slice and outputs as JSON.
