---
id: atd_crawl
human_name: "ATD Crawl"
type: SERVICE
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, crawl, graph, dependencies]
parents:
  - [[atd_cli]]
dependents: []
---

# ATD Crawl

## INTENT
To build a complete dependency graph from ATD docs and `@spec-link` tags in source code, and optionally report orphaned STABLE atoms that lack implementations.

## THE RULE / LOGIC
Walks the docs directory parsing `.atom.md` frontmatter (id, status, parents, dependents). Simultaneously walks the source directory for `@spec-link` references, mapping them as implementations. Outputs a JSON dependency graph. With `--gaps`, filters for STABLE atoms where `source_implementations` is empty — indicating specifications that exist without corresponding code.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd crawl [--src <path>] [--gaps]`
- **LLM Task:** None (deterministic)
- **Code Tag:** `@spec-link [[atd_crawl]]`
