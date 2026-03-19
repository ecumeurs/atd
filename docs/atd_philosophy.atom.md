---
id: atd_philosophy
human_name: "ATD Philosophy"
type: DOMAIN
version: 1.0
status: STABLE
priority: CORE
tags: [atd, philosophy, methodology]
parents: []
dependents: [[[atd_cli]]]]]]]]]
---

# ATD Philosophy

## INTENT
To define the foundational principles of Atomic Traceable Documentation (ATD) — a methodology where every system requirement, mechanic, and rule is expressed as a single-responsibility, version-controlled knowledge atom linked bidirectionally to its source code implementation and tests.

## THE RULE / LOGIC
ATD operates on five core principles:

1. **Minimum Atomic Scale**: Each atom describes exactly ONE state-changing rule. If an atom contains compound rules (e.g., "validates input AND applies transformation"), it must be split into separate atoms.

2. **Bidirectional Traceability**: Every atom links to its code via `@spec-link [[atom_id]]` tags embedded in source files, and every code module links back to its governing atom. This creates a verifiable chain from requirement → implementation → test.

3. **Fact-First Documentation**: The code is the ultimate source of truth. ATD atoms are extracted FROM implementations, not imposed on them. Domain-level documentation (READMEs, design docs) acts as a semantic overlay enriching the mechanical graph.

4. **LLM-Assisted, Human-Governed**: Local and remote LLMs handle bulk extraction, classification, and auditing tasks. The IDE Agent handles high-intelligence tasks (generation, reconciliation). Humans govern the final architecture.

5. **Token Economy**: Every LLM interaction is metered by task type. Cheap classification tasks (embedding, pass/fail) run locally. Expensive generation tasks (dissection, reconciliation) run on capable models. Deterministic tasks (weaving, updating, crawling) never touch an LLM.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[atd_philosophy]]`
- **Governed by:** The `.atd` configuration file at project root
