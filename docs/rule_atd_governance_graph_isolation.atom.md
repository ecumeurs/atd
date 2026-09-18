---
id: rule_atd_governance_graph_isolation
human_name: "Governance Atom Graph Isolation"
description: "CONTRACT and VISION atoms sit outside the ancestry graph: they declare no parents/dependents of their own, and no atom may name one in either link field. Reported by `atd lint`, repaired by `atd weave`."
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [atd, governance, contract, vision, graph, lint, weave]
parents:
  - [[domain_atd_structure]]
dependents: []
---

# Governance Atom Graph Isolation

## INTENT
To keep `CONTRACT` and `VISION` atoms out of the ancestry graph entirely, so that governance is read from the side rather than inherited through a `parents:` chain.

## THE RULE / LOGIC
Governance atoms (`CONTRACT`, `VISION` — ATD.md §1.4) are read for governance, never linked as structural ancestry. Four link shapes are therefore forbidden:

1. An ordinary atom naming a governance atom in its `parents:`.
2. An ordinary atom naming a governance atom in its `dependents:`.
3. A governance atom declaring any `parents:` of its own.
4. A governance atom declaring any `dependents:` of its own.

The rule holds across project boundaries: a `[[project:contract_x]]` reference is as forbidden as a bare one, and a governance atom in any workspace project is isolated in every other.

### Enforcement
- **`atd lint` reports.** Each of the four shapes is a distinct lint error naming the offending reference. Lint never edits.
- **`atd weave` repairs.** Weave drops a forbidden `parents:` entry instead of canonicalizing it, and empties a governance atom's own `parents:`. The dropped edge is excluded from the parent→dependents map, so neither end regains a `dependents:` entry for it. Every removal is named in weave's result text — weave rewrites files in place, so nothing is dropped silently.
- **Conservative by default.** An atom with no forbidden link has its `parents:` block left byte-identical; weave must not re-sort or re-render a clean corpus.
- **One predicate.** Both commands classify governance atoms through `atom.IsGovernanceType`, so the two can never disagree about what a governance atom is.

The `dependents:` side is weave-derived, so a governance atom holding dependents always mirrors a `parents:` violation on some other atom; lint reports both ends, and one weave pass clears both.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_atd_governance_graph_isolation]]`
- **Predicate:** `atd/pkg/atom/parse.go` — `IsGovernanceType`, `BareAtomID`
- **Reported by:** `atd/cmd/atd/cmd/lint.go` — `runLint`
- **Repaired by:** `atd/pkg/exploration/weave.go` — `stripGovernanceParents`, `weaveSingleProject`, `weaveWorkspace`

## EXPECTATION (For Testing)
- `atd lint` on a corpus where a `CONTRACT` declares `parents:` or `dependents:` exits non-zero and names the offending references.
- `atd lint` on a corpus where an ordinary atom names a governance atom in `parents:` or in `dependents:` exits non-zero for that atom.
- `atd weave` over an atom whose `parents:` names a governance atom removes that entry, keeps every other parent, and leaves the governance atom's `dependents:` empty.
- `atd weave` over a governance atom that declares `parents:` empties the block and reports the removal.
- `atd weave` over a corpus with no forbidden link rewrites zero files.
