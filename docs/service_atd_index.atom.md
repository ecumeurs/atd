---
id: service_atd_index
human_name: "ATD Index"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, index, nomic, embedding]
parents:
  - [[module_atd_cli]]
dependents:
  - [[mechanic_index_chunking]]
  - [[mechanic_index_schema]]
  - [[service_atd_search]]
layer: IMPLEMENTATION
---

# ATD Index

## INTENT
To build and maintain a semantic vector index of source code and ATD documents using Nomic embeddings stored in SQLite, enabling fast semantic search across the project.

## THE RULE / LOGIC
Crawls files (respecting `.gitignore` via `git ls-files`), chunks content (code by double-newline blocks; docs by `##` sections), and generates Nomic embeddings for each chunk. Uses mtime-based caching to skip unchanged files on re-indexing. Indexes both code and documentation into a single `.atd_index.db` database. The `embed` task has NO IDE fallback — Nomic must be available or the command fails.

## TECHNICAL INTERFACE (The Bridge)
- **MCP Tool:** `atd_index` (zero parameters — indexes entire project automatically)
- **CLI Command:** `atd index [--dir <path>] [--db <path>] [--mode code|docs|all]`
- **LLM Task:** `embed` (nomic-embed-text only)
- **Code Tag:** `@spec-link [[service_atd_index]]`

## EXPECTATION (For Testing)
- Running `atd index` should create/update `.atd_index.db` in the docs directory.
- Unchanged files should be skipped on re-indexing.
- The tool should fail with a clear error if no embedding model is available.
