# Issue: atd_verify MCP tool fails with git diff error

**ID:** `20260325_atd_verify_mcp_git_diff_failure`
**Ref:** `ISS-051`
**Date:** 2026-03-25
**Severity:** High
**Status:** Open
**Component:** `scripts/cmd/atd/cmd/verify.go`
**Affects:** ATD MCP Server, Verify Stage Workflow

---

## Summary

The `atd_verify` tool fails when invoked via the MCP server with the error: `error running git diff. Ensure you are in a git repository: exit status 129`. This prevents the automated compliance auditing of code changes against ATD atoms in agentic environments.

---

## Technical Description

### Background

The `atd verify` command is designed to run `git diff`, extract impacted `@spec-link` tags, and produce a structured audit prompt for compliance review. It must be run within a git repository to function.

### The Problem Scenario

1.  The ATD MCP server is running in a workspace that is a valid git repository.
2.  An agent calls the `atd_verify` MCP tool.
3.  The MCP tool handler invokes the internal `verify` logic.
4.  The `verify` logic attempts to execute `git diff` using `os/exec`.
5.  `git` returns exit status 129, indicating it cannot find the git repository or the command is malformed in the current environment context.

```bash
# Observed Error
Encountered error in step execution: error executing cascade step: CORTEX_STEP_TYPE_MCP_TOOL: error running git diff. Ensure you are in a git repository: exit status 129
```

### Where This Pattern Exists Today

- `scripts/cmd/atd/cmd/verify.go`: The implementation of the `verify` command which calls `git diff`.
- `scripts/cmd/atd/cmd/mcp_tools.go`: The entry point for the MCP tool which may not be correctly passing the working directory or environment to the verify logic.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High — breaks the core "Verify" loop for AI agents |
| Detectability | High — tool fails immediately with an error message |
| Current mitigant | Run `atd verify` manually via terminal instead of MCP |

---

## Recommended Fix

**Short term:** Ensure the MCP server correctly sets the working directory to the project root before executing commands, or allow passing the workspace path as a parameter to the verify tool.

**Medium term:** Update the `verify` implementation to accept an explicit `SearchPath` or `RepoPath` rather than relying on the current working directory.

**Long term:** Standardize how all MCP tools handle workspace context to ensure consistent behavior across different execution environments.

---

## References

- [scripts/cmd/atd/cmd/verify.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/verify.go)
- [scripts/cmd/atd/cmd/mcp_tools.go](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/mcp_tools.go)
