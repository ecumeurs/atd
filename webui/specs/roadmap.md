# Gemini Spec Builder — Development Roadmap

## Overview

This roadmap breaks the Gemini Spec Builder integration into **10 self-contained steps**. Each step file provides everything needed for a Gemini Flash model to implement it independently: full context, file paths, code snippets to modify, and expected outcomes.

## Architecture Summary

The Spec Builder is a **new tab** in the existing ATD WebUI that enables conversational ATD specification creation via the Gemini Chat API. The UI is split into:

- **Left column**: Chat panel (user ↔ Gemini conversation)
- **Right column**: Proposal review panel (Gemini-suggested ATD changes)

### Key Technical Decisions

| Decision | Choice | Rationale |
|---|---|---|
| API communication | Go backend via `google.golang.org/genai` SDK | Official Google SDK; keeps API key server-side; type-safe |
| Gemini model | User-selectable, default `gemini-2.5-flash` | Models listed via `client.Models.List()`; supports 2.5 + 3.x series |
| Available models | `gemini-2.5-flash` (default), `gemini-2.5-pro`, `gemini-2.5-flash-lite`, `gemini-3-flash-preview`, `gemini-3.1-pro-preview`, `gemini-3.1-flash-lite-preview` | Dynamically enumerated from API |
| Frontend framework | Vanilla JS (matches existing) | No build step; consistent with `app.js` |
| Proposal rendering | Markdown preview via `marked.js` | Already loaded in the project |
| State management | In-memory JS objects | Simple; session-scoped |
| Tab system | CSS + JS toggle (no framework) | Aligns with existing view-switcher pattern |

### Data Flow

```
User types message
    ↓
Frontend collects: message + action history + ATD context
    ↓
POST /api/gemini/chat  (Go backend)
    ↓
Go loads GEMINI_API_KEY from .env → sets GOOGLE_API_KEY env var
    ↓
genai.NewClient(ctx, nil)  — official Google Gen AI SDK
client.Models.GenerateContent(ctx, selectedModel, parts, config)
    (selectedModel from UI dropdown, default "gemini-2.5-flash")
    (System instruction = ATD Manifesto, ResponseMIMEType = "application/json")
    ↓
Parse structured JSON response: { message, proposals[] }
    ↓
Return to frontend
    ↓
Render message in chat + render proposals in side panel
```

## Step Sequence

| Step | File | Description | Dependencies |
|---|---|---|---|
| 1 | [step_01_backend_proxy.md](step_01_backend_proxy.md) | Go backend: `.env` loader, `/api/gemini/chat` endpoint, Gemini API call | None |
| 2 | [step_02_chat_ui_shell.md](step_02_chat_ui_shell.md) | HTML tab system + chat panel layout + CSS | Step 1 |
| 3 | [step_03_chat_logic.md](step_03_chat_logic.md) | JS message send/receive, streaming render, history management | Steps 1-2 |
| 4 | [step_04_proposal_panel.md](step_04_proposal_panel.md) | Proposal cards UI: accept/reject, diff preview, apply via `atd update` | Steps 1-3 |
| 5 | [step_05_atd_context.md](step_05_atd_context.md) | ATD context injection: recommend atoms, user confirmation, dedup | Steps 1-4 |
| 6 | [step_06_action_history.md](step_06_action_history.md) | Forward accepted/rejected proposal history in chat messages | Steps 4-5 |
| 7 | [step_07_context_drift.md](step_07_context_drift.md) | 10-exchange drift warning banner + session restart prompt | Completed |
| 8 | [step_08_session_mgmt.md](step_08_session_mgmt.md) | Session reset, conversation export, keyboard shortcuts | Completed |
| 9 | [step_09_polish.md](step_09_polish.md) | Animations, error states, loading skeletons, responsive layout | Steps 1-8 |
| 10 | [step_10_testing.md](step_10_testing.md) | Manual test plan, edge cases, walkthrough documentation | Steps 1-9 |

## Files Modified/Created Per Step

| File | Steps |
|---|---|
| `webui/main.go` | 1, 4 |
| `webui/static/index.html` | 2 |
| `webui/static/styles.css` | 2, 7, 9 |
| `webui/static/app.js` | 2 (tab init only) |
| `webui/static/spec-builder.js` (**NEW**) | 3, 4, 5, 6, 7, 8 |
| `webui/static/spec-builder.css` (**NEW**) | 2, 4, 7, 9 |
| `webui/.env` | 1 (document only, already exists) |

## Side Features (included in steps)

1. **Proposal diff preview** (Step 4): Show before/after when updating an existing atom.
2. **ATD search-as-you-type** (Step 5): Quick atom search to manually attach context.
3. **Export conversation** (Step 8): Download chat + proposals as Markdown.
4. **Keyboard shortcuts** (Step 8): `Ctrl+Enter` to send, `Esc` to dismiss.
5. **Typing indicator** (Step 9): Animated dots while waiting for Gemini response.
6. **Token usage display** (Step 9): Show approximate token usage from API response.
7. **Context drift warning** (Step 7): Visual indicator + session restart recommendation after 10 exchanges.
