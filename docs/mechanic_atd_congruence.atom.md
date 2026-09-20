---
id: mechanic_atd_congruence
human_name: "ATD Congruence"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, audit, congruence, consistency]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Congruence

## INTENT
To cross-examine a target atom against its related atoms (parents, dependents, tag-siblings) for logical contradictions in their INTENT and LOGIC sections before any code implementation.

## THE RULE / LOGIC
Loads all atoms from the docs directory, identifies the target atom's related context (parents via `[[id]]` links, dependents referencing the target, siblings sharing tags), builds a congruence audit prompt with only these relevant atoms, and routes through the tiered provider for task `congruence`. This targeted approach prevents context window bloat by excluding unrelated atoms.

A `--target` value may be workspace-qualified (`project:atom_id`); when the
`--workspace` flag is set (or the target carries a `project:` prefix), the
target's docs directory is resolved against that workspace project's own
`docs/` path (via the active `.atd.workspace` config) instead of the current
project's `--docs`/default docs directory, so a target belonging to a
sibling project in the same workspace can be audited without a manual
`--docs <other-project>/docs` override.

The LLM's response is validated, not merely echoed. `is_congruent: false`
requires a non-empty structured `findings` array (one entry per
contradiction, naming the specific contradicting atom, the section it lives
in, and what the contradiction is) — a bare title-only `audit_report` with no
findings is rejected with a non-nil error instead of being printed as if it
were a complete result. `is_congruent: true` still succeeds with an empty (or
absent) `findings` array.

## EXPECTATION
Response contract (`pkg/prompt.CongruenceFormat`):
```json
{
  "is_congruent": false,
  "audit_report": "free-text summary",
  "findings": [
    {"atom_id": "...", "section": "INTENT|THE RULE|LOGIC", "contradiction": "..."}
  ]
}
```
- If `is_congruent` is `false`, `findings` MUST contain at least one entry;
  `RunE` fails the command (non-zero exit) otherwise.
- If `is_congruent` is `true`, an empty `findings` array is acceptable.
- A response that fails to parse as this JSON shape is treated as a hard
  error, not silently printed.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd congruence --target <atom_id> [--docs <dir>] [--workspace]`
- **LLM Task:** `congruence`
- **Code Tag:** `@spec-link [[mechanic_atd_congruence]]`
- **Related known-defect record (now fixed):** see
  `failures/20260918_release_notes.md`'s "Third batch" entry — the prior
  behavior (bare `is_congruent: false` verdict accepted as complete, no
  `--workspace` resolution) previously lived in
  `failures/20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md`,
  now removed as fully processed.
