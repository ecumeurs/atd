# ATD Testing Plan Using upsilon-hub as Test Bed

**Date:** 2026-04-18  
**Purpose:** Comprehensive testing of implemented ATD features using upsilon-hub as real-world test case  
**Based on:** LOT 4 Core issues and investigation findings  

---

## Test Environment Setup

### Test Bed Configuration
- **Test Project:** `/home/bastien/work/skill/upsilon-hub`
- **ATD Tools:** `/home/bastien/work/skill/atd`
- **Documentation:** `/home/bastien/work/skill/upsilon-hub/docs`
- **Code:** Mixed Go, PHP, JavaScript/Vue codebase

### Test Goals
1. **Validate implemented features** that haven't been tested
2. **Verify bug fixes** from LOT 4 issues
3. **Test language-agnostic capabilities** across multiple languages
4. **Validate investigation findings** about false orphan reporting
5. **Performance benchmarking** for large-scale codebase

---

## Test Categories

### Category 1: Core Tooling Fixes (P0 Priority)

#### Test 1.1: ATD Indexing System Validation
**Issue:** ISS-071 (ATD Indexing System Failure)  
**Status:** Implementation Complete, Needs Validation

**Test Cases:**
- **T1.1.1:** Run `atd index` on upsilon-hub and verify file scanning
  - Expected: Index 250+ files across multiple directories
  - Actual: Record scanned file count
  - Success: Scan coverage ≥ 95% of known files

- **T1.1.2:** Verify multi-language file extension handling
  - Expected: Proper scanning of `.go`, `.php`, `.js`, `.vue` files
  - Test: Check indexed files by extension type
  - Success: All expected extensions indexed correctly

- **T1.1.3:** Test caching and incremental updates
  - Expected: Unchanged files skipped on re-index
  - Test: Modify one file, re-index, verify only changed file re-processed
  - Success: Incremental reindexing works correctly

**Validation Command:**
```bash
cd /home/bastien/work/skill/upsilon-hub
atd index --dir . --verbose
```

---

#### Test 1.2: ATD Link Resolution Validation  
**Issue:** ISS-073 (ATD Link Resolution Failure)  
**Status:** Implementation Complete, Needs Validation

**Test Cases:**
- **T1.2.1:** Verify @spec-link detection in Go files
  - Expected: Detect all 21 @spec-link tags in `upsilonbattle/battlearena/ruler/ruler.go`
  - Test: Run `atd trace mech_action_economy_action_cost_rules`
  - Success: Non-empty code_links array with correct file paths

- **T1.2.2:** Verify @spec-link detection in PHP files  
  - Expected: Detect all 12 @spec-link tags in `battleui/app/Http/Controllers/API/AuthController.php`
  - Test: Run `atd trace api_auth_login`
  - Success: Code links point to correct PHP controller methods

- **T1.2.3:** Verify @spec-link detection in Vue files
  - Expected: Proper detection in JavaScript/Vue components
  - Test: Run `atd trace ui_leaderboard`
  - Success: Code links point to correct Vue components

**Validation Command:**
```bash
cd /home/bastien/work/skill/upsilon-hub
atd trace mech_action_economy_action_cost_rules
atd trace api_auth_login  
atd trace ui_leaderboard
```

---

#### Test 1.3: ATD Orphan Detection Validation
**Issue:** ISS-072 (ATD Orphan Detection Logic Flaws)  
**Status:** Implementation Complete, Needs Validation

**Test Cases:**
- **T1.3.1:** Verify MODULE type exclusion
  - Expected: MODULE atoms not flagged as orphans
  - Test: Run `atd crawl --gaps` and check MODULE atoms
  - Success: No MODULE atoms in orphaned list

- **T1.3.2:** Verify CUSTOMER layer handling
  - Expected: Customer layer atoms with implemented children not flagged as orphans
  - Test: Run `atd trace uc_player_login`, check child coverage
  - Success: Customer atoms not flagged if children implemented

- **T1.3.3:** Validate true orphan count
  - Expected: Only ~8 true orphans (not 214)
  - Test: Run `atd crawl --gaps`, count results
  - Success: Orphan count ≤ 10 (allowing for some edge cases)

**Validation Command:**
```bash
cd /home/bastien/work/skill/upsilon-hub
atd crawl --gaps
atd stats
```

---

### Category 2: Language-Agnostic Features (P1 Priority)

#### Test 2.1: Language-Agnostic Verification  
**Issue:** ISS-044 (ATD Verify Hardcoded to Go Testing)  
**Status:** Implementation Complete, Needs Validation

