---
id: mechanic_atd_generate
human_name: "ATD Generate"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, generation, llm]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Generate

## INTENT
To take a dissect output file and send it to the LLM for structured extraction of atom boundaries as JSON.

## THE RULE / LOGIC
Reads a previously generated dissect prompt file and routes it through the tiered provider for task type `dissect`. The LLM must return structured JSON conforming to the dissect format schema. On IDE fallback, writes the prompt to pipeline_output for later processing.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd generate -dissect <path>`
- **LLM Task:** `dissect`
- **Code Tag:** `@spec-link [[mechanic_atd_generate]]`
