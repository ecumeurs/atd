---
id: ui_webui_waterfall_explorer
status: REVIEW
priority: 1
human_name: WebUI Waterfall of Intent Explorer
type: UI
tags: webui,explorer,visualization
version: 1.0
parents: [[ui_webui_explorer_layout]]
dependents: [[[mechanic_webui_connector_lines]]]
layer: ARCHITECTURE
---

# New Atom

## INTENT
Provide a three-column lane visualization showing atoms flowing from Customer through Architecture to Implementation, revealing architectural intent and traceability.

## THE RULE / LOGIC
Three-column lane layout (Customer → Architecture → Implementation):
- Each lane displays atom cards grouped by MODULE parent
- Active atoms are sorted into lanes by their `layer` field
- Stable+implemented+tested atoms are moved to a collapsed Foundation section
- Cards show type badge, status pill, intent preview, and coverage micro-indicators
- Clicking a card opens the detail panel and highlights the full ancestry path
- Module groups have clickable headers for quick navigation

## TECHNICAL INTERFACE

## EXPECTATION
The waterfall explorer renders all atoms in correctly-assigned lanes. Connector lines trace parent-child relationships. Foundation section starts collapsed. Clicking an atom shows its detail panel and ancestry path.
