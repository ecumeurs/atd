# ATD Enhancement Proposals

**Date:** 2026-04-18  
**Purpose:** Configurable enhancements for ATD system improvements  
**Approach:** Non-breaking, backward-compatible additions

---

## Enhancement 1: Enhanced File Discovery (ISS-071)

### Configuration Options
Add to `.atd` config:
```json
{
  "discovery": {
    "method": "walk|git-ls-files",
    "gitignore_patterns": [".vscode", "node_modules", "vendor"],
    "max_depth": 10
  }
}
```

### Implementation

**Modify `atd/cmd/atd/cmd/index.go`:**
- Add discovery method flag
- Support both filepath.Walk() and git ls-files approaches
- Add configurable gitignore patterns
- Add max depth option

**Benefits:**
- Users can choose discovery method based on project structure
- Git-based discovery remains default for backward compatibility
- Improved coverage for large projects without excessive depth

---

## Enhancement 2: Enhanced Link Resolution (ISS-073)

### Configuration Options
Add to `.atd` config:
```json
{
  "link_resolution": {
    "languages": {
      "go": {
        "spec_link_pattern": "// @spec-link \\[\\[([^\\]]+)\\]",
        "test_link_pattern": "// @test-link \\[\\[([^\\]]+)\\]",
        "comment_patterns": ["//", "/*", "*"]
      },
      "php": {
        "spec_link_pattern": "/** @spec-link \\[\\[([^\\]]+)\\] **/",
        "test_link_pattern": "/** @test-link \\[\\[([^\\]]+)\\] **/",
        "comment_patterns": ["//", "/*", "*"]
      },
      "javascript": {
        "spec_link_pattern": "// @spec-link \\[\\[([^\\]]+)\\]",
        "test_link_pattern": "// @test-link \\[\\[([^\\]]+)\\]",
        "comment_patterns": ["//", "/*", "*"]
      },
      "vue": {
        "spec_link_pattern": "// @spec-link \\[\\[([^\\]]+)\\]",
        "test_link_pattern": "// @test-link \\[\\[([^\\]]+)\\]",
        "comment_patterns": ["//", "/*", "*"]
      }
    }
  }
}
```

### Implementation

**Modify `atd/pkg/exploration/explorer.go`:**
- Add language-specific link parsing based on configuration
- Support configurable patterns for each language
- Line number tracking for precise linking

**Benefits:**
- Multi-language support becomes configurable
- Pattern matching can be tuned per language
- Better line number tracking for precise @spec-link location

---

## Enhancement 3: Type-Aware Orphan Detection (ISS-072)

### Configuration Options
Add to `.atd` config:
```json
{
  "orphan_detection": {
    "excluded_types": ["MODULE", "SPECIFICATION", "USECASE", "USER_STORY"],
    "hierarchical_check": true,
    "customer_layer_exception": true
  }
}
```

### Implementation

**Modify `atd/cmd/atd/cmd/crawl.go`:**
- Add orphan detection mode options
- Implement type-aware orphan logic
- Support hierarchical orphan checking

**Benefits:**
- Eliminates false positives from MODULE and high-level atoms
- Respects ATD layer hierarchy
- Configurable per project

---

## Implementation Priority

**Phase 1 (Week 1):** 
- Add discovery configuration to index command
- Implement enhanced link parsing with language support

**Phase 2 (Week 2):**
- Add orphan detection configuration to crawl command  
- Implement type-aware orphan detection logic

**Phase 3 (Week 3):**
- Testing with various configurations
- Documentation updates

**Testing Strategy:**
1. Test with existing projects (backward compatibility)
2. Test with new configurations (enhanced discovery)
3. Compare performance and accuracy
4. Document both testing sessions

---

## Rollout Plan

1. **Add configuration fields** to .atd schema
2. **Implement discovery options** in index command
3. **Add language patterns** to configuration parsing
4. **Update exploration package** with configurable link parsing
5. **Add orphan detection options** to crawl command
6. **Implement type-aware orphan logic** in crawler
7. **Test with upsilon-hub** using new configurations
8. **Compare with baseline** (upsilon-hub/.atd config)
9. **Document findings** in both testing sessions

---

## Success Metrics

### Coverage Improvements
- **Expected:** 250+ code files discovered per upsilon-hub
- **Target:** ≥95% file discovery rate

### Orphan Detection Accuracy
- **Expected:** ≤10% false positive rate (vs current 85%)
- **Target:** Proper exclusion of MODULE and hierarchy atoms

---

This approach provides **significant improvements** while maintaining **backward compatibility** and **avoiding major restructuring**.