---
id: mechanic_index_schema
human_name: "Index Database Schema"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 3
tags: [atd, index, sqlite, schema, database]
parents:
  - [[service_atd_index]]
dependents: []
---

# Index Database Schema

## INTENT
To define the SQLite database schema used by the ATD semantic vector index.

## THE RULE / LOGIC
The index is stored in a single SQLite database file (`.atd_index.db`) located in the docs directory. It contains one table:

```sql
CREATE TABLE IF NOT EXISTS atom_index (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT,
    chunk_text TEXT,
    embedding BLOB,
    last_modified INTEGER
);
```

| Column | Type | Purpose |
|---|---|---|
| `id` | INTEGER PK | Auto-incrementing row identifier. |
| `file_path` | TEXT | Absolute path to the source file this chunk came from. |
| `chunk_text` | TEXT | The raw text content of the chunk (prefixed with `File: <path>`). |
| `embedding` | BLOB | JSON-serialized `[]float64` vector from nomic-embed-text. |
| `last_modified` | INTEGER | Unix timestamp of the file's mtime at indexing time (used for cache invalidation). |

**Cache invalidation:** When re-indexing, all existing chunks for a file are `DELETE`d before inserting new ones (`DELETE FROM atom_index WHERE file_path = ?`).

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[mechanic_index_schema]]`
- **Database file:** `<docs_path>/.atd_index.db`
- **Implementation:** `index.go:runIndex()` lines 59–76

## EXPECTATION (For Testing)
- The `atom_index` table must exist after `atd index` runs.
- Each row must have a non-null `embedding` BLOB that deserializes to a `[]float64`.
- `last_modified` must match the file's mtime at indexing time.
