# Issue: ATD Creation Deficiencies (Task 03)

**ID:** `20260314_atd_creation_deficiencies`
**Ref:** `ISS-025`
**Date:** 2026-03-14
**Severity:** Medium
**Status:** Open
**Component:** `scripts/cmd/atd/update`
**Affects:** `Tests/03_atd_creation.md`

---

## Summary

The ATD creation process via `atd update` is currently insufficient. It produces sparse ATD files (often "oneliners") where only the `intent` field is populated, while critical metadata fields like `parents`, `dependents`, and `version` are missing or defaulted to empty. Furthermore, naming conventions for IDs and filenames are not consistently enforced during creation.

---

## Technical Description

### Background

The `atd update` command is intended to create or modify ATD atoms. A complete ATD atom should include a full YAML frontmatter and structured markdown sections (INTENT, LOGIC, INTERFACE, EXPECTATION).

### The Problem Scenario

During Task 03 of the test suite, the following deficiencies were observed:

1.  **Missing Metadata:** The `atd update` command does not populate `parents`, `dependents`, or `version` fields, even when creating a new file.
2.  **Sparse Content:** Only the `intent` field was populated, resulting in extremely minimal documentation.
3.  **Naming Convention Violation:** The requested naming convention `<type>_<camel_human_name>` for IDs was not strictly followed or suggested by the tool.
4.  **Template Requirement:** The tool initially failed to update files that did not already have a valid YAML block, requiring a manual workaround.

### Where This Pattern Exists Today

- Files in `upsilonbattle/docs/` created during Task 03:
    - [battle_arena.atom.md](file:///home/bastien/work/skill/upsilonbattle/docs/battle_arena.atom.md)
    - [ruler.atom.md](file:///home/bastien/work/skill/upsilonbattle/docs/ruler.atom.md)
    - [ruler_turn_flow.atom.md](file:///home/bastien/work/skill/upsilonbattle/docs/ruler_turn_flow.atom.md)

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium — leads to low-quality/incomplete documentation |
| Detectability | High — evident in generated `.atom.md` files |
| Current mitigant | Manual template initialization before running `atd update` |

---

## Recommended Fix

**Short term:** Update `atd update` to include default values for `version`, `parents`, and `dependents` when creating new files. Enforce the `id` naming convention in the `atd update` logic.

**Medium term:** Improve the `atd update` command to support populating more sections (LOGIC, INTERFACE, etc.) via flags or guided prompts.

**Long term:** Integrate a more robust templating engine into the `atd` binary that ensures all mandatory fields are present and valid before saving.

---

## References

- [Tests/03_atd_creation.md](file:///home/bastien/work/skill/Tests/03_atd_creation.md)
- [scripts/cmd/atd/update.go](file:///home/bastien/work/skill/scripts/cmd/atd/update.go)
