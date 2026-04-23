# ATD Workspace & Multi-Project Support - Testing Plan

**Version:** 1.0
**Date:** 2026-04-23
**Status:** Draft
**Target:** Full coverage of workspace feature (ISS-091)

---

## Overview

This plan covers testing for the ATD workspace/multi-project support feature implemented in:
- `atd/config/config.go` - WorkspaceConfig, LoadWorkspaceConfig(), SetProject()
- `atd/config/workspace_test.go` - Existing unit tests
- `atd/cmd/atd/cmd/workspace.go` - CLI workspace commands
- `test-workspace/` - Test workspace setup

**Test Workspace Structure:**
```
test-workspace/
├── .atd.workspace
├── project-a/
│   ├── .atd
│   ├── docs/
│   │   └── atom-a.atom.md
│   └── src/
│       └── main.go
├── project-b/
│   ├── .atd
│   ├── docs/
│   │   └── atom-b.atom.md
│   └── src/
│       └── lib.go
└── project-c/
```

---

## Test Matrix

| Category | Tests | Coverage Target | Status |
|----------|--------|----------------|--------|
| Unit Tests | Config & Workspace Core | 100% | ⚠️ Partial |
| Integration Tests | CLI commands | 100% | ❌ Missing |
| E2E Tests | Full workflows | 80% | ❌ Missing |
| MCP Tests | Workspace tools | 0% | ❌ Missing |
| Edge Cases | Error handling | 100% | ❌ Missing |

---

## 1. Unit Tests (`atd/config/workspace_test.go`)

### Existing Tests
| Test | Purpose | Status |
|------|---------|--------|
| `TestLoadWorkspaceConfig` | Load workspace from root and subproject | ✅ Implemented |
| `TestLoadInWorkspace` | Load config with workspace context | ✅ Implemented |

### Additional Unit Tests Required

#### 1.1 Workspace Config Loading

```go
func TestLoadWorkspaceConfig_Malformed(t *testing.T) {
    // Test invalid JSON
    // Test missing required fields
    // Test duplicate project names
}

func TestLoadWorkspaceConfig_PathResolution(t *testing.T) {
    // Test relative project paths
    // Test absolute project paths
    // Test workspace_root resolution
}

func TestLoadWorkspaceConfig_NoWorkspaceFound(t *testing.T) {
    // Test behavior when no .atd.workspace exists
    // Should return nil, nil (not an error)
}
```

#### 1.2 Project Detection

```go
func TestDetectActiveProject_FromRoot(t *testing.T) {
    // When CWD is workspace root
    // Should return empty active project
}

func TestDetectActiveProject_FromSubproject(t *testing.T) {
    // When CWD is inside a project directory
    // Should detect correct project
}

func TestDetectActiveProject_FromSubdir(t *testing.T) {
    // When CWD is deep within a project
    // Should still detect parent project
}

func TestDetectActiveProject_NestedPaths(t *testing.T) {
    // Test path resolution with ../ and ./
}
```

#### 1.3 SetProject

```go
func TestSetProject_Existing(t *testing.T) {
    // Set to valid project
    // Verify ActiveProject updates
    // Verify loadedFromDir updates
}

func TestSetProject_NonExisting(t *testing.T) {
    // Set to invalid project
    // Should return error
    // ActiveConfig should not change
}

func TestSetProject_WithCustomConfigPath(t *testing.T) {
    // Test project with custom config_path
    // Verify custom config is loaded
}

func TestSetProject_PreservesWorkspace(t *testing.T) {
    // SetProject should not lose workspace reference
    // WorkspaceConfig should remain intact
}
```

#### 1.4 Cross-Project References

```go
func TestParseCrossProjectReference(t *testing.T) {
    // Parse [[project-a:atom-id]]
    // Parse [[shared:atom-id]]
    // Parse [[local:atom-id]]
    // Parse [[atom-id]] (local, no prefix)
}

func TestResolveCrossProjectReference(t *testing.T) {
    // Resolve reference in same project (should work)
    // Resolve reference in other project (should work)
    // Resolve reference to non-existent project (should error)
    // Resolve reference to non-existent atom (should error)
}
```

---

## 2. Integration Tests (`atd/config/workspace_integration_test.go`)

### 2.1 Workspace Initialization

```bash
# Test: Workspace Init Creates Valid Config
cd /tmp
mkdir ws-test
cd ws-test
atd workspace init --name "test-ws"
# Verify: .atd.workspace created with correct content
# Verify: workspace_name matches
# Verify: projects array is empty

# Test: Workspace Init With No Name
cd /tmp
mkdir ws-test-2
cd ws-test-2
atd workspace init
# Verify: workspace_name matches directory name

# Test: Workspace Init Overwrites Existing
cd /tmp/ws-test
atd workspace init --name "new-name"
# Should: overwrite or ask for confirmation
```

### 2.2 Project Management

