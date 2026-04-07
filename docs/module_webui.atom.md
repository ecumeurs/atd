---
id: module_webui
human_name: "ATD WebUI Module"
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: REVIEW
priority: 4
tags: [webui, atd, visualization, gemini]
parents:
  - [[requirement_webui_platform]]
dependents:
  - [[api_webui_atd_router]]
  - [[api_webui_health_stats]]
  - [[mechanic_webui_gemini_proxy]]
  - [[ui_webui_document_viewer]]
  - [[ui_webui_global_theme]]
  - [[ui_webui_spec_builder]]
  - [[ui_webui_spec_builder]]
  - [[ui_webui_tab_system]]
  - [[ui_webui_traceability_explorer]]
---

# ATD WebUI Module

## INTENT
Provide a rich, interactive web interface for exploring the ATD graph, editing atoms in real-time, and building new specifications using Gemini AI.

## THE RULE / LOGIC
Decomposed Go backend with ES module frontend:

**Backend (Go/Gin):**
- `main.go` — Router-only entry point (~40 lines)
- `config.go` — Configuration types and loading
- `atoms.go` — Atom state management and intent extraction
- `handlers_atd.go` — ATD API handlers (tree, detail, update, search, summary)
- `handlers_gemini.go` — Gemini API handlers (chat, models, proposals)

**Frontend (ES Modules):**
- `js/app.js` — Entry point, tab switching, data loading
- `js/api.js` — Centralized fetch wrappers
- `js/state.js` — Shared reactive state with event bus
- `js/explorer.js` — Waterfall of Intent three-column layout
- `js/details.js` — Atom detail panel with edit form
- `js/treeview.js` — Hierarchical tree view with bulk selection
- `js/search.js` — Ctrl+K server-side search overlay
- `spec-builder.js` — Gemini chat interface (IIFE)

## TECHNICAL INTERFACE (The Bridge)
- **Repo Root**: `webui/`
- **Main Entry**: `webui/main.go`
- **Frontend Assets**: `webui/static/`
- **Code Tag**: `@spec-link [[module_webui]]`

## EXPECTATION (For Testing)
- The server starts on port `8081` (default).
- Root endpoint `/` serves `index.html`.
- `api/tree` returns full JSON of project atoms.
