# Issue: WebUI Explorer Display Readability

**ID:** `20260401_webui_explorer_readability`
**Ref:** `ISS-058`
**Date:** 2026-04-01
**Severity:** High
**Status:** Resolved
**Component:** `webui/explorer`
**Affects:** `webui/explorer`, `atd/visualization`

---

## Summary

The current WebUI Explorer display (using D3 Treemap) is suffering from severe readability issues. There are too many blocks on the test bed, and block names are often too long in unbroken lines, making the visualization cluttered and difficult to parse. The treemap approach treats all atoms with equal visual weight regardless of their hierarchical importance, burying the architecture in a "sea of boxes."

---

## Technical Description

### Background
The current explorer uses a D3-based treemap to visualize ATD atoms. This was intended to show the distribution of atoms by type or status, where the area of each rectangle represents the count or a similar metric.

### The Problem Scenario
1. **Visual Noise:** High-volume Implementation atoms (MECHANIC, BUILD) clutter the space, overwhelming high-level Customer atoms.
2. **Poor Typography:** Long atom names are not wrapped or truncated effectively within small treemap nodes.
3. **Loss of Hierarchy:** Treemaps are good for part-to-whole relationships but poor at showing the directional flow of intent (Customer -> Architecture -> Implementation).
4. **Information Density:** The "Jungle" view makes it impossible for an architect to see system flow or logic density at a glance.

### Where This Pattern Exists Today
- `webui/src/components/ExplorerTreemap.js` (or similar component)
- `atd/visualizers/treemap.go`

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High |
| Detectability | High — evident upon opening the explorer |
| Current mitigant | None |

---

## Recommended Fix

A complete redesign of the Explorer view moving away from a flat treemap toward a "Map of Intent and Health."

### 1. The "Waterfall of Intent" (Structural Layout)
- **Columnar Lanes:** Divide the screen into three vertical lanes (Customer → Architecture → Implementation).
- **Traceability Lines:** Draw connecting Bezier curves between atoms. Highlight downstream paths on click.
- **Module Grouping:** Cluster atoms inside visual containers to show system boundaries.

### 2. The "Readiness" Heatmap (Visual Signals)
- **Hollow Outline:** Vaporware (no implementation linked).
- **Pulsing Border:** Ready to implement (ancestry present, 0% coverage).
- **Red "Glow":** Logic Drift (atd_verify failure).
- **Solid/Dimmed:** Stable (100% coverage and sync).

### 3. The "Substrata" (Minimizing the Done)
- Move STABLE atoms with 100% coverage into a collapsed "Foundation" section at the bottom.
- Visual Treatment: Condensed micro-tiles or colored dots.

### 4. Search and Summarization
- **Semantic Lookup:** Ctrl+K command palette using `atd_search` (local embeddings).
- **On-the-Fly Summarization:** Cluster summaries using llama3.2 to generate contextually aware descriptions of groups of atoms.

---

## References

- [ATD.md](file:///home/bastien/work/skill/atd/ATD.md)
- [ISS-054_20260330_atd_summarization_feature.md](file:///home/bastien/work/skill/issues/ISS-054_20260330_atd_summarization_feature.md)
