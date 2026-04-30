---
id: rule_credit_earning_support
human_name: Support Credit Earning Rule
type: RULE
layer: ARCHITECTURE
version: 2.0
status: STABLE
priority: 5
tags: [economy, credits, support]
parents:
  - [[upsilonapi:domain_credit_economy]]
dependents:
  - [[rule_credit_earning_status_effects]]
---

# Support Credit Earning Rule

## INTENT
To implement support credit earning where damage mitigation (blocking, shielding) earns credits for the support player, and status effects provide flat credit rewards based on skill power.

## THE RULE / LOGIC
**Damage Mitigation Credits:**
- 1 HP mitigated = 1 credit earned
- Applies to shields, damage reduction, and other mitigation
- Credits go to the effect caster, not the shielded character

**Effect Caster Tracking:**
- All effects must track CasterID until effect ends
- Shields continue earning credits for caster even if caster dies
- Enables proper credit assignment for support play

**Status Effect Credits (Flat Rate):**
- Poison/Stun/Buff: SkillWeight/10 credits per application
- Example: 100 SW poison skill = 10 credits per poison application
- Credits awarded at moment of effect application, not per-turn
- Supports support playstyles and status effect users

**Credit Assignment Logic:**
```go
// When shield blocks damage
if shield.BlocksDamage > 0 {
    credits.Earned += shield.BlocksDamage
    credits.AssignedTo = shield.CasterID  // Original caster
}
```

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_credit_earning_support]]`
- **Test Names:** `TestShieldCreditEarning`, `TestStatusEffectCreditEarning`
