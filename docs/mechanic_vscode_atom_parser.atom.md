---
id: mechanic_vscode_atom_parser
status: DRAFT
human_name: "Lighweight Atom Parser"
type: MECHANIC
dependents: []
layer: IMPLEMENTATION
version: 1.0.0
priority: 3
parents: [[service_vscode_linker_features],[service_vscode_atd_ui]]
---

# New Atom

## INTENT
Provide a lightweight parser for extracting ATD metadata (layer, status, intent) from .atom.md files.

## THE RULE / LOGIC
1. Use regex to extract YAML frontmatter fields (layer, status, priority).
2. Extract ## INTENT section content.
3. Extract `parents` and `dependents` lists by matching `[[ID]]` within respective sections.

## TECHNICAL INTERFACE

## EXPECTATION
Parser returns accurate metadata and link lists from .atom.md files.
