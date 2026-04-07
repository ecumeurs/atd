---
id: mechanic_atd_weave
human_name: "ATD Weave"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, weave, links, dependencies]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Weave

## INTENT
To propagate bidirectional parent↔dependent links across all ATD atoms, ensuring that every atom listing a parent is also listed as a dependent of that parent.

## THE RULE / LOGIC
Reads all `.atom.md` files, builds a map of parent→dependent relationships from the `parents:` frontmatter field, then rewrites each atom's `dependents:` field to include all atoms that declare it as a parent. Only modifies files where the dependents list actually changes.

## TECHNICAL INTERFACE (The Bridge)
@spec-link [[mechanic_atd_weave]]
Code location: `scripts/pkg/atom/weave.go`
