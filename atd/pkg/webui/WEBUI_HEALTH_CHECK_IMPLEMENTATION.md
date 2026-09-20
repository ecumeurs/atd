# WebUI Health Check & Fixes Implementation
**Date**: 2026-04-19

---

## Changes Made

### 1. Fixed Document Generation (Configuration)

Updated `.atd` config files in 3 locations:
- `/home/bastien/work/atd/.atd`
- `/home/bastien/work/atd/atd_management_skill/.atd`
- `/home/bastien/work/atd/upsilon-hub/.atd`

**Changes**:
- Increased timeouts: remote (2s→10s), local (500ms→5s)
- Added missing assembly tasks to `llama3.2`:
  - `assemble`
  - `assemble_layer_BUSINESS`
  - `assemble_layer_ARCHITECTURE`
  - `assemble_layer_IMPLEMENTATION`
  - `assemble_final`
- Added priority to `qwen2.5-coder:14b`
- Added health/model TTLs (300000ms = 5 minutes)

### 2. Added Health Check Endpoint

**Backend** (`atd/pkg/webui/handlers.go`):
- New endpoint: `GET /api/health`
- Returns provider status and task availability
- Shows which provider each task is assigned to

**Frontend API** (`atd/pkg/webui/static/js/api.js`):
- Added `fetchHealth()` function

**Health Module** (`atd/pkg/webui/static/js/health.js` - new file):
- Auto-checks health on load and every 30 seconds
- Shows status indicator in header (🟢 All Ready, 🟡 Partial, 🔴 Offline)
- Click to open health details modal showing:
  - All providers with their status and available models
  - All tasks with their assigned provider:model

### 3. Fixed "Add Specific Atom" Bug

**File**: `atd/pkg/webui/static/js/documents.js`
- Changed `window.state?.atoms` to `state?.atoms` (uses imported state module)

### 4. Enhanced Stats Endpoint

**File**: `atd/pkg/webui/handlers.go`
- Fixed `handleStats` to use correct field name (`HasTests` instead of `Tests`)
- Added `ByLayer` and `ByStatus` breakdowns
- Properly counts orphans (non-BUSINESS atoms with no parents)

### 5. UI Enhancements

**File**: `atd/pkg/webui/static/index.html`
- Added health indicator badge in header
- Added health details modal

**File**: `atd/pkg/webui/static/styles.css`
- Added health badge styles (color-coded by status)
- Added modal overlay and content styles
- Added health provider and task item styles

**File**: `atd/pkg/webui/static/js/app.js`
- Imported and initialized health module

---

## Health Endpoint Response Example

```json
{
  "providers": [
    {
      "name": "remote",
      "type": "ollama",
      "base_url": "http://192.168.1.10:11434",
      "status": "available",
      "models": ["llama3.1:8b", "nomic-embed-text:latest", "llama3.2:latest", "llama3.2:1b", "deepseek-r1:7b", "qwen2.5-coder:14b"]
    },
    {
      "name": "local",
      "type": "ollama",
      "base_url": "http://localhost:11434",
      "status": "available",
      "models": ["nomic-embed-text:latest", "llama3.2:latest"]
    },
    {
      "name": "ide_agent",
      "type": "passthrough",
      "status": "available"
    }
  ],
  "tasks": {
    "assemble": "remote:llama3.2:latest",
    "assemble_final": "remote:llama3.2:latest",
    "assemble_layer_ARCHITECTURE": "remote:llama3.2:latest",
    "assemble_layer_BUSINESS": "remote:llama3.2:latest",
    "assemble_layer_IMPLEMENTATION": "remote:llama3.2:latest",
    "audit_bloat": "remote:llama3.2:latest",
    "audit_code": "remote:deepseek-r1:7b",
    "compare": "remote:deepseek-r1:7b",
    "congruence": "remote:deepseek-r1:7b",
    "dissect": "remote:qwen2.5-coder:14b",
    "embed": "remote:nomic-embed-text:latest",
    "fix_split": "remote:deepseek-r1:7b",
    "intent_extract": "remote:llama3.2:latest",
    "recon": "remote:qwen2.5-coder:14b",
    "reconcile": "remote:deepseek-r1:7b",
    "snapshot": "remote:qwen2.5-coder:14b"
  }
}
```

---

## Task Mapping Summary

| Task | Model | Provider |
|------|-------|----------|
| embed | nomic-embed-text:latest | remote |
| assemble, assemble_layer_*, assemble_final | llama3.2:latest | remote |
| audit_bloat, intent_extract | llama3.2:latest | remote |
| dissect, recon, snapshot | qwen2.5-coder:14b | remote |
| audit_code, compare, congruence, reconcile, fix_split | deepseek-r1:7b | remote |

---

## Testing

**Tested**:
- ✅ Health endpoint returns correct provider and task status
- ✅ All 15 tasks properly configured and assigned
- ✅ Search endpoint working (tested with "action" query)
- ✅ nomic-embed-text available on both remote and local

**To Test**:
- Document generation flow in WebUI
- Ctrl+K search in WebUI
- Health indicator in UI

---

## Next Steps

1. **Start WebUI** and verify health indicator shows "All Ready" (🟢)
2. **Test document generation** by generating a document from selected atoms
3. **Test Ctrl+K search** to ensure it finds atoms properly
4. **Test "Add Specific Atom"** to verify the picker works
