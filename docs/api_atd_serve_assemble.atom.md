---
id: api_atd_serve_assemble
human_name: "MCP Tool: atd_assemble"
description: "Stitch ATD atoms together into a single narrative or technical document."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_assemble
parents: [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_assemble

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_assemble`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Stitch ATD atoms together into a single narrative or technical document.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "starts": {
      "type": "string",
      "description": "Comma-separated list of Root Atom IDs (source atoms)."
    },
    "purpose": {
      "type": "string",
      "description": "Optional purpose to wrap the output in <System Objective> tags."
    },
    "snapshot": {
      "description": "If true, delegate narrative generation to IDE Agent (returns a task ID)."
    },
    "theme": {
      "type": "string",
      "description": "Theme for snapshot narrative (defaults to 'Executive Summary')."
    },
    "docs": {
      "type": "string",
      "description": "Override docs directory path."
    }
  }
}
```

### Required Fields
`starts`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_assemble`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
