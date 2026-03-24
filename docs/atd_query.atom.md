---
id: atd_query
human_name: "ATD Query"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, query, frontmatter]
parents:
  - [[atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Query

## INTENT
To search ATD atoms by frontmatter field values, enabling quick filtering by id, type, status, priority, or tags.

## THE RULE / LOGIC
Walks the docs directory, extracts YAML frontmatter from each `.atom.md` file, and matches the specified field against the search term (case-insensitive substring or regex). Returns matching file paths as a JSON array, suitable for piping into other tools.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd query -field <field_name> -search <term>`
- **LLM Task:** None (deterministic)
- **Code Tag:** `@spec-link [[atd_query]]`
