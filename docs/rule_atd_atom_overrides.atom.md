---
id: rule_atd_atom_overrides
human_name: "ATD Atom Metadata Overrides"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: DRAFT
priority: 5
tags: [core, audit, heatmap]
parents:
  - [[requirement_webui_documentation_management]]
dependents:
  - [[mechanic_atd_heatmap_calculation]]
  - [[ui_webui_heatmap_explorer]]
bloating: on
heatmap: all
---

# ATD Atom Metadata Overrides

## INTENT
Allow fine-grained control over documentation audits and health visualizations for specific atoms that deviate from standard granularity rules.

## THE RULE / LOGIC
- **bloating**: `on|off` (default `on`). If `off`, `atd audit` skips the bloat check for this atom. Useful for broad CUSTOMER requirements.
- **heatmap**: `all|no_dep|no_code|none` (default `all`).
    - `no_dep`: Ignore coupling heat (returns OPTIMAL).
    - `no_code`: Ignore linked code heat (returns OPTIMAL).
    - `none`: Ignore all thermal metrics (returns OPTIMAL/STABLE).

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_atd_atom_overrides]]`
- **Related Files:** `atd/pkg/atom/parse.go`, `atd/cmd/atd/cmd/audit.go`, `atd/pkg/exploration/heatmap.go`

## EXPECTATION
1. Atoms with `bloating: off` must not trigger LLM bloat analysis during a full audit.
2. Atoms with `heatmap: no_dep` must show an Optimal border in Dependency mode regardless of having >10 dependents.
