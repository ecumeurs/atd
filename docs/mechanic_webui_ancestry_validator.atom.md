---
id: mechanic_webui_ancestry_validator
status: DRAFT
human_name: Ancestry Validator Mechanic
layer: IMPLEMENTATION
version: 1.0
priority: 3
parents: [[api_webui_health_stats]]
dependents: []
type: MECHANIC
tags: [webui, mechanic, validation]
---

# Ancestry Validator Mechanic

## INTENT
Mathematically verify the "Rootedness" of the documentation graph.

## THE RULE / LOGIC
For any given atom:
1. Recursively traverse the `parents` field.
2. If a `DOMAIN` or `REQUIREMENT` type is found, return `true` (Rooted).
3. If the recursion depth exceeds 50 or the stack empties without finding a root, return `false` (Floating).
4. **Exception**: DOMAIN and REQUIREMENT types are allowed to be roots without parents.

## TECHNICAL INTERFACE (The Bridge)
- **Go Function**: `parser.IsRooted(atom)`
- **Code Tag**: `@spec-link [[mechanic_webui_ancestry_validator]]`
- **Test Names**: `TestAncestryValidator`

## EXPECTATION (For Testing)
The validator correctly identifies atoms that do not link back to the Customer layer.
