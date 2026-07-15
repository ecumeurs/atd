---
id: api_atd_serve_map
human_name: "MCP Tool: atd_map"
description: "Three-mode tool for linking source code to ATD atoms: find candidate atoms for undocumented code, confirm a specific match, or propose a new atom skeleton."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_map
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_map

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_map`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- **Three modes**, selected by which arguments are present:
  1. **Default mode** (`file` only): extracts architectural intent from an undocumented file, searches the ATD index, and recommends `@spec-link` tags to apply. Follows surgical attachment rules (no global headers, logic-boundary placement).
  2. **Confirm mode** (`file` + `atom`): validates whether the given file implements the specified atom. Returns a confidence score and rationale. Shorthand-equivalent to `atd_recon`.
  3. **Propose mode** (`file` + `new: true`): treats the file as entirely undocumented and returns a proposed new atom skeleton (id, type, layer, intent, logic) ready to be passed to `atd_update`.

## TECHNICAL INTERFACE (The Bridge)
### Description
Three-mode tool for linking source code to ATD atoms. Default mode (file only) extracts architectural intent from an undocumented file and recommends @spec-link tags. Confirm mode (file + atom) validates whether a specific file implements a given atom, returning a confidence score and rationale. Propose mode (file + new:true) returns a proposed new atom skeleton for an undocumented file.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "file": {
      "type": "string",
      "description": "Path to the source file to analyse."
    },
    "atom": {
      "type": "string",
      "description": "Confirm mode: atom ID or path to validate against the file."
    },
    "new": {
      "type": "boolean",
      "description": "Propose mode: return a new atom skeleton for the file instead of searching existing atoms."
    }
  },
  "required": ["file"]
}
```

### Required Fields
`file`

### Usage Examples
- `atd_map(file="src/foo.go")` — default mode, recommend `@spec-link` tags for undocumented code.
- `atd_map(file="src/foo.go", atom="rule_foo")` — confirm mode, validate a specific match.
- `atd_map(file="src/foo.go", new=true)` — propose mode, return a proposed new atom skeleton.

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_map`, it must furnish at least the required `file` argument above. The mode is determined by which optional arguments (`atom`, `new`) accompany `file`, returning a JSON-RPC response with block texts appropriate to that mode.
