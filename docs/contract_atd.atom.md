---
id: contract_atd
human_name: "ATD Project Contract"
description: "The project-wide mandatory invariants ATD must always uphold. Read whenever a BUSINESS-layer atom is added, removed, or updated; prevents removal of atoms mandatory to the current stable setup."
type: CONTRACT
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, governance, contract]
parents: []
dependents: []
layer: BUSINESS
---

# ATD Project Contract

## INTENT
To pin the hard, non-negotiable invariants of the ATD system in one place, so that any BUSINESS-layer change can be checked against them and no change can silently remove a guarantee the project's stable setup depends on.

## THE RULE / LOGIC
The following invariants are mandatory. A change that violates any of them is rejected unless this contract is updated in the same change to reflect a deliberate, human-ratified new state:

1. **Traceable origin.** Every IMPLEMENTATION-layer atom MUST have at least one BUSINESS-layer ancestor via `parents:`. Code without a business root is not allowed (the `has_customer_origin` invariant).
2. **Resolvable links.** Every `@spec-link`/`@test-link` and every `parents:`/`dependents:` reference MUST resolve to an existing atom in the workspace. No dangling references.
3. **Canonical taxonomy.** Every atom's `type` and `layer` MUST come from the sanctioned sets in ATD.md §1.3/§1.5. `layer` is one of BUSINESS / ARCHITECTURE / IMPLEMENTATION.
4. **Governance uniqueness.** Exactly one CONTRACT and one VISION atom exist per project; both live in the BUSINESS layer.
5. **Authoritative source of truth.** The atom graph — not any LLM output — is the source of truth for what the system should do. LLM-backed tools may inform but never override it.
6. **Stable means signed-off.** A STABLE BUSINESS atom is authoritative; modifying it requires explicit human confirmation (`atd update --force`), not silent agent edits.

## TECHNICAL INTERFACE (The Bridge)
- **Governance role:** Per ATD.md §1.4, this CONTRACT atom MUST be read whenever a BUSINESS-layer atom is added, removed, or updated.
- **Enforcement:** Invariants 2–4 and 6 are checked by `atd lint` and the `atd update` STABLE+BUSINESS guard; invariant 1 by `atd trace`'s `has_customer_origin` flag.
- **Uniqueness:** Exactly one CONTRACT atom per project (enforced by `atd lint`).
- **Override:** Relaxing any invariant REQUIRES updating this atom in the same change.

## EXPECTATION (For Testing)
`atd lint` reports exactly one CONTRACT atom and zero unresolved links / non-canonical types / non-canonical layers across the corpus. `atd update` on a STABLE BUSINESS atom without `--force` fails with a governance refusal. `atd trace` on any IMPLEMENTATION atom reports `has_customer_origin: true`.