**Test Cases:**
- **T2.1.1:** Verify Go project verification
  - Expected: Uses `go test` command for Go files
  - Test: Run `atd verify HEAD~1 HEAD` on upsilonapi directory
  - Success: Runs Go tests, produces correct audit output

- **T2.1.2:** Verify PHP project verification
  - Expected: Uses PHP test command (e.g., `php artisan test`) for PHP files
  - Test: Run `atd verify HEAD~1 HEAD` on battleui directory
  - Success: Runs PHP tests, produces correct audit output

- **T2.1.3:** Verify git diff integration
  - Expected: Proper git diffing between commits
  - Test: Run `atd verify HEAD~5 HEAD` (5 commits range)
  - Success: Analyzes only changes in that commit range

**Validation Commands:**
```bash
cd /home/bastien/work/skill/upsilon-hub
atd verify HEAD~1 HEAD  # Single commit diff
atd verify HEAD~5 HEAD  # Multi-commit diff
```

---

#### Test 2.2: Git Integration and Discovery
**Issues:** ISS-059, ISS-060 (Exploration Git Integration)  
**Status:** Implementation Complete, Needs Validation

**Test Cases:**
- **T2.2.1:** Verify git ls-files discovery
  - Expected: Uses `git ls-files` for language-agnostic file discovery
  - Test: Run `atd index` and check file discovery method
  - Success: Uses git for file discovery, respects .gitignore

- **T2.2.2:** Verify .gitignore respect
  - Expected: Ignores files matching .gitignore patterns
  - Test: Add test file to .gitignore, re-index, verify ignored
  - Success: Ignored files not indexed

- **T2.2.3:** Verify multi-language file discovery
  - Expected: Discovers Go, PHP, JS, Vue files correctly
  - Test: Check discovered files by language type
  - Success: All expected languages discovered

**Validation Commands:**
```bash
cd /home/bastien/work/skill/upsilon-hub
echo "test.tmp" >> .gitignore
atd index --verbose
# Verify test.tmp is not indexed
```

---

### Category 3: Performance and Scalability (P1 Priority)

#### Test 3.1: Indexing Performance  
**LOT 4 Theme:** Performance & Scalability (Status: COMPLETED)  
**Implementation:** Worker pool (10 concurrent threads) and SQLite caching

**Test Cases:**
- **T3.1.1:** Measure full indexing performance
  - Expected: Index 250+ files in reasonable time
  - Test: Time `atd index --force` on upsilon-hub
  - Success: Completes in <5 minutes for full reindex

- **T3.1.2:** Verify concurrent embedding generation
  - Expected: Uses worker pool for parallel processing
  - Test: Monitor CPU usage during indexing, check for parallelism
  - Success: Multiple CPU cores utilized during embedding

- **T3.1.3:** Test caching effectiveness
  - Expected: Unchanged files skipped via SQLite cache
  - Test: Run `atd index` twice, measure second run time
  - Success: Second run significantly faster (<30% of first run)

**Validation Commands:**
```bash
cd /home/bastien/work/skill/upsilon-hub
time atd index --force  # Full reindex
time atd index          # Cached reindex
```

---

#### Test 3.2: Audit Performance  
**LOT 4 Theme:** Performance & Scalability (Status: COMPLETED)  
**Implementation:** SQLite-based fallback cache for .atd_audit.db

**Test Cases:**
- **T3.2.1:** Measure audit performance on large atom set
  - Expected: Audit 243 atoms efficiently
  - Test: Time `atd audit` on upsilon-hub docs
  - Success: Completes in <2 minutes

- **T3.2.2:** Verify caching effectiveness
  - Expected: Skips unmodified atoms based on mtime
  - Test: Run `atd audit` twice, second run should be faster
  - Success: Second run significantly faster (<50% of first run)

- **T3.2.3:** Test SQLite database functionality
  - Expected: .atd_audit.db created and used properly
  - Test: Check database file creation and content
  - Success: Database contains proper cache entries

**Validation Commands:**
```bash
cd /home/bastien/work/skill/upsilon-hub
time atd audit
time atd audit  # Should be faster due to cache
ls -la .atd_audit.db
```

---

### Category 4: MCP Tools Integration (P2 Priority)

#### Test 4.1: MCP Tools Parameter Cleanup  
**Issue:** ISS-045 (MCP Tools Refactor and Cleanup)  
**Status:** Partially Implemented, Needs Validation

