---
id: rule_friendly_fire_match_type
human_name: Match Type Split
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: []
parents:
  - [[rule_friendly_fire]]
dependents: []
---
# Match Type Split

## INTENT
Applies rule differently based on match type.

## THE RULE / LOGIC
This applies to characters controlled by the identical player and characters controlled by an allied player in a 2v2 match.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_friendly_fire_match_type]]`
- **Implementation:** `upsilonbattle/battlearena/ruler/rules/skill.go`
