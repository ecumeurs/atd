# Ctrl+K + Document Generation - Workspace Integration Test Plan

**Related:** `webui_ctrlk_doc_integration.md`
**Issue:** ISS-091
**Date:** 2026-04-23
**Status:** Ready for Testing

---

## Test Environment Setup

### Prerequisites

1. **ATD CLI** built and available in PATH
2. **At least 2 test projects** configured with ATD atoms
3. **Workspace configuration** with 2+ projects
4. **WebUI server** running (`atd webui`)

### Test Data Structure

```
test-workspace/
├── workspace.yaml
├── project-a/
│   └── docs/
│       ├── req_auth.atom.md
│       ├── api_login.atom.md
│       └── mech_session.atom.md
└── project-b/
    └── docs/
        ├── req_payment.atom.md
        ├── api_checkout.atom.md
        └── mech_transaction.atom.md
```

### Test Workspace Config

```yaml
name: test-workspace
projects:
  - name: project-a
    path: ./project-a
  - name: project-b
    path: ./project-b
```

---

## Test Suite 1: Backend - Search with Workspace Scope

### Test 1.1: Single-Project Search (Baseline)

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `Search()` with `Workspace: false` | Returns atoms from single project only |
| 2 | Verify `SearchResult.Project` field | Empty or matches current project |
| 3 | Check limit enforcement | Returns at most `Limit` results |

**Go Test:**
```go
func TestSearch_SingleProject(t *testing.T) {
    opts := exploration.SearchOptions{
        Query: "authentication",
        Limit: 10,
        Workspace: false,
    }
    results, err := exploration.Search(opts)
    assert.NoError(t, err)
    assert.NotEmpty(t, results)
    for _, r := range results {
        assert.Empty(t, r.Project) // No project tag in single-project mode
    }
}
```

### Test 1.2: Workspace Search - All Projects

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `Search()` with `Workspace: true` | Returns atoms from all projects |
| 2 | Verify `SearchResult.Project` field | Populated for each result |
| 3 | Check project coverage | Results include atoms from project-a AND project-b |
| 4 | Verify no duplicates | No atom ID appears twice |

**Go Test:**
```go
func TestSearch_WorkspaceAllProjects(t *testing.T) {
    opts := exploration.SearchOptions{
        Query: "user",
        Limit: 20,
        Workspace: true,
        Projects: []string{}, // Empty = all
    }
    results, err := exploration.Search(opts)
    assert.NoError(t, err)
    assert.NotEmpty(t, results)

    seenIDs := make(map[string]bool)
    projectNames := make(map[string]bool)

    for _, r := range results {
        assert.NotEmpty(t, r.Project)
        assert.False(t, seenIDs[r.FilePath], "Duplicate result")
        seenIDs[r.FilePath] = true
        projectNames[r.Project] = true
    }
    assert.Greater(t, len(projectNames), 1, "Should have results from multiple projects")
}
```

### Test 1.3: Workspace Search - Specific Projects

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `Search()` with `Workspace: true`, `Projects: ["project-a"]` | Returns only from project-a |
| 2 | Verify all results | All `Project == "project-a"` |
| 3 | Verify no project-b results | No results with `Project == "project-b"` |

**Go Test:**
```go
func TestSearch_WorkspaceSpecificProject(t *testing.T) {
    opts := exploration.SearchOptions{
        Query: "api",
        Limit: 10,
        Workspace: true,
        Projects: []string{"project-a"},
    }
    results, err := exploration.Search(opts)
    assert.NoError(t, err)

    for _, r := range results {
        assert.Equal(t, "project-a", r.Project)
    }
}
```

### Test 1.4: Workspace Search - Invalid Project

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `Search()` with `Projects: ["nonexistent"]` | Returns empty or gracefully handles |
| 2 | No panic/error | Graceful degradation |

### Test 1.5: Semantic Search Fallback to Grep

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Remove `.atd_index.db` from project-a | Index missing |
| 2 | Call workspace search | Falls back to grep search |
| 3 | Verify results | Returns grep results with project tags |

---

## Test Suite 2: Backend - Assemble with Workspace Support

### Test 2.1: Single-Project Assembly (Baseline)

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `Assemble()` with `Workspace: false` | Assembles from single project |
| 2 | Provide atom ID from current project | Returns valid document |
| 3 | Check atom metadata | No project metadata in atoms |

**Go Test:**
```go
func TestAssemble_SingleProject(t *testing.T) {
    opts := exploration.AssembleOptions{
        Starts: "req_auth",
        Intent: "Describe authentication flow",
        Structured: true,
        AsJSON: true,
        DocsDir: "project-a/docs",
        Workspace: false,
    }
    result, err := exploration.Assemble(opts)
    assert.NoError(t, err)
    assert.Contains(t, result, "authentication")
}
```

