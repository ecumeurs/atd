---
id: requirement_customer_action_reporting
status: STABLE
human_name: Action Reporting & Rich Visualization
type: REQUIREMENT
layer: BUSINESS
version: 1.0
priority: 3
tags: [ui, protocol, customization, requirement]
parents:
  - [[battleui:req_player_experience]]
dependents:
  - [[battleui:ui_tactical_action_report]]
---

# New Atom

## INTENT
To provide explicit, structured action feedback (Action Report) for every state change to support rich UI animations and future premium customization features (skins, emotes).

## THE RULE / LOGIC
- Every state mutation MUST be accompanied by an `ActionFeedback` object.
- **Move Actions:** Must report the full path taken by the entity.
- **Combat Actions:** Must report damage dealt, remaining HP, and any triggered effects.
- **Pass Actions:** Must explicitly state that an entity passed its turn.
- **Protocol Extensibility:** The structure must allow additional fields for future effects (skins, emotes) without breaking existing clients.

## TECHNICAL INTERFACE
- **Protocol Field:** `BoardState.action`
- **Specification:** [[communication.md]]
- **Visual Mapping:** CSS-based animations and effects in [[BattleArena.vue]]
- **Future Integration:** Emote palette and skin selection in User Profile.

## EXPECTATION
The system must provide explicit data about moves (paths), combat (damage, results), and status changes (buffs, debuffs) in every state update. This data must support the integration of future premium features like custom animations, skins, and emotes.
