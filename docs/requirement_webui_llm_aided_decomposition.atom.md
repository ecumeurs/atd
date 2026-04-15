---
id: requirement_webui_llm_aided_decomposition
status: DRAFT
human_name: LLM Aided Decomposition
layer: CUSTOMER
version: 1.0
priority: 3
parents:
  - [[requirement_webui_platform]]
dependents: []
type: REQUIREMENT
tags: [webui, gemini, spec-builder]
---

# LLM Aided Decomposition

## INTENT
Assist architects in breaking down complex requirements into atomic units using Gemini.

## THE RULE / LOGIC
The system must take a "fuzzy" text input and propose a DAG of atoms (Parents -> Children) following the ATD manifesto.
- **Input**: Narrative requirement.
- **Output**: Structured JSON with proposed atoms, types, and logic.
- **Consistency**: All proposed atoms must link to the original requirement as a parent.

## TECHNICAL INTERFACE (The Bridge)
- **Endpoint**: `POST /api/gemini/chat`
- **Code Tag**: `@spec-link [[requirement_webui_llm_aided_decomposition]]`
- **Test Names**: `TestLlmDecomposition`

## EXPECTATION (For Testing)
Gemini proposes valid ATD fragments for a given high-level requirement.
