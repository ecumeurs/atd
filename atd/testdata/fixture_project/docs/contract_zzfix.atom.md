---
id: contract_zzfix
human_name: "zzfix Fixture Contract"
type: CONTRACT
version: 1.0
status: STABLE
priority: CORE
tags: [zzfix, governance]
parents: []
dependents: []
layer: BUSINESS
---

# zzfix Fixture Contract

## INTENT
To pin the one mandatory invariant this fixture corpus exists to exercise: exactly one CONTRACT and one VISION atom per project, so `atd lint`'s governance checks have a real, always-clean baseline to regress against.

## THE RULE / LOGIC
1. Exactly one CONTRACT atom and one VISION atom exist in this fixture project.
2. Every atom's `type` and `layer` come from the canonical sets ATD.md documents.
3. Every `parents:`/`dependents:` reference resolves to an existing atom in this fixture.

## TECHNICAL INTERFACE
- **Enforcement:** `atd lint`'s project-governance check (exactly-one CONTRACT/VISION).

## EXPECTATION
`atd lint` against this fixture's docs/ reports exactly one CONTRACT and one VISION atom, with zero governance errors.
