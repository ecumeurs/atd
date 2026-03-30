---
id: rule_webui_health_classification
human_name: "Health Classification Rules"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [logic, classification, health]
parents:
  - [[domain_atd_structure]]
dependents: [[[mechanic_webui_health_categorization]]]
---

# Health Classification Rules

## INTENT
Define the four bucket criteria for categorizing ATD atoms within the Explorer view.

## THE RULE / LOGIC
Atoms are bucketed into four distinct health categories:
1. **Done:**
    - `status=STABLE`
    - `len(linked_codes) > 0` (implementation exists)
    - `has_tests=true` (tests exist)
2. **Almost Done:**
    - `status=STABLE`
    - `len(linked_codes) > 0`
    - `has_tests=false`
3. **WIP (Work In Progress):**
    - `status=STABLE`
    - `len(linked_codes) == 0` (no implementation)
    - **Ancestry Requirement:** Must have at least one **Architecture** AND one **Customer** ancestor in its `parents` array.
4. **Doc Jungle:**
    - Everything else (orphans, drafts, missing parents, unstable documents).

## TECHNICAL INTERFACE (The Bridge)
- **JS Function:** `categorizeAtoms(atoms)`
- **Code Tag:** `@spec-link [[rule_webui_health_classification]]`

## EXPECTATION (For Testing)
An atom with `status=STABLE`, no implementation, a 'CUSTOMER' parent, and an 'ARCHITECTURE' parent must be categorized as 'WIP'.
