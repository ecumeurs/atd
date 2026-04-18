# Issue: ATD Indexing System Failure

**ID:** `20260418_atd_indexing_system_failure`
**Ref:** `ISS-071`
**Date:** 2026-04-18
**Severity:** Critical
**Status:** Open
**Component:** `scripts/pkg/`, `scripts/cmd/atd/cmd/index.go`
**Affects:** ATD System Accuracy, Coverage Reporting, Orphan Detection

---

## Summary

The ATD indexing system fails to scan code directories properly, resulting in critical false positives: reports 214 orphaned atoms (88%) with 0% coverage ratio, while reality shows 421 @spec-link occurrences across 250 files (~82% actual coverage). This causes severe tooling credibility issues.

---

## Technical Description

### Background
ATD indexing should build a searchable database of all code files containing @spec-link tags to enable accurate coverage tracking and orphan detection.

### The Problem Scenario
1. **Incomplete Directory Scanning**: `atd_index` only scanned 28 chunks across 1 file instead of 250+ files
2. **Path Resolution Issues**: Indexer may only scan `docs/` directory, missing code paths (`upsilonapi/`, `upsilonbattle/`, `battleui/`, `upsiloncli/`)
3. **File Extension Filtering**: File extension patterns may be too restrictive, missing `.go`, `.php`, `.js`, `.vue` files
4. **Caching Problems**: Caching mechanism may prevent re-scanning after code changes

### Where This Pattern Exists Today
- `scripts/pkg/indexer/`: Core indexing logic
- `scripts/cmd/atd/cmd/index.go`: CLI command implementation
- `.atd` configuration: Missing comprehensive code paths configuration

### Evidence from Investigation
- **Expected**: Index 250+ files across multiple directories and languages
- **Actual**: Index 28 chunks across 1 file
- **Result**: False orphan reporting (214 vs 8 true orphans)

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Current indexing behavior confirmed via investigation) |
| Impact if triggered | Critical (Complete loss of ATD system credibility and usefulness) |
| Detectability | Medium (Requires manual verification against grep results) |
| Current mitigant | Manual grep searches for @spec-link tags as workaround |

---

## Recommended Fix

**Short term**: Update indexing configuration to include all code directories and file extensions. Force re-indexing with `atd_index --force` to rebuild database.

**Medium term**: Refactor indexer to support configurable scan paths and file patterns via `.atd` configuration. Add verbose logging to show exactly which files are being scanned.

**Long term**: Implement incremental indexing with proper mtime-based cache invalidation to handle large codebases efficiently.

---

## References

- [Traceability Analysis](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/traceability_analysis.md)
- [Actual Traceability State](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/actual_traceability_state.md)
- [Final Summary](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/final_summary.md)
- [Current indexing implementation](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/index.go)