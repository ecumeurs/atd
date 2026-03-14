---
id: atd_serve
human_name: "ATD MCP Server"
type: SERVICE
version: 1.0
status: STABLE
priority: CORE
tags: [atd, cli, mcp, server, json-rpc]
parents:
  - [[atd_cli]]
dependents: []
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
- Implements MCP methods: `initialize`, `notifications/initialized`, `tools/list`, `tools/call`, `ping`
- Registered tools (14 as of v1.0):
  - Deterministic: `atd_query`, `atd_crawl`, `atd_weave`, `atd_update`, `atd_roadmap`, `atd_verify`, `atd_assemble`, `atd_test_links`
  - LLM-backed: `atd_dissect`, `atd_index`, `atd_search`, `atd_audit`, `atd_recon`, `atd_discover`
- LLM tools that print progress use `captureStdout()` redirect to avoid polluting the JSON-RPC stdio channel
- HTTP transport validates `Origin` header against `localhost` / `127.0.0.1` to prevent DNS rebinding attacks
- `atd init` is intentionally **not** a MCP tool — it is a one-time filesystem bootstrap

## TECHNICAL INTERFACE (The Bridge)
- **Binary:** `scripts/cmd/atd/cmd/serve.go`
- **MCP Infrastructure:** `scripts/pkg/mcp/` (types, registry, handler, stdio, http)
- **Tool Registrations:** `scripts/cmd/atd/cmd/mcp_tools.go`
- **Code Tag:** `@spec-link [[atd_serve]]`

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