### Test 2.2: Workspace Assembly - Cross-Project Atoms

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `Assemble()` with `Workspace: true` | Loads all workspace atoms |
| 2 | Provide atom IDs from multiple projects | Assembles including cross-project dependencies |
| 3 | Verify output | Document includes content from all specified projects |

**Go Test:**
```go
func TestAssemble_WorkspaceCrossProject(t *testing.T) {
    opts := exploration.AssembleOptions{
        Starts: "req_auth,req_payment",
        Intent: "Describe user journey from login to payment",
        Structured: true,
        AsJSON: true,
        DocsDir: "project-a/docs",
        Workspace: true,
    }
    result, err := exploration.Assemble(opts)
    assert.NoError(t, err)
    assert.Contains(t, result, "authentication") // From project-a
    assert.Contains(t, result, "payment")        // From project-b
}
```

### Test 2.3: Atom Project Metadata

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Load atom via workspace crawl | Atom has metadata |
| 2 | Call `atom.GetProject()` | Returns correct project name |
| 3 | Verify metadata persistence | Project name preserved in graph |

**Go Test:**
```go
func TestAtomProjectMetadata(t *testing.T) {
    graph := &exploration.DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
    err := exploration.CrawlDocsWithTag("project-b/docs", graph, "project-b")
    assert.NoError(t, err)

    for _, a := range graph.Atoms {
        assert.Equal(t, "project-b", a.GetProject())
        assert.Equal(t, "project-b", a.Metadata["project"])
    }
}
```

### Test 2.4: Assembly with Cross-Project Dependencies

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create atom in project-a that depends on atom in project-b | Cross-project dependency |
| 2 | Assemble with `Workspace: true` | Includes both atoms |
| 3 | Assemble with `Workspace: false` | Shows missing dependency warning |

---

## Test Suite 3: MCP Server Tools

### Test 3.1: MCP Search - Workspace Parameter

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `atd_search` with `workspace=true` | Returns workspace-scoped results |
| 2 | Check response JSON | Each result has `project` field |
| 3 | Call `atd_search` with `workspace=false` | Returns single-project results |

**Manual Test:**
```bash
# Via Claude Code
mcp__atd__atd_search(query="user", workspace=true)

# Expected: Results from all projects with project tags
```

### Test 3.2: MCP Assemble - Workspace Parameter

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Call `atd_assemble` with `workspace=true` | Cross-project assembly enabled |
| 2 | Provide cross-project atom IDs | Assembles from multiple projects |
| 3 | Verify JSON output | Includes atoms from all projects |

**Manual Test:**
```bash
mcp__atd__atd_assemble(
  starts="req_auth,req_payment",
  intent="Full user flow",
  workspace=true
)
```

### Test 3.3: MCP Tool Schema Validation

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Check `atd_search` schema | Has `workspace` and `projects` fields |
| 2 | Check `atd_assemble` schema | Has `workspace` field |
| 3 | Verify defaults | `workspace` defaults to `false` |

---

## Test Suite 4: WebUI Backend API

### Test 4.1: GET /api/search - No Workspace Parameter

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | `GET /api/search?q=user` | Returns single-project results |
| 2 | Check response body | No `project` field or empty |

**Curl Test:**
```bash
curl "http://localhost:8080/api/search?q=auth"
# Expected: Results from active project only
```

### Test 4.2: GET /api/search - With Workspace Parameter

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | `GET /api/search?q=user&workspace=true` | Returns workspace results |
| 2 | Check response body | Each result has `project` field |

**Curl Test:**
```bash
curl "http://localhost:8080/api/search?q=user&workspace=true"
# Expected: Results from all projects with project field
```

### Test 4.3: GET /api/search - With Projects Filter

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | `GET /api/search?q=api&workspace=true&projects=project-a` | Only project-a results |
| 2 | Verify all projects | All `project == "project-a"` |

**Curl Test:**
```bash
curl "http://localhost:8080/api/search?q=api&workspace=true&projects=project-a"
```

### Test 4.4: POST /api/generate-document - No Workspace

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST with single-project atom IDs | Generates from single project |
| 2 | Verify assembly | Only includes current project atoms |

**Curl Test:**
```bash
curl -X POST http://localhost:8080/api/generate-document \
  -H "Content-Type: application/json" \
  -d '{"intent":"Auth flow","starts":["req_auth"]}'
```

### Test 4.5: POST /api/generate-document - With Workspace

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | POST with `workspace=true` | Cross-project assembly |
| 2 | Provide cross-project atom IDs | Includes all specified projects |

**Curl Test:**
```bash
curl -X POST http://localhost:8080/api/generate-document \
  -H "Content-Type: application/json" \
  -d '{"intent":"Full flow","starts":["req_auth","req_payment"],"workspace":true}'
```

