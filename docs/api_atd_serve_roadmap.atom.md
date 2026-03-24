---
id: api_atd_serve_roadmap
human_name: "MCP Tool: atd_roadmap"
description: "Scan a source directory and build a complexity roadmap JSON identifying high-density files."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_roadmap
parents:
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_roadmap

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_roadmap`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Scan a source directory and build a complexity roadmap JSON identifying high-density files.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "dir": {
      "type": "string",
      "description": "Directory to scan."
    },
    "out": {
      "type": "string",
      "description": "Optional output file path for roadmap.json."
    }
  }
}
```

### Required Fields
`dir`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_roadmap`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
