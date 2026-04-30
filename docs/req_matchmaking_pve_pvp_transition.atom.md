---
id: req_matchmaking_pve_pvp_transition
human_name: Transition for PVE/PVP Matchmaking Queues
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
# Transition for PVE/PVP Matchmaking Queues

## INTENT
For both 1v1 and 2V2 queue options, players instantly spawn onto the Board after joining the Waiting Room.

## THE RULE / LOGIC
The player must immediately spawn onto the Board upon joining the Waiting Room.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[req_matchmaking_pve_pvp_transition]]`
