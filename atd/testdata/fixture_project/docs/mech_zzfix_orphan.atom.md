---
id: mech_zzfix_orphan
human_name: "zzfix Orphan Mechanic"
type: MECHANIC
version: 1.0
status: STABLE
priority: 3
tags: [zzfix]
parents:
  - [[req_zzfix_tech_debt_backlog]]
dependents: []
layer: IMPLEMENTATION
---

# zzfix Orphan Mechanic

## INTENT
To be the deliberate orphan in this fixture: a STABLE IMPLEMENTATION-layer atom with no dependents and no `@spec-link` anywhere in `src/`, so `atd crawl --gaps` has exactly one atom it must report. Parented only to the [[req_zzfix_tech_debt_backlog]] escape hatch -- required by the repo's own pre-commit structural hook (every ARCHITECTURE/IMPLEMENTATION atom needs *a* parent) -- which does not affect orphan status: `atd crawl --gaps`'s IsOrphan looks at `Implementations` and dependents, never at `parents` (pkg/exploration/orphan.go).

## THE RULE / LOGIC
This mechanic is documented but was never implemented — it exists only to be caught by orphan detection.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[mech_zzfix_orphan]]` (deliberately NOT applied anywhere in src/ — that is the point of this atom).

## EXPECTATION
`atd crawl --gaps` reports mech_zzfix_orphan in `orphaned_stable_atoms`, and no other atom in this fixture appears there.
