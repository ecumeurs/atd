---
id: mechanic_vscode_webview_graph
status: DRAFT
type: MECHANIC
priority: 3
parents:
  - [[service_vscode_atd_ui]]
dependents: []
human_name: "Vis.js Graph Webview"
layer: IMPLEMENTATION
version: 1.0.0
---

# Vis.js Graph Webview

## INTENT
Implement a Webview panel using Vis.js to visualize the local neighborhood graph of an atom.

## THE RULE / LOGIC
1. Create `vscode.WebviewPanel` on demand or editor change.
2. Fetch trace data via `atd trace`.
3. Generate HTML with Vis.js configuration.
4. Map graph_slice data (parents, dependents, code_links) to vis.DataSet nodes and edges.

## TECHNICAL INTERFACE

## EXPECTATION
Webview renders a Vis.js network graph showing parents, dependents, and code links.
