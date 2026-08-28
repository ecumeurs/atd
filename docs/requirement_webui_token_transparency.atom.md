---
id: requirement_webui_token_transparency
status: DRAFT
parents:
  - [[requirement_webui_platform]]
human_name: Token Usage Transparency
priority: 3
version: 1.0
dependents: []
type: REQUIREMENT
layer: BUSINESS
---

# Token Usage Transparency

## INTENT
Ensure full transparency regarding token consumption and cost for every LLM interaction.

## THE RULE / LOGIC
1. Capture Prompt and Response token counts from Gemini SDK. 2. Forward counts to the frontend for each message. 3. Maintain an accumulate session-wide total in the UI.

## TECHNICAL INTERFACE

## EXPECTATION
The user can see the exact token cost after each message and the session-wide total.
