---
id: mechanic_webui_gemini_chat_orchestration
status: REVIEW
type: MECHANIC
layer: IMPLEMENTATION
priority: 2
tags: webui,gemini,chat
version: 1.0
parents:
  - [[mechanic_webui_gemini_proxy]]
human_name: WebUI Gemini Chat Orchestration
dependents: []
---

# New Atom

## INTENT
Orchestrate multi-turn conversational chat with Gemini for ATD spec creation, injecting system instructions and ATD context.

## THE RULE / LOGIC
1. Receive ChatRequest with messages, model, atd_context, actions, omit_history.
2. Build system instruction from atdManifesto + ATD context atoms + action history.
3. Construct Gemini content array from message history.
4. Call Gemini API with structured JSON output schema enforcing proposals format.
5. Parse response into GeminiResponse with message, proposals, and usage metadata.
6. Handle quota errors (429) by returning appropriate status for client-side model switching.

## TECHNICAL INTERFACE

## EXPECTATION
POST /api/gemini/chat returns a structured JSON response with a conversational message and optional ATD proposals. Quota errors return HTTP 429.
