---
id: vision_atd
human_name: "ATD Project Vision"
description: "The project-wide philosophical scope of ATD: what the system is for and where its boundaries lie. Read whenever a BUSINESS-layer atom is added or updated."
type: VISION
version: 1.0
status: STABLE
priority: CORE
tags: [atd, governance, vision]
parents: []
dependents: []
layer: BUSINESS
---

# ATD Project Vision

## INTENT
To state, once and unambiguously, what ATD (Atomic Traceable Documentation) exists to do — so that every proposed BUSINESS-layer atom can be checked against a single scope boundary and scope creep is caught before it enters the graph.

## THE RULE / LOGIC
ATD is a **development governance substrate**, not an application. Its purpose is to keep documentation, code, and tests provably synchronized through bidirectional traceability, and to make drift between them visible and actionable.

In scope:
- Expressing every requirement, rule, and mechanic of the ATD tool itself as a single-responsibility atom linked to code (`@spec-link`) and tests (`@test-link`).
- Deterministic, token-free tooling for navigating and validating that graph (query, crawl, weave, check, trace, lint, heatmap).
- LLM-assisted extraction, classification, and auditing where a deterministic answer is impossible — always subordinate to human governance.
- Workspace/multi-project operation over several `docs/` corpora.

Out of scope (scope-creep guard):
- Hosting atoms that describe the behaviour of any project ATD is used to document. Those belong in that project's own corpus, referenced cross-project via `[[project:atom_id]]`. This corpus is for the ATD tool only (CLI, MCP server, WebUI, VS Code extension, skill).
- Becoming a general project-management, issue-tracking, or CI platform.
- Replacing human architectural judgment; ATD surfaces evidence, humans decide.
- Any feature that makes an LLM the source of truth for what the system *should* do — the atom graph is that source of truth.

## TECHNICAL INTERFACE (The Bridge)
- **Governance role:** Per ATD.md §1.4, this VISION atom MUST be read whenever a BUSINESS-layer atom is added or updated, to confirm the change stays within the project's intended purview — read for governance, never referenced as an atom's `parents:` ancestor. VISION governs scope (is this still within purview?); CONTRACT (see `contract_atd`) governs the currently-guaranteed surface (does this break what's already promised?) — the two are separate axes and a change can trip either independently.
- **Uniqueness:** Exactly one VISION atom per project (enforced by `atd lint`).
- **Override:** Widening scope REQUIRES updating this atom in the same change.

## EXPECTATION (For Testing)
`atd lint` reports exactly one VISION atom for the project. Any BUSINESS-layer addition whose intent falls under an "out of scope" clause above should be rejected or should carry a corresponding update to this atom's scope statement.
