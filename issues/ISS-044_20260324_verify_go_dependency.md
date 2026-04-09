# Issue: ATD Verify Tool is Hardcoded to Go Testing

**ID:** `20260324_verify_go_dependency`
**Ref:** `ISS-044`
**Date:** 2026-03-24
**Severity:** High
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/verify.go`
**Affects:** Non-Go projects using ATD framework

---

## Summary

The `atd verify` tool currently has a hard dependency on Go, specifically executing `go test` under the hood. This hardcoded assumption makes the verification process broken or inappropriate for projects built with other languages. We need a flexible approach where the Agent LLM determines how testing should be handled or relies on repository-level `.atd` configuration, before delegating the test execution back to the MCP/ATD tool.

---

## Technical Description

### Background
The `atd verify` command is designed to validate atomic traceable documentation against the source repository's tests to ensure requirements and expectations are correctly verified.

### The Problem Scenario
1. **Langauge Lock-in:** When running `atd verify` on a non-Go project (e.g., Python, PHP, JS), the command fails or executes irrelevantly because it attempts to invoke `go test`.
2. **CI Diff Constraints:** In Continuous Integration (CI) environments, simply fetching current git changes might be insufficient. The tool must be able to compare changes between two specific git references (e.g., base branch vs. PR branch) rather than just assuming a working tree state.

### Where This Pattern Exists Today
The hardcoded dependency resides in the execution logic within `scripts/cmd/atd/cmd/verify.go`. 

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (for any non-Go codebase) |
| Impact if triggered | High (blocks verification entirely) |
| Detectability | High — verify command outright fails or complains about missing test setup. |
| Current mitigant | None structure-wise for non-Go repositories. |

---

## Recommended Fix

**Short term:**
- Delegate test discovery to the Agent LLM, asking it how to handle the testing execution, and feeding those instructions to the MCP / ATD tool.
- Pass explicit git references to facilitate correct diffing in CI cases.

**Medium/Long term:**
- Introduce a `tests` structure within the root `.atd` config file.
- Store commands mappings to handle complex, multi-language/framework projects (e.g., `*.go` triggers `go test`, `*.php` triggers `php artisan test`).
- allow `.atd` configuration file to store multiple test commands and strategies for file types & langages. 
- Modify `atd verify` to either test these means sequentially or rely on tags to dynamically discover which testing methodology to apply based on the files being verified.
- Ensure the MCP tool `atd_verify` is updated to allow the agent to specify the testing methodology to apply based on the files being verified.

---

## References

- `scripts/cmd/atd/cmd/verify.go`
