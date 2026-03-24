---
id: mechanic_index_chunking
human_name: "Index Chunking & Embedding Pipeline"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 3
tags: [atd, index, chunking, embedding, nomic]
parents:
  - [[service_atd_index]]
dependents: []
---

# Index Chunking & Embedding Pipeline

## INTENT
To describe the file discovery, content chunking, and embedding generation pipeline used by `atd index` to build the semantic vector index.

## THE RULE / LOGIC
1. **File Discovery:** Uses `git ls-files -c -o --exclude-standard` to enumerate tracked and untracked files while respecting `.gitignore`.
2. **Extension Filtering:** Files are included based on mode:
   - `code`: Only files matching `supported_extensions` from `.atd` config, excluding `.atom.md`.
   - `docs`: Only `.atom.md` files.
   - `all`: Both code and atom files.
3. **Mtime Caching:** Each file's `last_modified` timestamp is stored. On re-indexing, files with unchanged mtime are skipped entirely.
4. **Chunking Strategy:**
   - **Atom files:** Split by `## ` (H2 headers). Each section becomes one chunk, prefixed with `File: <path>`.
   - **Code files:** Split by `\n\n` (double newline). Chunks shorter than 20 characters are discarded. Each chunk is prefixed with `File: <path>`.
5. **Embedding:** Each chunk is embedded via `ollama.QueryEmbed()` using the model assigned to the `embed` task (typically `nomic-embed-text`). No IDE Agent fallback — a local or remote Ollama provider is required.
6. **Concurrency:** 10 worker goroutines consume chunk jobs from a buffered channel, embedding and inserting in parallel.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[mechanic_index_chunking]]`
- **Implementation:** `index.go:runIndex()` lines 92–216
- **Depends on:** `ollama.QueryEmbed()`, `config.ActiveConfig.SupportedExtensions`

## EXPECTATION (For Testing)
- Given a directory with 3 Go files and 2 `.atom.md` files, mode `all` should index all 5.
- Files unchanged since last indexing should be skipped (mtime check).
- Chunks shorter than 20 characters should not appear in the database.
