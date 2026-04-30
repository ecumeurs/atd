---
id: rule_friendly_fire
human_name: Friendly Immunity Rule
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: []
parents: []
dependents:
  - [[rule_friendly_fire_match_type]]
  - [[rule_friendly_fire_team_validation]]
---
# Friendly Immunity Rule

## INTENT
To aggregate the constituent rules of Friendly Immunity Rule.

## THE RULE / LOGIC
Prohibits self-inflicted harm between players and characters on the same side.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_friendly_fire]]`
- **Implementation:** `upsilonbattle/battlearena/ruler/rules/attack.go`, `upsilonbattle/battlearena/ruler/rules/skill.go`
