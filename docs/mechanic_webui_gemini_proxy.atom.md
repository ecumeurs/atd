---
id: mechanic_webui_gemini_proxy
human_name: "WebUI Gemini API Proxy"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 3
tags: [gemini, proxy, backend, atd-context]
parents:
  - [[module_webui]]
dependents: []
---

# WebUI Gemini API Proxy

## INTENT
Provide a secure, structured backend proxy to the Gemini API using the `google.golang.org/genai` SDK, ensuring all AI interactions are rooted in ATD principles via manifesto injection.

## THE RULE / LOGIC
1. **API Keys**: Load `GEMINI_API_KEY` from `.env`. Set `GOOGLE_API_KEY` for the SDK.
2. **Context Injection**:
   - Prepend `atdManifesto` as system instruction.
   - Append `atd_context` (JSON-serialized atoms) to provide localized knowledge.
   - Append `actions` (action history) to track acceptance/rejection patterns.
3. **Structured Response**:
   - Force response format using SDK's `ResponseMIMEType: "application/json"`.
   - Use `ResponseSchema` to ensure fields: `message` (string) and `proposals` (array of `action`, `atom_id`, `content`, `impact_summary`).
4. **Model Selection**: Default to `gemini-2.5-flash`, allow overriding via `model` field.

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
