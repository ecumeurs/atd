---
id: atd_recon
human_name: "ATD Recon"
type: MECHANIC
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, recon, validation, tagging]
parents:
  - [[atd_cli]]
dependents: []
---

# ATD Recon

## INTENT
To validate whether a specific code file implements the rules defined in a given ATD atom, returning a confidence score and listing any mismatches.

## THE RULE / LOGIC
Reads the atom file (extracting INTENT + LOGIC) and the candidate source code file, builds a recon prompt asking the LLM (task `recon`) to compare the code against the specification. Returns JSON `{Confidence: 0-100, Mismatches: "description of gaps"}`. Used during the archaeology phase to surgically inject `@spec-link` tags.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd recon -atom <atom.md> -candidate <code_file>`
- **LLM Task:** `recon`
- **Code Tag:** `@spec-link [[atd_recon]]`
