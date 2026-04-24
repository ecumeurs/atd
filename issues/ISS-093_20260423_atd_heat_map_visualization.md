# Issue: ATD Heat Map Visualization

**ID:** `20260423_atd_heat_map_visualization`
**Ref:** `ISS-093`
**Date:** 2026-04-23
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/pkg/webui`, `scripts/pkg/atd`
**Affects:** WebUI Explorer, ATD CLI/MCP tools

---

## Summary

The ATD documentation system lacks a visual heat map to identify unhealthy documentation patterns. This feature will provide three layers of thermal analysis:
1. **Dependency Heat**: Atoms with excessive or insufficient parent/dependent relationships
2. **Code Heat**: Code files with too many or too few `@spec-link` references
3. **Update Heat**: Atoms that are being frequently modified (potential instability)

The primary value will be in the WebUI Explorer, which will support 4 toggles to color ATD cards according to thermal state. CLI/MCP tools will provide file-based analysis for specific atoms or code files.

---

## Technical Description

### Background

Currently, there's no visual mechanism to detect:
- Atoms with excessive coupling (too hot) or isolation (too cold)
- Code files that are overloaded with documentation or under-documented
- Documentation thrashing (frequent consecutive updates indicating instability)

### Heat Map Layers

#### Layer 1: Atom Dependency Heat

| State | Parents | Dependents | Color |
|-------|---------|------------|-------|
| Too Cold | 0 | 0 | ❄️ Pale Blue |
| Optimal | 1-3 | 1-3 | ✅ Green |
| Getting Hot | 4-10 | 4-10 | ⚡ Orange |
| Too Hot | 10+ | 10+ | 🔴 Red |

**Special Rules:**
- **IMPLEMENTATION layer**: May have 0 dependents (leaf nodes acceptable)
- **CUSTOMER layer**: May have 0 parents (top-level requirements acceptable)

#### Layer 2: Code File Heat

| State | Atoms Linked | Color |
|-------|--------------|-------|
| Too Cold | 0-1 | ❄️ Pale Blue |
| Optimal | 2-5 | ✅ Green |
| Getting Hot | 6-10 | ⚡ Orange |
| Too Hot | 10+ | 🔴 Red |

**Measured by**: Count of distinct `@spec-link [[atom_id]]` tags per code file

#### Layer 3: Recent Update Heat

Based on ATD updates within the last 20 commits. Consecutive updates are cumulative.

| State | Recent Updates | Color |
|-------|----------------|-------|
| Cold (Good) | 0 | ⚪ White/Grey |
| Optimal | 1-3 | ✅ Green |
| Getting Hot | 4-6 | ⚡ Orange |
| Too Hot | 7+ | 🔴 Red |

### WebUI Explorer Integration

The Explorer view will have 4 toggle buttons in the toolbar:

```
[None] [Dep Heat] [Code Heat] [Update Heat]
```

**Behavior:**
- **None**: Default view, no heat map coloring
- **Dep Heat**: Colors ATD cards based on dependency heat (Layer 1)
- **Code Heat**: Colors ATD cards based on code file heat (Layer 2)
  - Requires computing heat for each code file that links to the atom
  - Atom takes the maximum heat of all its linked code files
- **Update Heat**: Colors ATD cards based on update frequency (Layer 3)

**Card Styling:**
- Thermal state applied as a colored border or subtle background tint
- Tooltip on hover showing detailed metrics (e.g., "Parents: 12, Dependents: 3 - TOO HOT")
- Badge indicator in corner for hot items (⚡ or 🔴)

### CLI/MCP Tools (File-based)

#### CLI Commands
```bash
# Heat map for a specific atom
atd heatmap atom docs/mechanic_atom.atom.md

# Heat map for a code file (analyzes @spec-link count)
atd heatmap code upsilonapi/controllers/auth.go

# Heat map for the entire project
atd heatmap project --layer=dependency
atd heatmap project --layer=code
atd heatmap project --layer=updates
atd heatmap project --all

# Generate a summary report
atd heatmap report --output=heatmap.json
```

#### MCP Tool
```typescript
// Heat map for specific atom
atd_heatmap(atom: string): HeatMapResult

// Heat map for code file
atd_heatmap_code(path: string): HeatMapResult

// Project-wide heat map
atd_heatmap_project(layer: HeatLayer): HeatMapSummary

// Types
type HeatLayer = "dependency" | "code" | "updates" | "all"
type HeatState = "cold" | "optimal" | "warm" | "hot" | "stable"

interface HeatMapResult {
  atom: string
  layer: string
  state: HeatState
  metrics: {
    parents: number
    dependents: number
    codeFilesLinked: number
    recentUpdates: number
    lastUpdated: string
  }
  recommendations: string[]
}
```

### Where This Pattern Exists Today

- `scripts/pkg/webui/static/js/explorer.js`: ATD card rendering (needs heat map styling)
- `scripts/pkg/atd/pkg/atom/atom.go`: Atom data structure (needs heat metrics)
- `scripts/pkg/atd/pkg/exploration/graph.go`: Dependency graph (for parent/dependent counting)
- `scripts/cmd/atd/cmd/root.go`: CLI command registration

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High — requested feature for documentation health |
| Impact if triggered | Medium — improves documentation quality awareness |
| Detectability | High — visual indicators in Explorer and CLI output |
| Current mitigant | Manual review of `atd trace` and `git log` for heat analysis |

---

## Recommended Fix

**Phase 1: Backend Heat Calculation**
1. Add heat metrics to `atom.AtomData` struct
2. Implement heat calculation logic in `exploration` package:
   - `CalculateDependencyHeat(atom)`: Count parents and dependents
   - `CalculateCodeHeat(atom)`: Count code files via `@spec-link` index
   - `CalculateUpdateHeat(atom)`: Analyze git log for last 20 commits
3. Cache heat metrics to avoid recomputation

**Phase 2: CLI Implementation**
1. Add `atd heatmap` command group
2. Implement subcommands: `atom`, `code`, `project`, `report`
3. Output formatted tables with color-coded thermal states

**Phase 3: MCP Tool**
1. Implement `atd_heatmap()` MCP function
2. Support atom, code file, and project-wide queries
3. Return structured JSON with metrics and recommendations

**Phase 4: WebUI Integration** (Primary Value)
1. Add 4 toggle buttons to Explorer toolbar
2. Implement heat state styling in CSS classes (`.heat-cold`, `.heat-optimal`, `.heat-warm`, `.heat-hot`, `.heat-stable`)
3. Update `createAtomCard()` to apply thermal styling based on active toggle
4. Add tooltips showing detailed heat metrics
5. Add visual badges for warm/hot items

**Phase 5: Refinement**
1. Add heat map legend to Explorer
2. Implement click-to-filter by thermal state
3. Add export of heat map report to CSV/JSON

**Phase 6: Audit Integration**
Update audit so that it takes heatmap result into account.

---

## References

- [ISS-068](file:///home/bastien/work/skill/issues/ISS-068_20260403_webui_atd_health_indicators.md) - Related health indicators work
- [explorer.js](file:///home/bastien/work/skill/scripts/pkg/webui/static/js/explorer.js) - ATD card rendering
- [graph.go](file:///home/bastien/work/skill/scripts/pkg/atd/pkg/exploration/graph.go) - Dependency graph logic
