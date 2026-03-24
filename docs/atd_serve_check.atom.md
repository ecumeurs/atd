---
id: atd_serve_check
human_name: "ATD Serve: Check Tool"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, mcp, check, health]
parents:
  - [[atd_serve]]
  - [[atd_check]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Serve: Check Tool

## INTENT
To expose the `atd check` functionality as an MCP tool (`atd_check`), allowing IDE agents to automatically diagnose environment issues.

## THE RULE / LOGIC
Calls the core `atd check` logic while capturing its output. Returns the status report as a string to the MCP client.

## TECHNICAL INTERFACE (The Bridge)
- **MCP Tool:** `atd_check`
- **Input Schema:** `{}` (no parameters)
- **Code Tag:** `@spec-link [[atd_serve_check]]`
