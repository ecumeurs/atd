# ATD P0/P1 Test Execution Results

## Test Overview
**Date**: 2026-04-19  
**Target Project**: `upsilon-hub` (reverted modification state)  
**ATD Version**: Unified CLI (2026-04-19 build)  
**Method**: Execution of [ATD_TESTING_PLAN_UPSILON_HUB.md](file:///home/bastien/work/skill/claude_plans/ATD_TESTING_PLAN_UPSILON_HUB.md)

---

## 1. Executive Summary 🚨

The test execution highlights a **fundamental gap between tool completion and operational readiness**. While the infrastructure (MCP, Git integration, performance) is solid, the core value proposition—**traceability from code to documentation**—is currently failing in the baseline configuration.

### Critical Blockers:
- **Index Blindness**: `atd index` failed to discover 80% of the project (missed submodules).
- **Tag Parsing Failure**: Even for indexed files, the tool **failed to capture `@spec-link` tags** in chunks.
- **Reporting Stagnation**: Stats remain at the "Broken" baseline (0% coverage, 214 orphans).

---

## 2. Success Metrics Comparison 📊

| Metric | Baseline (Investigation) | Target (Success Criteria) | Actual (Current Run) | Status |
|---|---|---|---|---|
| **Indexed Files** | 1 (28 chunks) | **250+ files** | 43 files (55 chunks) | 🔴 |
| **@spec-link Detection Rate**| 0% | **100%** | 0% (Tags ignored) | 🔴 |
| **Reported Orphans** | 214 (88%) | **< 10 (3%)** | 214 (88%) | 🔴 |
| **Coverage Ratio** | 0% | **~82%** | 0% | 🔴 |
| **MCP Schema Cleanliness** | Leaky | **Clean (No paths)** | Clean | 🟢 |
| **Git Integration** | Manual/Broken | **Automatic (.gitignore)**| Verified (.gitignore) | 🟢 |

---

## 3. Detailed Category Results

### Category 1: Core Tooling (P0)
- **T1.1 Indexing System**: 🔴 **FAILED**. Discovered files only in `upsilonbattle`. Submodules (`battleui`, `upsilonapi`) were excluded.
- **T1.2 Link Resolution**: 🔴 **FAILED**. `atd trace` failed to detect links in `ruler.go` despite them being present in the source.
- **T1.3 Orphan Detection**: 🔴 **FAILED**. Inherited the "blind" state of the index.

### Category 2: Language-Agnostic (P1)
- **T2.1 Verification**: 🔴 **FAILED**. "No @spec-link tags found" error due to parsing failure.
- **T2.2 Git Integration**: 🟢 **PASSED**. Correctly respects `.gitignore` during `atd index`.

### Category 3: Performance (P1)
- **T3.1 Indexing Performance**: 🟢 **PASSED**. Incremental re-indexing takes <1s for 43 files.
- **T3.2 Audit Performance**: 🟢 **PASSED**. Audits 246 atoms with LLM-backed bloat detection.

### Category 4: MCP Integration (P2)
- **T4.1 Parameter Cleanup**: 🟢 **PASSED**. Internal paths (src, db, docs) hidden from Agent LLM. descriptions are agent-centric.

---

## 4. Root Cause Analysis

### Discovery Method Limitation
The current default discovery method (relying on `git ls-files` without submodule recursion) prevents the tool from seeing 80% of the project's source code. Reverting the `.atd` config removed the mandatory `code_paths` that provided the workaround for this.

### Tag Extraction Bug
There is a confirmed bug in the `atd index` chunking logic. While it correctly segments code blocks, it appears to **strip or ignore comments containing `@spec-link` tags** during the embedding phase, rendering the semantic index useless for traceability.

---

## 5. Immediate Recommendations 🛠️

1. **Restore Repo-Specific Config**: Re-apply `discovery_method: walk` and explicit `code_paths` to `.atd`.
2. **Fix Parsing Engine**: Update the indexing logic to ensure `@spec-link` and `@test-link` tags are preserved within chunks.
3. **Submodule Awareness**: Update the `git` discovery method to optionally recurse into submodules.

---

## 6. Conclusion

The "ATD System" as a set of MCP tools is **structurally sound**, but its **data acquisition layer** is currently incompatible with the complex multi-repo nature of `upsilon-hub` in its clean state. The tools work, but they are currently "blind" to the documentation links they are meant to track.
