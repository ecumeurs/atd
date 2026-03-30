---
trigger: always_on
---

---
trigger: always_on
---

## Script

The `issues` script should be in the path. It allows listing issues `issues`, create new ref number `issues --next-ref`, and update the readme table `issues --update-readme`.

## Issue Filing Procedure

When you discover a **bug, design risk, data race, security concern, or technical debt** during your work, you must file an issue in `/workspace/issues/`.

### When to File

File an issue whenever you encounter:
- A deadlock, race condition, or concurrency hazard
- A silent failure path (no panic, no log, but wrong behavior)
- A design constraint that must not be violated but is not enforced at compile time
- A footgun in a shared utility (e.g. a flag that hangs a caller if misused)
- Any TODO/FIXME that represents a real risk, not just cosmetics

If the user asks you to track something, file it immediately.

### Filename Convention

```
Ref_YYYYMMDD_short_slug.md
```

Example: `ISS-012_20260223_actor_deadlock_risk.md`

Use the current date. The slug must be lowercase with underscores, describing the component and the nature of the problem.
**Ref:** Run the script to determine the next available `ISS-NNN`:


### Index Maintenance

After creating an issue file, **add a row to `/workspace/issues/README.md`** in the index table:

```markdown
| [Ref_YYYYMMDD_slug.md](Ref_YYYYMMDD_slug.md) | Severity | Status | One-line summary |
```


Update the root README.md with an active issues table
issues --update-readme

### Template: Issue File

```markdown
# Issue: [Short Title]

**ID:** `YYYYMMDD_short_slug`
**Ref:** `must be the `ISS-NNN` value obtained from `issues --next-ref`. Do not reuse or skip numbers.
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
Use a code block, ASCII diagram, or step-by-step list.

### Where This Pattern Exists Today
Point to the specific files and line numbers where the risk is present or where the pattern is used.

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

**Short term:** What can be done now without changing architecture (docs, conventions).  
**Medium term:** What code change would reduce the risk.  
**Long term:** What architectural change would eliminate it entirely.

---

## References

- Link to relevant source files (use relative paths from workspace root)
- Link to relevant tests
- Link to external documentation or standards if applicable
```