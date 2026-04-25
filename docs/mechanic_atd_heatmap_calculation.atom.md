---
id: mechanic_atd_heatmap_calculation
human_name: "Heat Map Calculation Logic"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 3
tags: [heatmap, logic]
parents:
  - [[rule_atd_atom_overrides]]
dependents: []
bloating: on
heatmap: all
---

# Heat Map Calculation Logic

## INTENT
Calculate documentation and code thermal state to identify overloaded or isolated components.

## THE RULE / LOGIC
- **Dependency Heat**: Based on the maximum of parent/dependent counts. HOT if >= 10, WARM if >= 4, OPTIMAL if >= 1, COLD if 0.
- **Code Heat**: Maximum heat of all linked source files. A source file is HOT if >= 10 linked atoms, WARM if >= 6, OPTIMAL if >= 2.
- **Update Heat**: Based on git history (last 20 commits). HOT if >= 7 commits, WARM if >= 4, OPTIMAL if >= 1, STABLE if 0.
- **Overrides**: Atoms can override these states using `heatmap: no_dep|no_code|none` in YAML header.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[mechanic_atd_heatmap_calculation]]`
- **Related File:** `atd/pkg/exploration/heatmap.go`

## EXPECTATION
1. Atoms with 15 dependents must report as HOT for dependency.
2. Atoms with `heatmap: none` must report as OPTIMAL/STABLE regardless of metrics.
