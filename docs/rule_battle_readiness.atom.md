---
id: rule_battle_readiness
human_name: "Battle Readiness Protocol"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [protocol, handshake, initialization]
parents:
  - [[upsilonapi:api_ruler_methods]]
dependents: []
---

# Battle Readiness Protocol

## INTENT
To ensure that all participating controllers are fully synchronized and ready before the combat simulation officially begins.

## THE RULE / LOGIC
- **Registration Phase**: The Ruler accepts controllers until `len(Controllers) == NbControllers`.
- **Match Start**: Once registration is complete, the Ruler broadcasts `BattleStart` to all controllers and transitions internal state to `InProgress`.
- **Readiness Wait**: The Ruler stays in a "Ready-Waiting" sub-state. It will NOT hand out the first turn yet.
- **Handshake**: Each controller must explicitly send a `ControllerBattleReady` notification after they have successfully fetched and initialized the initial game state (grid, entities).
- **Activation**: Only when **EVERY** registered controller has signaled `ControllerBattleReady` does the Ruler trigger the first turn (`ControllerNextTurn`) and start the turn shot clock.

## TECHNICAL INTERFACE (The Bridge)
- **API Message:** `rulermethods.ControllerBattleReady`
- **Code Tag:** `@spec-link [[rule_battle_readiness]]`
- **Related Logic:** `isBattleReadyToExecute()` in `ruler.go`

## EXPECTATION (For Testing)
- Verify that `ControllerNextTurn` is NOT sent until the final `ControllerBattleReady` message is received by the Ruler.
