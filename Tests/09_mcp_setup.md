# Task 09 — MCP Setup: Protocol Handshake + tools/list

## Objective
Configure and verify the MCP connection to `atd serve`. Confirm that all 14 tools are listed and the protocol handshake succeeds.

## Prerequisites
- Task 01 complete (binary compiled)
- User has configured `.mcp.json` in the IDE (see below)
- MCP is enabled in the IDE Agent settings

## .mcp.json Configuration (stdio, recommended)
Place this file at the project root (`/home/bastien/work/skill/.mcp.json`):
```json
{
  "servers": {
    "atd": {
      "type": "stdio",
      "command": "/usr/local/bin/atd",
      "args": ["serve"]
    }
  }
}
```

## Manual Smoke Tests (Outside IDE)

### 1. Test initialize via stdio
```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}' \
  | atd serve
```
**Expected:** `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25","capabilities":{"tools":{}},"serverInfo":{"name":"atd","version":"1.0.0"},...}}`

### 2. Test tools/list via stdio
```bash
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | atd serve | python3 -c "
import sys, json
d = json.load(sys.stdin)
tools = d['result']['tools']
print(f'Tool count: {len(tools)}')
for t in tools: print(f'  - {t[\"name\"]}')
"
```
**Expected:** `Tool count: 14` with all tool names listed.

### 3. Test ping
```bash
echo '{"jsonrpc":"2.0","id":3,"method":"ping","params":{}}' | atd serve
```
**Expected:** `{"jsonrpc":"2.0","id":3,"result":{}}`

### 4. Test unknown method error
```bash
echo '{"jsonrpc":"2.0","id":4,"method":"unknown/method","params":{}}' | atd serve
```
**Expected:** `{"jsonrpc":"2.0","id":4,"error":{"code":-32601,"message":"method not found: unknown/method"}}`

### 5. In IDE Agent: request tools/list
Ask the IDE Agent (with MCP enabled): *"List all available ATD tools via MCP."*
> Agent should call `tools/list` and report back 14 tools.

## Acceptance Criteria
- [ ] Manual stdio `initialize` returns correct `protocolVersion: "2025-11-25"`
- [ ] `tools/list` returns exactly 14 tools
- [ ] `ping` returns `{}`
- [ ] Unknown method returns JSON-RPC error code -32601
- [ ] IDE Agent successfully connects and lists tools
