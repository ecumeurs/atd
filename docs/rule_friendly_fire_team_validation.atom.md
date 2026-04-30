---
id: rule_friendly_fire_team_validation
human_name: Team Validation Split
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
# Team Validation Split

## INTENT
Validates that characters are not part of the same team before applying destructive behavior.

## THE RULE / LOGIC
Target Validation: Characters identified as belonging to the same team cannot apply destructive behavior (e.g., attacks, negative modifiers) to one another.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_friendly_fire_team_validation]]`
- **Implementation:** `upsilonbattle/battlearena/ruler/rules/attack.go`, `upsilonbattle/battlearena/ruler/rules/skill.go`
