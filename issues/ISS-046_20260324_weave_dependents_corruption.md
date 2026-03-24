# Issue: Weave Regex Corrupts Dependents Field

**ID:** `20260324_weave_dependents_corruption`
**Ref:** `ISS-046`
**Date:** 2026-03-24
**Severity:** Critical
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/weave.go`
**Affects:** All `.atom.md` files processed by `atd weave`

---

## Summary

The `atd weave` command corrupts the `dependents` YAML field on every run, producing malformed arrays with triple brackets and duplicated entries (e.g., `dependents: [[[service_atd_search]]]]]]]]]]]]]]]`). This is caused by a non-greedy regex that fails to match the full `dependents: [...]` line when it contains `[[...]]` references.

---

## Technical Description

### Background
`atd weave` scans all atoms for `parents` references and then overwrites the `dependents` field in each parent atom file. The expected output format is `dependents: [[[child_a]], [[child_b]]]`.

### The Problem Scenario
The regex on line 72 of `weave.go`:
```go
dependentsRegex := regexp.MustCompile(`(?m)^dependents:\s*\[.*?\]`)
```

Uses `.*?` (non-greedy), which stops at the **first** `]` it encounters. In a value like `dependents: [[[service_atd_search]]]`, the first `]` appears inside the `[[service_atd_search]]` reference, so the regex only matches `dependents: [[[atd_search]` — leaving the trailing `]]` in the file.

On replacement, the new value `dependents: [[[service_atd_search]]]` is inserted **before** the leftover `]]`, producing: `dependents: [[[service_atd_search]]]]]`.

Each subsequent `atd weave` run compounds the problem further, adding more trailing brackets and duplicate entries.

### Where This Pattern Exists Today
- `weave.go:72` — the core regex
- 7 atoms confirmed corrupted: `atd_index`, `atd_philosophy`, `atd_structure`, `atd_audit`, `atd_config`, `atd_check`, `atd_usage_protocol`

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High — triggers on every `atd weave` run |
| Impact if triggered | Critical — produces invalid YAML, breaks parsing tools |
| Detectability | High — visible in file contents |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** Fix the regex to greedily match the entire `dependents: [...]` line, accounting for nested `[[...]]` brackets:
```go
dependentsRegex := regexp.MustCompile(`(?m)^dependents:\s*\[.*\]`)
```
Using greedy `.*` ensures the match extends to the **last** `]` on the line, correctly capturing the full value.

**Medium term:** After fixing the regex, run `atd weave` once to repair all corrupted atoms. Alternatively, add a cleanup step that normalizes the dependents field before replacement.

**Long term:** Replace regex-based YAML mutation with a proper YAML parser (e.g., `gopkg.in/yaml.v3`) for all frontmatter operations.

---

## References

- [weave.go](scripts/cmd/atd/cmd/weave.go)
- [atom/parse.go](scripts/pkg/atom/parse.go)
