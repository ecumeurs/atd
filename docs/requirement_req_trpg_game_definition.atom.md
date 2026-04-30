---
id: requirement_req_trpg_game_definition
status: STABLE
layer: BUSINESS
version: 1.0
type: REQUIREMENT
priority: 5
tags: [trpg, combat, game-definition]
parents:
  - [[battleui:req_player_experience]]
dependents: []
human_name: TRPG Game Definition
---

# New Atom

## INTENT
Define Upsilon Battle as a turn-based Tactical RPG where players deploy character rosters onto a procedurally generated grid and compete via initiative-driven combat.

## THE RULE / LOGIC
- **Genre:** Turn-based Tactical RPG (TRPG).
- **Core Loop:** Players deploy 3 characters onto a grid board. Characters act in order of initiative (delay-based ticker). On each turn, a character may Move, Attack, Pass, or Forfeit.
- **Board:** A procedurally generated rectangular grid (5-15 tiles per side) with obstacle tiles.
- **Teams:** Players are grouped into 2 opposing teams. Allies share a TeamID. Victory is achieved when all opposing entities are defeated or forfeit.
- **Action Economy:** Each action incurs a delay cost that determines when the character next acts. Move (+20/tile), Attack (+100), Pass (+300), Timeout (+400).
- **Shot Clock:** Each turn has a strict 30-second limit. Exceeding it triggers an auto-pass with penalty.
- **Modes:** 1v1 PvE, 1v1 PvP, 2v2 PvE, 2v2 PvP.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[req_trpg_game_definition]]`
- **Related Atoms:** `[[module_game]]`, `[[uc_combat_turn]]`, `[[mech_action_economy]]`

## EXPECTATION
- A player understands the game is a turn-based tactical RPG with initiative-driven combat on a grid.
- The combat loop enforces fair time constraints and strategic depth through the action economy.
