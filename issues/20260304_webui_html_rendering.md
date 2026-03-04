# Issue: WebUI ATD Preview HTML Rendering

**ID:** `20260304_webui_html_rendering`
**Ref:** `ISS-005`
**Date:** 2026-03-04
**Severity:** Low
**Status:** Open
**Component:** `webui`
**Affects:** `webui/app.js`

---

## Summary

The ATD preview in the WebUI is not rendered as HTML. It shows plain, trimmed markdown, which makes it difficult to read and loses the structure of the Atom.

---

## Technical Description

### Background
ATDs are stored as Markdown files with YAML frontmatter.

### The Problem Scenario
The WebUI simply pulls the raw text and strips headers/frontmatter without converting the remaining content to a readable HTML format (e.g., using a markdown parser).

### Where This Pattern Exists Today
- `webui/` preview pane.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Low |
| Detectability | High |
| Current mitigant | Raw text view |

---

## Recommended Fix

**Short term:** Integrate a lightweight markdown library (like `marked`) into the WebUI.
**Medium term:** Add syntax highlighting for code blocks within ATD previews.
**Long term:** Support custom rendering for ATD-specific fields (like formulas or rules).

---

## References

- [webui directory](../webui/)
