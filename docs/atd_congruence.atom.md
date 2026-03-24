---
id: atd_congruence
human_name: "ATD Congruence"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, audit, congruence, consistency]
parents:
  - [[atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Congruence

## INTENT
To cross-examine a target atom against its related atoms (parents, dependents, tag-siblings) for logical contradictions in their INTENT and LOGIC sections before any code implementation.

## THE RULE / LOGIC
Loads all atoms from the docs directory, identifies the target atom's related context (parents via `[[id]]` links, dependents referencing the target, siblings sharing tags), builds a congruence audit prompt with only these relevant atoms, and routes through the tiered provider for task `congruence`. This targeted approach prevents context window bloat by excluding unrelated atoms.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd congruence -target <atom_id>`
- **LLM Task:** `congruence`
- **Code Tag:** `@spec-link [[atd_congruence]]`
