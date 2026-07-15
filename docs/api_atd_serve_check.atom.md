---
id: api_atd_serve_check
human_name: "MCP Tool: atd_check"
description: "Unified coverage report: lists impl links (@spec-link) and test links (@test-link) for every atom touched by the current diff or the full project, with optional LLM-based semantic compliance checking."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_check
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_check

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_check`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- **Default behavior**: audits uncommitted changes via `git diff`, extracting `@spec-link`/`@test-link` tags touched by the diff.
- **`base`/`target`**: optionally compare a specific commit/ref range instead of the working tree diff.
- **`full`**: audit the entire project's coverage regardless of what has changed.
- **`file`** (+ optional `line`): narrow the check to a specific source file (and line within it).
- **`semantic`**: additionally runs an LLM compliance check per impl link — the linked code is compared against the atom specification and returns PASS/FAIL per link. This consumes tokens via the configured LLM provider.
- **CLI counterpart**: the standalone `atd check` CLI command (see `check_coverage.go`) exposes an overlapping but distinct flag set — `--atom` (check one atom's full coverage), `--file`, `--full`, `--semantic`, `--out` (write report to file), and `--docs` (override docs directory). The MCP tool schema below reflects only the arguments accepted by the `atd_check` MCP tool itself.

## TECHNICAL INTERFACE (The Bridge)
### Description
Unified coverage report: lists impl links (@spec-link) and test links (@test-link) for every atom touched by the current diff or the full project. Default mode audits uncommitted changes (git diff); pass base/target to compare commits; pass full:true to audit the entire project. When semantic:true is added, each impl link is also checked for LLM compliance.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "base": {
      "type": "string",
      "description": "Optional: base commit/ref to compare from (e.g. 'HEAD~5')."
    },
    "target": {
      "type": "string",
      "description": "Optional: target commit/ref to compare to (defaults to working tree)."
    },
    "full": {
      "type": "boolean",
      "description": "Optional: audit the entire project instead of just the diff."
    },
    "file": {
      "type": "string",
      "description": "Optional: target a specific file for verification."
    },
    "line": {
      "type": "integer",
      "description": "Optional: target a specific line for verification (requires 'file')."
    },
    "semantic": {
      "type": "boolean",
      "description": "Optional: add LLM compliance check per impl link (consumes tokens). Returns PASS/FAIL per @spec-link."
    }
  }
}
```

### Required Fields
None — default mode is git-diff driven.

### Usage Examples
- `atd_check()` — coverage report for the current uncommitted diff.
- `atd_check(base="HEAD~5")` — coverage report comparing against a specific ref.
- `atd_check(full=true)` — coverage report for the whole project.
- `atd_check(file="src/foo.go", line=42)` — coverage for a specific file/line.
- `atd_check(semantic=true)` — add LLM-based compliance check per `@spec-link`.

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_check`, all arguments are optional; it must return a JSON-RPC response with block texts reporting impl/test link coverage for the selected scope (diff, ref range, single file, or full project).
