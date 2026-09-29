---
id: vision_zzfix
human_name: "zzfix Fixture Vision"
type: VISION
version: 1.0
status: STABLE
priority: CORE
tags: [zzfix, governance]
parents: []
dependents: []
layer: BUSINESS
---

# zzfix Fixture Vision

## INTENT
To state the scope of this fixture project: a small, complete, self-consistent ATD corpus used only by the ATD toolkit's own scenario test suite, never a real product.

## THE RULE / LOGIC
In scope: exactly the states the scenario harness needs to classify — a covered atom, an atom implemented but untested, an orphan, a cross-layer chain, and an atom whose docs use `###` subheadings inside an H2 section.

Out of scope: anything resembling real UpsilonBattle product functionality. Every id in this corpus carries a `zzfix` marker so repo-wide rename/propagation tooling can never mistake it for real content.

## TECHNICAL INTERFACE
- **Enforcement:** `atd lint`'s project-governance check (exactly-one CONTRACT/VISION).

## EXPECTATION
`atd lint` against this fixture's docs/ reports exactly one VISION atom, with zero governance errors.
