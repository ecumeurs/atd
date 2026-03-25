---
id: api_atd_serve_update
human_name: "MCP Tool: atd_update"
description: "Surgically update fields in an ATD atom file without rewriting it. Pass set as 'key=value' pairs. File and filter are mutually exclusive."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_update
parents: [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_update

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_update`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.

## TECHNICAL INTERFACE (The Bridge)
### Description
Surgically update fields in an ATD atom file without rewriting it. Pass set as 'key=value' pairs.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "Absolute or relative path to the .atom.md file. Optional if filter is provided."
    },
    "filter": {
      "type": "string",
      "description": "Filter atoms to update instead of a single file (e.g. 'status=DRAFT,type=RULE'). Optional if file is provided."
    },
    "set": {
      "description": "Frontmatter edits as 'key=value' strings, e.g. [\"status=STABLE\",\"priority=CORE\"]."
    },
    "intent": {
      "type": "string",
      "description": "New INTENT section text."
    },
    "logic": {
      "type": "string",
      "description": "New THE RULE / LOGIC section text."
    },
    "interface": {
      "type": "string",
      "description": "New TECHNICAL INTERFACE section text."
    },
    "expectation": {
      "type": "string",
      "description": "New EXPECTATION section text."
    },
    "spec_link": {
      "type": "string",
      "description": "Atom ID to prepend as @spec-link in a source file (requires spec_link_file)."
    },
    "spec_link_file": {
      "type": "string",
      "description": "Source file path for --spec-link injection."
    }
  }
}
```

### Required Fields
None (either `file` or `filter` must be provided).

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_update`, it must furnish the required arguments above, returning a JSON-RPC response with block texts.
