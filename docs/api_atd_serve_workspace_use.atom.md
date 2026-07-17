---
id: api_atd_serve_workspace_use
human_name: "MCP Tool: atd_workspace_use"
description: "Switch the active project context in the current ATD workspace."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_workspace_use
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_workspace_use

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_workspace_use`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- Calls `config.SetProject(project)`, switching which project's `.atd`/`docs/` subsequent tool calls are scoped to.
- `project` is required; an unknown project name is refused with the underlying `config.SetProject` error.

## TECHNICAL INTERFACE (The Bridge)
### Description
Switch the active project context in the current workspace. Subsequent tool calls will be scoped to this project.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "project": {
      "type": "string",
      "description": "Name of the project to switch to."
    }
  }
}
```

### Required Fields
`project`

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_workspace_use` with a valid `project` name, subsequent tool calls are scoped to that project's docs/config. An unknown project name returns a loud error rather than silently keeping the previous active project.
