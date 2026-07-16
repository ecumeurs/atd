---
id: req_zzfix_ws_alpha
human_name: "zzfix Workspace Alpha Requirement"
type: REQUIREMENT
version: 1.0
status: STABLE
priority: 3
tags: [zzfix, workspace]
parents: []
dependents:
  - [[zzfix_b:api_zzfix_ws_beta]]
layer: BUSINESS
---

# zzfix Workspace Alpha Requirement

## INTENT
To be a local (same-project) atom in project zzfix_a, implemented and tested within zzfix_a itself, used by scenario S1's workspace-context id-form test (canonical/bare vs a redundant `requirement_`-prefixed id, per test_atd_07_26.md Addendum B's strip-and-retry) and by scenario S9 (future work) as the cross-project parent of `zzfix_b:api_zzfix_ws_beta`.

## THE RULE / LOGIC
The system must expose the zzfix workspace alpha capability.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[req_zzfix_ws_alpha]]`
- **Test Tag:** `@test-link [[req_zzfix_ws_alpha]]`

## EXPECTATION
`atd check --atom req_zzfix_ws_alpha` (run from zzfix_a) reports 1 impl link and 1 test link, whether the id is given bare (`req_zzfix_ws_alpha`) or with a redundant `requirement_` prefix (`requirement_req_zzfix_ws_alpha`) — the workspace resolver strips the known type-word prefix and retries.
