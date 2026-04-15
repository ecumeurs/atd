---
id: mechanic_webui_connector_lines
status: DRAFT
human_name: WebUI SVG Connector Lines
type: MECHANIC
layer: IMPLEMENTATION
priority: 4
tags: webui,svg,visualization
parents:
  - [[ui_webui_waterfall_explorer]]
version: 1.0
dependents: []
---

# New Atom

## INTENT
Draw SVG Bézier connector lines between parent-child atom cards in the Waterfall of Intent layout to visualize cross-layer traceability. Currently deferred pending scroll-sync implementation.

## THE RULE / LOGIC
1. After atom cards are rendered, iterate all atoms with parent links.
2. For each parent-child pair, find their card DOM positions.
3. Calculate Bezier control points for smooth horizontal curves across lanes.
4. Draw SVG path elements in an overlay SVG that spans the waterfall container.
5. On click of an atom, highlight its full ancestry path and dim other lines.
6. Recalculate positions on scroll or window resize.

## TECHNICAL INTERFACE

## EXPECTATION
SVG lines correctly connect parent cards to child cards across lanes. Lines update on scroll/resize. Selected atom highlights its full path.
