---
id: requirement_req_ui_session_timeout
status: STABLE
type: REQUIREMENT
layer: BUSINESS
version: 1.0
priority: 3
parents:
  - [[req_security_token_ttl]]
human_name: UI Session Timeout Requirement
tags: [ui,auth,session,ux]
dependents: []
---

# New Atom

## INTENT
Enforce a clear, immersive user experience when the neural synchronization (JWT) has been terminated or expired.

## THE RULE / LOGIC
- The system MUST monitor all incoming HTTP responses for status code 401.
- Upon 401 detection, the system MUST:
  1. Set the global `isSessionExpired` state to TRUE.
  2. Clear all local authentication artifacts (tokens, user data).
  3. Clear persistent tactical identities to ensure a clean slate.
  4. Trigger the Display of the Session Expired Modal.
- The Modal MUST bridge the gap between technical failure and the "Neon in the Dust" aesthetic.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[req_ui_session_timeout]]`
- **Global State:** `isSessionExpired` (reactive bool)
- **Modal Component:** `SessionExpiredModal.vue`

## EXPECTATION
- When an API request returns a 401 Unauthorized status, the application MUST NOT perform any further retry logic.
- A non-closable modal MUST appear explaining the synchronization failure.
- The user MUST only have an option to re-establish the connection (redirect to landing/login).
