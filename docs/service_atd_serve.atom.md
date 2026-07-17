---
id: service_atd_serve
human_name: "ATD MCP Server"
type: SERVICE
version: 1.0
status: STABLE
priority: 5
tags: [atd, cli, mcp, server, json-rpc]
parents:
  - [[module_atd_cli]]
dependents:
  - [[api_atd_mcp_ops]]
  - [[api_atd_serve_assemble]]
  - [[api_atd_serve_audit]]
  - [[api_atd_serve_check]]
  - [[api_atd_serve_config]]
  - [[api_atd_serve_crawl]]
  - [[api_atd_serve_dissect]]
  - [[api_atd_serve_env]]
  - [[api_atd_serve_heatmap]]
  - [[api_atd_serve_heatmap_code]]
  - [[api_atd_serve_heatmap_project]]
  - [[api_atd_serve_index]]
  - [[api_atd_serve_lint]]
  - [[api_atd_serve_map]]
  - [[api_atd_serve_query]]
  - [[api_atd_serve_recon]]
  - [[api_atd_serve_roadmap]]
  - [[api_atd_serve_search]]
  - [[api_atd_serve_stats]]
  - [[api_atd_serve_test_links]]
  - [[api_atd_serve_trace]]
  - [[api_atd_serve_update]]
  - [[api_atd_serve_weave]]
  - [[api_atd_serve_workspace_list]]
  - [[api_atd_serve_workspace_stats]]
  - [[api_atd_serve_workspace_use]]
layer: ARCHITECTURE
---

# ATD MCP Server

## INTENT
Expose all ATD operations as MCP (Model Context Protocol) tools over JSON-RPC 2.0, enabling IDE agents (VS Code, Claude Desktop) and other MCP hosts to invoke ATD commands without a subprocess shell.

## THE RULE / LOGIC
- Protocol: JSON-RPC 2.0, MCP spec 2025-11-25
- **Transports**:
  - `stdio` (primary): server reads newline-delimited JSON from stdin, writes to stdout. Launched by MCP host as a subprocess.
  - `HTTP` (secondary): single `/mcp` endpoint, `POST` for requests, `GET` for SSE stream, `DELETE` for session teardown. Tracks sessions via `MCP-Session-Id` header.
- **Flags**: `atd serve [--http] [--port 7474]`
- Implements server MCP methods: `initialize`, `notifications/initialized`, `tools/list`, `tools/call`, `ping`
- Emits client MCP requests: `roots/list` (if client declares `roots` capability during initialization, to locate the `.atd` config accurately)
- Registered tools (16 as of v1.0):
  - Deterministic: `atd_query`, `atd_crawl`, `atd_weave`, `atd_update`, `atd_roadmap`, `atd_assemble`, `atd_test_links`, `atd_check`, `atd_config`, `atd_stats`
  - LLM-backed: `atd_dissect`, `atd_index`, `atd_search`, `atd_audit`, `atd_recon`, `atd_map`
- LLM tools that print progress use `captureStdout()` redirect to avoid polluting the JSON-RPC stdio channel
- HTTP transport validates `Origin` header against `localhost` / `127.0.0.1` to prevent DNS rebinding attacks
- `atd init` is intentionally **not** a MCP tool — it is a one-time filesystem bootstrap

## TECHNICAL INTERFACE (The Bridge)
- **Binary:** `scripts/cmd/atd/cmd/serve.go`
- **MCP Infrastructure:** `scripts/pkg/mcp/` (types, registry, handler, stdio, http)
- **Tool Registrations:** `scripts/cmd/atd/cmd/mcp_tools.go`
- **Code Tag:** `@spec-link [[service_atd_serve]]`

## EXPECTATION (For Testing)
```bash
# stdio — initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}' | atd serve
# Expected: {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25",...}}

# stdio — tools/list should return 14 tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | atd serve

# HTTP — initialize with session ID
curl -X POST http://localhost:7474/mcp -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}'
# Response headers must include: MCP-Session-Id: <uuid>
```

### VS Code `.mcp.json` (stdio)
```json
{
  "servers": {
    "atd": {
      "type": "stdio",
      "command": "/path/to/atd",
      "args": ["serve"]
    }
  }
}
```

### VS Code `.mcp.json` (HTTP)
```json
{
  "servers": {
    "atd": {
      "type": "http",
      "url": "http://localhost:7474/mcp"
    }
  }
}
```
