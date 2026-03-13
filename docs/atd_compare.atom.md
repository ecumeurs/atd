---
id: atd_compare
human_name: "ATD Compare"
type: MECHANIC
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, audit, compare, collision]
parents:
  - [[atd_audit]]
dependents: []
---

# ATD Compare

## INTENT
To resolve semantic collisions between two atoms by analyzing their shared responsibilities and proposing MERGE, REFACTOR, or PARENT extraction via LLM.

## THE RULE / LOGIC
Parses both atom files, extracts shared keywords through deterministic set intersection, then builds a comparison prompt feeding both atoms' INTENT, LOGIC, and tags to the LLM (task `compare`). The LLM returns a plaintext diagnostic recommending one of: MERGE (combine into one), REFACTOR (extract shared logic into new parent), or KEEP (false positive). Output is a markdown report.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd compare -a <atom_a> -b <atom_b> [--out <report.md>]`
- **LLM Task:** `compare`
- **Code Tag:** `@spec-link [[atd_compare]]`
