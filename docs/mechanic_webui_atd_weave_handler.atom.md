---
id: mechanic_webui_atd_weave_handler
status: DRAFT
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
parents:
  - [[api_webui_atd_weave]]
dependents: []
human_name: WebUI Weave Handler Mechanic
priority: 3
---

# WebUI Weave Handler Mechanic

## INTENT
Define the backend handler for the weave action in the WebUI.

## THE RULE / LOGIC
The handler must ensure that the ATD graph is initialized, call `atom.Weave(docsDir)`, log the action, and then call `s.refreshAtoms()` to update the WebUI's internal memory state. It should also return a clear success or error message to the client.

## TECHNICAL INTERFACE
@spec-link [[mechanic_webui_atd_weave_handler]]

## EXPECTATION
Successfully triggers atom.Weave and refreshes the internal WebUI atom cache.
