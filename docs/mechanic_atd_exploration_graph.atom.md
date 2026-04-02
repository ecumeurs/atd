---
id: mechanic_atd_exploration_graph
human_name: "Exploration Graph Logic"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 3
tags: [atd, exploration, graph, ast]
parents:
  - [[module_atd_exploration]]
dependents: []
---

# Exploration Graph Logic

## INTENT
To implement robust structs and walker functions capable of ingesting the disk files into an interoperable in-memory graph, mapped by Atom IDs.

## THE RULE / LOGIC
- Export a `DependencyGraph` structure containing an `Atoms` map linking IDs to their specific metadata and topology structure (`AtomNode`).
- **`CrawlDocs(dir)`:** Walk all `.atom.md` files handling broken structures safely but mapping as much parent and directional data as mathematically possible.
- **`CrawlSrc(dir)`:** Walk non-ignored source files, sniffing for `@spec-link` references via simple string split logic or regex, appending positive strikes to `AtomNode.Implementations`.
- **`WalkUp(id, visited...)` / `WalkDown(id, visited...)`:** Accept a starting ID and yield or map ancestors or dependents respectively, avoiding infinite repetition through an internal or passed `visited` set.

## TECHNICAL INTERFACE (The Bridge)
- **Functions:** `CrawlDocs`, `CrawlSrc`, `DependencyGraph.WalkUp`, `DependencyGraph.WalkDown`
- **Code Tag:** `@spec-link [[mechanic_atd_exploration_graph]]`

## EXPECTATION (For Testing)
When generating the graph, no cycles cause panic. A search on `@spec-link [[id]]` correctly appends the implementation file and line location precisely to the graph.
