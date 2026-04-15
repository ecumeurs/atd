---
id: mechanic_webui_summary_aggregation
status: REVIEW
parents:
  - [[api_webui_health_stats]]
dependents: []
human_name: WebUI Summary Aggregation
priority: 2
type: MECHANIC
layer: IMPLEMENTATION
tags: webui,summary,ollama
version: 1.0
---

# New Atom

## INTENT
Generate context-aware summaries of ATD atoms by traversing the dependency graph and aggregating intent sections, optionally enhanced by local LLM.

## THE RULE / LOGIC
Graph-traversing contextual summarization:
1. Receives atom ID, walks parent/dependent graph to collect context
2. Groups context by layer (Customer, Architecture, Implementation)
3. Attempts Ollama summarization with structured prompt
4. Falls back to intent-concatenation if Ollama is unavailable
5. Returns JSON with summary text and llm_used flag

## TECHNICAL INTERFACE

## EXPECTATION
GET /api/summary/:id returns a structured summary with context from ancestor and descendant atoms. Works without LLM (aggregation mode) and is enhanced when Ollama is reachable.
