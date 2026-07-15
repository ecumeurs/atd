---
id: req_zzfix_tech_debt_backlog
human_name: "zzfix Tech Debt Backlog (escape hatch)"
type: REQUIREMENT
version: 1.0
status: DRAFT
priority: 5
tags: [zzfix, tech-debt]
parents: []
dependents:
  - [[mech_zzfix_orphan]]
layer: BUSINESS
---

# zzfix Tech Debt Backlog (escape hatch)

## INTENT
To be this fixture's own copy of the repo-wide "escape hatch" parent the checked-out repo's `.git/hooks/pre-commit` suggests for a deliberately unlinked ARCHITECTURE/IMPLEMENTATION atom (`parents: [[req_tech_debt_backlog]]`) -- [[mech_zzfix_orphan]] needs *a* parent to pass that structural hook, without acquiring real implementation or being excluded from `atd crawl --gaps`'s orphan detection (which looks at `Implementations`/dependents, not `parents`, so this placeholder parent does not change mech_zzfix_orphan's orphan status at all -- see pkg/exploration/orphan.go's IsOrphan).

## THE RULE / LOGIC
Atoms parented here are acknowledged tech debt: documented but deliberately unimplemented, tracked rather than silently missing a parent.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[req_zzfix_tech_debt_backlog]]` (intentionally not applied -- this is a bookkeeping placeholder, not a real capability).

## EXPECTATION
Exists solely so [[mech_zzfix_orphan]] can declare a non-empty `parents:` list; carries no coverage expectation of its own.
