# Issue: Implement `map-impact` sub-mode for `atd crawl`

**ID:** `20260323_crawl_map_impact_submode`
**Ref:** `ISS-030`
**Date:** 2026-03-23
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/crawl.go`
**Affects:** `Architect Mode`, `atd crawl-graph consumers`

---

## Summary

The `atd_map_impact` tool is mentioned in the `ATD.md` rules as a "Ripple check" required before committing updates, but it does not exist as a standalone CLI tool or subcommand. This functionality should be implemented as a sub-mode or flag for the `atd crawl` command to streamline impact analysis.

---

## Technical Description

### Background
Currently, `atd crawl` generates a full dependency graph of all atoms and their implementations. While this provides the necessary data for impact analysis, it requires the user or agent to manually filter the graph to find the downstream "ripple effect" of changing a specific atom.

### The Problem Scenario
When an architect updates a high-level atom (e.g., a `DOMAIN` or `MODULE`), they need to know exactly which sub-atoms and code files are impacted.
Running `atd crawl` returns the *entire* graph, which can be overwhelming for large projects.

### Where This Pattern Exists Today
- `scripts/cmd/atd/cmd/crawl.go`: Currently only supports `--gaps`.
- `atd_management_skill/.agent/rules/ATD.md`: References a non-existent `atd_map_impact`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | High — users will notice the tool is missing |
| Current mitigant | Manual traversal of `atd crawl` output |

---

## Recommended Fix

**Short term:** Update documentation to explain how to use `atd crawl` for impact analysis (Done).
**Medium term:** Add a `--target <atom_id>` flag to `atd crawl` that filters the output to show only the target atom and its recursive dependents.
**Long term:** Implement a visual ripple-effect map in the WebUI using this filtered graph data.

---

## References

- [crawl.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/crawl.go)
- [ATD.md](file:///home/bastien/work/skill/atd_management_skill/.agent/rules/ATD.md)
