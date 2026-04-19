# Issue: Decouple Code Compliance from Audit Service

**ID:** `20260419_decouple_audit_code_compliance`
**Ref:** `ISS-085`
**Date:** 2026-04-19
**Severity:** Low
**Status:** Open
**Component:** `atd/cmd/atd/cmd/audit.go`
**Affects:** User clarity on feature sets, maintenance of compliance logic

---

## Summary

The `atd audit` command currently contains a "Code Compliance Mode" (via `--code` and `--atom` flags) that overlaps significantly with the `atd verify` command. This duplication creates confusion about which tool to use for compliance checks and splits the maintenance of audit prompting logic. This issue tracks the removal of code compliance from `atd audit` to keep it focused strictly on ATD-level structural integrity (bloat detection and semantic collisions).

---

## Technical Description

### Background
`atd audit` was originally designed as a Swiss Army knife for both internal documentation health and external code compliance. `atd verify` has since evolved as the primary tool for change-based compliance.

### The Problem Scenario
1. A developer wants to verify a code snippet.
2. They see both `atd verify` and `atd audit --code` in the documentation.
3. They use `atd audit --code`, which lacks the ancestry context and surgical discovery features being added to `atd verify`.
4. Maintenance teams must update two different prompt builders (`prompt.AuditCodeBuild` and the logic in `verify.go`) when compliance rules evolve.

### Where This Pattern Exists Today
- `atd/cmd/atd/cmd/audit.go`: `runCodeAudit` function and associated flags.
- `docs/service_atd_audit.atom.md`: Documentation of the dual modes.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Low |
| Detectability | High |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** 
- Remove `runCodeAudit` from `atd/cmd/atd/cmd/audit.go`.
- Remove `--code` and `--atom` flags from `auditCmd`.
- Update `docs/service_atd_audit.atom.md` to remove references to code compliance.

**Medium term:** 
- Ensure `atd verify` (via ISS-083) covers all use cases that were previously handled by `atd audit --code`.

---

## References

- [audit.go](file:///home/bastien/work/skill/atd/cmd/atd/cmd/audit.go)
- [verify.go](file:///home/bastien/work/skill/atd/cmd/atd/cmd/verify.go)
- [service_atd_audit.atom.md](file:///home/bastien/work/skill/docs/service_atd_audit.atom.md)
