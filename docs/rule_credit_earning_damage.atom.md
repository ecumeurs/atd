---
id: rule_credit_earning_damage
human_name: Credit Earning from Damage Rule
type: RULE
layer: ARCHITECTURE
version: 2.0
status: STABLE
priority: 5
tags: [economy, credits, combat]
parents:
  - [[upsilonapi:domain_credit_economy]]
dependents: []
---

# Credit Earning from Damage Rule

## INTENT
To establish the base credit earning mechanism where 1 HP of absolute damage dealt equals 1 credit earned, with healing also earning credits.

## THE RULE / LOGIC
**Base Credit Rule:** 1 HP reduction = 1 credit earned.

**Damage Credits:**
- Credits awarded based on actual HP loss of the target.
- Example: Deal 15 damage to a target with 0 shield = 15 credits.
- Example: Deal 15 damage to a target with 10 shield = 5 credits (10 absorbed by shield).
- Applies to direct attacks and damaging skills.

**Healing Credits:**
- Credits awarded based on actual HP recovered by the target (capped by MaxHP).
- Example: Heal 15 HP to a target missing 5 HP = 5 credits.
- Supports active playstyles and rewards healing contribution.

**Credit Assignment:**
- Credits assigned to the caster of the effect (attacker or healer).
- Tracked per character per match.
- Balance updated via webhooks to the central user account.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_credit_earning_damage]]`
- **Test Names:** `TestDamageCreditEarning`, `TestHealingCreditEarning`
