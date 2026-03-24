---
id: specification_vscode_atd_linker
type: SPECIFICATION
status: STABLE
priority: 5
version: 1.0.0
parents: [atd_philosophy]
dependents: []
layer: CUSTOMER
---

# VS Code ATD Linker

**Intent:** 
Provide seamless navigation and traceability within Atomic Traceable Documentation (ATD) enabled codebases directly from the VS Code editor.

---

## Technical Interface

The extension implements standard VS Code language features:

1.  **DocumentLinkProvider**: Recognizes `[[atomic_id]]` patterns in any document and transforms them into clickable links (Ctrl+Click) that resolve to the corresponding `.atom.md` file in the project's documentation directory.
2.  **DefinitionProvider**: Allows users to "Go to Definition" or "Peek Definition" on `[[atomic_id]]` patterns, jumping directly to the source atom document.

### Configuration
The extension reads the `.atd` configuration file at the workspace root to determine the `docs_path`. It watches this file for changes to ensure the link resolution path is always up-to-date.

---

## The Rule / Logic

1.  **Link Detection**: Matches the regex `/\[\[([^\]]+)\]\]/g`.
2.  **Path Resolution**: 
    - Loads `docs_path` from `.atd` JSON.
    - Resolves `[[ID]]` to `<workspaceRoot>/<docs_path>/ID.atom.md`.
3.  **Activation**: The extension activates when the workspace contains a `.atd` file.

---

## Expectation

- Developers can navigate between atoms and from source code to atoms without manual searching.
- Documentation becomes a live, navigable map of the system architecture.

---

## Code Mapping

- **Implementation:** [extension.js](file:///home/bastien/work/skill/extension/extension.js)
- **Code Tag:** `@spec-link [[specification_vscode_atd_linker]]`
