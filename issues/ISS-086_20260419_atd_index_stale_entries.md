# Issue: ATD Index Stale Entries and Chunking Failures

**ID:** `20260419_atd_index_stale_entries`
**Ref:** `ISS-086`
**Date:** 2026-04-19
**Severity:** Medium
**Status:** Resolved
**Component:** `atd/cmd/atd/cmd/index.go`
**Affects:** `atd index` command, semantic search accuracy

---

## Summary

The ATD embedding index (`.atd_index.db`) suffers from two major issues:
1. It never removes entries for files that have been deleted from the disk, leading to database bloat and false positive search results.
2. The chunking strategy for code files (double-newline) can produce chunks that exceed the context length of embedding models, causing indexing failures on large or poorly formatted files.

---

## Technical Description

### Background
The `atd index` command crawls the project, splits files into chunks, and stores their embeddings in a SQLite database. It uses modification times to skip unchanged files.

### The Problem Scenario

#### 1. Stale Entries (Index Bloat)
In `runIndex` within `atd/cmd/atd/cmd/index.go`, the tool iterates over current files on disk. It deletes existing entries for a file only if that file is about to be re-indexed (i.e., it changed since the last run).
However, if a file is deleted from the project, it is no longer returned by the explorer, and thus the tool never encounters it to perform a removal. This results in the database knowing about files that no longer exist (e.g., 610 files in index vs 250 on disk).

#### 2. Excessive Chunk Length
Embedding models (like `nomic-embed-text`) have a fixed context window (e.g., 512 or 2048 tokens). The current code-splitting logic in `index.go` splits by `\n\n`. If a block of code (e.g., a large configuration array or a generated file like `battleui/bootstrap/cache/services.php`) is very long without double-newlines, or if the block itself is simply too large, the embedding API returns a 500 error: `the input length exceeds the context length`.

### Where This Pattern Exists Today
- **Stale Entries**: `atd/cmd/atd/cmd/index.go` around lines 78-90 (fetching existing paths) and 150-222 (iterating over current files).
- **Chunking**: `atd/cmd/atd/cmd/index.go` around lines 207-216.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium (Stale results, failed indexing for some files) |
| Detectability | Medium — manifested by index count > file count and explicit warning messages. |
| Current mitigant | None. Users must manually delete the `.atd_index.db` to refresh completely. |

---

## Recommended Fix

**Short term:** 
- Add a "cleanup" loop after the main indexing loop that finds all `file_path` entries in the DB that were not seen during the current crawl and deletes them.
- Add a check for chunk length and possibly skip or further subdivide chunks that are too large.

**Medium term:** 
- Implement a more robust chunking strategy (e.g., recursive character splitting or token-aware splitting).
- Improve error handling for embedding failures to log the exact line range that failed.

**Long term:** 
- Use a dedicated vector database or a more managed indexing service if the project scale grows significantly.

---

## References

- [atd/cmd/atd/cmd/index.go](file:///home/bastien/work/skill/atd/cmd/atd/cmd/index.go)
- [docs/service_atd_index.atom.md](file:///home/bastien/work/skill/docs/service_atd_index.atom.md)
