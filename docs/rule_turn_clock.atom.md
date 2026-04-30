---
id: rule_turn_clock
human_name: 30-Second Max Turn Clock
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [rule, combat, clock]
parents:
  - [[uc_combat_turn]]
  - [[upsilonapi:domain_ruler_state]]
dependents: []
---
# 30-Second Max Turn Clock

## INTENT
To ensure combat maintains a brisk pace and prevents AFK/abandoned matches from locking up server resources or stalling opponents.

## THE RULE / LOGIC
- When an entity's turn officially begins (represented by `ControllerNextTurn` under [[api_ruler_methods]]), a timer of exactly **30.0 seconds** begins counting down on the server.
- The Turn Clock is absolute. Client-side animations or lag do not extend this base duration.
- **Interruption:** If the active entity manually executes `EndOfTurn` (ends their actions early) before the 30 seconds elapse, the timer is immediately cancelled and torn down.
- **Expiration:** If the timer reaches 0, the engine automatically forces an `EndOfTurn` command for the active entity. 
- *Follow-Up Rule Constraint:* If a player allows their timer to expire 3 consecutive times without taking any action, they should be flagged as AFK and potentially auto-forfeited (Implementable under a separate AFK mechanics atom in the future).

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** N/A (Internal Engine Logic)
- **Code Tag:** `@spec-link [[rule_turn_clock]]`
- **Related Issue:** `#12` (Assume issue filed for the missing clock mechanic)
- **Test Names:** `TestShotClockExpiry`, `TestShotClockCancellation`

## EXPECTATION (For Testing)
- Engine announces `ControllerNextTurn`.
- System waits > 30 seconds.
- Engine automatically broadcasts the *next* entity's `ControllerNextTurn` because the previous one was forced to end.
