---
id: mechanic_search_semantic
human_name: "Semantic Search Mechanism"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 3
tags: [atd, search, semantic, cosine, embedding]
parents:
  - [[service_atd_search]]
dependents: []
---

# Semantic Search Mechanism

## INTENT
To describe how `atd search` performs semantic similarity search against the vector index.

## THE RULE / LOGIC
1. **Query Embedding:** The user's query string is embedded via `ollama.QueryEmbed()` using the `embed` task model (nomic-embed-text). No IDE Agent fallback — Ollama must be available.
2. **Database Load:** All rows from `atom_index` are loaded from the SQLite database.
3. **Scope Filtering:** Rows are filtered by `scope` parameter:
   - `code`: Exclude rows where `file_path` ends with `.atom.md`.
   - `docs`: Include only rows where `file_path` ends with `.atom.md`.
   - `all`: No filtering.
4. **Cosine Similarity:** Each chunk's stored embedding is compared to the query embedding using `cosine.Similarity()`. This produces a score between 0.0 (no similarity) and 1.0 (identical).
5. **Ranking:** Results are sorted by descending similarity score.
6. **Top-N:** Only the top `limit` results are kept (default: 5).
7. **Result Formatting:**
   - Default: Returns chunks with metadata and similarity scores.
   - Paths Only: Returns unique absolute file paths of the matching files, one per line.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[mechanic_search_semantic]]`
- **Implementation:** `search.go:runSemanticSearch()`
- **Depends on:** `cosine.Similarity()`, `ollama.QueryEmbed()`, `atom_index` table

## EXPECTATION (For Testing)
- A query semantically similar to a documented atom should return that atom's chunks with high similarity scores (>0.7), or its absolute file path if `paths_only` is true.
- `scope=code` must never return `.atom.md` file results.
- `scope=docs` must only return `.atom.md` file results.
- Results must be sorted by descending similarity.
- When `paths_only` is true, the response must contain only unique absolute file paths.
