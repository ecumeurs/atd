---
id: mechanic_vscode_link_provider
status: DRAFT
human_name: "Link and Definition Resolver"
type: MECHANIC
dependents: []
layer: IMPLEMENTATION
version: 1.0.0
priority: 3
parents: [[service_vscode_linker_features]]
---

# New Atom

## INTENT
Implement DocumentLinkProvider and DefinitionProvider to resolve [[ID]] tokens to atom files.

## THE RULE / LOGIC
1. Use regex `/\[\[([^\]]+)\]\]/g` to find tokens.
2. Resolve ID to file path: `<workspaceRoot>/<docs_path>/<ID>.atom.md`.
3. Map matches to vscode.DocumentLink or vscode.Location objects.

## TECHNICAL INTERFACE

## EXPECTATION
Clicking on [[ID]] correctly navigates to the corresponding .atom.md file.
