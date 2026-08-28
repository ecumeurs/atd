---
id: ui_webui_details_weave_button
status: DRAFT
type: UI
priority: 3
version: 1.0
parents:
  - [[requirement_webui_force_weave]]
dependents:
  - [[api_webui_atd_weave]]
human_name: WebUI Detail Weave Button
layer: ARCHITECTURE
---

# WebUI Detail Weave Button

## INTENT
Provide a visual entry point for the force weave action in the ATD Detail side panel.

## THE RULE / LOGIC
A 'Force Weave' button should be added to the `.detail-actions` section of the side panel, alongside 'Edit' and 'AI Summary'. It should be a secondary button but easily visible.

## TECHNICAL INTERFACE
@spec-link [[ui_webui_details_weave_button]]

## EXPECTATION
The button appears in the detail-actions div and is visible only when an atom is selected.
