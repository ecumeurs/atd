---
id: api_webui_atd_router
status: STABLE
human_name: WebUI ATD Router
type: API
layer: ARCHITECTURE
priority: 4
parents: [[module_webui]], [[requirement_webui_documentation_management]]
tags: webui,router,api
version: 1.0
dependents: [[[api_webui_atom_code]], [[api_webui_atom_detail]], [[api_webui_atom_tests]], [[api_webui_atom_update]], [[api_webui_bulk_update]], [[api_webui_info]], [[api_webui_search]], [[api_webui_summary]], [[api_webui_tree]]]
---

# New Atom

## INTENT
Encapsulate all ATD related API endpoints served by the WebUI.

## THE RULE / LOGIC
Groups the sub-routes under the ATD group in gin.

## TECHNICAL INTERFACE

## EXPECTATION
All endpoints defined are correctly grouped and routed.
