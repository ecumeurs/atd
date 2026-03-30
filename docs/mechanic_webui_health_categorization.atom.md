---
id: mechanic_webui_health_categorization
human_name: "Health Categorization Algorithm"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 3
tags: [logic, classification, implementation]
parents:
  - [[rule_webui_health_classification]]
dependents: []
---

# Health Categorization Algorithm

## INTENT
Implement the 4-bucket logic for atom health categorization in the web explorer.

## THE RULE / LOGIC
- **Input:** Flat array of `parser.Atom` objects.
- **Process:**
    - Build a map of all atoms for ancestor lookup.
    - Check each atom:
        - `Done`: STABLE + Implement (+1 code tag) + Tests (+1 test tag).
        - `Almost Done`: STABLE + Implement + NO Tests.
        - `WIP`: STABLE + NO Implement + (Arch AND Customer Ancestors).
        - `Doc Jungle`: Fallback.
- **Output:** Object with four arrays of atoms.

## TECHNICAL INTERFACE (The Bridge)
- **Function:** `categorizeAtoms(flatData)` in `app.js`
- **Code Tag:** `@spec-link [[mechanic_webui_health_categorization]]`

## EXPECTATION (For Testing)
When the function is called with the project data, it must return precisely four buckets, and no atom should be missing or duplicated.
