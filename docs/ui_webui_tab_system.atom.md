---
id: ui_webui_tab_system
human_name: "Tab Navigation System"
type: UI
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [ui, navigation, tab]
parents:
  - [[module_webui]]
dependents: [[[mechanic_webui_breadcrumb_navigation]]]
---

# Tab Navigation System

## INTENT
Ensure a consistent and intuitive navigation experience across all feature tabs in the Web UI.

## THE RULE / LOGIC
- **Header:** Contains the global logo, project info, and a horizontal tab switcher.
- **Tab Bar:** Active tab must be highlighted with an accent-colored underline and higher contrast.
- **Content Area:** All tabs must render within a unified `main-content` flexbox container to ensure side-panels and tools appear consistently.
- **Tab Switching:** Must be handled via a state-change in the JS; the current view is swapped without a full-page reload.

## TECHNICAL INTERFACE (The Bridge)
- **HTML Class:** `.tab-btn`, `.tab-content`
- **JS Event Listener:** `click` on `.tab-btn`
- **Code Tag:** `@spec-link [[ui_webui_tab_system]]`

## EXPECTATION (For Testing)
When switching from 'Explorer' to another tab, the content of the 'Explorer' tab must be hidden (CSS: `display: none`) and the new tab content displayed (CSS: `display: block`).
