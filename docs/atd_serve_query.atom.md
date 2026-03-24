---
id: atd_serve_query
human_name: "MCP Tool: atd_query"
description: "Search ATD atoms by frontmatter field value. Returns JSON array of matching atoms."
type: TECHNICAL_CONTRACT
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_query
parents:
  - [[atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_query

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_query`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Search ATD atoms by frontmatter field value. Returns JSON array of matching atoms.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "field": {
      "type": "string",
      "description": "Frontmatter field to search (e.g. 'type', 'status', 'id')."
    },
    "search": {
      "type": "string",
      "description": "Value to match (case-insensitive substring)."
    }
  }
}
```

### Required Fields
`search`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_query`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
