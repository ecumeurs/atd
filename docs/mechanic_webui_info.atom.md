---
id: mechanic_webui_info
status: STABLE
layer: IMPLEMENTATION
parents: [[api_webui_info]]
priority: 3
tags: webui,mechanic,info
version: 1.0
dependents: []
human_name: WebUI Info Mechanic
type: MECHANIC
---

# New Atom

## INTENT
Implement the /info endpoint by reading the global AppConfig and Atoms map length.

## THE RULE / LOGIC
Retrieve `AppConfig.ProjectPath`, `AppConfig.ATDPath`, and `len(Atoms)`.

## TECHNICAL INTERFACE

## EXPECTATION
Correctly marshals global config state into the API response.