**Test Cases:**
- **T4.1.1:** Verify parameter leak removal
  - Expected: No leaky parameters (docs, src, threshold, db) in tool schemas
  - Test: Inspect `atd serve` MCP tool registration
  - Success: Internal paths not exposed to agents

- **T4.1.2:** Verify agent-centric descriptions
  - Expected: Tool descriptions optimized for agent LLMs
  - Test: Review tool descriptions for clarity and agent-friendliness
  - Success: Descriptions are clear, concise, and agent-focused

- **T4.1.3:** Verify automatic configuration
  - Expected: Tools use .atd config without manual path specification
  - Test: Call `atd_search` without db parameter
  - Success: Uses correct database path from config

**Validation Commands:**
```bash
cd /home/bastien/work/skill/upsilon-hub
atd serve --http --port 7474 &
# Call MCP tools and inspect schemas
```

---

### Category 5: Feature Completeness (P2 Priority)

#### Test 5.1: Missing Feature Detection  
**Issues:** ISS-040, ISS-049, ISS-052 (Still Open)  
**Status:** Not Implemented, Document Gaps Expected

**Test Cases:**
- **T5.1.1:** Verify mermaid export absence
  - Expected: No `--format mermaid` option in `atd crawl`
  - Test: Run `atd crawl --help` and check for format options
  - Success: Confirms mermaid export is missing (expected per open issue)

- **T5.1.2:** Verify audit-trace integration absence
  - Expected: Audit doesn't integrate trace for coverage
  - Test: Run `atd audit` and check for coverage recommendations
  - Success: Confirms audit-trace integration is missing (expected per open issue)

- **T5.1.3:** Verify stats coverage reporting absence
  - Expected: No detailed coverage/ancestry metrics in `atd stats`
  - Test: Run `atd stats` and examine output
  - Success: Confirms enhanced coverage reporting is missing (expected per open issue)

**Validation Commands:**
```bash
cd /home/bastien/work/skill/upsilon-hub
atd crawl --help | grep -i mermaid
atd audit | grep -i "trace\|coverage"
atd stats | grep -i "ancestry\|coverage.*implementation"
```

---

## Test Execution Plan

### Phase 1: Environment Setup (Day 1)
1. **Configure ATD for upsilon-hub testing**
   - Create test-specific .atd configuration
   - Verify all paths and extensions configured correctly
   - Set up test databases and caching

2. **Baseline measurements**
   - Run all tools and record baseline performance
   - Document current behavior for comparison
   - Create test result templates

### Phase 2: Core Tooling Validation (Days 2-3)
1. **Execute Category 1 tests** (Core Tooling Fixes)
   - T1.1: Indexing system validation
   - T1.2: Link resolution validation
   - T1.3: Orphan detection validation

2. **Record detailed results**
   - Pass/fail status for each test case
   - Performance measurements
   - Error logs and debugging information

### Phase 3: Language-Agnostic Validation (Days 4-5)
1. **Execute Category 2 tests** (Language-Agnostic Features)
   - T2.1: Language-agnostic verification
   - T2.2: Git integration and discovery

2. **Multi-language testing**
   - Test Go, PHP, JavaScript/Vue separately
   - Verify mixed-language project handling
   - Document any language-specific issues

### Phase 4: Performance Benchmarking (Days 6-7)
1. **Execute Category 3 tests** (Performance and Scalability)
   - T3.1: Indexing performance
   - T3.2: Audit performance

2. **Performance analysis**
   - Compare against expected performance targets
   - Identify bottlenecks and optimization opportunities
   - Document scaling characteristics

### Phase 5: Integration Testing (Day 8)
1. **Execute Category 4 tests** (MCP Tools Integration)
   - T4.1: MCP tools parameter cleanup

2. **Integration validation**
   - Test MCP server with real agent interactions
   - Verify tool schemas and parameter handling
   - Document any agent workflow issues

### Phase 6: Feature Gap Analysis (Day 9)
1. **Execute Category 5 tests** (Feature Completeness)
   - T5.1: Missing feature detection

2. **Gap documentation**
   - Document confirmed missing features
   - Prioritize based on investigation findings
   - Update issue statuses as appropriate

---

## Success Criteria

### Tooling Accuracy (Category 1)
- **Indexing:** ≥95% file coverage, proper multi-language support
- **Link Resolution:** 100% of known @spec-link tags detected
- **Orphan Detection:** ≤10 reported orphans (vs 214 false positives)

