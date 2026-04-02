---
id: mechanic_webui_gemini_proxy
human_name: "WebUI Gemini API Proxy"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 3
tags: [gemini, proxy, backend, atd-context]
parents: [[ui_webui_spec_builder]]
  - [[module_webui]]
dependents: [[[mechanic_webui_gemini_chat_orchestration]], [[mechanic_webui_gemini_model_list]], [[rule_webui_context_history_management]]]
---

# WebUI Gemini API Proxy

## INTENT
Provide a secure, structured backend proxy to the Gemini API using the `google.golang.org/genai` SDK, ensuring all AI interactions are rooted in ATD principles via manifesto injection.

## THE RULE / LOGIC
1. API Keys: Load GEMINI_API_KEY from .env. 2. Context Injection: Manifesto, ATD context, actions, and conditionally history based on OmitHistory flag. 3. Structured Response: message, proposals, and usage metadata. 4. Model Selection: Allow override.

## TECHNICAL INTERFACE (The Bridge)
- **Endpoint**: `POST /api/gemini/chat`
- **Request Schema**: `ChatRequest` (Messages, Model, AtdContext, Actions)
- **Response Schema**: `GeminiResponse` (Message, Proposals)
- **SDK**: `google.golang.org/genai`
- **Handler**: `handleGeminiChat` in `webui/main.go`
- **Code Tag**: `@spec-link [[mechanic_webui_gemini_proxy]]`

## EXPECTATION (For Testing)
- `POST /api/gemini/chat` returns HTTP 400 for malformed JSON.
- `POST /api/gemini/chat` returns structured JSON even when no proposals are made (`"proposals": []`).
- Gemini receives the `atdManifesto` in the system instructions on every call.
