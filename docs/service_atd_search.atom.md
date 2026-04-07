---
id: service_atd_search
human_name: "ATD Search"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, search, semantic, grep]
parents:
  - [[service_atd_index]]
dependents:
  - [[mechanic_search_grep]]
  - [[mechanic_search_semantic]]
layer: IMPLEMENTATION
---

# ATD Search

## INTENT
To find relevant code or documentation by semantic meaning (Nomic embedding + cosine similarity) or by keyword (grep mode), returning ranked results from the vector index.

## THE RULE / LOGIC
In semantic mode: embeds the query via Nomic, loads all stored embeddings from the SQLite index, computes cosine similarity for each chunk, and returns the top N results sorted by score. In grep mode: walks the project directory doing string matching, returning file matches. Supports `--scope code|docs|all` to filter results by type within the single index database. If `--paths-only` or `-p` is provided, the command returns only unique absolute file paths of the matching files, one per line.

## TECHNICAL INTERFACE (The Bridge)
- **MCP Tool:** `atd_search` with params: `query`, `grep`, `scope`, `limit`, `paths_only`
- **CLI Command:** `atd search --query <text> [--limit N] [--grep <keyword>] [--scope code|docs|all] [--paths-only]`
- **LLM Task:** `embed` (semantic mode only)
- **Code Tag:** `@spec-link [[service_atd_search]]`

## EXPECTATION (For Testing)
- Semantic search should return results ranked by cosine similarity, or unique absolute paths if `--paths-only` is provided.
- Grep search should find literal string matches across project files, or unique absolute paths if `--paths-only` is provided.
- Scope filtering should correctly exclude/include results by file type.
