---
id: mechanic_vscode_codelens_provider
status: DRAFT
layer: IMPLEMENTATION
priority: 3
parents:
  - [[service_vscode_linker_features]]
human_name: "Health CodeLens Provider"
type: MECHANIC
version: 1.0.0
dependents: []
---

# New Atom

## INTENT
Implement a CodeLensProvider that displays live health metrics at the top of .atom.md files using atd trace.

## THE RULE / LOGIC
1. Watch markdown files ending in `.atom.md`.
2. Extract ID from frontmatter.
3. Execute `atd trace <ID>` as child process.
4. Format `HealthSnapshot` into a readable CodeLens string at line 0.

## TECHNICAL INTERFACE

## EXPECTATION
Markdown files display a CodeLens with accurate implementation and test coverage percentages.
