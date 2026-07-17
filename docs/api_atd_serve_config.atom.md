---
id: api_atd_serve_config
human_name: "MCP Tool: atd_config"
description: "View or modify the .atd project configuration: list the full config, query bloating factors, or reassign task-to-model mappings."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_config
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_config

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_config`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- **`list`**: if true, returns the full `.atd` configuration as JSON.
- **`bloating_factor`**: an atom type name (e.g. `RULE`, `USECASE`) to query its granularity tolerance before creating atoms of that type.
- **`task`+`model`**: both required together to reassign which LLM model handles a given task (e.g. `dissect`, `embed`, `audit_bloat`).
- Exactly one of the three modes above must be selected per call; calling with none of `list`, `bloating_factor`, or `task`+`model` is refused.

## TECHNICAL INTERFACE (The Bridge)
### Description
View or modify the .atd project configuration. Use 'list':true to see the full config. Use 'bloating_factor' to check the granularity tolerance for a specific atom type before creating atoms. Use 'task'+'model' to reassign which LLM model handles a specific task type.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "list": {
      "type": "boolean",
      "description": "If true, return the full .atd configuration as JSON."
    },
    "bloating_factor": {
      "type": "string",
      "description": "Atom type to query for its bloating factor (e.g. 'RULE', 'USECASE'). Check this before creating atoms."
    },
    "task": {
      "type": "string",
      "description": "Task name to reassign (requires 'model'). E.g. 'dissect', 'embed', 'audit_bloat'."
    },
    "model": {
      "type": "string",
      "description": "Model name to assign to the task (requires 'task')."
    }
  }
}
```

### Required Fields
None -- but at least one of `list`, `bloating_factor`, or `task`+`model` must be provided, or the call errors.

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_config` with `list:true`, it must return the full `.atd` config as JSON. With `bloating_factor` set, it returns that type's granularity tolerance. With `task`+`model` set, it reassigns the task's model and confirms the change. With none of the three, it returns an error naming the required parameter combinations.
