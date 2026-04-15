---
id: api_atd_serve_crawl
human_name: "MCP Tool: atd_crawl"
description: "Crawl ATD docs and source code. Returns a dependency graph JSON. Set gaps=true to list STABLE atoms with no implementations."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_crawl
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_crawl

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_crawl`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Crawl ATD docs and source code. Returns a dependency graph JSON. Set gaps=true to list STABLE atoms with no implementations.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "src": {
      "type": "string",
      "description": "Path to source code directory (optional)."
    },
    "gaps": {
      "description": "If true, return only orphaned STABLE atoms."
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
When the MCP client sends a `tools/call` request for `atd_crawl`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
