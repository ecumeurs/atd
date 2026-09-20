---
id: service_atd_lint
human_name: "ATD Structural Linter"
type: MECHANIC
version: 0.1.0
status: STABLE
priority: 3
tags: [atd, cli, lint, validation]
parents:
  - [[module_atd_cli]]
dependents:
  - [[api_atd_serve_lint]]
layer: IMPLEMENTATION
---

# ATD Structural Linter

## INTENT
Perform a fast, deterministic structural validation across all ATD atoms in a documentation directory. This ensures that all atoms adhere to the mandatory schema and reference integrity before they are used by other tools or agents.

## THE RULE / LOGIC
- Scans for all `*.atom.md` files in the specified directory (default: configured docs dir).
- **Mandatory Fields**: Verifies presence of `id`, `human_name`, `type`, `layer`, `version`, `status`, and `priority`.
- **Enum Validation**:
    - `layer` must be one of: `CUSTOMER`, `ARCHITECTURE`, `IMPLEMENTATION`.
    - `priority` must be one of: `1`, `2`, `3`, `4`, `5`, `CORE`.
- **Mandatory Sections**: Verifies non-empty content for `## INTENT`, `## THE RULE / LOGIC`, `## TECHNICAL INTERFACE`, and `## EXPECTATION`.
- **Reference Integrity**: Verifies that all IDs listed in `parents` and `dependents` exist within the scanned set of atoms.
- **Governance Isolation** (ATD.md §1.4): `CONTRACT`/`VISION` atoms sit outside the ancestry graph. Flags a governance atom that declares a non-empty `parents` or `dependents`, and flags any atom that names a governance atom in either of its own link fields.
- **Project Governance** (ATD.md §1.4): Flags more than one `CONTRACT` or `VISION` atom, and — once the corpus holds at least one `BUSINESS`-layer atom — a missing one.
- **Deterministic**: No LLM or network calls.

## TECHNICAL INTERFACE (The Bridge)
- **Binary:** `scripts/cmd/atd/cmd/lint.go`
- **Usage:** `atd lint [dir]`
- **Code Tag:** `@spec-link [[service_atd_lint]]`

## EXPECTATION (For Testing)
- Running `atd lint` on a directory with a missing `id` field in an atom returns a non-zero exit code and a descriptive error message.
- Running `atd lint` on a directory with an unresolved parent link returns an error identifying the broken reference.
- Running `atd lint` on a directory where a `CONTRACT` or `VISION` atom declares any `parents` or `dependents` returns an error naming the offending references.
- Running `atd lint` on a fully compliant documentation directory returns "All atoms passed structural validation." and exit code 0.
