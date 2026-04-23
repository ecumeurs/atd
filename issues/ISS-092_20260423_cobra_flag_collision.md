# Issue: Cobra Flag Collision on Serve Command

**ID:** `20260423_cobra_flag_collision`
**Ref:** `ISS-092`
**Date:** 2026-04-23
**Severity:** High
**Status:** Resolved
**Component:** `atd/cmd/atd/cmd/serve.go`
**Affects:** MCP Server startup (`atd serve`)

---

## Summary

The MCP server fails to start due to a panic caused by a cobra flag collision. The `serve` command attempts to use the `-p` shorthand for the `--port` flag, which collides with the global `--project` flag's `-p` shorthand defined in `root.go`.

---

## Technical Description

### Background
The ATD CLI provides a `serve` command to start an MCP server. This command has a `--port` flag to define the HTTP port if `--http` is used. Additionally, the CLI has a global `--project` flag to specify the workspace project.

### The Problem Scenario
When the `serve` command was initialized, it registered the `--port` flag with the `-p` shorthand. Recently, workspace support was introduced, adding a global `--project` flag to `root.go` which also uses the `-p` shorthand. During command initialization, Cobra merges the persistent flags from `root.go` with the local flags of `serve.go`, leading to a panic: `unable to redefine 'p' shorthand in "serve" flagset: it's already used for "port" flag`.

### Where This Pattern Exists Today
The issue was specifically located in `atd/cmd/atd/cmd/serve.go` around line 67 where the port flag was defined using `IntP`.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High |
| Detectability | High — causes immediate panic on `atd serve` execution |
| Current mitigant | Removed the shorthand |

---

## Recommended Fix

**Short term:** Remove the shorthand from the `--port` flag in `serve.go`. (Fixed).  
**Medium term:** Introduce stricter flag registration policies to avoid single-letter collisions.  
**Long term:** Implement automated tests for all commands to verify successful initialization without flag collisions.

---

## References

- [serve.go](../atd/cmd/atd/cmd/serve.go)
- [root.go](../atd/cmd/atd/cmd/root.go)
