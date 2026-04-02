---
id: module_atd_exploration
human_name: "ATD Graph Exploration"
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: DRAFT
priority: 3
tags: [atd, exploration, ast, graph, dependency]
parents:
  - [[module_atd_cli]]
dependents:
  - [[mechanic_atd_exploration_graph]]
---

# ATD Graph Exploration

## INTENT
To centralize and provide a unified, predictable method for traversing, mapping, and reading the ATD atom dependency graph and corresponding source implementations.

## THE RULE / LOGIC
The CLI tools must not implement their own filepath walking or custom parsing for the graph. Instead, they must rely on the `pkg/exploration` module. This module guarantees consistent generation of the bidirectional `.atom.md` matrix, handling circular dependencies, validating formats gracefully (or creating stubs for malformed elements), and establishing implementations via `@spec-link` references. 

Future upgrades (like strict `.gitignore` parsing) apply globally to the ecosystem when implemented here.

## TECHNICAL INTERFACE (The Bridge)
- **Library Package:** `scripts/pkg/exploration`
- **Exposed Types:** `DependencyGraph`, `AtomNode`
- **Code Tag:** `@spec-link [[module_atd_exploration]]`

## EXPECTATION (For Testing)
When generating an exploration graph, the engine correctly assembles all nodes, attributes parents and dependents bidirectionally, links specified source implementations accurately across the given search path, and returns deterministic sizes.
