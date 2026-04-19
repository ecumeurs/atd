# ATD Issues Summary
**Date**: 2026-04-19
**Action**: Reviewed WebUI issues, archived 7 resolved issues, analyzed critical WebUI problems

---

## Issues Archived (Resolved)

Moved to `issues/resolved/`:
- ISS-071: ATD Indexing System Failure
- ISS-073: ATD Link Resolution Failure
- ISS-083: Granular Verify JSON Recap
- ISS-084: Project-Wide Verify Mode
- ISS-086: ATD Index Stale Entries
- ISS-087: WebUI Foundation Customer Missing
- ISS-088: WebUI Selection Visibility

**Remaining**: ~50 issues in `issues/` (includes summary files)

---

## Critical WebUI Issues Identified

### 1. Ctrl+K Search Broken (HIGH)

**Root Cause**: Semantic search requires `nomic-embed-text` model, but:
- Provider timeouts are too short (2s remote, 500ms local)
- No fallback to grep search when embedding fails
- Network/provider issues cause silent failures

**Fix**: Increase timeouts, add grep fallback, add provider health UI

### 2. Document Generation Broken (HIGH)

**Root Cause**: Assembly tasks not configured in `.atd`:
- Missing: `assemble`, `assemble_layer_CUSTOMER`, `assemble_layer_ARCHITECTURE`, `assemble_layer_IMPLEMENTATION`, `assemble_final`
- Falls back to IDE agent which returns task delegation message instead of content

**Fix**: Add missing tasks to `.atd` models configuration for llama3.2 or better model

### 3. Missing Health Information (MEDIUM)

**Root Cause**:
- `/api/stats` exists but returns incomplete data (test coverage/orphans always 0)
- No UI component to display health metrics
- No aggregate project health dashboard

**Fix**: Complete stats handler, add health dashboard UI

### 4. "Add Specific Atom" Bug (MEDIUM)

**Root Cause**: Uses `window.state?.atoms` but state is a named export, not global

**Fix**: Import `state` from `./state.js` instead of using `window.state`

---

## Configuration Issues

Current `.atd` LLM config has problems:
1. **Missing assembly tasks** - Document generation broken
2. **Short timeouts** - 2s/500ms too short for LLM operations
3. **No health checks** - Can't verify provider status

---

## Outdated Issues

Some issues may be outdated due to:
- Implicit fixes from other changes (e.g., ISS-087/ISS-088 may have been fixed by UI refactors)
- Changed requirements (some features may have been deprioritized)
- Moved to different implementation approach

**Recommendation**: Review remaining 50 issues and mark outdated ones as "Superseded" or move to resolved with explanation.

---

## Analysis Document

Full technical analysis saved to: `issues/WEBUI_CRITICAL_ISSUES_ANALYSIS_2026-04-19.md`

---

## Quick Fixes Needed

1. **Update `.atd` config** - Add missing assembly tasks
2. **Increase timeouts** - 10s remote, 5s local
3. **Fix documents.js** - Import state properly
4. **Complete handleStats** - Calculate real test coverage and orphans
