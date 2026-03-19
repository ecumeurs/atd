---
id: atd_index
human_name: "ATD Index"
type: SERVICE
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, index, nomic, embedding]
parents:
  - [[atd_cli]]
dependents: [[[atd_search]]]]]]]]]
---

# ATD Index

## INTENT
To build and maintain a semantic vector index of source code and/or ATD documents using Nomic embeddings stored in SQLite, enabling fast semantic search across the project.

## THE RULE / LOGIC
Crawls files (respecting `.gitignore` via `git ls-files`), chunks content (code by function boundaries or 50-line blocks; docs by `##` sections), and generates Nomic embeddings for each chunk. Uses mtime-based caching to skip unchanged files on re-indexing. Supports three modes: `code` (source files matching `supported_extensions`), `docs` (`.atom.md` files), and `all` (both). The `embed` task has NO IDE fallback — Nomic must be available or the command fails.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd index --dir <path> [--db <path>] [--mode code|docs|all]`
- **LLM Task:** `embed` (nomic-embed-text only)
- **Code Tag:** `@spec-link [[atd_index]]`
