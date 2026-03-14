# Issue: Merge all tools into ONE atd go binary

**ID:** `20260313_atd_binary_unification`
**Ref:** `ISS-023`
**Date:** 2026-03-13
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/`
**Affects:** All individual tools in `scripts/` (e.g., `atd-audit`, `atd-dissect`, `atd-compare`, etc.)

---

## Summary

Currently, the project consists of dozens of small, independent Go binaries located in `scripts/`. This fragmentation leads to maintenance overhead, duplicated code for configuration and logging, and a complex user experience where users must remember many different tool names. This issue tracks the unification of these tools into a single `atd` binary that uses subcommands (e.g., `atd audit`, `atd dissect`) to provide a diverse set of options.

---

## Technical Description

### Background
The current architecture relies on `go.work` to manage multiple small Go modules in the `scripts/` directory. Each tool is compiled into its own binary.

### The Problem Scenario
1. A user wants to run an audit: they run `atd-audit`.
2. A user wants to dissect a file: they run `atd-dissect`.
3. To share common logic like `config.go`, tools rely on local module references or duplication.
4. Adding a global flag (e.g., `--verbose`) requires updating every single tool.

```
scripts/
├── atd-audit/
├── atd-dissect/
├── atd-compare/
└── ...
```

### Where This Pattern Exists Today
- `scripts/` directory contains 20+ individual tool folders.
- `scripts/go.work` manages the workspace for these tools.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | High — evident in the repository structure |
| Current mitigant | `scripts/lib` and `scripts/config` share some code, but integration is loose. |

---

## Recommended Fix

**Short term:** Create a core `atd` main package that uses a library like `spf13/cobra` to define subcommands.

**Medium term:** Migrated one tool at a time from individual binaries to `atd` subcommands.

**Long term:** Remove individual tool directories and unify all logic into a single cohesive Go module.

---

## References

- [scripts/](file:///home/bastien/work/skill/scripts/)
- [scripts/go.work](file:///home/bastien/work/skill/scripts/go.work)