---

## Test Suite 5: WebUI Frontend - Ctrl+K Search

### Test 5.1: Ctrl+K in Non-Workspace Mode

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Open WebUI outside workspace | UI loads normally |
| 2 | Press Ctrl+K | Search overlay opens |
| 3 | Verify UI | NO workspace scope toggle visible |
| 4 | Search "auth" | Results from current project only |
| 5 | Check results | No project badges shown |

### Test 5.2: Ctrl+K in Workspace Mode - Default State

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Open WebUI in workspace | UI shows workspace header |
| 2 | Press Ctrl+K | Search overlay opens |
| 3 | Verify UI | Workspace scope toggle visible |
| 4 | Check default selection | "Current Project" radio selected |
| 5 | Search "api" | Results from active project only |

### Test 5.3: Ctrl+K - Switch to Workspace Scope

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Press Ctrl+K | Search overlay opens |
| 2 | Click "All Projects" radio | Scope changes to workspace |
| 3 | Search "user" | Results from all projects |
| 4 | Verify cross-project results | Non-active project atoms show project badge |
| 5 | Check active project results | No badge (or different style) |

### Test 5.4: Ctrl+K - Click Cross-Project Result

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Search in workspace scope | Results from all projects |
| 2 | Click result from inactive project | Modal appears asking to switch projects |
| 3 | Click "Switch" in modal | Project switches, atom loads |
| 4 | Click "Cancel" in modal | Modal closes, stays on current project |

### Test 5.5: Ctrl+K - Search Persistence

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Set scope to "All Projects" | Preference stored |
| 2 | Close and reopen Ctrl+K | Scope remains "All Projects" |
| 3 | Switch projects | Scope preference resets to "Current Project" |

### Test 5.6: Ctrl+K - Keyboard Navigation

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Press Ctrl+K | Search overlay opens |
| 2 | Type query | Results appear |
| 3 | Press Arrow Down | First result highlighted |
| 4 | Press Arrow Down repeatedly | Moves through results |
| 5 | Press Enter | Opens highlighted atom |
| 6 | Press Escape | Overlay closes |

---

## Test Suite 6: WebUI Frontend - Document Generation

### Test 6.1: Document Generation - No Workspace Mode

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click "Generate Document" button | Setup modal opens |
| 2 | Verify UI | NO workspace toggle visible |
| 3 | Search for atoms | Only active project results |
| 4 | Select atoms and generate | Document from single project |

### Test 6.2: Document Generation - Workspace Mode Default

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click "Generate Document" | Setup modal opens |
| 2 | Verify UI | Workspace toggle visible, OFF by default |
| 3 | Check hint text | Shows "Current project only" |
| 4 | Search for atoms | Only active project results |

### Test 6.3: Document Generation - Enable Workspace Scope

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Click "Generate Document" | Setup modal opens |
| 2 | Toggle "Allow cross-project atoms" ON | Toggle visual state changes |
| 3 | Check hint text | Shows "All projects in workspace" |
| 4 | Search for atoms | Results from all projects |
| 5 | Verify results | Project badges shown on all results |
| 6 | Select cross-project atoms | Selection allowed |
| 7 | Click Generate | Document includes all selected atoms |

### Test 6.4: Document Generation - Cross-Project Assembly

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Enable workspace toggle | Cross-project mode active |
| 2 | Select atoms from project-a and project-b | Mixed selection |
| 3 | Enter intent and generate | Document assembles from both |
| 4 | Verify generated content | Includes content from all projects |
| 5 | Check document metadata | Shows source projects |

### Test 6.5: Document Generation - Toggle State Persistence

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Enable workspace toggle | Toggle ON |
| 2 | Close modal | Modal closes |
| 3 | Reopen document generation | Toggle remains ON |
| 4 | Switch projects | Toggle resets to OFF

### Test 6.6: Document Generation - Error Handling

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Enable workspace toggle | Cross-project mode |
| 2 | Select only atoms from inactive project | Selection allowed |
| 3 | Generate document | Document generates successfully |
| 4 | Verify content | Content from inactive project included

---

## Test Suite 7: Integration & Edge Cases

### Test 7.1: Empty Workspace

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create workspace with 0 projects | Workspace config exists |
| 2 | Try workspace search | Returns empty results, no error |
| 3 | Try workspace assembly | Returns error "no projects in workspace" |

### Test 7.2: Single Project in Workspace

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create workspace with 1 project | Workspace with single project |
| 2 | Try workspace search | Returns same results as project search |
| 3 | Verify behavior | No errors, graceful handling |

### Test 7.3: Project Without Docs Directory

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Add project to workspace without `docs/` | Missing docs |
| 2 | Try workspace search | Skips that project, searches others |
| 3 | Check logs | Warning logged about missing directory |

