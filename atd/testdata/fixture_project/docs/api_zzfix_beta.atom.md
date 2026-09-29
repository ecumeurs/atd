---
id: api_zzfix_beta
human_name: "zzfix Beta API"
type: API
version: 1.0
status: STABLE
priority: 3
tags: [zzfix]
parents:
  - [[req_zzfix_alpha]]
dependents:
  - [[mech_zzfix_gamma]]
layer: ARCHITECTURE
---

# zzfix Beta API

## INTENT
To be the "covered atom" state in this fixture: implemented (two `@spec-link` sites in `src/beta.go`, exercising S2's file-mode dedup) and tested (one `@test-link` in `src/beta_test.go`), and the ARCHITECTURE midpoint of the [[req_zzfix_alpha]] -> api_zzfix_beta -> [[mech_zzfix_gamma]] cross-layer chain.

## THE RULE / LOGIC
Implements the capability required by [[req_zzfix_alpha]] and delegates its procedural detail to [[mech_zzfix_gamma]].

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[api_zzfix_beta]]`
- **Test Tag:** `@test-link [[api_zzfix_beta]]`
- Deliberately tagged twice in the same file (`src/beta.go`) so `atd check --file src/beta.go` has a real dedup case to prove (scenario S2).

## EXPECTATION
`atd check --atom api_zzfix_beta` reports 2 impl links and 1 test link. `atd check --file src/beta.go` lists api_zzfix_beta exactly once, not twice.
