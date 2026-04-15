---
id: mechanic_vscode_hover_provider
status: DRAFT
human_name: "Spec-link Hover Provider"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0.0
dependents: []
priority: 3
parents:
  - [[service_vscode_linker_features]]
---

# New Atom

## INTENT
Implement a HoverProvider that shows atom details when hovering over @spec-link tags in source code.

## THE RULE / LOGIC
1. Match `@spec-link [[ID]]` at cursor position.
2. Parse local metadata for ID.
3. Fetch health metrics via `atd trace`.
4. Render vscode.MarkdownString with metadata, health icons, and intent.

## TECHNICAL INTERFACE

## EXPECTATION
Hovering over @spec-link displays a rich markdown summary of the atom and its health.
