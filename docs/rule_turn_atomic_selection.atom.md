---
id: rule_turn_atomic_selection
status: STABLE
human_name: Atomic Turn Selection Rule
type: RULE
priority: 5
parents:
  - [[uc_combat_turn]]
dependents: []
layer: ARCHITECTURE
version: 1.0
---

# New Atom

## INTENT
To ensure that entities are only removed from the Turn Queue when their controlling actor is ready and the engine is guaranteed to execute the turn.

## THE RULE / LOGIC
- The Ruler must treat the `Turner` queue as an immutable peek-source until the Turn Handover is finalized.
- Before calling `Turner.NextTurn()`, the Ruler must verify that the candidate entity's controller has signaled readiness (`ControllerBattleReady`).
- If readiness is not verified, the Ruler must pause and wait for the signal, leaving the entity at the head of the queue.
- This prevents "Entity Leaks" where an entity is popped but no notification is sent, leading to an empty queue and match hangs.

## TECHNICAL INTERFACE
- **Ruler Method:** `triggerFirstTurn`
- **Code Tag:** `@spec-link [[rule_turn_atomic_selection]]`
- **Check Constraint:** `isBattleReadyToExecute` must be true.

## EXPECTATION
- Turner queue has Entity A -> Entity A's controller is NOT ready -> Ruler does NOT call NextTurn().
- Turner queue has Entity A -> Entity A's controller IS ready -> Ruler calls NextTurn() -> Entity A is popped and Turn begins.
- Result: Turner queue matches actual active turns; no entities are "lost."
