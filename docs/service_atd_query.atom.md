---
id: service_atd_query
human_name: "ATD Query"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, query, frontmatter]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Query

## INTENT
To search ATD atoms by frontmatter field values, enabling quick filtering by id, type, status, priority, or tags.

## THE RULE / LOGIC
- **Search Capability**: Extracts YAML frontmatter from each `.atom.md` file and matches the specified field against the search term (case-insensitive substring or regex).
- **Paths Only Mode**: If `--paths-only` or `-p` is provided, the command returns only a JSON array of strings (the absolute file paths) instead of full atom objects.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd query -field <field_name> -search <term> [-paths-only]`
- **LLM Task:** None (deterministic)
- **Code Tag:** `@spec-link [[service_atd_query]]`