### Language Support (Category 2)  
- **Go:** Full verification support with `go test`
- **PHP:** Full verification support with `php artisan test`
- **JavaScript/Vue:** Proper file discovery and link detection
- **Git:** Proper diffing and .gitignore respect

### Performance (Category 3)
- **Indexing:** Full reindex <5 minutes, cached reindex <30% of full
- **Audit:** Full audit <2 minutes, cached audit <50% of full
- **Scalability:** Linear performance scaling with project size

### Integration (Category 4)
- **MCP Tools:** No leaky parameters, agent-centric descriptions
- **Configuration:** Automatic path resolution from .atd config
- **Agent Workflows:** Smooth tool interaction with minimal errors

---

## Test Infrastructure

### Test Result Recording
Create test result template:
```markdown
# Test Results: [Test Category]
**Date:** [Date]  
**Tester:** [Name]  
**Environment:** upsilon-hub test bed  

## Test Case Results
| Test ID | Description | Expected | Actual | Status | Notes |
|---|---|---|---|---|
| T1.1.1 | Index 250+ files | Scanned X files | Pass/Fail | Details |

## Performance Metrics
- Indexing Time: X seconds
- Audit Time: X seconds  
- Memory Usage: X MB
- CPU Utilization: X%

## Issues Found
- [Description of any issues discovered]
- [Severity assessment]
- [Recommendation for fixes]
```

### Automated Testing Script
Create automated test runner:
```bash
#!/bin/bash
# atd_test_runner.sh

echo "=== ATD Testing Suite ==="
echo "Test Environment: upsilon-hub"
echo "Date: $(date)"
echo ""

# Run all test categories
./test_category1_core_tooling.sh
./test_category2_language_agnostic.sh  
./test_category3_performance.sh
./test_category4_integration.sh
./test_category5_completeness.sh

echo "=== Test Suite Complete ==="
echo "Results saved to test_results.md"
```

---

## Issue Status Updates

### Expected Status Changes Based on Testing

**Implementation Complete, Needs Testing:**
- **ISS-044** (verify.go): Expected to resolve as "Implemented & Validated"
- **ISS-059/060** (exploration): Expected to resolve as "Implemented & Validated"  
- **Performance improvements:** Expected to confirm "COMPLETED" status

**Still Open Issues:**
- **ISS-040** (mermaid export): Remains open, implementation needed
- **ISS-049** (audit-trace integration): Remains open, implementation needed
- **ISS-052** (stats coverage): Remains open, implementation needed
- **ISS-045** (mcp tools): Status update based on test results

### Investigation Validation
**Confirmed Findings:**
- **False orphan reporting:** Should confirm ~96% false positive rate
- **Coverage accuracy:** Should confirm ~82% actual implementation coverage
- **Multi-language support:** Should confirm proper Go/PHP/JS/Vue handling

**Unexpected Findings:**
- Any new issues discovered during testing
- Performance characteristics not anticipated
- Integration challenges not identified in issues

---

## Next Steps After Testing

### Immediate Actions (Post-Testing)
1. **Update issue statuses** based on test results
2. **Document test findings** in comprehensive report
3. **Prioritize remaining issues** based on testing outcomes

### Medium Term (Based on Test Results)
1. **Implement missing features** (mermaid export, audit-trace integration, stats coverage)
2. **Performance optimizations** based on benchmarking results
3. **Documentation updates** reflecting validated capabilities

### Long Term (Production Readiness)
1. **CI/CD integration** for automated ATD testing
2. **Monitoring and alerting** for ATD system health
3. **User feedback integration** for continuous improvement

---

## Risk Mitigation

### Testing Risks
- **Time Constraints:** 9-day timeline may be aggressive for comprehensive testing
  - *Mitigation:* Prioritize P0/P1 tests first, defer P2/P3 if needed
- **Environment Differences:** upsilon-hub may not represent all use cases
  - *Mitigation:* Document limitations, plan additional test beds if needed
- **Test Result Validity:** Manual testing may miss edge cases
  - *Mitigation:* Include automated tests where possible, peer review results

### Rollback Planning
- Keep baseline measurements for comparison
- Maintain ability to revert test configuration changes
- Document all test modifications for reproducibility

---

## References

- [LOT 4 Core Issues](issues/LOT_SUMMARY_4_Core.md)
- [Individual Issues](issues/)
- [Investigation Findings](upsilon-hub/atd_investigation/)
- [ATD Upgrade Plan](ATD_UPGRADE_PLAN.md)
- [Current ATD Documentation](ATD.md)