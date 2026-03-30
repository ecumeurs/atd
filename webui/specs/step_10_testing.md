# Step 10: Integration Testing & Walkthrough

## Objective
Verify the full Spec Builder integration end-to-end. This step is a manual test plan, not code changes.

## Prerequisites
- Steps 1-9 completed
- `GEMINI_API_KEY` set in `webui/.env`
- WebUI server built and running

## Setup

```bash
cd webui
go build -o webui_bin .
./webui_bin
# Open http://localhost:8081 in browser
```

## Test Cases

### T1: Backend Health

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 1.1 | API info | `curl http://localhost:8081/api/info` | JSON with `project_path`, `atd_path`, `atd_count` |
| 1.2 | Atoms search | `curl http://localhost:8081/api/gemini/atoms` | JSON array of atoms |
| 1.3 | Atoms filter | `curl http://localhost:8081/api/gemini/atoms?q=module` | Filtered results |
| 1.4 | Chat endpoint | `curl -X POST http://localhost:8081/api/gemini/chat -H "Content-Type: application/json" -d '{"messages":[{"role":"user","content":"hello"}]}'` | JSON with `message` and `proposals` |

### T2: Tab Navigation

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 2.1 | Default tab | Load page | Explorer tab active, treemap/treeview visible |
| 2.2 | Switch to Spec Builder | Click "Spec Builder" tab | Chat UI visible, Explorer hidden |
| 2.3 | Switch back | Click "Explorer" tab | Explorer visible, header tools reappear |
| 2.4 | Header tools | On Explorer tab | Toggle layout, Select Mode, Refresh visible |
| 2.5 | Header tools hidden | On Spec Builder tab | Explorer tools hidden |

### T3: Chat Flow

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 3.1 | Welcome screen | Open Spec Builder tab | Welcome message with 3 suggestion chips |
| 3.2 | Suggestion chip | Click "Define a feature" | Input pre-filled, chip text in input |
| 3.3 | Send message | Type "I need a login system" → Ctrl+Enter | User bubble appears, typing indicator shows |
| 3.4 | Response | Wait for Gemini response | Model bubble with markdown, typing indicator removed |
| 3.5 | Auto-scroll | Send multiple messages | Chat scrolls to bottom automatically |
| 3.6 | Textarea grow | Type multiline text | Textarea expands up to 150px |
| 3.7 | Empty send | Click send with empty input | Nothing happens |

### T4: Proposals

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 4.1 | Proposal appears | After response with proposals | Cards appear in right panel |
| 4.2 | Pending count | Multiple proposals | Badge shows "N pending" |
| 4.3 | Preview | Click 👁 on proposal | Detail section expands with fields |
| 4.4 | Accept | Click ✓ on proposal | Card turns green, atom created/updated |
| 4.5 | Reject | Click ✗ on proposal | Card turns red and fades |
| 4.6 | Accept All | Click "Accept All" | All pending proposals applied |
| 4.7 | Disabled buttons | After accept/reject | Action buttons grayed out |

### T5: ATD Context

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 5.1 | Open search | Click "+" in context bar | Modal with search input appears |
| 5.2 | Search atoms | Type in search | Results filter in real-time |
| 5.3 | Add context | Click an atom | Chip appears in context bar |
| 5.4 | Remove context | Click × on chip | Chip removed |
| 5.5 | Confirmation | Send with new context | Confirmation dialog with checkboxes |
| 5.6 | Dedup | Send again | Already-sent atoms show ✓, grayed out |
| 5.7 | Close modal | Press Escape | Modal closes |

### T6: Action History

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 6.1 | Accept notice | Accept a proposal | "✅ Accepted: CREATE @atom_id" in chat |
| 6.2 | Reject notice | Reject a proposal | "❌ Rejected: CREATE @atom_id" in chat |
| 6.3 | Forward to API | Send next message after accept | Request payload includes actions array |

### T7: Context Drift

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 7.1 | Counter badge | After 1st exchange | "1/10" badge in chat header |
| 7.2 | 10 exchanges | Send 10 messages with responses | Orange caution banner in chat |
| 7.3 | 15 exchanges | Continue to 15 | Yellow warning banner |
| 7.4 | Restart from warning | Click "New Session" on warning | Session resets, counter clears |

### T8: Session Management

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 8.1 | New Session | Click 🔄 New Session | Confirmation dialog, then full reset |
| 8.2 | Export | Click 📥 Export | Markdown file downloaded |
| 8.3 | Export content | Open exported file | Contains messages, proposals table, context list |
| 8.4 | Shortcut: send | Ctrl+Enter | Message sent |
| 8.5 | Shortcut: session | Ctrl+Shift+N | New session dialog |
| 8.6 | Shortcut: export | Ctrl+Shift+E | Export triggered |
| 8.7 | Shortcut: context | Ctrl+Shift+A | Context search opens |

### T9: Error Handling

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 9.1 | No API key | Remove GEMINI_API_KEY from .env | Error bubble: "GEMINI_API_KEY not configured" |
| 9.2 | Network error | Stop server, try sending | Error bubble with retry button |
| 9.3 | Retry | Click retry button | Re-sends the last message |

### T10: Visual Polish

| # | Test | Steps | Expected |
|---|------|-------|----------|
| 10.1 | Dark theme | All elements | Consistent dark palette |
| 10.2 | Message animation | Send/receive | Fade-in-up animation |
| 10.3 | Proposal animation | New proposals | Slide-in from right |
| 10.4 | Timestamps | Each message | Time displayed under bubble |
| 10.5 | Responsive (900px) | Resize window | Proposals move below chat |
| 10.6 | Responsive (600px) | Resize further | Compact headers, stacked chips |

## Verification Command

```bash
# Quick smoke test
cd webui && go build -o webui_bin . && ./webui_bin &
sleep 2
curl -s http://localhost:8081/api/info | jq .
curl -s http://localhost:8081/api/gemini/atoms | jq '. | length'
curl -s -X POST http://localhost:8081/api/gemini/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"Create a spec for a user registration flow"}]}' | jq .
kill %1
```

## Success Criteria

- [ ] All T1-T10 tests pass
- [ ] No JavaScript console errors
- [ ] Response times < 10s for Gemini calls
- [ ] All proposals can be applied without backend errors
- [ ] Session export produces valid Markdown
- [ ] Context drift warning appears at exchange 10
- [ ] Dark theme consistent with existing Explorer tab
