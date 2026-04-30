---
id: req_security_authorization
human_name: Authorization Requirement
type: REQUIREMENT
layer: BUSINESS
version: 1.1
status: STABLE
priority: 5
tags: [auth, authorization]
parents:
  - [[req_security]]
dependents: []
---
# Authorization Requirement

## INTENT
All other UI pages and backend API calls require a valid Sanctum Token.

## THE RULE / LOGIC
Every UI page (dashboard, waiting room, board page) and all backend API calls (except public ones) require a valid Personal Access Token via the Authorization header.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[req_security_authorization]]`
