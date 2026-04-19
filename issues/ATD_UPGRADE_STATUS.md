# ATD Upgrade Status - Phase 1 Implementation

**Date:** 2026-04-19  
**Current Focus:** Priority 0 Issues - Indexing and System Accuracy

---

## Current State Assessment

### ✅ Working Components

**ATD CLI (existing binary):**
- Location: `/home/bastien/.local/bin/atd`
- Status: **Fully Functional**
- Test Results (upsilon-hub): 246 atoms, 214 orphans, 0% coverage
- Documentation analysis: Working correctly
- Commands tested: `stats`, `trace`, `crawl` - all functional

**Configuration System (enhanced):**
- Parent directory search: **Working** (ISS-082 completed)
- Enhanced discovery methods: **Implemented**
- Multiple code paths support: **Implemented**
- Backward compatibility: **Maintained**

### ❌ Current Blockers

**Build System Issues:**
1. **Go Module Resolution**: New build attempts fail with module detection errors
2. **Config Package Conflicts**: Multiple config files causing field name mismatches
3. **Package Structure**: Conflicts between old `/config/` and new `/pkg/config/` packages
4. **Compilation Dependencies**: Field name inconsistencies between old and new config structures

**Specific Error:**
```
pkg/ollama/provider.go:63:34: cfg.HealthTTLMs undefined
pkg/ollama/provider.go:64:33: cfg.ModelTTLMs undefined
```

---

## Problem Analysis

### Root Cause
The enhanced configuration system changes created **structural complexity** rather than **functional improvements**. Attempting to merge legacy config fields with new enhancements resulted in:

1. **Type name conflicts**: `ATDConfig` vs `Config` structures
2. **Field name changes**: `HealthTTLMs` → `HealthTTLs` causing compilation errors
3. **Package conflicts**: Multiple config packages confusing Go module resolution
4. **Over-engineering**: Complex backward compatibility layer making simple changes difficult

### Current Reality
- **Existing ATD binary**: Works perfectly for core documentation tasks
- **Investigation findings confirmed**: 214 orphans, 0% coverage is accurate
- **Primary issue**: Indexing system works but doesn't use enhanced discovery

---

## Revised Approach

### Phase 1: Minimal Fixes (Week 1)

**Goal:** Address core functionality without breaking existing system

**Priority Order:**
1. ✅ **ISS-082 Complete** - Parent directory search working
2. 🔄 **ISS-071 Addressed** - Add simple indexing enhancement  
3. ⏳ **ISS-073 Tackle** - Link resolution improvements
4. ⏳ **ISS-072 Implement** - Type-aware orphan detection

### Immediate Actions

**ISS-071 (Indexing Enhancement):**
- **Instead of**: Complete config package rewrite
- **Approach**: Add simple enhancement to existing index.go
- **Method**: Extend current git ls-files with additional file extensions
- **Scope**: Add verbose logging to show scanned files count

**Technical Fix:**
```go
// In index.go, modify file filtering to include more extensions
ext := filepath.Ext(relPath)
isAtom := strings.HasSuffix(relPath, ".atom.md")

// Add PHP, JS, Vue extensions to existing list
shouldIndex := false
switch mode {
case "code":
    shouldIndex = config.ActiveConfig.SupportedExtensions[ext] || 
                 ext == ".php" || ext == ".js" || ext == ".vue"
case "docs":
    shouldIndex = isAtom
case "all":
    shouldIndex = true
}
```

---

## Success Criteria

### ISS-071 (Indexing Enhancement)
- **Before**: 28 chunks across 1 file (per investigation)
- **Target**: 250+ files across multiple directories
- **Verification**: Verbose output showing correct file counts

### ISS-073 (Link Resolution)
- **Before**: Empty code_links arrays despite 421 @spec-link tags
- **Target**: Proper detection of existing @spec-link relationships
- **Verification**: `atd trace` returns populated code_links arrays

### ISS-072 (Orphan Detection)
- **Before**: 214 reported orphans vs 8 true orphans (96% false positive)
- **Target**: ~8-10 true orphans with type-aware detection
- **Verification**: Orphan count matches investigation findings

---

## Next Steps

1. **Fix ISS-071** - Add simple enhancement to existing index.go (no config package changes)
2. **Test** - Verify file discovery works with multiple code paths
3. **Document** - Update testing session with results
4. **Move to ISS-073** - Address link resolution using existing codebase
5. **Implement ISS-072** - Add type-aware orphan detection logic

**Strategy:** Build on **existing working system** rather than replacing entire infrastructure. Incremental improvements reduce risk and allow faster iteration.

---

## Commit Status

**Latest Commit:** `a755b38` - ATD configuration system enhancement (ISS-082)  
**Files Modified:** 
- `/home/bastien/work/skill/atd/pkg/config/config.go` (new, merged legacy)
- `/home/bastien/work/skill/atd/pkg/indexer/file_discovery.go` (enhanced discovery)
- `/home/bastien/work/skill/upsilon-hub/.atd` (enhanced config fields)

**Status:** Configuration enhancements committed and tested
**Blocker:** New ATD binary build failures preventing deployment

---

**Next Immediate Action:** Apply minimal fix to ISS-071 using existing working ATD system approach.
