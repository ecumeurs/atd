---
id: requirement_customer_api_first
status: STABLE
human_name: API-First Experience & Documentation
type: REQUIREMENT
layer: BUSINESS
version: 1.0
tags: api,documentation,developer-experience
parents:
  - [[battleui:req_player_experience]]
priority: 3
dependents:
  - [[usecase_api_flow_game_turn]]
  - [[usecase_api_flow_matchmaking]]
  - [[upsilonapi:api_help_endpoint]]
---

# New Atom

## INTENT
To ensure the entire game is accessible, discoverable, and fully playable via a standalone, self-documenting API.

## THE RULE / LOGIC
1. **Full Playability**: 100% of game-critical actions (Auth, Matchmaking, Combat, Progression) must have equivalent API endpoints.
2. **Self-Discovery**: The API must expose a JSON `/help` endpoint listing every available URI, Verb, and its contract.
3. **Frontend Accessibility**: The UI must render the API documentation to ensure transparency and ease of third-party tool integration.
4. **Synchronization**: Documentation and implementation must remain in lockstep; any update to endpoint logic requires an update to the corresponding [[API]] atom and help registry.

## TECHNICAL INTERFACE
- **Help Endpoint:** `GET /api/v1/help`
- **Frontend Page:** `/api-docs` (proposed)
- **Code Tag:** `@spec-link [[requirement_customer_api_first]]`

## EXPECTATION
- The API is fully usable without a browser.
- The /help endpoint returns a complete JSON map of the API.
- Documentation in the UI matches the /help output.
