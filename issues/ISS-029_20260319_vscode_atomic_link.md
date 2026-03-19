# Issue: VSCode Extension for Atomic Link Following

**ID:** `20260319_vscode_atomic_link`
**Ref:** `ISS-029`
**Date:** 2026-03-19
**Severity:** Medium
**Status:** Open
**Component:** `vscode-extension`
**Affects:** `developers`

---

## Summary

Currently, developers using the ATD system in VS Code cannot easily navigate between atoms using the `[[atomic_id]]` syntax. A VS Code extension should be created or enhanced to treat these identifiers as clickable links (Ctrl+Click), allowing seamless navigation within the documented architecture.

---

## Technical Description

### Background
The ATD system uses `[[atomic_id]]` syntax in documentation and source code @spec-links to reference atomic units of logic. Standard markdown link following works for file paths, but not for these custom identifiers.

### The Problem Scenario
A developer sees `[[ATD-001]]` in a source file or another atom. To see the definition, they must manually search for the file with that ID, which breaks their flow and slows down architectural exploration.

### Where This Pattern Exists Today
This pattern is universal across all ATD documentation (`docs/**/*.atom.md`) and any source file using `@spec-link` or `@test-link` tags.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | High |
| Current mitigant | Manual search using `atd search` or IDE grep. |

---

## Recommended Fix

**Short term:** A basic VS Code extension that implements `DocumentLinkProvider` to recognize `[[...]]` regex and resolve it to the corresponding `.atom.md` file using the ATD index or simple file search.
**Medium term:** Integrate with the ATD MCP server to provide richer tooltips and navigation.
**Long term:** A full-featured ATD Language Server Protocol (LSP) implementation.

---

## References

- [atd/docs](docs/)
- [atd index](.atd_index.db)
