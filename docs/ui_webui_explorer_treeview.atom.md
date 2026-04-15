---
id: ui_webui_explorer_treeview
status: REVIEW
type: UI
human_name: WebUI Explorer Tree View
layer: ARCHITECTURE
priority: 3
tags: webui,explorer,treeview
version: 1.0
parents:
  - [[ui_webui_explorer_layout]]
dependents: []
---

# New Atom

## INTENT
Provide a hierarchical indented tree view of atoms organized by parent-child relationships for bulk selection and status management.

## THE RULE / LOGIC
Hierarchical list view of all atoms:
- Builds parent-child hierarchy from flat atom list
- Each tree item shows status dot, version badge, status pill, name, and type badge
- Selection mode adds checkboxes for bulk status updates
- Bulk bar appears at top with status dropdown and Apply/Cancel actions
- Clicking an item opens it in the detail panel

## TECHNICAL INTERFACE

## EXPECTATION
Tree view renders all atoms in correct hierarchy. Selection mode enables multi-select with bulk status change. Click opens detail panel.
