# Issue: Project-Wide Verification Mode for ATD

**ID:** `20260419_project_wide_verify_mode`
**Ref:** `ISS-084`
**Date:** 2026-04-19
**Severity:** Low
**Status:** Open
**Component:** `atd/cmd/atd/cmd/verify.go`
**Affects:** Project health visibility, initial onboarding audits

---

## Summary

The current `atd verify` command is strictly diff-based (uncommitted changes or between two refs). There is no way to perform a "sanity check" on the entire codebase to ensure that *all* existing @spec-link implementations still comply with their linked atoms. This represents a risk of documentation rot where code refactors that weren't audited at the time (due to bypasses or old processes) remain non-compliant.

---

## Technical Description

### Background
`atd verify` uses `git diff --name-only` to seed its file discovery.

### The Problem Scenario
1. A developer wants to run a full compliance audit on the entire project.
2. They run `atd verify`, but it returns "No differences found."
3. They are forced to use a fake `base` ref (like an empty tree) or manually touch every file to trigger the audit.

### Where This Pattern Exists Today
`atd/cmd/atd/cmd/verify.go`: The `runVerify` function lacks a non-git file discovery path.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Low |
| Detectability | Low — documentation rot is silent. |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** 
- Add a `--full` flag to the `verify` command.
- If `--full` is provided, skip the `git diff` call and use `config.GetCodePaths()` and `exploration.Explorer.ListFiles()` to seed the discovery.

**Medium term:** 
- Integrate with the granular audit logic (ISS-083) so that a "full" audit is still surgical and efficient.

---

## References

- [verify.go](file:///home/bastien/work/skill/atd/cmd/atd/cmd/verify.go)
- [config.go](file:///home/bastien/work/skill/atd/config/config.go)
- [exploration.go](file:///home/bastien/work/skill/atd/pkg/exploration/exploration.go)
