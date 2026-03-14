# Issue: Add a tool to handle atds indexing and research through atds (nomic)

**ID:** `20260313_atd_indexing_nomic`
**Ref:** `ISS-024`
**Date:** 2026-03-13
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/atd-ollama-indexer`, `scripts/atd-ollama-search`
**Affects:** `atd-audit`, and general ATD exploration workflows.

---

## Summary

While the project has `atd-ollama-indexer` and `atd-ollama-search` using Nomic embeddings, these are currently designed for code parsing and general search. There is a need for a unified tool specifically designed to index and research *through ATDs* (Atomic Technical Documents). This tool should allow for semantic navigation, gap detection, and research across the ATD corpus, and should be unified with other tools like `atd-audit` to provide a cohesive auditing experience.

---

## Technical Description

### Background
Currently, `atd-ollama-indexer` crawls a directory and generates embeddings for code chunks using the `nomic-embed-text` model. Search is performed by calculating cosine similarity between a query embedding and the stored chunk embeddings.

### The Problem Scenario
1. A user wants to find all ATDs related to "concurrency" across the entire corpus.
2. The current search tools might return code blocks instead of ATD summaries or structural information.
3. `atd-audit` does not currently leverage semantic search to find related ATDs when identifying gaps or inconsistencies.

### Where This Pattern Exists Today
- `scripts/atd-ollama-indexer/main.go`
- `scripts/atd-ollama-search/main.go`
- `scripts/atd-audit/main.go` (needs integration)

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | High |
| Detectability | Medium — manifests as difficulty in navigating large ATD repositories |
| Current mitigant | Manual grep and basic semantic search on code. |

---

## Recommended Fix

**Short term:** Enhance `atd-ollama-indexer` to specifically prioritize ATD files and metadata.

**Medium term:** Integrate semantic search capabilities into `atd-audit` to allow "research" mode where the auditor can query the ATD corpus.

**Long term:** Unify indexing and search into the core `atd` binary and provide a high-level API for other tools to perform semantic research on ATDs.

---

## References

- [scripts/atd-ollama-indexer/main.go](file:///home/bastien/work/skill/scripts/atd-ollama-indexer/main.go)
- [scripts/atd-ollama-search/main.go](file:///home/bastien/work/skill/scripts/atd-ollama-search/main.go)
- [scripts/atd-audit/](file:///home/bastien/work/skill/scripts/atd-audit/)
