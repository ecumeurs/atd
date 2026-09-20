# Issue: ATD Graph Visualization Improvements

**ID:** `20260325_atd_graph_visualization_improvements`
**Ref:** `ISS-048`
**Date:** 2026-03-25
**Severity:** Medium
**Status:** Open
**Component:** `extension/extension.js`, `scripts/cmd/atd/cmd/trace.go`
**Affects:** VS Code Extension Graphical Explorer, ATD Trace CLI output

---

## Summary

The current ATD graph visualization in the VS Code extension (`atd.showFullGraph`) is functionally limited and aesthetically basic. It fails to show deeper ranks of the dependency graph (only immediate parents/dependents), lacks a visual legend for node states (layer, type, health), and does not support interactive exploration (hover for intent, click to re-center). 

Furthermore, the `atd trace` command provides ancestry and descendants as flattened blocks (`[]string`), which prevents the extension from easily reconstructing the hierarchical tree structure for multi-rank visualization. Improving this is critical for navigating complex ATD graphs and quickly assessing system health.

---

## Technical Description

### Background

The `atd.showFullGraph` command opens a Webview that renders a node-link diagram using `vis-network`. It calls `atd trace <atomId>` to get a `TraceSnapshot`. Currently, the `GraphSlice` in this snapshot flattens all related atoms into simple string slices (`Parents`, `Dependents`), losing the adjacency information necessary for a proper multi-level tree.

### The Problem Scenario

1.  Open an `.atom.md` file.
2.  Run `ATD: Show Full Graph`.
3.  The graph displays only one level of ancestry/descendants because `atd trace` does not provide the structure needed to link ancestors beyond the immediate parents.
4.  Nodes are basic ellipses or boxes with minimal stylistic differentiation.
5.  Hovering over a node does nothing (no intent displayed).
6.  Clicking a node does not change the focus or redraw the graph.

```
[ Legend Missing ]
      │
      ▼
  [ Parent A ]  <-- Basic grey ellipse
      │
      ▼
[ CENTER ATOM ] <-- Basic blue box (Layer info only)
      │
      ▼
 [ Dependent B ] <-- Basic grey ellipse
```

### Where This Pattern Exists Today

- `extension/extension.js:278-393`: `updateGraphPanel` handles the Webview rendering but receives flattened data.
- `extension/extension.js:332-386`: vis-network initialization lacks interactivity.
- `scripts/cmd/atd/cmd/trace.go:115-157`: `walkUp` and `walkDown` flatten the graph into string slices.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — degrades developer experience and system observability |
| Detectability | High — immediately visible in the UI |
| Current mitigant | Sidebar TreeView provides some health info but lacks the bird's-eye view of a graph. |

---

## Recommended Fix

**Short term:** 
- Add a CSS-based legend to the Webview.
- Implement `hoverNode` listener to show `intent` (from atom metadata).
- Implement `click` (or `selectNode`) listener to send a message back to the extension host to re-trigger `updateGraphPanel` with the clicked `atomId`.

**Medium term:**
- Update `atd trace` to return a structured dependency map (adjacency list) or a nested tree instead of flat slices in `GraphSlice`.
- Update `updateGraphPanel` to support recursive fetching or use the new structured trace data to render multiple ranks.
- Style nodes based on `layer` (color), `type` (shape), and `health` (border color/glow or icons).

**Long term:**
- Move graph rendering logic to a separate React/Vue component within the extension for better state management and interactivity.

---

## References

- [extension/extension.js](file:///home/bastien/work/atd/extension/extension.js)
- [vis-network documentation](https://visjs.github.io/vis-network/docs/network/)
