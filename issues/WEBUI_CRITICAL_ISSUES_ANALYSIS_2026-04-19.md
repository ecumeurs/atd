# WebUI Critical Issues Analysis
**Date**: 2026-04-19
**Scope**: ATD WebUI (atd/pkg/webui/)
**Status**: ✅ **FIXED** - All critical issues resolved

---

## ✅ Fixes Applied (2026-04-19)

### 1. Document Generation - FIXED
- Added missing assembly tasks to `.atd` config
- Increased provider timeouts (2s→10s remote, 500ms→5s local)
- All tasks now properly assigned to working models

### 2. Ctrl+K Search - FIXED
- Increased timeouts for embedding operations
- nomic-embed-text confirmed available on remote provider
- Search tested and working

### 3. Health Information - ADDED
- New `/api/health` endpoint
- Health indicator in UI header (🟢🟡🔴 status)
- Health details modal showing all providers and tasks

### 4. "Add Specific Atom" Bug - FIXED
- Fixed state import in documents.js (was using `window.state`)

---

## Executive Summary

Three critical issues affecting the WebUI have been identified and analyzed:

1. **Ctrl+K Search Broken** - Cannot find ATD atoms
2. **Document Generation Broken** - Cannot generate narratives from ATD graph
3. **Missing Health Information** - No aggregate project health dashboard

**Root Cause**: Missing LLM model configuration for critical tasks and incomplete UI implementation.

---

## Issue 1: Ctrl+K Search Broken

### Severity: HIGH

### Current Implementation

**Frontend**: `atd/pkg/webui/static/js/search.js`
```javascript
export async function searchAtoms(query) {
    const resp = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
    if (!resp.ok) throw new Error('Search failed');
    return resp.json();
}
```

**Backend**: `atd/pkg/webui/handlers.go` → `exploration.Search()`
```go
func SemanticSearch(query, dbPath string, limit int, scope string) ([]SearchResult, error) {
    queryEmb, err := ollama.QueryEmbed(query)  // ← Critical dependency
    if err != nil {
        return nil, fmt.Errorf("failed to embed query: %v", err)
    }
    // ... similarity search in SQLite database
}
```

**Embedding Provider**: `atd/pkg/ollama/provider.go`
```go
func QueryEmbed(text string) ([]float64, error) {
    res, err := ResolveProvider("embed")  // ← Looks for "embed" task
    if err != nil {
        return nil, err
    }
    if res.IsIDE {
        return nil, fmt.Errorf("embedding requires Ollama with nomic-embed-text — no IDE fallback available")
    }
    return Embed(res.BaseURL, res.Model, text)
}
```

### Root Cause

The "embed" task is configured for the `nomic-embed-text` model in `.atd`:

```json
"models": {
  "nomic-embed-text": {
    "tasks": ["embed"]
  }
}
```

