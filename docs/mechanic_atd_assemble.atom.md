---
id: mechanic_atd_assemble
human_name: "ATD Assemble"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, assemble, snapshot]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Assemble

## INTENT
To recursively stitch atoms together following their dependent chains, producing a unified document for human review or an LLM-ready narrative.

## THE RULE / LOGIC
Starting from one or more root atom IDs, recursively gathers content through the dependents graph (preventing cycles via visited set). In default mode, outputs raw assembled fragments. With `--purpose`, prepends a system prompt instructing an LLM to rewrite the fragments into flowing prose. With `--snapshot`, generates a themed narrative rewrite prompt (task `snapshot`) from the assembled content — absorbing the former atd-generate-snapshot functionality.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd assemble --starts <ids> [--purpose <text>] [--snapshot --theme <text>]`
- **LLM Task:** `snapshot` (only with --snapshot flag)
- **Code Tag:** `@spec-link [[mechanic_atd_assemble]]`
