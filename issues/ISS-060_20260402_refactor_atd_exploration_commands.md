# Issue: Refactor ATD Commands to Use Unified Exploration Package with Caching

**ID:** `20260402_refactor_atd_exploration_commands`
**Ref:** `ISS-060`
**Date:** 2026-04-02
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd`, `scripts/pkg/exploration`
**Affects:** All ATD operations

---

## Summary

This issue is a follow-up to ISS-059. The goal is to refactor all remaining ATD subcommands to leverage the newly extracted `pkg/exploration` package as a unified front for ATD crawling and codebase exploration. The package itself must be updated to accommodate the diverse needs of all commands. Additionally, a caching mechanism must be introduced into the exploration package to reduce redundant file reads and parsing across operations.

---

## Technical Description

### Background
Currently, some graph logic and crawling functionalities have been extracted into `pkg/exploration` (see ISS-059). However, not all ATD commands fully utilize this unified entry point, and some commands still run fragmented file traversal algorithms. 

### The Problem Scenario
Without a unified exploration package, different ATD commands parse atom files inconsistently. The lack of a centralized caching layer for crawled atoms and explored paths means that navigating large codebases remains slow, as file system crawls are repeated from scratch instead of relying on a shared cache.

### Where This Pattern Exists Today
Various commands within `scripts/cmd/atd/cmd`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium |
| Detectability | Medium — Code duplication visible in commands; noticeable performance delays for large workspaces. |
| Current mitigant | Some commands share `pkg/exploration`, but it is not comprehensive. |

---

## Recommended Fix

**Short term:** Implement an internal caching method within `pkg/exploration` so that multiple requests for the graph or file sets can reuse previously parsed results, improving speed across the tooling ecosystem.
**Medium term:** Identify all ATD commands not currently utilizing `pkg/exploration`. With priority to trace and health-related commands.
**Long term:** Refactor those commands to use `pkg/exploration`. Update the exploration package with whatever new filtering/query options those commands need to work.

---

## References

- `scripts/pkg/exploration/exploration.go`
- `scripts/cmd/atd/cmd/`
- [ISS-059](ISS-059_20260402_refactor_atd_exploration_gitignore.md)
