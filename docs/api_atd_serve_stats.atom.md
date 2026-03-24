---
id: api_atd_serve_stats
human_name: "MCP Tool: atd_stats"
description: "Produce quantitative documentation health metrics: total atoms, atoms by type, status, domain, coverage ratio, and orphan count."
type: API
version: 1.0
status: STABLE
priority: 5
tags:
  - mcp
  - tool
  - atd_stats
parents:
  - [[service_atd_serve]]
dependents: []
layer: ARCHITECTURE
---

# MCP Tool: atd_stats

## INTENT
Exposes the ATD project health metrics to MCP clients (such as IDEs or AI agents) allowing them to gauge documentation completeness.

## THE RULE / LOGIC
- **JSON-RPC Method**: `tools/call` with name `atd_stats`.
- **Logic**: Invokes the internal `runStats` function which crawls atoms and source code to aggregate metrics.

## TECHNICAL INTERFACE (The Bridge)
### Description
Produce quantitative documentation health metrics: total atoms, atoms by type, status, domain, coverage ratio, and orphan count.

### Input Schema
```json
{
  "type": "object",
  "properties": {
    "src": {
      "type": "string",
      "description": "Path to source code directory (optional)."
    },
    "docs": {
      "type": "string",
      "description": "Override docs directory path."
    }
  }
}
```

### Required Fields
None

## EXPECTATION (For Testing)
Returns a JSON-RPC response containing a `StatsReport` object with `total_atoms`, `by_type`, `by_status`, `by_domain`, `coverage_ratio`, and `orphan_count`.
