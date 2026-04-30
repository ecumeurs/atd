---
id: rule_pvp_stalemate_draw
human_name: "PvP Stalemate Protocol"
type: RULE
layer: BUSINESS
version: 1.0
status: REVIEW
priority: 3
tags: [pvp, ruling, stalemate]
parents:
  - [[battleui:req_ui_look_and_feel]]
dependents: []
---

# PvP Stalemate Protocol

## INTENT
To prevent infinite, "soft-locked" matches where no remaining combatant has enough Attack power to overcome an opponent's Defense.

## THE RULE / LOGIC
- **Condition Check:** At the end of every turn, perform a "Vulnerability Audit".
- **Stalemate Logic:** If for all remaining players (Team A and Team B):
  - `MAX(Attack_Team_A) <= MIN(Defense_Team_B)` AND
  - `MAX(Attack_Team_B) <= MIN(Defense_Team_A)`
- **Resolution:** The game MUST immediately end in a **DRAW**.
- **Exception:** If any side can still use a tactic/ability (if implemented) that bypasses static defense, the stalemate condition is suppressed. (Currently not applicable until Tactics are implemented).

## TECHNICAL INTERFACE (The Bridge)
- **Logic Location:** Go Engine's Turn Evaluator (and mirrored in Laravel Webhook handler for state reconciliation).
- **Code Tag:** `@spec-link [[rule_pvp_stalemate_draw]]`
- **Issue Reference:** `ISS-029`

## EXPECTATION (For Testing)
This rule is superseded by the 'Minimum 1 Damage' logic implemented in [[mech_combat_attack_computation]], ensuring that combat eventually concludes. No stalemate detection is required.
