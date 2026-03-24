---
id: atd_serve_config
human_name: "ATD Serve: Config Tool"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, mcp, config]
parents:
  - [[atd_serve]]
  - [[atd_config]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Serve: Config Tool

## INTENT
To expose `.atd` configuration management as an MCP tool (`atd_config`), enabling IDE agents to read and modify tool behavior dynamically.

## THE RULE / LOGIC
Accepts parameters for listing the full config (`list: true`), querying specific bloating factors (`bloating_factor: "TYPE"`), or re-mapping tasks to models (`task: "T", model: "M"`).

## TECHNICAL INTERFACE (The Bridge)
- **MCP Tool:** `atd_config`
- **Input Schema:** `{"list": bool, "bloating_factor": string, "task": string, "model": string}`
- **Code Tag:** `@spec-link [[atd_serve_config]]`
