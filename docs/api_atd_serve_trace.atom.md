---
id: api_atd_serve_trace
human_name: "MCP Tool: atd_trace"
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [mcp, tool, atd_trace]
parents:
  - [[api_atd_mcp_ops]]
  - [[service_atd_serve]]
dependents:
  - [[service_atd_trace]]
---

# MCP Tool: atd_trace

## INTENT
Expose the ATD graph tracing and health snapshot capabilities to MCP clients.

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_trace`.
- **Logic**: Invokes the internal `runTrace` function to traverse the atom graph and compute coverage metrics.

## TECHNICAL INTERFACE (The Bridge)
### Description
Get a structured Health Snapshot JSON for a specific atom by traversing its graph ancestry and descendants. Includes warnings for layer compliance and metrics for testing and implementation coverage.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "atom": {
      "type": "string",
      "description": "Target ID of the atom to trace."
    }
  },
  "required": ["atom"]
}
```

## EXPECTATION (For Testing)
Returns a `TraceSnapshot` JSON object containing `target_id`, `layer`, `health_summary`, `metrics`, `graph_slice`, and any `warnings`.
