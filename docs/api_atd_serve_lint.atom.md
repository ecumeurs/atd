---
id: api_atd_serve_lint
human_name: "MCP Tool: atd_lint"
type: API
version: 0.1.0
status: STABLE
priority: 3
tags: [atd, mcp, lint, validation]
parents: [[api_atd_mcp_ops]]
  - [[mechanic_atd_lint]]
  - [[service_atd_serve]]
dependents: []
layer: IMPLEMENTATION
---

# MCP Tool: atd_lint

## INTENT
Expose the structural linting capability of the ATD toolkit via the Model Context Protocol (MCP). This allows IDE agents to validate documentation state during automated workflows.

## THE RULE / LOGIC
- Maps the `atd_lint` tool name to the deterministic `runLint` function in the ATD binary.
- **Inputs**:
    - `docs` (optional): Path to the documentation directory to validate. Defaults to the configured docs path.
- **Outputs**:
    - Returns a success message if all atoms pass.
    - Returns a detailed list of linting errors (missing fields, invalid enums, broken links) if any atom fails validation.

## TECHNICAL INTERFACE (The Bridge)
- **Implementation:** `scripts/cmd/atd/cmd/mcp_tools.go`
- **Tool Name:** `atd_lint`
- **Code Tag:** `@spec-link [[api_atd_serve_lint]]`

## EXPECTATION (For Testing)
- Calling the `atd_lint` MCP tool with a valid `docs` path returns a successful validation message.
- Calling the tool on a directory containing invalid atoms returns the structural errors as a string within the MCP response.
