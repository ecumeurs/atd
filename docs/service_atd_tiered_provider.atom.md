---
id: service_atd_tiered_provider
human_name: "ATD Tiered LLM Provider"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, llm, provider, tiered]
parents:
  - [[specification_atd_config]]
dependents: []
layer: ARCHITECTURE
---

# ATD Tiered LLM Provider

## INTENT
To automatically resolve which LLM provider and model to use for any given task type, by checking providers in priority order and falling back gracefully to IDE Agent passthrough when no Ollama instance is available.

## THE RULE / LOGIC
Given a task type (e.g. `audit_code`, `audit_bloat`, `embed`):
1. Look up which model handles this task from `llm.models` in `.atd`
2. For each provider in `llm.providers` order:
   - If type is `passthrough` → return IDE Agent fallback
   - Call `/api/tags` on the provider's `base_url` with `timeout_ms`
   - If desired model is available → return this provider + model
   - If desired model is not found but `fallback_model` is → return provider + fallback
3. If no provider reachable → return IDE Agent fallback
4. Exception: `embed` task has NO IDE fallback — returns error if Nomic is unavailable

## TECHNICAL INTERFACE (The Bridge)
- **Package:** `scripts/internal/ollama/provider.go`
- **Function:** `ResolveProvider(taskType string) → Resolution`
- **Code Tag:** `@spec-link [[service_atd_tiered_provider]]`
