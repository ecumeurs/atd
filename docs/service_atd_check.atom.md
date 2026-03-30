---
id: service_atd_check
human_name: "ATD Check"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, check, health, environment]
parents:
  - [[module_atd_cli]]
dependents:
  - [[service_atd_serve_check]]
layer: IMPLEMENTATION
---

# ATD Check

## INTENT
To provide a unified environment smoke test that validates the `.atd` configuration, verifies connectivity to all configured LLM providers, and ensures model availability for all defined tasks.

## THE RULE / LOGIC
Pings each provider defined in `.atd`. For Ollama providers, it calls `/api/tags` to list available models. It then simulates the priority-based model resolution logic for every task type (e.g., `dissect`, `audit_code`) and reports whether the task is "Ready" or "Missing Model".

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd check`
- **LLM Task:** None (deterministic connectivity/listing)
- **Code Tag:** `@spec-link [[service_atd_check]]`
