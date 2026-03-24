---
id: atd_serve_index
human_name: "MCP Tool: atd_index"
description: "Build a semantic vector index of source code and/or ATD documents using nomic-embed-text. Requires a local or remote Ollama provider."
type: TECHNICAL_CONTRACT
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_index
parents:
  - [[atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_index

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_index`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Build a semantic vector index of source code and/or ATD documents using nomic-embed-text. Requires a local or remote Ollama provider.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "dir": {
      "type": "string",
      "description": "Directory to crawl and index. Defaults to current directory."
    },
    "db": {
      "type": "string",
      "description": "Path to SQLite database. Defaults to <docs_path>/.atd_index.db."
    },
    "mode": {
      "type": "string",
      "description": "What to index: 'code', 'docs', or 'all'. Defaults to 'code'."
    }
  }
}
```

### Required Fields
None

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_index`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
