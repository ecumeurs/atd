# Issue: [Short Title]

**ID:** `YYYYMMDD_short_slug`
**Ref:** `ISS-NNN`
**Date:** YYYY-MM-DD
**Severity:** Critical / High / Medium / Low
**Status:** Open / In Progress / Resolved / Wont Fix
**Component:** `path/to/affected/package`
**Affects:** `path/to/callers/or/consumers`

---

## Summary

One paragraph. What is the problem, where does it live, and why does it matter.

---

## Technical Description

### Background

Briefly describe the normal expected behavior of the component.

### The Problem Scenario

Walk through the exact sequence of events that triggers the issue.
Use a code block, ASCII diagram, or numbered step list — not plain prose alone.

```
Actor A                      Actor B
──────                       ──────
example diagram here
```

### Where This Pattern Exists Today

Point to specific files and line numbers where the risk is present or where the pattern is used.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Low / Medium / High |
| Impact if triggered | Low / Medium / High |
| Detectability | Low / Medium / High — explain how it manifests |
| Current mitigant | Any existing guard or workaround |

---

## Recommended Fix

**Short term:** What can be done now without changing architecture (docs, conventions, guard clauses).

**Medium term:** What code change would reduce the risk.

**Long term:** What architectural change would eliminate it entirely.

---

## References

- Link to relevant source files (relative paths from workspace root)
- Link to relevant tests
- Link to external documentation or standards if applicable
