# Issue: Update ATD Exploration to Use .gitignore

**ID:** `20260402_refactor_atd_exploration_gitignore`
**Ref:** `ISS-059`
**Date:** 2026-04-02
**Severity:** Medium
**Status:** Open
**Component:** `scripts/pkg/exploration`
**Affects:** All ATD operations (crawl, trace, assemble, etc.)

---

## Summary

The current crawler and exploration methods used by the ATD tooling ignore hidden files and "vendor" repositories via hardcoded regex checks, but they do not actively parse or strictly adhere to standard `.gitignore` rules in the repository. As a result, the tool crawls unnecessary files, runs slower, and may produce false positives. This issue tracks the requirement to adopt `.gitignore` rule parsing within the newly extracted `pkg/exploration` libraries.

---

## Technical Description

### Background
The system explores sources using `filepath.Walk` and ignores directories containing `/.` or `vendor`. This was sufficient early on but is fragile and non-standard.

### The Problem Scenario
When users have large non-hidden but git-ignored directories (e.g., build artifacts like `dist/`, `.next/`, or auto-generated docs), the ATD engine still wastes cycles processing those files. This drags down CLI responsiveness and could accidentally link to compiled/generated specification files rather than the primary source of truth.

### Where This Pattern Exists Today
The logic is now encapsulated within `scripts/pkg/exploration/exploration.go`, specifically within the `CrawlSrc` and `CrawlDocs` functions. 

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | Medium — users notice CLI slowdowns or noisy output when analyzing complex node/front-end projects. |
| Current mitigant | Hardcoded exclusion filters. |

---

## Recommended Fix

**Short term:** Continue expanding the hardcoded exclusion filters if explicit directories cause extreme slowdowns.
**Medium term:** Utilize an established go package (like `ignore`) inside `pkg/exploration` to filter out files accurately during `filepath.Walk`.
**Long term:** Shift to a streaming or delta-index system that only crawls files actually managed by or known to git (e.g., reading via `git ls-files`).

---

## References

- `scripts/pkg/exploration/exploration.go`
- `issues/README.md`
