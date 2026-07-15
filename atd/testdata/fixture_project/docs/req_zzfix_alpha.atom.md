---
id: req_zzfix_alpha
human_name: "zzfix Alpha Requirement"
type: REQUIREMENT
version: 1.0
status: STABLE
priority: 3
tags: [zzfix]
parents: []
dependents:
  - [[api_zzfix_beta]]
  - [[rule_zzfix_untested]]
layer: BUSINESS
---

# zzfix Alpha Requirement

## INTENT
To anchor the top of this fixture's cross-layer chain: a BUSINESS-layer requirement with an ARCHITECTURE child ([[api_zzfix_beta]]) and its own directly-linked RULE ([[rule_zzfix_untested]]), so scenario tests have a real ancestry to walk.

## THE RULE / LOGIC
The system must expose the zzfix alpha capability described by [[api_zzfix_beta]], subject to the constraint recorded in [[rule_zzfix_untested]].

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[req_zzfix_alpha]]` (intentionally not applied — BUSINESS-layer atoms link down, not to code directly).

## EXPECTATION
`atd trace req_zzfix_alpha` reports two dependents ([[api_zzfix_beta]], [[rule_zzfix_untested]]) and `has_business_origin: true`.
