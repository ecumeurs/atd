---
id: atd_serve_weave
human_name: "MCP Tool: atd_weave"
description: "Populate the dependents[] array in ATD atoms by scanning parents references. Bi-directional link weaving."
type: TECHNICAL_CONTRACT
version: 1.0
status: STABLE
priority: CORE
tags:
  - mcp
  - tool
  - atd_weave
parents:
  - [[atd_serve]]
dependents: []
---

# MCP Tool: atd_weave

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_weave`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Populate the dependents[] array in ATD atoms by scanning parents references. Bi-directional link weaving.

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
When the MCP client sends a `tools/call` request for `atd_weave`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