### Test 7.4: Large Workspace (Performance)

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create workspace with 10+ projects | Many projects |
| 2 | Run workspace search | Completes in < 2 seconds |
| 3 | Check result count | Respects limit parameter |

### Test 7.5: Atom ID Collision Across Projects

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Create same atom ID in project-a and project-b | Collision |
| 2 | Run workspace search | Both atoms returned with different project tags |
| 3 | Select and assemble | Uses both atoms correctly |

### Test 7.6: Workspace Configuration Changes

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Remove project from workspace config | Workspace updated |
| 2 | Reload WebUI | Project no longer in scope |
| 3 | Search workspace | No results from removed project |

---

## Test Suite 8: UI/UX Validation

### Test 8.1: Visual Indicators - Project Badges

| Check | Expected |
|-------|----------|
| Badge color | Matches project theme or neutral purple |
| Badge visibility | Only shown for cross-project results |
| Badge tooltip | Shows "From project: {name}" on hover |
| Badge positioning | Consistent across search and document UI |

### Test 8.2: Search Scope Toggle

| Check | Expected |
|-------|----------|
| Visibility | Only shown in workspace mode |
| Default state | "Current Project" selected |
| Switch animation | Smooth transition |
| Active state | Clear visual indication of selected scope |

### Test 8.3: Document Generation Toggle

| Check | Expected |
|-------|----------|
| Toggle design | Switch-style toggle |
| Label text | Clear "Allow cross-project atoms" |
| Hint text | Updates based on toggle state |
| Position | Below Document Length field |

### Test 8.4: Accessibility

| Check | Expected |
|-------|----------|
| Keyboard navigation | All UI elements reachable via Tab |
| Screen reader | Proper ARIA labels on toggles |
| Color contrast | Meets WCAG AA standards |
| Focus states | Clear visual focus indicators |

---

## Test Suite 9: Regression Testing

### Test 9.1: Existing Single-Project Workflow

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | Run standard ATD workflow (non-workspace) | Everything works as before |
| 2 | Ctrl+K search | Project-scoped search works |
| 3 | Document generation | Single-project assembly works |
| 4 | Atom navigation | All existing navigation works |

### Test 9.2: Existing ATD Commands

| Step | Action | Expected Result |
|------|--------|-----------------|
| 1 | `atd search query` | Works as before |
| 2 | `atd assemble --starts atom` | Works as before |
| 3 | `atd trace atom` | Works as before |
| 4 | `atd stats` | Works as before |

---

## Test Execution Checklist

### Pre-Test
- [ ] Test workspace configured with 2+ projects
- [ ] Each project has unique atoms
- [ ] WebUI server running
- [ ] Browser dev tools open for console inspection

### Backend Tests
- [ ] Test Suite 1: Search with Workspace Scope (5 tests)
- [ ] Test Suite 2: Assemble with Workspace Support (4 tests)
- [ ] Test Suite 3: MCP Server Tools (3 tests)
- [ ] Test Suite 4: WebUI Backend API (5 tests)

### Frontend Tests
- [ ] Test Suite 5: Ctrl+K Search (6 tests)
- [ ] Test Suite 6: Document Generation (6 tests)

### Integration Tests
- [ ] Test Suite 7: Integration & Edge Cases (6 tests)

### UI/UX Tests
- [ ] Test Suite 8: UI/UX Validation (4 tests)

### Regression Tests
- [ ] Test Suite 9: Regression Testing (2 tests)

---

## Acceptance Criteria

The feature is considered complete when:

1. **Workspace Search**: Ctrl+K can search across all projects with visual project indicators
2. **Scope Toggle**: Users can toggle between project and workspace scope in Ctrl+K
3. **Cross-Project Navigation**: Clicking cross-project results prompts for project switch
4. **Document Assembly**: Document generation can include atoms from multiple projects
5. **Workspace Toggle**: Document generation has a toggle for cross-project assembly
6. **MCP Integration**: MCP tools support workspace parameters
7. **Backward Compatibility**: All existing single-project workflows work unchanged
8. **No Regressions**: All existing tests pass

---

## Bug Report Template

When reporting test failures, include:

```markdown
### Bug Report: [Title]

**Test Suite:** [Test Suite Number]
**Test Case:** [Test Case Title]

**Steps to Reproduce:**
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Expected Behavior:**
[What should happen]

**Actual Behavior:**
[What actually happened]

**Environment:**
- OS: [e.g., Linux 6.12]
- Go Version: [e.g., 1.22]
- Browser: [e.g., Chrome 123]

**Logs/Errors:**
```
[Paste relevant logs or error messages]
```

**Screenshot (if applicable):**
[Attach screenshot]
```

---

## Sign-Off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| Developer | | | |
| QA Tester | | | |
| Product Owner | | | |
