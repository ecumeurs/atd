---
id: service_atd_audit
human_name: "ATD Audit"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, audit, integrity]
parents:
  - [[module_atd_cli]]
dependents:
  - [[mechanic_atd_compare]]
  - [[mechanic_atd_fix]]
layer: IMPLEMENTATION
---

# ATD Audit

## INTENT
To systematically assess the structural integrity of the ATD ecosystem — detecting bloated atoms, semantic collisions, and missing abstractions — and optionally verifying code compliance against ATD rules.

## THE RULE / LOGIC
Operates in two modes:

**Default mode (Bloat + Collision):**
Phase 1 feeds each atom's INTENT and LOGIC sections to the LLM (task `audit_bloat`) for a binary YES/NO bloat assessment. Phase 2 computes Nomic embeddings for all atoms, builds a pairwise cosine similarity matrix, and for pairs exceeding `diff_similarity_threshold`, performs an ancestry BFS walk to determine if they share a parent (SOUND) or represent a missing abstraction (COLLISION).

**Code mode (`--code` flag):**
Feeds an atom rule and a code snippet to the LLM (task `audit_code`) for a JSON pass/fail compliance verdict.

Results are cached in SQLite keyed by file mtime to avoid redundant LLM queries.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd audit [--code <snippet> --atom <path>] [--threshold <float>]`
- **LLM Tasks:** `audit_bloat`, `embed`, `audit_code`
- **Code Tag:** `@spec-link [[service_atd_audit]]`
