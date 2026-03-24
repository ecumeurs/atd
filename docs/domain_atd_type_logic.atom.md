---
id: domain_atd_type_logic
human_name: "Logic Atoms: RULE, MECHANIC, DOMAIN"
type: DOMAIN
version: 1.0
status: STABLE
priority: 5
tags: [atd, types, logic, rules]
parents:
  - [[domain_atd_structure]]
dependents: []
layer: CUSTOMER
---

# Logic Atoms: RULE, MECHANIC, DOMAIN

## INTENT
To define the usage and specificities of logic-level atoms.

## THE RULE / LOGIC
These types represent the "Meat" of the system:

- **`RULE`**: Strict constraints or boolean checks (e.g., "Max file size is 5MB", "User must be authenticated"). It defines "What" must be true.
- **`MECHANIC`**: Procedural logic and algorithms (e.g., "How the diff is calculated", "Minimax search protocol"). It defines "How" it works.
- **`DOMAIN`**: The high-level context, intent, and "The Why". It fills the gap between design and technical implementation.

### Specificity
- **Granularity**: Strict "One Rule" enforcement. Split if compound logic is found.
- **Linking**: `@spec-link` for a `RULE` is placed at the validation point. `MECHANIC` links go above the implementation function/class. `DOMAIN` links are often in documentation or orchestration files.

## TECHNICAL INTERFACE
- **Type Tag:** `type: RULE | MECHANIC | DOMAIN`.
- **Code Tag Placement:** Validation logic, algorithmic blocks.

## EXPECTATION
- `RULE` atoms should have specific `EXPECTATION` fields for binary pass/fail testing.
- `MECHANIC` atoms must detail the algorithm or procedure step-by-step.
