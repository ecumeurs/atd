---
id: api_atd_serve_test_links
human_name: "MCP Tool: atd_test_links"
description: "Audit @test-link tags in source code to find which atoms are verified by which tests."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_test_links
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_test_links

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_test_links`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Audit @test-link tags in source code to find which atoms are verified by which tests.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "atom": {
      "type": "string",
      "description": "Optional: Filter for a specific Atom ID."
    },
    "docs": {
      "type": "string",
      "description": "Override docs directory path."
    }
  }
}
```

### Required Fields
None

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_test_links`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
