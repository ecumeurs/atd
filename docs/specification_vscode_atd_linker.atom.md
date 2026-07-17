---
id: specification_vscode_atd_linker
human_name: "VS Code ATD Linker"
type: SPECIFICATION
status: STABLE
priority: 5
version: 1.0.0
parents:
  - [[domain_atd_philosophy]]
dependents:
  - [[service_vscode_atd_ui]]
  - [[service_vscode_linker_features]]
layer: ARCHITECTURE
---

# VS Code ATD Linker

## INTENT
Provide seamless navigation and traceability within Atomic Traceable Documentation (ATD) enabled codebases directly from the VS Code editor.

## TECHNICAL INTERFACE (The Bridge)

The extension implements standard VS Code language features:

### Language Features
1.  **DocumentLinkProvider**: Recognizes `[[atomic_id]]` patterns and transforms them into clickable links resolving to `.atom.md` files.
2.  **DefinitionProvider**: Allows "Go to Definition" on `[[atomic_id]]` patterns.
3.  **HoverProvider**: When hovering over `@spec-link [[ID]]` in source code, displays atom metadata (layer, status, priority), intent, and live health metrics (implementation/test rates) fetched via `atd trace`.
4.  **CodeLensProvider**: In `.atom.md` files, displays a status bar at the top with ancestry completeness, implementation coverage, and test coverage metrics.

### Visualizations
1.  **ATD Graph Explorer**: A sidebar tree view (in the "ATD" activity bar container) that displays the local graph slice (parents and dependents) for the currently active atom file. It uses `atd trace` to fetch child health status.
2.  **ATD System Graph**: A webview panel showing the full system architecture graph, generated using `atd crawl` and rendered via Vis.js.

## THE RULE / LOGIC

1.  **Link Detection**: Matches the regex `/\[\[([^\]]+)\]\]/g`.
2.  **Path Resolution**: 
    - Loads `docs_path` from `.atd` JSON.
    - Resolves `[[ID]]` to `<workspaceRoot>/<docs_path>/ID.atom.md`.
3.  **Health Integration**: 
    - Invokes `atd trace <ID>` as a subprocess to retrieve the `HealthSnapshot` JSON.
    - Maps the snapshot to UI elements (Hover, CodeLens, Tree Items).
4.  **Graph Visualization**: 
    - Invokes `atd crawl` to get the full project graph.
    - Renders the graph in a Webview using a Vis.js network.
5.  **Activation**: The extension activates when the workspace contains a `.atd` file.

---

## EXPECTATION (For Testing)

- Developers can navigate between atoms and from source code to atoms without manual searching.
- Documentation becomes a live, navigable map of the system architecture.

---

## Code Mapping

- **Implementation:** [extension.js](file:///home/bastien/work/skill/extension/extension.js)
- **Code Tag:** `@spec-link [[specification_vscode_atd_linker]]`