```bash
# Test: Add Project to Workspace
cd /tmp/ws-test
mkdir -p proj-a/docs
atd workspace add --name "proj-a" --path ./proj-a
# Verify: .atd.workspace updated
# Verify: proj-a in projects list

# Test: Add Project With Duplicate Name
atd workspace add --name "proj-a" --path ./proj-b
# Should: Error "project 'proj-a' already exists"

# Test: Add Project With Invalid Path
atd workspace add --name "proj-b" --path ./nonexistent
# Should: Error about path not found

# Test: List Projects
atd workspace list
# Verify: Shows workspace name
# Verify: Lists all projects
# Verify: Shows active marker for current project

# Test: Add Multiple Projects
mkdir -p proj-{b,c}/docs
atd workspace add --name "proj-b" --path ./proj-b
atd workspace add --name "proj-c" --path ./proj-c
atd workspace list
# Verify: All three projects listed
```

### 2.3 Command Context Awareness

```bash
# Test: Stats from Workspace Root
cd /tmp/ws-test
atd stats
# Expected: Error - specify --project or --workspace

# Test: Stats from Project Root
cd proj-a
atd stats
# Verify: Uses proj-a's config
# Verify: Shows only proj-a's atoms

# Test: Stats from Subdirectory
cd proj-a/docs
atd stats
# Verify: Still uses proj-a's config
# Verify: Respects project boundaries

# Test: Stats with Project Flag
cd /tmp/ws-test
atd stats --project proj-a
# Verify: Shows proj-a stats regardless of CWD
```

### 2.4 Workspace-Wide Operations

```bash
# Test: Workspace Stats
cd /tmp/ws-test
atd stats --workspace
# Verify: Aggregates stats from all projects
# Verify: Shows project breakdown
# Verify: Shows total atom count

# Test: Workspace Search
atd search --workspace "authentication"
# Verify: Searches across all projects
# Verify: Results show project name for each atom

# Test: Workspace Crawl
atd crawl --workspace
# Verify: Builds workspace-wide graph
# Verify: Cross-project links are visible
```

### 2.5 Config File Priority

```bash
# Test: Project Config Overrides Workspace Settings
# .atd.workspace has common llm settings
# project-a/.atd has different llm settings
cd project-a
# Verify: project-a's settings take precedence
```

---

## 3. End-to-End Tests

### 3.1 Complete Workspace Workflow

```bash
# Scenario: User sets up new workspace from scratch
cd /tmp
mkdir my-monorepo
cd my-monorepo

# Initialize workspace
atd workspace init --name "my-monorepo"

# Create first project
mkdir -p frontend/{docs,src}
cat > frontend/.atd <<EOF
{"docs_path":"docs/","code_paths":["src/"]}
EOF
atd workspace add --name "frontend" --path ./frontend

# Create an atom
cat > frontend/docs/component.atom.md <<EOF
---
id: button_component
type: UI
layer: IMPLEMENTATION
status: STABLE
---
# Button Component
## INTENT
Reusable button for all pages.
EOF

# Verify atom is found
cd frontend
atd query --search button_component
# Should: Return the atom

# Create second project
mkdir -p backend/{docs,src}
cat > backend/.atd <<EOF
{"docs_path":"docs/","code_paths":["src/"]}
EOF
atd workspace add --name "backend" --path ./backend

# Create cross-project dependency
cat > backend/docs/api_endpoint.atom.md <<EOF
---
id: user_api
type: API
layer: ARCHITECTURE
status: STABLE
parents:
  - [[frontend:user_form]]
---
# User API
## INTENT
Endpoint for user CRUD operations.
EOF

# Switch to frontend, add referenced atom
cd frontend
cat > docs/user_form.atom.md <<EOF
---
id: user_form
type: UI
layer: ARCHITECTURE
status: STABLE
dependents:
  - [[backend:user_api]]
---
# User Form
## INTENT
Form for creating/updating users.
EOF

# Test workspace-wide crawl
cd /tmp/my-monorepo
atd crawl --workspace
# Should: Show dependency graph spanning both projects
```

### 3.2 Migration from Single-Config

```bash
# Scenario: Existing monorepo with single .atd
# Migrate to workspace structure
```

### 3.3 Developer Workflow Simulation

```bash
# Scenario: Developer A works on frontend, Developer B works on backend
# Verify: Each developer sees only their project's changes
# Verify: atd verify doesn't fail for other project's changes
```

---

## 4. MCP Server Tests

### 4.1 Session Initialization

```go
// Test: Initialize with workspace context
func TestMCPInit_WithWorkspace(t *testing.T) {
    // Send initialize request from workspace directory
    // Verify response includes workspace_root
    // Verify response includes active_project
}

// Test: Initialize without workspace
func TestMCPInit_NoWorkspace(t *testing.T) {
    // Send initialize from non-workspace directory
    // Verify workspace fields are null
}
```

### 4.2 Workspace MCP Tools

```go
// Test: atd_workspace_list
func TestMCPWorkspaceList(t *testing.T) {
    // Call tool
    // Verify returns all projects
    // Verify includes active marker
}

// Test: atd_workspace_stats
func TestMCPWorkspaceStats(t *testing.T) {
    // Call tool
    // Verify aggregates all projects
    // Verify includes per-project breakdown
}

// Test: atd_workspace_switch
func TestMCPWorkspaceSwitch(t *testing.T) {
    // Call tool with new project
    // Verify active project changes
    // Verify subsequent calls use new context
}
```

