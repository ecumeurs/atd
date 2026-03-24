---
id: module_atd_cli
human_name: "ATD CLI Tool"
type: MODULE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, unified]
parents:
  - [[domain_atd_philosophy]]
dependents:
  - [[mechanic_atd_assemble]]
  - [[service_atd_audit]]
  - [[service_atd_check]]
  - [[specification_atd_config]]
  - [[mechanic_atd_congruence]]
  - [[mechanic_atd_continue]]
  - [[service_atd_crawl]]
  - [[service_atd_discover]]
  - [[mechanic_atd_dissect]]
  - [[mechanic_atd_generate]]
  - [[service_atd_index]]
  - [[mechanic_atd_init]]
  - [[service_atd_query]]
  - [[mechanic_atd_recon]]
  - [[mechanic_atd_reconcile]]
  - [[service_atd_roadmap]]
  - [[service_atd_serve]]
  - [[service_atd_stats]]
  - [[specification_atd_test_links]]
  - [[mechanic_atd_update]]
  - [[service_atd_verify]]
  - [[mechanic_atd_weave]]
layer: ARCHITECTURE
---

# ATD CLI Tool

## INTENT
To provide a single unified command-line binary (`atd`) aggregating all ATD management operations — replacing 23 individual Go tools with one cobra-based CLI that reads its configuration from the `.atd` file at project root.

## THE RULE / LOGIC
The `atd` binary is the sole entry point for all ATD operations. It discovers the `.atd` config file by walking up from the current working directory, resolves all paths relative to it, and routes LLM requests through the tiered provider chain. Every subcommand follows a consistent pattern: load config → validate flags → execute logic → output result (stdout, file, or pipeline_output for IDE Agent fallback).

## TECHNICAL INTERFACE (The Bridge)
- **Binary:** `scripts/cmd/atd/main.go`
- **Config:** `.atd` file at project root
- **Code Tag:** `@spec-link [[module_atd_cli]]`
