---
id: api_atd_serve_search
human_name: "MCP Tool: atd_search"
description: "Search the indexed codebase semantically (requires a built index) or via keyword grep."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_search
parents: [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_search

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_search`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Search the indexed codebase semantically (requires a built index) or via keyword grep.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Semantic search query (uses Nomic embeddings)."
    },
    "grep": {
      "type": "string",
      "description": "Literal keyword search across project files."
    },
    "db": {
      "type": "string",
      "description": "Path to SQLite index database."
    },
    "limit": {
      "description": "Number of semantic results to return. Defaults to 5."
    },
    "scope": {
      "type": "string",
      "description": "Search scope: 'code', 'docs', or 'all'. Defaults to 'all'."
    }
  }
}
```

### Required Fields
None

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_search`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
