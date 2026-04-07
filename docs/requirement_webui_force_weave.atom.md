---
id: requirement_webui_force_weave
status: DRAFT
layer: CUSTOMER
version: 1.0
human_name: Force Weave via WebUI
type: REQUIREMENT
priority: 3
parents: []
dependents:
  - [[ui_webui_details_weave_button]]
---

# New Atom

## INTENT
Allow users to manually trigger the ATD graph weaving process from the WebUI to ensure link consistency after manual edits.

## THE RULE / LOGIC
The WebUI must provide a clearly accessible action to trigger the 'atd weave' logic without requiring the user to switch to a terminal or CLI.

## TECHNICAL INTERFACE
@spec-link [[requirement_webui_force_weave]]

## EXPECTATION
User can click a button in the WebUI and the ATD graph links are synchronized across all files.
