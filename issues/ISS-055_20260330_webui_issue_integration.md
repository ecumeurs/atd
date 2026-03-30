# Issue: WebUI Issue Integration

**ID:** `20260330_webui_issue_integration`
**Ref:** `ISS-055`
**Date:** 2026-03-30
**Severity:** Medium
**Status:** Open
**Component:** `webui/static`
**Affects:** `webui/main.go`

---

## Summary

The WebUI currently focuses on ATD visualization but lacks visibility into the project's tracked issues in `/workspace/issues/`. When an `issues` folder is present in the repository, the WebUI should provide a dedicated interface to browse, view, and potentially filter the active issues, similar to how it handles ATDs.

---

## Technical Description

### Background

The `issue_management` skill tracks technical debt, bugs, and risks in Markdown files within the `/issues/` directory. These files follow a strict format and are indexed by `issues/README.md`.

### The Problem Scenario

Users working with the WebUI to understand the system's architecture (via ATDs) are currently disconnected from the known issues and risks associated with that architecture unless they manually check the filesystem.

1.  Open WebUI to explore ATDs.
2.  Identify a component of interest.
3.  No visual indication or tab exists to see if there are open `ISS-NNN` linked or general issues for the project.
4.  User must switch to an IDE or terminal to run `issues` or check the directory.

### Where This Pattern Exists Today

- `webui/main.go` serves ATD data but does not yet scan or serve the `issues` directory.
- `webui/static/index.html` (and associated JS) only has views for ATD graph and details.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Medium |
| Impact if triggered | Medium — reduced situational awareness for developers using the WebUI. |
| Detectability | High — users will notice the missing "Issues" section. |
| Current mitigant | Manual lookup via terminal/IDE. |

---

## Recommended Fix

**Short term:** 
- Update `webui/main.go` to check for the existence of an `issues/` directory at startup.
- If present, serve the list of issues (parsed from the Markdown files or the README index) via a new API endpoint (e.g., `/api/issues`).

**Medium term:**
- Add an "Issues" tab or navigation item to the WebUI.
- Implement a list view for issues, showing Ref, Severity, Status, and Summary.
- Provide a detail view for individual issues.

**Long term:**
- Bidirectional linking: allow navigating from an ATD to its linked issues and vice versa within the WebUI.

---

## References

- [/issues/README.md](file:///home/bastien/work/skill/issues/README.md)
- [webui/main.go](file:///home/bastien/work/skill/webui/main.go)
- [ISS-014_20260304_issue_atd_integration.md](file:///home/bastien/work/skill/issues/ISS-014_20260304_issue_atd_integration.md)
