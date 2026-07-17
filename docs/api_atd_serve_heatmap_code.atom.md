---
id: api_atd_serve_heatmap_code
human_name: "MCP Tool: atd_heatmap_code"
description: "Get heat map metrics for a specific source file based on @spec-link density."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_heatmap_code
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_heatmap_code

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_heatmap_code`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- `file` (required) is a path to a source file; the report is based on the density of `@spec-link` tags found in it.

## TECHNICAL INTERFACE (The Bridge)
### Description
Get heat map metrics for a specific source file based on @spec-link density.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "Source file path."
    }
  }
}
```

### Required Fields
`file`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_heatmap_code` with a valid `file` path, it must return that file's @spec-link density heat metrics.
