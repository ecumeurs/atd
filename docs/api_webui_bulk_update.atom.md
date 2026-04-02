---
id: api_webui_bulk_update
status: STABLE
human_name: WebUI Bulk Update API
parents: [[api_webui_atd_router]]
tags: webui,api,bulk
dependents: [[[mechanic_webui_bulk_update]]]
type: API
layer: ARCHITECTURE
priority: 3
version: 1.0
---

# New Atom

## INTENT
Allow frontend clients to apply a status update to many atoms simultaneously.

## THE RULE / LOGIC
Endpoint receives an array of IDs and a target status.

## TECHNICAL INTERFACE

## EXPECTATION
Status updates are applied sequentially to all requested IDs.
