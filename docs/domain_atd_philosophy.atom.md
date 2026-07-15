---
id: domain_atd_philosophy
human_name: "ATD Philosophy"
type: DOMAIN
version: 1.0
status: STABLE
priority: 5
tags: [atd, philosophy, methodology]
parents: []
dependents:
  - [[domain_atd_structure]]
  - [[domain_atd_usage_protocol]]
  - [[module_atd_cli]]
  - [[requirement_webui_platform]]
layer: BUSINESS
---

# ATD Philosophy

## INTENT
To define the foundational principles of Atomic Traceable Documentation (ATD) — a methodology where every system requirement, mechanic, and rule is expressed as a single-responsibility, version-controlled knowledge atom linked bidirectionally to its source code implementation and tests.

## THE RULE / LOGIC
ATD operates on five core principles:

1. **Minimum Atomic Scale**: Each atom describes exactly ONE state-changing rule. If an atom contains compound rules (e.g., "validates input AND applies transformation"), it must be split into separate atoms.

2. **Bidirectional Traceability**: Every atom links to its code via `@spec-link [[atom_id]]` tags and to tests via `@test-link [[atom_id]]` tags embedded in source files. This creates a verifiable chain from Customer requirement → Architecture → Implementation → Test.

3. **Doc-Code Co-evolution**: During cold-start (bootstrapping an undocumented codebase), atoms are extracted FROM existing implementations — the code is the initial source of truth. Once the initial ATD base is established, documentation and code evolve together: new features begin as DRAFT atoms (requirements, specs, design) before implementation, and implementation feeds back into atom refinement. Neither side is subordinate; they are kept in sync through the verification loop.

4. **LLM-Assisted, Human-Governed**: Local and remote LLMs handle bulk extraction, classification, and auditing tasks. The IDE Agent handles high-intelligence tasks (generation, reconciliation). Humans govern the final architecture.

5. **Token Economy**: Every LLM interaction is metered by task type. Cheap classification tasks (embedding, pass/fail) run locally. Expensive generation tasks (dissection, reconciliation) run on capable models. Deterministic tasks (weaving, updating, crawling) never touch an LLM.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[domain_atd_philosophy]]`
- **Governed by:** The `.atd` configuration file at project root
