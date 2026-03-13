---
id: atd_search
human_name: "ATD Search"
type: SERVICE
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, search, semantic, grep]
parents:
  - [[atd_index]]
dependents: []
---

# ATD Search

## INTENT
To find relevant code or documentation by semantic meaning (Nomic embedding + cosine similarity) or by keyword (grep mode), returning ranked results from the vector index.

## THE RULE / LOGIC
In semantic mode: embeds the query via Nomic, loads all stored embeddings from the SQLite index, computes cosine similarity for each chunk, and returns the top N results sorted by score. In grep mode (absorbed from atd-tag-sweep): walks the project directory doing string matching, returning file:line matches. Supports `--scope code|docs|all` to filter by index type.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd search --query <text> [--db <path>] [--limit N] [--grep <keyword>] [--scope code|docs|all]`
- **LLM Task:** `embed` (semantic mode only)
- **Code Tag:** `@spec-link [[atd_search]]`
