---
id: api_atd_serve_heatmap_project
human_name: "MCP Tool: atd_heatmap_project"
description: "Get a project-wide heat map summary across dependency, code, and update layers."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_heatmap_project
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_heatmap_project

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_heatmap_project`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- `layer` (optional) selects one heat layer to report: `dependency`, `code`, `updates`, or `all` (default).

## TECHNICAL INTERFACE (The Bridge)
### Description
Get a project-wide heat map summary.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "layer": {
      "type": "string",
      "description": "Heat layer: dependency, code, updates, all (default)."
    }
  }
}
```

### Required Fields
None

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_heatmap_project`, it must return a project-wide heat map summary for the requested layer (or all layers if `layer` is omitted).
