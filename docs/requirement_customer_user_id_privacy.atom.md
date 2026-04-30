---
id: requirement_customer_user_id_privacy
status: STABLE
type: REQUIREMENT
layer: BUSINESS
priority: 3
tags: security,privacy,auth
version: 1.2
parents: []
dependents:
  - [[upsilonapi:arch_api_id_masking_gateway]]
  - [[upsilonbattle:entity_player]]
human_name: Internal Identity & Ownership Privacy
---

# New Atom

## INTENT
To ensure that internal database identifiers (User) are never exposed to the frontend, protecting the system from primary key enumeration and ensuring that all actions are validated against explicit ownership.

## THE RULE / LOGIC
- **ID Masking:** The primary UUID (database ID) of a User MUST NOT be sent to the client (frontend) in public contexts.
- **Ownership Enforcement:** Every API request targeting a specific Character or Match MUST verify that the authenticated User is the legitimate owner or participant before processing the request.
- **WebSocket Security:** WebSocket private channels MUST be keyed using a secure, persistent pseudonym (`ws_channel_key`) instead of User IDs.
- **Identity Resolution:** The backend MUST resolve user identity purely via JWT/Session context. Client-provided user IDs MUST NOT be trusted for identity or authorization.

## TECHNICAL INTERFACE
- **PHP Resource:** `UserResource` (must exclude `id` field)
- **Frontend Utility:** `tactical_id.js`
- **Code Tag:** `@spec-link [[requirement_customer_user_id_privacy]]`

## EXPECTATION
