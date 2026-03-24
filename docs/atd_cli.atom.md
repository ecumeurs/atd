---
id: atd_cli
human_name: "ATD CLI Tool"
type: MODULE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, unified]
parents:
  - [[atd_philosophy]]
dependents:
  - [[atd_assemble]]
  - [[atd_audit]]
  - [[atd_check]]
  - [[atd_config]]
  - [[atd_congruence]]
  - [[atd_continue]]
  - [[atd_crawl]]
  - [[atd_discover]]
  - [[atd_dissect]]
  - [[atd_generate]]
  - [[atd_index]]
  - [[atd_init]]
  - [[atd_query]]
  - [[atd_recon]]
  - [[atd_reconcile]]
  - [[atd_roadmap]]
  - [[atd_serve]]
  - [[atd_stats]]
  - [[atd_test_links]]
  - [[atd_update]]
  - [[atd_verify]]
  - [[atd_weave]]
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
- **Code Tag:** `@spec-link [[atd_cli]]`
