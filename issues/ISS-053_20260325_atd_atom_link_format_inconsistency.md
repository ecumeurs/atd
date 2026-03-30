# Issue: ATD Atom Link Format Inconsistency

**ID:** `20260325_atd_atom_link_format_inconsistency`
**Ref:** `ISS-053`
**Date:** 2026-03-25
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/cmd/atd/cmd/weave.go`
**Affects:** All `.atom.md` files, VS Code ATD Linker extension

---

## Summary

The `parents` and `dependents` fields in ATD atom files sometimes use a triple-bracketed string format (e.g., `dependents: [[[atom_id]]]`) instead of a proper YAML list of double-bracketed links (e.g., `- [[atom_id]]`). This inconsistency breaks VS Code extension logic that expects standard ATD link formats, preventing developers from following links between atoms.

---

## Technical Description

### Background

The ATD specification requires `parents` and `dependents` to be YAML lists for multi-value support and consistency with other Markdown link patterns used in the system. Standard tools and extensions (like the VS Code ATD Linker) rely on the `[[atom_id]]` format within these lists to provide navigation and traceability.

### The Problem Scenario

When `atd weave` or other automated tools update these fields, they sometimes produce a format that uses triple brackets or fails to use the list structure.

1.  A tool (e.g., `atd weave`) identifies a relationship.
2.  Instead of appending to a true YAML list, it replaces the field with `[[[ atom_id ]]]`.
3.  The VS Code extension reads this as `[atom_id` (incorrectly stripping only two brackets) or fails to recognize it as a valid link at all.
4.  Navigation between atoms in the IDE is broken.

```yaml
# Expected Format (Standard)
dependents:
  - [[atom_id_1]]
  - [[atom_id_2]]

# Observed Format (Buggy)
dependents: [[[atom_id_1]]]
```

### Where This Pattern Exists Today

- `ISS-046` documents a related corruption in `weave.go` that produces trailing brackets and malformed fields.
- Multiple atoms in `docs/` may currently be affected by this formatting issue.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High — persistent in many existing atoms |
| Impact if triggered | Medium — breaks dev experience and IDE navigation |
| Detectability | High — clearly visible in YAML frontmatter and IDE failure |
| Current mitigant | Manual fix of the YAML format in affected atoms |

---

## Recommended Fix

**Short term:** Standardize the expected format in `ATD.md` and ensure all manual edits follow the YAML list pattern with `[[atom_id]]`.

**Medium term:** Fix the `atd weave` tool (related to `ISS-046`) to ensure it always produces or maintains a true YAML list format. Update the regex to support list items.

**Long term:** Migrate to a full YAML parser for all ATD frontmatter manipulations to avoid regex-based formatting errors.

---

## References

- [ATD.md](file:///home/bastien/work/skill/ATD.md)
- [ISS-046_20260324_weave_dependents_corruption.md](file:///home/bastien/work/skill/issues/ISS-046_20260324_weave_dependents_corruption.md)
- [weave.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/weave.go)

## Change Log
- **2026-03-25**: Standardized `parents` and `dependents` fields to use bullet-point lists. Updated both `atd weave` and `atd update` to robustly handle this format, ensuring compatibility with the VS Code ATD Linker extension.
