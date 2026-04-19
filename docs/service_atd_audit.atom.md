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
Operates in two modes focusing on internal ATD health:

**Mode 1 (Bloat Detection):**
Feeds each atom's INTENT and LOGIC sections to the LLM (task `audit_bloat`) for a binary YES/NO bloat assessment based on the atom type's strictness.

**Mode 2 (Collision Detection):**
Computes Nomic embeddings for all atoms and builds a pairwise cosine similarity matrix. For pairs exceeding the threshold, performs an ancestry BFS walk to determine if they represent a COLLISION (missing abstraction) or are structurally related.

Code-level compliance is handled by the [[service_atd_verify]] tool.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd audit [--code <snippet> --atom <path>] [--threshold <float>]`
- **LLM Tasks:** `audit_bloat`, `embed`, `audit_code`
- **Code Tag:** `@spec-link [[service_atd_audit]]`
