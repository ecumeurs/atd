---
id: rule_credit_earning_status_effects
human_name: Status Effect Credit Earning Rule
type: RULE
layer: ARCHITECTURE
version: 2.0
status: STABLE
priority: 5
tags: [economy, credits, status-effects]
parents:
  - [[rule_credit_earning_support]]
  - [[upsilonapi:domain_credit_economy]]
dependents: []
---

# Status Effect Credit Earning Rule

## INTENT
To establish credit earning rules for status effects (poison, stun, buffs) using a flat rate based on skill power (SkillWeight/10 per application), as these effects don't provide immediate HP changes.

## THE RULE / LOGIC
**Status Effect Credit Formula:**
- **Credits Earned:** SkillWeight / 10 per application
- **Example:** 100 SW poison skill = 10 credits per poison application
- **Example:** 200 SW stun skill = 20 credits per stun application
- **Example:** 50 SW buff skill = 5 credits per buff application

**Credit Timing:**
- **Application Moment:** Credits awarded when effect is first applied
- **No Per-Term Credits:** Status effects don't earn ongoing credits
- **One-Time Reward:** Single credit reward per application, not per-turn duration

**Multiple Source Handling:**
- **Overlapping Effects:** Each application earns credits separately
- **Stacking Rules:** Multiple sources can apply same status effect, each earns credits
- **Credit Ownership:** Credits go to effect caster, not effect target

**Status Effect Types:**

**Poison (Damage Over Time):**
- **Credit:** SkillWeight / 10 when poison is applied
- **No Ongoing Credits:** Poison damage per turn does not earn additional credits

**Stun (Disable Actions):**
- **Credit:** SkillWeight / 10 when stun is applied
- **No Ongoing Credits:** Stun duration per turn does not earn additional credits

**Buffs (Stat Enhancements):**
- **Credit:** SkillWeight / 10 when buff is applied
- **No Ongoing Credits:** Buff duration per turn does not earn additional credits

**Debuffs (Stat Reductions):**
- **Credit:** SkillWeight / 10 when debuff is applied
- **No Ongoing Credits:** Debuff duration per turn does not earn additional credits

**Complex Status Effects:**
- **Multi-Effect Skills:** Apply credit formula to each effect component
- **Compound Effects:** Credit based on total SkillWeight of skill
- **Duration Extensions:** No credit for extended duration (cost already in SW)

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_credit_earning_status_effects]]`
- **Test Names:** `TestPoisonCreditEarning`, `TestStunCreditEarning`, `TestBuffCreditEarning`