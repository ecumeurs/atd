---
id: service_vscode_atd_ui
status: DRAFT
priority: 3
human_name: "VS Code ATD UI Components"
version: 1.0.0
parents:
  - [[specification_vscode_atd_linker]]
dependents:
  - [[mechanic_vscode_sidebar_tree]]
  - [[mechanic_vscode_webview_graph]]
type: SERVICE
layer: ARCHITECTURE
---

# VS Code ATD UI Components

## INTENT
Orchestrate ATD-specific UI components including the Sidebar Explorer and the Webview Graph.

## THE RULE / LOGIC
1. Register ATDGraphProvider as a TreeDataProvider.
2. Manage a WebviewPanel for Vis.js graph rendering.
3. Synchronize both views with onDidChangeActiveTextEditor events.

## TECHNICAL INTERFACE

## EXPECTATION
Sidebar and Webview components correctly display the ATD graph and update based on editor selection.