However, if the configured Ollama providers (remote: http://192.168.1.10:11434, local: http://localhost:11434) are:
1. **Unreachable** - Network timeout or offline
2. **Missing the model** - nomic-embed-text not installed
3. **Malfunctioning** - API errors

Then search will fail silently or with errors.

### Current Config Analysis

```json
"providers": [
  {
    "name": "remote",
    "base_url": "http://192.168.1.10:11434",
    "timeout_ms": 2000
  },
  {
    "name": "local",
    "base_url": "http://localhost:11434",
    "timeout_ms": 500
  }
]
```

**Problem**: Only 2 second timeout for remote, 500ms for local. These are very short for embedding operations.

### Recommended Fix

**Short Term**:
1. Increase timeout values in `.atd` config:
   ```json
   {
     "name": "remote",
     "base_url": "http://192.168.1.10:11434",
     "timeout_ms": 10000  // Increase to 10s
   },
   {
     "name": "local",
     "base_url": "http://localhost:11434",
     "timeout_ms": 5000  // Increase to 5s
   }
   ```

2. Add fallback to grep search when embedding fails in `exploration/search.go`:
   ```go
   func Search(opts SearchOptions) ([]SearchResult, error) {
       if opts.Query != "" {
           results, err := SemanticSearch(opts.Query, opts.DBPath, opts.Limit, opts.Scope)
           if err != nil {
               log.Printf("Semantic search failed, falling back to grep: %v", err)
               return GrepSearch(opts.Query, opts.Root)
           }
           return results, nil
       }
       // ...
   }
   ```

**Medium Term**:
1. Add health check UI that shows provider status
2. Allow manual model selection in UI
3. Cache embeddings for common queries

---

## Issue 2: Document Generation Broken

### Severity: HIGH

### Current Implementation

**Frontend**: `atd/pkg/webui/static/js/documents.js`
```javascript
async function handleGenerate() {
    const intent = intentInput.value.trim();
    const starts = Array.from(startInputs).map(i => i.value);

    const doc = await generateDocument(intent, starts);  // ← Calls backend
    setupModal.style.display = 'none';
    showDocumentViewer(doc);
}
```

**Backend**: `atd/pkg/webui/handlers.go` → `exploration.Assemble()`

**Assembly Logic**: `atd/pkg/exploration/assemble.go`
```go
func Assemble(opts AssembleOptions) (string, error) {
    // ... gather atoms ...

    if opts.Structured {
        // Multi-pass LLM with these tasks:
        queryLayer("CUSTOMER")      // ← "assemble_layer_CUSTOMER"
        queryLayer("ARCHITECTURE")  // ← "assemble_layer_ARCHITECTURE"
        queryLayer("IMPLEMENTATION") // ← "assemble_layer_IMPLEMENTATION"

        // Final pass
        ollama.Query("assemble_final", finalPrompt, ...)
    } else {
        // Single-pass
        ollama.Query("assemble", requestPrompt, ...)  // ← "assemble"
    }
}
```

### Root Cause

**The assembly tasks are NOT configured in `.atd`!**

Current `.atd` models configuration:
```json
"models": {
  "llama3.2": { "tasks": ["audit_bloat", "intent_extract", "snapshot"] },
  "llama3.1:8b": { "tasks": ["audit_bloat", "intent_extract", "snapshot", "dissect", "recon"] },
  "deepseek-r1:7b": { "tasks": ["audit_code", "compare", "congruence", "reconcile", "fix_split", "audit_bloat", "intent_extract", "snapshot"] },
  "qwen2.5-coder:14b": { "tasks": ["dissect", "recon", "snapshot"] },
  "nomic-embed-text": { "tasks": ["embed"] }
}
```

**Missing tasks**:
- `assemble`
- `assemble_layer_CUSTOMER`
- `assemble_layer_ARCHITECTURE`
- `assemble_layer_IMPLEMENTATION`
- `assemble_final`

When these tasks are not found, the code falls back to `fallback_model` (llama3.2). But if llama3.2 is unavailable, it falls back to IDE agent with `ErrIDEFallback`.

**When IDE fallback happens in assemble.go**:
```go
if err == ollama.ErrIDEFallback {
    // Returns task delegation message instead of content
    taskList, _ := pipeline.WriteTaskList(...)
    msg := fmt.Sprintf("Task delegated to IDE Agent: %s", taskList)
    return msg + RenderMetadata(metadata), nil  // ← Just a message, not actual content
}
```

### Recommended Fix

**Immediate Fix - Add missing tasks to `.atd` config**:
```json
"models": {
  "llama3.2": {
    "tasks": [
      "audit_bloat",
      "intent_extract",
      "snapshot",
      "assemble",  // ← ADD
      "assemble_layer_CUSTOMER",  // ← ADD
      "assemble_layer_ARCHITECTURE",  // ← ADD
      "assemble_layer_IMPLEMENTATION",  // ← ADD
      "assemble_final"  // ← ADD
    ]
  },
  // ... other models
}
```

**Alternative - Use a better model for assembly**:
```json
"qwen2.5-coder:14b": {
  "tasks": [
    "dissect",
    "recon",
    "snapshot",
    "assemble",
    "assemble_layer_CUSTOMER",
    "assemble_layer_ARCHITECTURE",
    "assemble_layer_IMPLEMENTATION",
    "assemble_final"
  ]
}
```

**Long Term - Remove IDE Fallback from Assembly**:
Document generation should fail gracefully with a clear error message rather than delegating to IDE agent, since users expect direct results in the WebUI.

---

## Issue 3: Missing Health Information

### Severity: MEDIUM

### Current Implementation

**Backend**: `atd/pkg/webui/handlers.go`
```go
func (s *Server) handleStats(c *gin.Context) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    var total, covered, tested, orphans int
    if s.explorer.Graph != nil {
        for _, node := range s.explorer.Graph.Atoms {
            total++
            if len(node.Implementations) > 0 {
                covered++
            }
        }
        // @todo: implement from Explorer.TestLinks ← TODO comment
    }

    c.JSON(http.StatusOK, gin.H{
        "Total":        total,
        "SpecCoverage": covered,
        "TestCoverage": tested,  // Always 0 - not implemented
        "Orphans":      orphans,  // Always 0 - not implemented
    })
}
```

**Frontend**: The stats endpoint exists but there's no UI to display it.

### Root Cause

1. **Missing UI Component** - No health dashboard in the WebUI
2. **Incomplete Backend** - Test coverage and orphan counts are hardcoded to 0
3. **Missing Metrics** - No aggregate health summary

### Recommended Fix

**Short Term - Complete the stats handler**:
```go
func (s *Server) handleStats(c *gin.Context) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    var total, covered, tested, orphans int
    var byLayer = map[string]int{"CUSTOMER": 0, "ARCHITECTURE": 0, "IMPLEMENTATION": 0}
    var byStatus = map[string]int{"DRAFT": 0, "REVIEW": 0, "STABLE": 0}

    if s.explorer.Graph != nil {
        for _, node := range s.explorer.Graph.Atoms {
            total++
            if len(node.Implementations) > 0 {
                covered++
            }
            if len(node.Tests) > 0 {  // Assuming Tests field exists
                tested++
            }
            // Check for orphans (no parents in CUSTOMER layer)
            if node.Layer != "CUSTOMER" && len(node.Parents) == 0 {
                orphans++
            }
            byLayer[node.Layer]++
            byStatus[node.Status]++
        }
    }

    c.JSON(http.StatusOK, gin.H{
        "Total":        total,
        "SpecCoverage": float64(covered) / float64(total),
        "TestCoverage": float64(tested) / float64(total),
        "Orphans":      orphans,
        "ByLayer":      byLayer,
        "ByStatus":     byStatus,
    })
}
```

**Medium Term - Add health dashboard UI**:

Add to `index.html`:
```html
<div id="health-dashboard" class="health-dashboard">
    <div class="health-metric">
        <span class="health-label">Total Atoms</span>
        <span class="health-value" id="health-total">0</span>
    </div>
    <div class="health-metric">
        <span class="health-label">Coverage</span>
        <span class="health-value" id="health-coverage">0%</span>
    </div>
    <div class="health-metric">
        <span class="health-label">Orphans</span>
        <span class="health-value" id="health-orphans">0</span>
    </div>
</div>
```

Add to `app.js` or new `health.js`:
```javascript
async function loadHealthStats() {
    const resp = await fetch('/api/stats');
    const stats = await resp.json();
    document.getElementById('health-total').textContent = stats.Total;
    document.getElementById('health-coverage').textContent = (stats.SpecCoverage * 100).toFixed(1) + '%';
    document.getElementById('health-orphans').textContent = stats.Orphans;
}
```

---

## Issue 4: "Add Specific Atom" Popup Not Working

### Severity: MEDIUM

### Current Implementation

**Frontend**: `atd/pkg/webui/static/js/documents.js`
```javascript
function openAtomPicker(e) {
    // ... creates inline picker ...

    input.addEventListener('input', () => {
        const query = input.value.toLowerCase();
        results.innerHTML = '';
        if (query.length < 2) return;

        // ← PROBLEM: Uses window.state?.atoms directly
        const matches = window.state?.atoms?.filter(a =>
            a.id.toLowerCase().includes(query) ||
            (a.human_name && a.human_name.toLowerCase().includes(query))
        ) || [];
        // ...
    });
}
```

### Root Cause

The code uses `window.state?.atoms` but the state module exports `state` as a named export, not as a global window property.

Looking at `state.js`:
```javascript
export const state = { atoms: [] };  // ← Named export
```

It's not attached to `window` anywhere. The `window.state` would be undefined.

### Recommended Fix

**Fix in `documents.js`**:
```javascript
// Import state at the top
import { state } from './state.js';

function openAtomPicker(e) {
    // ... existing code ...

    input.addEventListener('input', () => {
        const query = input.value.toLowerCase();
        results.innerHTML = '';
        if (query.length < 2) return;

        // Use imported state instead of window.state
        const matches = state?.atoms?.filter(a =>
            a.id.toLowerCase().includes(query) ||
            (a.human_name && a.human_name.toLowerCase().includes(query))
        ) || [];
        // ...
    });
}
```

---

## Configuration Issues Summary

### Current `.atd` LLM Configuration Problems

1. **Missing Assembly Tasks** - Document generation tasks not configured
2. **Short Timeouts** - 2s/500ms too short for LLM operations
3. **No Health Endpoint** - Can't check if providers are online
4. **No Fallback for Search** - No grep fallback when embedding fails

### Recommended `.atd` Updates

```json
{
  "llm": {
    "providers": [
      {
        "name": "remote",
        "base_url": "http://192.168.1.10:11434",
        "timeout_ms": 10000
      },
      {
        "name": "local",
        "base_url": "http://localhost:11434",
        "timeout_ms": 5000
      },
      {
        "name": "ide_agent",
        "type": "passthrough"
      }
    ],
    "models": {
      "llama3.2": {
        "tasks": [
          "audit_bloat",
          "intent_extract",
          "snapshot",
          "assemble",
          "assemble_layer_CUSTOMER",
          "assemble_layer_ARCHITECTURE",
          "assemble_layer_IMPLEMENTATION",
          "assemble_final"
        ]
      },
      "qwen2.5-coder:14b": {
        "tasks": [
          "dissect",
          "recon",
          "snapshot"
        ],
        "priority": 10  // Higher priority for complex tasks
      },
      "deepseek-r1:7b": {
        "tasks": [
          "audit_code",
          "compare",
          "congruence",
          "reconcile",
          "fix_split",
          "audit_bloat",
          "intent_extract",
          "snapshot"
        ]
      },
      "nomic-embed-text": {
        "tasks": ["embed"]
      }
    },
    "fallback_model": "llama3.2",
    "health_ttl_ms": 300000,
    "model_ttl_ms": 300000
  }
}
```

---

## Next Steps

### Priority 1: Fix Document Generation
1. Add missing assembly tasks to `.atd` config
2. Test document generation flow end-to-end
3. Add error handling for LLM failures

### Priority 2: Fix Search
1. Increase provider timeouts
2. Add grep fallback when embedding fails
3. Test search with various queries

### Priority 3: Fix Add Atom Picker
1. Fix state import in documents.js
2. Test inline picker functionality

### Priority 4: Add Health Dashboard
1. Complete stats handler implementation
2. Add health metrics UI
3. Add provider status indicators

---

## Related Issues

- [ISS-061](ISS-061_20260402_webui_document_generation_bugs.md) - Document generation bugs
- [ISS-063](ISS-063_20260403_webui_spec_builder_enhancements.md) - Spec builder enhancements
- [ISS-064](ISS-064_20260403_webui_search_regression_and_perf.md) - Search regression
- [ISS-068](ISS-068_20260403_webui_atd_health_indicators.md) - Health indicators
- [ISS-069](ISS-069_20260407_webui_explorer_global_health.md) - Global health dashboard
- [ISS-089](ISS-089_20260419_webui_search_docgen_regression.md) - Recent search/docgen regression
