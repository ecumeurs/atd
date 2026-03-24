---
id: api_atd_serve_audit
human_name: "MCP Tool: atd_audit"
description: "Audit ATD atoms for documentation bloat and semantic collisions. Can also check code compliance against a specific atom."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_audit
parents:
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_audit

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_audit`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Audit ATD atoms for documentation bloat and semantic collisions. Can also check code compliance against a specific atom.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "docs": {
      "type": "string",
      "description": "Override docs directory path."
    },
    "threshold": {
      "description": "Cosine similarity threshold for collision detection (0.0\u20131.0). Defaults to config value."
    },
    "code": {
      "type": "string",
      "description": "Path to code file for compliance mode (requires 'atom')."
    },
    "atom": {
      "type": "string",
      "description": "Path to atom file for compliance mode (requires 'code')."
    }
  }
}
```

### Required Fields
None

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_audit`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
