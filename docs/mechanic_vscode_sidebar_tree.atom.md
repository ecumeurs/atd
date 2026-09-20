---
id: mechanic_vscode_sidebar_tree
status: DRAFT
human_name: "Sidebar Graph Explorer"
type: MECHANIC
version: 1.0.0
priority: 3
layer: IMPLEMENTATION
parents:
  - [[service_vscode_atd_ui]]
dependents: []
---

# Sidebar Graph Explorer

## INTENT
Implement a TreeDataProvider for the VS Code sidebar to explore atom parent/dependent relationships.

## THE RULE / LOGIC
1. Track active atom via `onDidChangeActiveTextEditor`.
2. Use `mechanic_vscode_atom_parser` to get raw links.
3. Use `atd trace` to get health status (pass/warning) for each child node.
4. Update TreeView items with ThemeIcons.

## TECHNICAL INTERFACE

## EXPECTATION
The sidebar tree accurately lists parents and dependents with their health icons.
