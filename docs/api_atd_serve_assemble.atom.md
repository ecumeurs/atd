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
parents:
  - [[api_atd_mcp_ops]]
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
      "description": "Comma-separated list of root Atom IDs to begin assembly from."
    },
    "intent": {
      "type": "string",
      "description": "The intent the LLM should focus on (e.g., summarize, executive summary). Defaults to 'Executive Summary'."
    },
    "length": {
      "type": "string",
      "description": "Length constraint: 'short', 'default', 'extended', 'long'."
    },
    "structured": {
      "type": "boolean",
      "description": "If true, group atoms by layer and perform multi-pass summarization."
    },
    "json": {
      "type": "boolean",
      "description": "If true, outputs the result as a structured JSON object along with involved atoms metadata."
    }
  },
  "required": ["starts"]
}
```

### Required Fields
`starts`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_assemble`, it must furnish the required arguments above. It will receive either plain text or a JSON string (if `--json` is active) containing the gathered and summarized information layer-by-layer.
