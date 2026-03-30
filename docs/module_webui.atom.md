---
id: module_webui
human_name: "ATD WebUI Module"
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: DRAFT
priority: 4
tags: [webui, atd, visualization, gemini]
parents:
  - [[requirement_webui_platform]]
dependents: [[[api_webui_health_stats]], [[mechanic_webui_gemini_proxy]], [[ui_webui_global_theme]], [[ui_webui_spec_builder]], [[ui_webui_spec_builder]], [[ui_webui_tab_system]], [[ui_webui_traceability_explorer]]]
---

# ATD WebUI Module

## INTENT
Provide a rich, interactive web interface for exploring the ATD graph, editing atoms in real-time, and building new specifications using Gemini AI.

## THE RULE / LOGIC
The WebUI is a Go-based Gin server serving a single-page application (SPA). 
- **Tree Visualization**: Use D3.js and `atd_query` (simulated via local parser) to render current ATDs.
- **Spec Builder**: Interface with Gemini/models to facilitate interactive Spec decomposition.
- **Traceability**: All interactions must maintain or propose atomic documentation links.

## TECHNICAL INTERFACE (The Bridge)
- **Repo Root**: `webui/`
- **Main Entry**: `webui/main.go`
- **Frontend Assets**: `webui/static/`
- **Code Tag**: `@spec-link [[module_webui]]`

## EXPECTATION (For Testing)
- The server starts on port `8081` (default).
- Root endpoint `/` serves `index.html`.
- `api/tree` returns full JSON of project atoms.
