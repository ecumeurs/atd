---
id: api_atd_serve_env
human_name: "MCP Tool: atd_env"
description: "Unified environment smoke test: validates .atd config, checks provider connectivity, and verifies model availability."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_env
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_env

## INTENT
This atom describes the JSON schema and functionality as exposed to the MCP client (IDE or AI agent).

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_env`.
- **Endpoint Responsibility**: Acts as a bridge between the MCP protocol and the internal ATD CLI functionality.
- Runs the same diagnostic as the `atd check` CLI command (`runCheck`): loads the `.atd` config, checks LLM provider connectivity, and verifies model availability.
- Takes no parameters -- always runs against the active project configuration.

## TECHNICAL INTERFACE (The Bridge)
### Description
Unified environment smoke test: validates .atd config, checks provider connectivity, and verifies model availability. Use to diagnose 'Connection Refused' or 'Model Not Found' errors, or to verify a new provider/model is correctly configured.

### Input Schema
```json
{
  "type": "object",
  "properties": {}
}
```

### Required Fields
None

## EXPECTATION (For Testing)
When the MCP client sends a `tools/call` request for `atd_env`, it must return a JSON-RPC response reporting config validity, provider connectivity, and model availability, with no arguments required.
