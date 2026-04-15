---
id: api_atd_serve_dissect
human_name: "MCP Tool: atd_dissect"
description: "Dissect a source code or documentation file into atomic boundaries. Set llm=true to route through Ollama; false returns the prompt for IDE Agent passthrough."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_dissect
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_dissect

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_dissect`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Dissect a source code or documentation file into atomic boundaries. Set llm=true to route through Ollama; false returns the prompt for IDE Agent passthrough.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "Path to the file to dissect."
    },
    "llm": {
      "description": "If true, route through tiered Ollama provider. Defaults to false (IDE Agent passthrough)."
    }
  }
}
```

### Required Fields
`file`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_dissect`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
