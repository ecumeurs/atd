# Issue: MCP Server Needs Roots Capability Support

**ID:** `20260318_mcp_roots_capability`
**Ref:** `ISS-028`
**Date:** 2026-03-18
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/pkg/mcp`
**Affects:** `scripts/cmd/atd/cmd/serve.go`

---

## Summary

The ATD MCP server currently lacks the ability to ask the client for its workspace roots (`roots/list`). Without workspace roots, the server must rely on `os.Getwd()` to locate the `.atd` configuration file, which is unreliable when the server is launched natively by an IDE (like VS Code) or when using the HTTP transport. We need bidirectional JSON-RPC to fetch roots and accurately configure the ATD runtime.

---

## Technical Description

### Background
Currently, the MCP server in `scripts/pkg/mcp` only acts as a responder: it accepts client requests and returns responses. 

### The Problem Scenario
1. Client establishes transport (stdio or HTTP).
2. Client sends `initialize` with `capabilities.roots` present.
3. Server responds to `initialize` but has no way to dispatch its own `roots/list` request because bidirectional JSON-RPC routing is not implemented.
4. Server falls back to `os.Getwd()` to locate `.atd`, which may point to an irrelevant or system-level path instead of the actual workspace root being interrogated by the client.

### Where This Pattern Exists Today
- `scripts/pkg/mcp/handler.go`: Assumes all messages with an `id` and `method` are requests from the client. Does not handle client responses.
- `scripts/pkg/mcp/stdio.go` and `http.go`: Transports do not multiplex sending server requests and receiving client responses gracefully.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | Medium |
| Current mitigant | None |

---

## Recommended Fix

**Short term:**
- Implement bidirectional JSON-RPC in the Registry by adding a pending requests map and waiting on response channels.
- Update `stdio` and `http` to allow server-initiated writes.
- Detect `roots` capability during `initialize`. On receiving `notifications/initialized`, send `roots/list` and parse the result.

**Medium term:** / **Long term:** n/a

---

## References

- `scripts/pkg/mcp/handler.go`
- `scripts/pkg/mcp/stdio.go`
- `scripts/pkg/mcp/http.go`

## Change Log
- **2026-03-18**: Implemented server-initiated roots/list request across stdio and http transports; resolved.
