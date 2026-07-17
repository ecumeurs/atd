---
id: api_atd_mcp_ops
human_name: "MCP Operations"
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [mcp, tools, api]
parents:
  - [[service_atd_serve]]
dependents:
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
  - [[service_atd_trace]]
---

# MCP Operations

## INTENT
Grouping atom for all Model Context Protocol (MCP) operations exposed by the ATD server, providing a unified architecture layer for tool definitions.

## THE RULE / LOGIC
- All atoms representing MCP-exposed tools must list `api_atd_mcp_ops` as their primary parent.
- This atom acts as a gateway between the `service_atd_serve` (the protocol handler) and individual tool logic.

## TECHNICAL INTERFACE
- **Registry**: `scripts/cmd/atd/cmd/mcp_tools.go`
- **Spec Link**: `@spec-link [[api_atd_mcp_ops]]`

## EXPECTATION
All registered MCP tools are traceable back to this atom in the ATD graph.
