---
id: atd_serve_verify
human_name: "MCP Tool: atd_verify"
description: "Run git diff, extract @spec-link tags, and produce an audit prompt for the IDE Agent."
type: TECHNICAL_CONTRACT
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_verify
parents:
  - [[atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_verify

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_verify`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Run git diff, extract @spec-link tags, and produce an audit prompt for the IDE Agent.

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
When the MCP client sends a `tools/call` request for `atd_verify`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
