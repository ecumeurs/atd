---
id: ui_webui_heatmap_explorer
human_name: "WebUI Heat Map Visualization"
type: UI
layer: ARCHITECTURE
version: 1.0
status: DRAFT
priority: 3
tags: [webui, heatmap]
parents:
  - [[rule_atd_atom_overrides]]
dependents: []
bloating: on
heatmap: all
---

# WebUI Heat Map Visualization

## INTENT
Provide a visual layer in the Explorer to surface documentation health and complexity hotspots.

## THE RULE / LOGIC
- **Layer Toggle**: 4-state button group (Off, Dep, Code, Log).
- **Visual Feedback**:
    - **HOT**: Red pulsing border + `🔴` badge.
    - **WARM**: Orange pulsing border + `⚡` badge.
    - **COLD**: Blue border + inner glow.
- **Interactions**: Tooltips on buttons, badges, and metrics explain the meaning of the thermal state and provide raw data.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[ui_webui_heatmap_explorer]]`
- **Related Files:** `atd/pkg/webui/static/index.html`, `atd/pkg/webui/static/js/explorer.js`, `atd/pkg/webui/static/styles.css`

## EXPECTATION
1. Users must see a description of the heat layer when hovering over the toggle buttons.
2. Clicking 'Dep' must immediately update card borders based on coupling metrics.
