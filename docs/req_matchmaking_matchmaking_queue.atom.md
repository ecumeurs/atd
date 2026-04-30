---
id: req_matchmaking_matchmaking_queue
human_name: Matchmaking Queue Requirement
type: REQUIREMENT
layer: BUSINESS
version: 1.1
status: STABLE
priority: 5
tags: []
parents:
  - [[req_matchmaking]]
dependents: []
---
# Matchmaking Queue Requirement

## INTENT
The system must offer four explicit matchmaking queue options:

## THE RULE / LOGIC
The system must offer exactly four distinct matching queue options:
- 1v1_PVP
- 1v1_PVE
- 2v2_PVP
- 2v2_PVE

Any other mode must be rejected with a 4xx error.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[req_matchmaking_matchmaking_queue]]`
