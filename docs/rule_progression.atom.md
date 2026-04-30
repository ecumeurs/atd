---
id: rule_progression
human_name: Character Progression Rule
type: RULE
layer: ARCHITECTURE
version: 2.1
status: STABLE
priority: 5
tags: [progression, character]
parents:
  - [[requirement_customer_player_profile]]
  - [[upsilonbattle:entity_character]]
dependents:
  - [[battleui:ui_character_full_stat_panel]]
  - [[rule_stat_taxonomy]]
  - [[uc_progression_stat_allocation]]
---
# Character Progression Rule

## INTENT
Governs how character attributes improve after participating in a successful game using the V2 CP Point-Buy System, enforcing mathematical balance.

## THE RULE / LOGIC
- **Post-Win Reward:** After each game win, the player's account gains exactly +10 Character Points (CP) to its allowed progression cap.
- **Point-Buy System Constraints:**
  - **Global Cap:** A character's total `spent_cp` on upgrades MUST NOT exceed `100 + (total_wins * 10)`.
  - **Non-Negativity:** No attribute is allowed to have a negative value.
- **Stat Taxonomy (CP-upgradable vs item-only):**
  - **Class A — Character-leveled (CP-upgradable, persisted on `characters`):** HP, MP, SP, Attack, Defense, Movement, JumpHeight, CritChance, CritDamage.
  - **Class B — Effective-only (granted by items / buffs only, NEVER CP-upgradable):** AttackRange, Shield. These properties exist on the engine entity but are not selectable from the upgrade UI; they appear on the dashboard only as item / buff contributions.
- **Attribute Costs (Class A — Standard):**
  - HP (+1): Costs 1 CP
  - **MP (+1): Costs 1 CP** *(resource counter, parity with HP)*
  - **SP (+1): Costs 1 CP** *(resource counter, parity with HP)*
  - Attack (+1): Costs 5 CP
  - Defense (+1): Costs 5 CP
  - Movement (+1): Costs 30 CP
- **Attribute Costs (Class A — Exotic):**
  - JumpHeight (+1): 15 CP
  - CritChance (+1%): 10 CP
  - CritDamage / CritMultiplier (+5%): 5 CP
- **Unrestricted Spend:** The old V1 movement restriction (once every 5 wins) is entirely removed because Movement inherently self-balances via its 30 CP cost.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_progression]]`
- **Test Names:** `TestPostWinStatAllocation`, `TestMovementProgressionRestriction`, `TestGlobalAttributeCap`

## EXPECTATION (For Testing)
- Character wins a game -> Gains 10 CP to their maximum allowed progression cap.
- Player assigns 1 Attack to a character -> The upgrade costs 5 CP; operation is successful if `spent_cp + 5 <= 100 + (total_wins * 10)`.
- Player assigns 1 MP to a character -> The upgrade costs 1 CP and increments both `mp` and `max_mp`.
- Player assigns 1 SP to a character -> The upgrade costs 1 CP and increments both `sp` and `max_sp`.
- Player attempts to upgrade AttackRange or Shield via the progression endpoint -> Operation rejected (Class B stats are item / buff only).
- Player tries to upgrade attributes exceeding the allowed total CP cap -> Operation rejected.
