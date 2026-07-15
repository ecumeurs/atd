---
id: specification_atd_config
human_name: "ATD Configuration Schema"
type: SPECIFICATION
version: 1.0
status: DRAFT
priority: 5
tags: [atd, config, specification]
parents:
  - [[module_atd_cli]]
dependents:
  - [[service_atd_serve_config]]
  - [[service_atd_tiered_provider]]
layer: ARCHITECTURE
---

# ATD Configuration Schema

## INTENT
To define the `.atd` configuration file format that governs all ATD tool behavior — including docs path, LLM providers, model-to-task mappings, bloating thresholds, and logging.

## THE RULE / LOGIC
The `.atd` file is a JSON file placed at the project root. It contains:
- `docs_path`: relative path to the ATD docs folder
- `diff_similarity_threshold`: cosine similarity threshold for collision detection (0.0-1.0)
- `bloating_factor`: per-type thresholds for bloat detection
- `logging`: log file path configuration
- `llm.providers`: ordered list of LLM providers (remote → local → ide_agent passthrough)
- `llm.models`: map of model names to their assigned task types
- `llm.fallback_model`: default model when no task-specific model is found

No tool may hardcode an Ollama URL, model name, or path. All must read from this config.

## TECHNICAL INTERFACE (The Bridge)
- **File:** `.atd` at project root
- **Command:** `atd config list`, `atd config bloating-factor <type>`
- **Loaded by:** `scripts/config/config.go`
- **Code Tag:** `@spec-link [[specification_atd_config]]`
