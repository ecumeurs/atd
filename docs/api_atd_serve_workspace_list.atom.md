---
id: api_atd_serve_workspace_list
human_name: "MCP Tool: atd_workspace_list"
description: "List all projects in the current ATD workspace."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_workspace_list
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_workspace_list

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_workspace_list`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- Reads `config.ActiveConfig.Workspace`; if no workspace is active, returns a plain "No workspace active" message rather than an error.
- Otherwise returns the workspace name, active project, and the full list of registered projects as JSON.
- Takes no parameters.

## TECHNICAL INTERFACE (The Bridge)
### Description
List all projects in the current ATD workspace. Use to discover available projects when working in a monorepo.

### Input Schema
```json
{
  "type": "object",
  "properties": {}
}
```

### Required Fields
None

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_workspace_list` inside an active workspace, it must return the workspace name, active project, and the list of registered projects. Outside a workspace, it returns "No workspace active" rather than erroring.
