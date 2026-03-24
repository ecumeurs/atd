---
id: atd_serve_discover
human_name: "MCP Tool: atd_discover"
description: "Extract architectural intent from an undocumented source file, search the ATD index, and recommend @spec-link tags to apply."
type: TECHNICAL_CONTRACT
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_discover
parents:
  - [[atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_discover

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_discover`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Extract architectural intent from an undocumented source file, search the ATD index, and recommend @spec-link tags to apply.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "Path to the undocumented source file."
    },
    "docs": {
      "type": "string",
      "description": "Override docs directory path."
    }
  }
}
```

### Required Fields
`file`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_discover`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
