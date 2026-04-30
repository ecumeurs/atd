---
id: rule_ruler_test_robustness
human_name: "Ruler Testing Robustness"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [testing, robustness, non-determinism]
parents:
  - [[upsilonapi:api_ruler_methods]]
dependents: []
---

# Ruler Testing Robustness

## INTENT
To prevent flaky tests when validating Ruler logic involving multi-controller scenarios and randomized initiative.

## THE RULE / LOGIC
- **Non-Determinism**: Be aware that `NewCompleteRuler` generates random entity delays. The first turn order is therefore non-deterministic between controllers.
- **Robust Detection**: Tests MUST NOT assume a specific controller is picked first.
- **Wait Pattern**: When waiting for a turn event (`ControllerNextTurn`), tests must monitor the inboxes of **all** participating controllers (e.g., using a `select` loop or a robust helper).
- **Context Identification**: Once a turn is detected, the test should identify the `activeCtrl` and use that reference for subsequent assertions in that turn segment.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_ruler_test_robustness]]`
- **Reference Pattern:** Loop detection used in `TestShotClockExpiry` and `TestShotClockCancellation`.

## EXPECTATION (For Testing)
- Test suite passes consistently even across multiple runs with different random seeds (`go test -count 10`).
