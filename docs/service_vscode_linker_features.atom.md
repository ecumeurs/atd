---
id: service_vscode_linker_features
status: DRAFT
type: SERVICE
version: 1.0.0
priority: 3
parents: [[specification_vscode_atd_linker]]
human_name: "VS Code Linker Language Features"
layer: ARCHITECTURE
dependents:
  - [[mechanic_vscode_atd_config]]
  - [[mechanic_vscode_atom_parser]]
  - [[mechanic_vscode_codelens_provider]]
  - [[mechanic_vscode_hover_provider]]
  - [[mechanic_vscode_link_provider]]
---

# New Atom

## INTENT
Orchestrate VS Code language features (links, definitions, hovers, codelenses) for ATD integration.

## THE RULE / LOGIC
1. Register Link and Definition providers for all files to resolve `[[ID]]`.
2. Register CodeLens provider for markdown files to show health at line 0.
3. Register Hover provider for all files to show meta/health for `@spec-link [[ID]]`.

## TECHNICAL INTERFACE

## EXPECTATION
VS Code language features correctly resolve ATD links, show definitions, hovers with health data, and codelenses in .atom.md files.
