---
id: atd_dissect
human_name: "ATD Dissect"
type: MECHANIC
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, extraction, dissect]
parents:
  - [[atd_cli]]
dependents: []
---

# ATD Dissect

## INTENT
To parse a document or source code file and identify atomic boundaries where single-responsibility rules begin and end, outputting either an LLM prompt for IDE Agent processing or directly extracting boundaries via Ollama.

## THE RULE / LOGIC
Reads the target file, prepends line numbers for precise referencing, and constructs a dissect prompt. In default mode, prints the prompt to stdout for the IDE Agent. With `--llm`, routes through the tiered provider for task type `dissect` — returning structured JSON boundaries or falling back to the IDE pipeline output protocol.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd dissect -file <path> [--llm]`
- **LLM Task:** `dissect`
- **Output (LLM):** `{"atoms": [{"id": "string", "responsibility": "string", "line_range": [int, int]}]}`
- **Code Tag:** `@spec-link [[atd_dissect]]`
