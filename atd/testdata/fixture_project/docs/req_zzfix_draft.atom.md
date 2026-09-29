---
id: req_zzfix_draft
human_name: "zzfix Draft Requirement"
type: REQUIREMENT
version: 1.0
status: DRAFT
priority: 4
tags: [zzfix]
parents: []
dependents: []
layer: BUSINESS
---

# zzfix Draft Requirement

## INTENT
To give scenario S4 a BUSINESS-layer atom that is NOT STABLE, so `atd update` on it can be shown to proceed with no friction — in direct contrast to this fixture's STABLE BUSINESS requirement, which the same guard must refuse without `--force`.

## THE RULE / LOGIC
This requirement is still under discussion and carries no enforceable rule yet.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[req_zzfix_draft]]` (intentionally not applied — DRAFT, no implementation expected yet).

## EXPECTATION
`atd update` against this atom's file succeeds without `--force`, since its status is DRAFT, not STABLE.
