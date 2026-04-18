# Issue: ATD Link Resolution Failure

**ID:** `20260418_atd_link_resolution_failure`
**Ref:** `ISS-073`
**Date:** 2026-04-18
**Severity:** Critical
**Status:** Open
**Component:** `scripts/pkg/`, `scripts/cmd/atd/cmd/crawl.go`, `scripts/cmd/atd/cmd/trace.go`
**Affects:** Code-Atom Traceability, Dependency Graph Accuracy, Blast Radius Analysis

---

## Summary

ATD link resolution (`atd_crawl` and `atd_trace`) fails to detect existing @spec-link relationships in code, resulting in empty code_links arrays despite 421 @spec-link occurrences across 250 files. This breaks traceability verification and impact analysis features.

---

## Technical Description

### Background
Link resolution should parse code files to find @spec-link [[atom_id]] tags and build the dependency graph between atoms and code implementations.

### The Problem Scenario
1. **Multi-Language Parsing Failure**: Link parser may not handle different comment syntax for Go, PHP, JavaScript/Vue
2. **Path Resolution Issues**: Path resolution between docs directory and code directories may be broken
3. **Regular Expression Problems**: Pattern matching for @spec-link tags may be incorrect or too restrictive
4. **Empty Results**: `atd_crawl` and `atd_trace` return empty code_links despite extensive tag usage

### Where This Pattern Exists Today
- `scripts/pkg/parser/`: Link parsing logic
- `scripts/cmd/atd/cmd/crawl.go`: Dependency graph building
- `scripts/cmd/atd/cmd/trace.go`: Atom tracing functionality
- File extension handling across multiple languages

### Evidence from Investigation
- **Expected**: Parse 421 @spec-link tags across Go, PHP, JavaScript, Vue files
- **Actual**: Empty code_links arrays in trace results
- **Impact**: Unable to verify implementation coverage, broken blast radius analysis

### High-Density Link Files with Proper @spec-link Tags:
- `upsilonbattle/battlearena/ruler/ruler.go`: 21 tags
- `battleui/app/Http/Controllers/API/AuthController.php`: 12 tags
- `battleui/app/Http/Controllers/API/MatchMakingController.php`: 11 tags
- `battleui/app/Http/Controllers/API/ProfileController.php`: 11 tags

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Confirmed via investigation - tags exist but aren't detected) |
| Impact if triggered | Critical (Complete failure of core ATD traceability features) |
| Detectability | High (Empty results in atd_trace and atd_crawl) |
| Current mitigant | Manual grep searches for @spec-link tags as workaround |

---

## Recommended Fix

**Short term**: Fix link parsing regex patterns to handle multi-language comment syntax. Verify path resolution between docs and code directories works correctly.

**Medium term**: Implement language-specific link parsing with proper handling of Go (`//`), PHP (`/* */`), JavaScript (`//`, `/* */`), and Vue (`//`) comment styles. Support both inline and block comments.

**Long term**: Add robust error handling and logging for link parsing failures. Implement line number tracking for precise link location mapping.

---

## References

- [Traceability Analysis](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/traceability_analysis.md)
- [Actual Traceability State](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/actual_traceability_state.md)
- [Current crawl implementation](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/crawl.go)
- [Current trace implementation](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/trace.go)