---
id: api_atd_serve_query
human_name: "MCP Tool: atd_query"
description: "Search ATD atoms by frontmatter field value. Returns JSON array of matching atoms."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_query
parents: [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_query

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_query`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- **Paths Only Mode**: If `paths_only` is true, the response is a JSON array of strings containing absolute file paths.

## TECHNICAL INTERFACE (The Bridge)
### Description
Search ATD atoms by frontmatter field value. Returns JSON array of matching atoms or paths.

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
    },
    "paths_only": {
      "type": "boolean",
      "description": "If true, return only a JSON array of absolute file paths."
    }
  }
}
```

### Required Fields
`search`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_query`, it must furnish the required arguments above, returning a JSON-RPC response with block texts or absolute file paths if `paths_only` is true.
