---
id: mechanic_webui_bulk_update
status: STABLE
priority: 3
parents: [[api_webui_bulk_update]]
dependents: []
type: MECHANIC
layer: IMPLEMENTATION
tags: webui,mechanic,bulk
version: 1.0
human_name: WebUI Bulk Update Mechanic
---

# New Atom

## INTENT
Implement the fast bulk mutation via loop on the atd update CLI.

## THE RULE / LOGIC
Iterate IDs, executing `atd update --set status=X` for each, then refresh graph.

## TECHNICAL INTERFACE

## EXPECTATION
Successfully sets all items and ignores unknown IDs.
