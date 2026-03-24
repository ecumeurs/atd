# Issue: Crawl Summary and Mermaid Graph Export

**ID:** `20260324_crawl_summary_mermaid`
**Ref:** `ISS-040`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/crawl.go`
**Affects:** Architecture reviews, documentation reports

---

## Summary

The dependency graph from `atd crawl` is a raw JSON blob. There is no human-readable summary of findings and no standard way to render the graph visually. The tool should produce both a structured summary (total atoms, link counts, orphans, coverage) and export the graph in Mermaid format for embedding in Markdown.

---

## Technical Description

### Background
`atd crawl` outputs raw JSON with atoms, parent/dependent edges, and code links. Architects must parse this manually.

### The Problem Scenario
1. An architect runs `atd crawl` before a design review.
2. The output is a large JSON blob. They need to manually count atoms, identify orphans, and draw diagrams.
3. There is no way to generate a visual graph for the review presentation.

### Where This Pattern Exists Today
- `scripts/cmd/atd/cmd/crawl.go` — JSON output only.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — architecture reviews are slower |
| Detectability | High — users immediately want visual output |
| Current mitigant | Manual parsing and diagram drawing |

---

## Recommended Fix

**Short term:** Add summary data to `atd crawl` default output: total atoms, link counts, orphan count, coverage percentage, key findings.
**Medium term:** Add `--format mermaid` flag that outputs the dependency graph as a Mermaid diagram, directly embeddable in Markdown documents. Mermaid format is sufficient; Graphviz/DOT is not required.
**Long term:** WebUI should consume this graph data for interactive exploration with filtering by domain, type, and status.

---

## References

- [ATD.md §3.2.9](file:///home/bastien/work/skill/ATD.md)
- [ISS-030](ISS-030_20260323_crawl_map_impact_submode.md)
