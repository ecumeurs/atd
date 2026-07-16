---
id: mech_zzfix_gamma
human_name: "zzfix Gamma Mechanic"
type: MECHANIC
version: 1.0
status: STABLE
priority: 3
tags: [zzfix]
parents:
  - [[api_zzfix_beta]]
dependents: []
layer: IMPLEMENTATION
---

# zzfix Gamma Mechanic

## INTENT
To be the IMPLEMENTATION leaf of this fixture's cross-layer chain ([[req_zzfix_alpha]] -> [[api_zzfix_beta]] -> mech_zzfix_gamma), and the atom whose docs use `###` subheadings inside an H2 section — the exact shape that once made `atom.Parse` silently swallow content past the first subheading (test_atd_07_26.md E4).

## THE RULE / LOGIC
Implements the procedural detail [[api_zzfix_beta]] delegates to it.

## TECHNICAL INTERFACE (The Bridge)
### Description
This subsection, and the ones below it, are H3 (`###`) headings living inside the single H2 `## TECHNICAL INTERFACE` section. `atom.Parse` must retain all of their text as part of `Interface` — none of it belongs to a later top-level section.

### Input Schema
Not applicable — this is a fixture atom, not a real MCP tool schema. This subsection exists purely to prove the parser doesn't stop reading at the first `###`.

### Code Tag
`@spec-link [[mech_zzfix_gamma]]`, `@test-link [[mech_zzfix_gamma]]`

## EXPECTATION
`atd check --atom mech_zzfix_gamma` reports 1 impl link and 1 test link. Parsing this file must yield a non-empty `Interface` field containing all three subsection headings above, not just "Description".
