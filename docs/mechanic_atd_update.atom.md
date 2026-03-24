---
id: mechanic_atd_update
human_name: "ATD Update"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, update, frontmatter]
parents:
  - [[module_atd_cli]]
dependents: []
layer: CUSTOMER
---

# ATD Update

## INTENT
To programmatically modify ATD atom files — updating YAML frontmatter fields, body sections, and injecting `@spec-link` tags into source code.

## THE RULE / LOGIC
Reads an atom file (or multiple atom files matched via `--filter`), applies field-level updates to the YAML frontmatter (`-set key=value`), replaces body sections (`-intent`, `-logic`, `-interface`), and optionally propagates status changes to linked atoms. The `--spec-link` mode (absorbed from atd-legacy-wrapper) prepends `// @spec-link [[id]]` to a target source file without touching the atom.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd update [--file <atom.md> | --filter <query>] [-set key=value] [-intent <text>] [-logic <text>] [--spec-link <id> <source_file>]`
- **LLM Task:** None (deterministic)
- **Code Tag:** `@spec-link [[mechanic_atd_update]]`
