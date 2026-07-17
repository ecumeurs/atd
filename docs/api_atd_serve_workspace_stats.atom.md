---
id: api_atd_serve_workspace_stats
human_name: "MCP Tool: atd_workspace_stats"
description: "Aggregate documentation health metrics across all projects in the workspace."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_workspace_stats
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_workspace_stats

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_workspace_stats`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- Shorthand for `atd_stats` with `workspace:true` always forced on -- aggregates metrics across every project in the active workspace rather than just the current one.
- Takes no parameters.

## TECHNICAL INTERFACE (The Bridge)
### Description
Aggregate documentation health metrics across all projects in the workspace.

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
When the MCP client sends a `tools/call` request for `atd_workspace_stats`, it must return a `StatsReport` aggregated across every project in the active workspace, equivalent to `atd_stats(workspace=true)`.
