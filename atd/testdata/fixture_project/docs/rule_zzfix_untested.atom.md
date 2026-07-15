---
id: rule_zzfix_untested
human_name: "zzfix Untested Rule"
type: RULE
version: 1.0
status: STABLE
priority: 3
tags: [zzfix]
parents:
  - [[req_zzfix_alpha]]
dependents: []
layer: ARCHITECTURE
---

# zzfix Untested Rule

## INTENT
To be the "implemented but untested" state in this fixture: a `@spec-link` exists (`src/untested.go`) but no `@test-link` anywhere references it, as required by [[req_zzfix_alpha]]. Also carries a body reference to [[req_zzfix_alpha]] (in addition to this atom's own `parents:` entry above) so scenario S7's rename-propagation test has a genuine second inbound reference to rewrite — one in another atom's frontmatter (see [[api_zzfix_beta]]'s `parents:`), one here in prose.

## THE RULE / LOGIC
The constraint required by [[req_zzfix_alpha]]: zzfix beta must never run without this rule's precondition holding.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_zzfix_untested]]` (applied in `src/untested.go`; deliberately no matching `@test-link` anywhere).

## EXPECTATION
`atd check --atom rule_zzfix_untested` reports 1 impl link and 0 test links (status NO_TESTS).