### 4.3 Tool Context Preservation

```go
// Test: Project context persists across tool calls
func TestMCPContextPersistence(t *testing.T) {
    // Initialize in project-a
    // Call atd_query - should query project-a's atoms
    // Switch to project-b via atd_workspace_switch
    // Call atd_query again - should query project-b's atoms
}
```

---

## 5. Edge Cases & Error Handling

### 5.1 Path Resolution

| Scenario | Expected Behavior | Test |
|----------|------------------|-------|
| Workspace root with trailing `/` | Normalized | Test with `./test-workspace/` |
| Project path with `../` | Resolved relative to workspace | Test |
| Absolute paths | Used as-is | Test |
| Non-existent project path | Error during `add` | Test |
| Symlinked directories | Follow symlinks | Test |

### 5.2 Config Conflicts

| Scenario | Expected Behavior | Test |
|----------|------------------|-------|
| Multiple `.atd.workspace` in tree | Use highest (closest) | Test |
| Project config missing `.atd` | Use workspace defaults | Test |
| Malformed JSON in workspace config | Graceful error with message | Test |
| Malformed JSON in project config | Graceful error, workspace still valid | Test |

### 5.3 Cross-Project References

| Scenario | Expected Behavior | Test |
|----------|------------------|-------|
| Reference to non-existent project | Error indicating which project | Test |
| Circular cross-project references | Detection and warning | Test |
| Self-reference in same project | Allowed, no prefix needed | Test |
| Deeply nested project names | Resolved correctly | Test |

### 5.4 Concurrent Operations

| Scenario | Expected Behavior | Test |
|----------|------------------|-------|
| Two sessions with different projects | Each maintains own context | Test via MCP |
| File lock on `.atd.workspace` | Wait or fail gracefully | Test |

---

## 6. Performance Tests

### 6.1 Scalability

```bash
# Test: Large Workspace (100+ projects)
# Create workspace with 100 projects
# Measure: workspace list time
# Expected: <1 second

# Test: Large Atom Count (1000+ atoms per project)
# Measure: workspace stats time
# Expected: <5 seconds
```

### 6.2 Memory

```go
// Test: Memory usage doesn't grow with workspace switches
func TestWorkspaceSwitchMemory(t *testing.T) {
    // Load workspace
    // Switch between projects 100 times
    // Verify: No memory leaks
}
```

---

## 7. Test Execution Checklist

### Before Implementation
- [ ] Review existing `workspace_test.go` coverage
- [ ] Identify gaps in unit tests
- [ ] Set up CI for workspace tests

### During Implementation
- [ ] Run unit tests after each function
- [ ] Run integration tests after each command
- [ ] Document edge case behavior

### Before Release
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Manual E2E tests completed
- [ ] MCP tests completed
- [ ] Performance benchmarks run
- [ ] Documentation updated
- [ ] CLAUDE.md updated with workspace examples

---

## 8. Test Data

### Required Test Fixtures

```
test-fixtures/
├── workspace-basic/           # Simple 2-project workspace
├── workspace-large/            # 10+ projects
├── workspace-nested/           # Projects in subdirectories
├── workspace-malformed/        # Invalid JSON configs
├── workspace-conflict/         # Overlapping project paths
└── workspace-crosslinks/       # Projects with cross-project refs
```

---

## 9. Success Criteria

A workspace feature is considered complete when:

1. **Unit Test Coverage**: ≥80% for workspace-related code
2. **Integration Tests**: All CLI workspace commands tested
3. **E2E Tests**: At least 3 complete workflows validated
4. **MCP Tests**: All workspace MCP tools tested
5. **Edge Cases**: All identified edge cases have tests
6. **Performance**: Workspace operations complete within SLA
7. **Documentation**: All new behaviors documented

---

## 10. Open Questions for Implementation

1. **Migration Path**: What's the script to migrate existing single-config workspaces?
2. **Shared Libraries**: How should `shared_libraries` be resolved (copy vs symlink)?
3. **Git Integration**: Should workspace detection use `.git` boundaries?
4. **VS Code Extension**: Does VS Code need awareness of workspace mode?
5. **Conflict Resolution**: What if project config contradicts workspace config?

---

## Appendix: Test Commands Quick Reference

```bash
# Unit tests
cd /home/bastien/work/skill/atd/config
go test -v -run TestLoadWorkspace
go test -v -run TestSetProject

# Integration tests (to be created)
cd /home/bastien/work/skill
go test -v -run TestWorkspace

# Manual E2E tests
cd test-workspace
atd workspace list
cd project-a && atd stats
cd .. && atd stats --workspace

# MCP tests (to be created)
# Run MCP server with workspace context
curl -X POST http://localhost:7474/mcp -d '{"method":"tools/list","params":{}}'
```

---

**Next Steps**:
1. Implement missing unit tests in `workspace_test.go`
2. Create `workspace_integration_test.go`
3. Set up test fixtures in `test-fixtures/`
4. Write MCP integration tests
5. Execute manual E2E test scenarios
6. Run performance benchmarks
