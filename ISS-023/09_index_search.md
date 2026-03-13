# Task 09 — Index & Search

**Depends on:** Task 03 (ollama client), Task 04 (provider for embed task), Task 03 (cosine)  
**Produces:** `atd index` and `atd search` subcommands

## Context

`atd-ollama-indexer` (232 lines) chunks code files and stores Nomic embeddings in SQLite. `atd-ollama-search` (167 lines) embeds a query and cosine-ranks results. `atd-tag-sweep` (58 lines) does keyword search. These merge into two commands with modes.

## Steps

### 9.1 Create `cmd/index.go`

**Flags:** `--dir <path>`, `--db <path>` (default: `<docsDir>/.atd_index.db`), `--mode code|docs|all`

**Execution flow:**

```
Resolve provider for task "embed" (nomic-embed-text)
├── Must succeed — embed has NO IDE fallback
│
Open/create SQLite DB with schema:
  CREATE TABLE IF NOT EXISTS code_index (
    id TEXT, file_path TEXT, chunk_text TEXT, embedding BLOB,
    last_modified INTEGER, PRIMARY KEY (id)
  );
│
Git ls-files (respect .gitignore) in --dir
│
For each file matching extensions (or *.atom.md if mode=docs/all):
├── Check mtime vs DB → skip if unchanged
├── Chunk content:
│   ├── Code: split by func/struct boundaries or fixed-line chunks (50 lines)
│   └── Docs: split by ## sections (INTENT, LOGIC as separate chunks)
├── Embed each chunk via ollama.QueryEmbed()
└── Upsert into SQLite
│
Report: "Indexed N chunks across M files (S skipped unchanged)"
```

**Source reference:** `atd-ollama-indexer/main.go` — full logic. Keep the `sync.WaitGroup` concurrent embedding from the current implementation.

**Help text:**
```
Build a semantic vector index of source code and/or ATD documents.

Uses nomic-embed-text to generate embeddings stored in SQLite.
Files unchanged since last indexing are automatically skipped.

Examples:
  atd index --dir ./src                    # Index code only
  atd index --mode docs                    # Index ATD atoms only
  atd index --dir ./src --mode all         # Index both
  atd index --dir ./src --db ./custom.db   # Custom DB path
```

### 9.2 Create `cmd/search.go`

**Flags:** `--query <text>`, `--db <path>`, `--limit <N>` (default 5), `--grep <keyword>`, `--scope code|docs|all`

**Execution flow:**

```
If --grep is set:
│ Walk --dir (or config.ProjectRoot()), string match keyword
│ Output: each match as "file:line: content"
│ (Absorbed from atd-tag-sweep)
│
Else (semantic search):
│ Resolve provider for "embed"
│ Embed the query via ollama.QueryEmbed()
│ Load all embeddings from DB
│ Compute cosine similarity for each
│ Sort descending, take top --limit
│ Output:
│   [Match 1] Score: 0.89 File: path/to/file.go Lines: 45-95
│   ...
```

**Source reference:** `atd-ollama-search/main.go` — full logic. `atd-tag-sweep/main.go` for grep mode.

### 9.3 Write tests

**`cmd/index_test.go`:**
- Test: `--dir` with nonexistent dir → error
- Test: `--mode` validation (only code|docs|all accepted)
- Test: SQLite schema creation (in-memory DB)
- Test: mtime skip logic with mock data

**`cmd/search_test.go`:**
- Test: `--grep` mode finds keyword in a temp file
- Test: `--query` without `--db` → error
- Test: Result formatting includes score and file path

## Acceptance Criteria

- [ ] `atd index --help` and `atd search --help` show all flags
- [ ] `atd index --dir ../upsilonbattle/` creates `.atd_index.db` (requires nomic-embed-text)
- [ ] `atd index --dir ../upsilonbattle/` second run skips unchanged files
- [ ] `atd search --query "entity movement" --db <path>` returns ranked results
- [ ] `atd search --grep "Turner"` returns file:line matches
- [ ] Tests pass
