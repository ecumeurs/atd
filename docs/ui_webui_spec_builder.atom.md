---
id: ui_webui_spec_builder
human_name: "Spec Builder UI"
type: UI
layer: ARCHITECTURE
version: 1.0
status: DRAFT
priority: 4
tags: [webui, ui, spec-builder, gemini, chat]
parents: [[module_webui]]
  - [[module_webui]]
dependents:
  - [[mechanic_webui_gemini_proxy]]
---

# Spec Builder UI

## INTENT
Provide an interactive, conversational interface for architects to decompose high-level requirements into atomic ATD units using Gemini AI.

## THE RULE / LOGIC
The Spec Builder is a tabbed interface within the WebUI providing a split-column layout:
1. **Chat Panel (Left)**: Conversational thread using Gemini API. Supports manual and automatic ATD context injection.
2. **Proposal Review (Right)**: Visual cards for AI-suggested ATD modifications (CREATE, UPDATE, DELETE).
3. **Draft Control**: Manage session history, exports, and model selection.
4. **Safety**: Require manual confirmation (Accept/Reject) before any ATD change is persisted to disk.

## TECHNICAL INTERFACE (The Bridge)
- **Frontend**: `webui/static/spec-builder.js`
- **Styles**: `webui/static/spec-builder.css`
- **API Link**: `POST /api/gemini/chat`
- **Code Tag**: `@spec-link [[ui_webui_spec_builder]]`

## EXPECTATION (For Testing)
- Users can toggle between the Explorer and Spec Builder tabs.
- Chat messages are rendered in real-time with markdown support.
- Proposal cards appear when Gemini returns structured `proposals` JSON.
- Accepting a proposal invokes `POST /api/gemini/apply-proposal`.
