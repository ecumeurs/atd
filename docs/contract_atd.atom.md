---
id: contract_atd
human_name: "ATD Project Contract"
description: "The versioned set of guarantees ATD currently makes about its own behavior — its public contract, not an enumeration of every BUSINESS atom. Read whenever a change touches an atom whose behavior is part of this guaranteed surface; breaking changes require a MAJOR bump here, additive ones a MINOR bump."
type: CONTRACT
version: 1.0
status: STABLE
priority: CORE
tags: [atd, governance, contract]
parents: []
dependents: []
layer: BUSINESS
---

# ATD Project Contract

## INTENT
To state, once and versioned, what ATD currently guarantees about its own behavior to anything that adopts it — so that a change to a guaranteed property is always a deliberate, human-ratified version bump, never a silent drift. For this project, ATD's own product *is* the graph-integrity discipline below, so the surface enumerated here is exhaustive rather than a subset. A consumer project instead classifies each of its own BUSINESS/ARCHITECTURE atoms as either surface (a behavior promised to adopters) or internal, and only the surface ones belong in its contract.

## THE RULE / LOGIC
The following invariants are mandatory. A change that violates any of them is rejected unless this contract's `version` takes a MAJOR bump in the same change, reflecting a deliberate, human-ratified new state:

1. **Traceable origin.** Every IMPLEMENTATION-layer atom MUST have at least one BUSINESS-layer ancestor via `parents:`. Code without a business root is not allowed (the `has_customer_origin` invariant).
2. **Resolvable links.** Every `@spec-link`/`@test-link` and every `parents:`/`dependents:` reference MUST resolve to an existing atom in the workspace. No dangling references.
3. **Canonical taxonomy.** Every atom's `type` and `layer` MUST come from the sanctioned sets. `type` is one of CONTRACT, VISION, REQUIREMENT, USER_STORY (synonyms USECASE, WORKFLOW), RULE, DOMAIN, MODULE, SERVICE, API, ENTITY, UI, SPECIFICATION, DATA, BUILD, USAGE, MECHANIC. `layer` is one of BUSINESS / ARCHITECTURE / IMPLEMENTATION.
4. **Governance uniqueness.** Exactly one CONTRACT and one VISION atom exist per project; both live in the BUSINESS layer.
5. **Authoritative source of truth.** The atom graph — not any LLM output — is the source of truth for what the system should do. LLM-backed tools may inform but never override it.
6. **Stable means signed-off.** A STABLE BUSINESS atom is authoritative; modifying it requires explicit human confirmation (`atd update --force`), not silent agent edits.
7. **Self-hosting boundary.** This corpus documents the ATD tool itself ONLY. Atoms describing the behaviour, rules, requirements, or entities of any *consumer* project (e.g. a game ATD is used to document) belong in that project's own `docs/`, referenced cross-project via `[[project:atom_id]]` — never hosted here. An atom whose subject matter is not the ATD tool, its CLI, its MCP server, its WebUI, or its VS Code extension is out of place and MUST be relocated, regardless of how its parents are wired.

## TECHNICAL INTERFACE (The Bridge)
- **Governance role:** This CONTRACT atom MUST be read whenever a change would alter, weaken, or remove one of the invariants below — no `parents:` link needed to trigger it; this atom is read for governance, never referenced as an atom's ancestor.
- **Enforcement:** Invariants 2–4 and 6 are checked by `atd lint` and the `atd update` STABLE+BUSINESS guard; invariant 1 by `atd trace`'s `has_customer_origin` flag; invariant 7 by review against this project's module list (CLI, MCP server, WebUI, VS Code extension, skill) — atoms outside those modules are rejected.
- **Uniqueness:** Exactly one CONTRACT atom per project (enforced by `atd lint`).
- **Override:** Relaxing any invariant REQUIRES a MAJOR bump on this atom's `version` in the same change; adding a new invariant (widening the guarantee, not narrowing it) takes a MINOR bump instead.

## EXPECTATION (For Testing)
`atd lint` reports exactly one CONTRACT atom and zero unresolved links / non-canonical types / non-canonical layers across the corpus. `atd update` on a STABLE BUSINESS atom without `--force` fails with a governance refusal. `atd trace` on any IMPLEMENTATION atom reports `has_customer_origin: true`. No atom in this corpus describes the behaviour of a consumer project; any that does is flagged as a self-hosting violation.
