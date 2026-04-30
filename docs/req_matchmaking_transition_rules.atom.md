---
id: req_matchmaking_transition_rules
human_name: Transition Rules for Matchmaking
type: REQUIREMENT
layer: BUSINESS
version: 1.0
status: STABLE
priority: 5
tags: []
parents:
  - [[req_matchmaking]]
dependents: []
---
# Transition Rules for Matchmaking

## INTENT
Players join the Waiting Room until the required human count is met, then instantly spawn onto the Board.

## THE RULE / LOGIC
The player must be in the Waiting Room when their queue option is selected.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[req_matchmaking_transition_rules]]`
