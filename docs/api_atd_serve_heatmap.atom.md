---
id: api_atd_serve_heatmap
human_name: "MCP Tool: atd_heatmap"
description: "Get heat map metrics for a specific atom: dependency coupling, code implementation density, and update instability."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_heatmap
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_heatmap

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_heatmap`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- `atom` (required) may be an atom ID or a path to its `.atom.md` file.
- Reports three layers: dependency (coupling -- fan-in/fan-out in the atom graph), code (implementation density -- @spec-link count), and updates (instability -- change frequency).

## TECHNICAL INTERFACE (The Bridge)
### Description
Get heat map metrics for a specific atom. Layers: dependency (coupling), code (implementation density), updates (instability).

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "atom": {
      "type": "string",
      "description": "Atom ID or file path."
    }
  }
}
```

### Required Fields
`atom`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_heatmap` with a valid `atom` ID or path, it must return the atom's dependency/code/update heat metrics. An unknown atom id returns a loud error naming the id.
