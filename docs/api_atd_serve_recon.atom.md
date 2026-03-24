---
id: api_atd_serve_recon
human_name: "MCP Tool: atd_recon"
description: "Semantic archaeology: validate whether a candidate source file implements a specific ATD atom."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_recon
parents:
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_recon

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_recon`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Semantic archaeology: validate whether a candidate source file implements a specific ATD atom.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "atom": {
      "type": "string",
      "description": "Path to the target .atom.md file."
    },
    "candidate": {
      "type": "string",
      "description": "Path to the candidate source code file to validate."
    }
  }
}
```

### Required Fields
`atom`, `candidate`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_recon`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
