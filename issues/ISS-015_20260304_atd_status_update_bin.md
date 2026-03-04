# Issue: ATD Status Update Uses Token-Heavy LLM Rewrites

**ID:** `20260304_atd_status_update_bin`
**Ref:** `ISS-015`
**Date:** 2026-03-04
**Severity:** Medium
**Status:** Open
**Component:** `scripts/atd-update` or equivalent
**Affects:** LLM Context / Token Usage

---

## Summary

When asking the LLM to update the status of an ATD file, the current process relies on the LLM to rewrite the entire file to alter the YAML header. This is unnecessarily token-heavy, costly, and risks unintended modifications. We need a dedicated binary or script that handles updating the frontmatter/YAML header directly in ATD files, bypassing the need for the LLM to rewrite the contents.

---

## Technical Description

### Background
The ATD files rely on a strictly structured YAML frontmatter block for metadata like `status`, `version`, `priority`, etc.

### The Problem Scenario
1. The user asks the LLM to update the status of an Atom from `[DRAFT]` to `[STABLE]`.
2. The LLM processes the request, opens the file, and rewrites the content.
3. This consumes a significant amount of tokens simply to update one or two values in the frontmatter, creating a risk of losing context or corrupting the file body.

### Where This Pattern Exists Today
This pattern is currently present in workflows involving the LLM editing or updating ATD tags, metadata, or status natively. 

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium (Token bloat and potential for unintended changes by LLM during rewrite) |
| Detectability | High (Visible in logs and context usage overhead) |
| Current mitigant | None |

---

## Recommended Fix

**Short term:** Write a utility binary or script (e.g., `atd_update_header`) that parses the file, safely updates specific key-value pairs strictly within the YAML frontmatter section, and saves it while preserving all content below the frontmatter intact.
**Medium term:** Ensure the LLM refers to and delegates to this component/binary whenever an ATD property change is requested.
**Long term:** Standardize this binary's usage across all tools/agents that touch ATD properties.

---

## References

- ATD structure specifications
