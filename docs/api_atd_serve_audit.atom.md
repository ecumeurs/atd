---
id: api_atd_serve_audit
human_name: "MCP Tool: atd_audit"
description: "Audit ATD atoms for documentation bloat and semantic collisions across the docs directory."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_audit
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_audit

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_audit`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Audit ATD atoms for documentation bloat and semantic collisions across the docs directory.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "docs": {
      "type": "string",
      "description": "Override docs directory path."
    },
    "threshold": {
      "type": "number",
      "description": "Cosine similarity threshold for collision detection (0.0\u20131.0). Defaults to the configured diff_similarity_threshold (or 0.85)."
    }
  }
}
```

### Required Fields
None

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_audit`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
