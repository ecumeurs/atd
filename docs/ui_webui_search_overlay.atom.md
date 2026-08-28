---
id: ui_webui_search_overlay
status: REVIEW
human_name: WebUI Search Overlay
type: UI
layer: ARCHITECTURE
version: 1.0
dependents: []
priority: 2
tags: webui,search
parents:
  - [[ui_webui_explorer_layout]]
---

# WebUI Search Overlay

## INTENT
Provide a Ctrl+K command palette for fast atom lookup by name, ID, or content using server-side search.

## THE RULE / LOGIC
Ctrl+K command palette for server-side atom search:
- Opens a modal overlay with text input and results list
- Debounced input triggers GET /api/search?q=... server-side
- Results show type, name, layer, status, and intent preview
- Keyboard navigation (↑↓ arrows, Enter to select)
- Selecting a result navigates to the atom and closes the overlay
- ESC or clicking backdrop closes the overlay

## TECHNICAL INTERFACE

## EXPECTATION
Ctrl+K opens search. Typing filters atoms server-side. Selecting navigates to atom in explorer. Overlay closes on Escape.
