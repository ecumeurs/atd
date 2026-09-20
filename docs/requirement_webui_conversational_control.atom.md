---
id: requirement_webui_conversational_control
status: DRAFT
human_name: Conversational History Control
type: REQUIREMENT
layer: BUSINESS
priority: 3
parents:
  - [[requirement_webui_platform]]
version: 1.0
dependents: []
---

# Conversational History Control

## INTENT
Allow users to toggle between context-aware 'Chat' and context-less 'Single-shot' modes.

## THE RULE / LOGIC
1. User can toggle 'Chat Mode' via header switch. 2. When disabled, the conversation history is omitted from subsequent LLM calls. 3. System context and ATD context remain active in both modes.

## TECHNICAL INTERFACE

## EXPECTATION
Switching modes properly clears/restores history forwarding to the backend.
