---
id: ui_webui_explorer_layout
human_name: "Explorer Tab Layout"
type: UI
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [ui, layout, explorer]
parents:
  - [[ui_webui_traceability_explorer]]
dependents:
  - [[rule_webui_blob_sizing]]
  - [[ui_webui_explorer_treeview]]
  - [[ui_webui_search_overlay]]
  - [[ui_webui_waterfall_explorer]]
---

# Explorer Tab Layout

## INTENT
Define the spatial distribution of the zoomed health blobs and the detail-view interface within the Explorer tab.

## THE RULE / LOGIC
- **Main Container:** A flexbox row covering the full viewport below the header.
- **Visualization Zone (Left):** Occupies 80% of the horizontal space; hosts the D3 Treemap SVG/Canvas.
- **Details Side-Panel (Right):** Occupies 20% of the space; collapsible; contains the metadata for the selected atom.
- **Fluidity:** The visualization zone must re-render via a resize observer whenever the details panel is collapsed or expanded.

## TECHNICAL INTERFACE (The Bridge)
- **CSS Selectors:** `#treemap-container`, `.details-panel`, `.health-blob`
- **D3 Layout:** Squarified Treemap for the initial view; Partition layout for zoomed-in hierarchy.
- **Code Tag:** `@spec-link [[ui_webui_explorer_layout]]`

## EXPECTATION (For Testing)
The Explorer tab area must be divided into a main visualization zone (80%) and a details side-panel (20%, expandable). The visualization zone must hold the four health blobs.
