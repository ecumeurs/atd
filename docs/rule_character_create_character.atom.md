---
id: rule_character_create_character
human_name: Character creation base allocation
type: RULE
layer: ARCHITECTURE
version: 2.0
status: STABLE
priority: 5
tags: [character, creation]
parents:
  - [[upsilonbattle:entity_character]]
dependents: []
---
# Character creation base allocation

## INTENT
Define the V2 base attributes (x10 baseline) and the initial 100 CP point-buy allocation for new characters.

## THE RULE / LOGIC
- **Base Attributes (V2 x10 Baseline):** Every character starts with:
  - HP: 30
  - Attack: 10
  - Defense: 5
  - Movement: 3
- **Initial Allocation (Point-Buy):** Characters are initialized with a starting pool of exactly **100 Character Points (CP)** to spend on top of their base attributes.
- **Unspent CP:** Upon creation, all 100 CP are kept in the character's pool (`spent_cp: 0`), leaving the character "ready for progression" via the progression allocation rules.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_character_create_character]]`
- **Test Names:** `TestInitialStatConsistency`
