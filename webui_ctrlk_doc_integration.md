# Ctrl+K and Document Generation - Workspace Integration Plan

**Related Issue:** ISS-091 - ATD Workspace & Multi-Project Support
**Related:** `webui_workspace_integration.md`
**Date:** 2026-04-23
**Status:** Planning

---

## Overview

This document outlines the integration of workspace/multi-project support into two key WebUI features:

1. **Ctrl+K Search** - Should work across workspace boundaries with visual hints
2. **Document Generation** - Needs a toggle to allow/disallow project crossing

Both features require updates to:
- Backend search/assemble tools
- MCP server tools
- Frontend UI

---

## Current State Analysis

### Ctrl+K Search (`search.js`)

**Current Behavior:**
- Opens overlay with search input
- Calls `/api/search?q=query`
- Returns atoms from **current project only**
- Displays results with layer badges
- Clicking result opens atom details

**Limitations in Workspace:**
- No indication which project each atom belongs to
- Cannot search across all projects
- No visual distinction between local and cross-project results

---

### Document Generation (`documents.js`)

**Current Behavior:**
- Opens modal with prompt input and atom selection
- Searches for atoms via `searchAtoms(query)` (project-scoped)
- User selects atoms for document assembly
- Calls `/api/generate-document` with selected atom IDs
- Backend uses `exploration.Assemble()` with single `DocsDir`

**Limitations in Workspace:**
- Can only select atoms from active project
- No option to include atoms from other projects
- No visual indication of cross-project dependencies
- Assembly is limited to single `DocsDir`

---

## Implementation Plan

### Phase 1: Backend Tool Updates

#### 1.1 Update Search with Workspace Scope

**File:** `atd/pkg/exploration/search.go`

**Current `SearchOptions`:**
```go
type SearchOptions struct {
    Query   string
    Grep    string
    DBPath  string
    Limit   int
    Scope   string  // "all", "code", "docs"
    Root    string
}
```

**New `SearchOptions`:**
```go
type SearchOptions struct {
    Query      string
    Grep       string
    DBPath     string
    Limit      int
    Scope      string  // "all", "code", "docs"
    Root       string
    Workspace  bool    // NEW: Search across workspace projects
    Projects   []string // NEW: Specific projects to search (empty = all)
}
```

**Changes to `Search()`:**
```go
func Search(opts SearchOptions) ([]SearchResult, error) {
    // If workspace mode is enabled and we're in a workspace
    if opts.Workspace && config.ActiveConfig.Workspace != nil {
        return WorkspaceSearch(opts)
    }

    // Existing logic for single-project mode
    // ...
}

func WorkspaceSearch(opts SearchOptions) ([]SearchResult, error) {
    var allResults []SearchResult
    workspace := config.ActiveConfig.Workspace

    // Determine which projects to search
    var projectsToSearch []ProjectConfig
    if len(opts.Projects) > 0 {
        // Filter to specified projects
        for _, p := range workspace.Projects {
            for _, name := range opts.Projects {
                if p.Name == name {
                    projectsToSearch = append(projectsToSearch, p)
                    break
                }
            }
        }
    } else {
        // Search all projects
        projectsToSearch = workspace.Projects
    }

    // Search each project's docs directory
    for _, project := range projectsToSearch {
        projDocsPath := filepath.Join(workspace.LoadedFrom, project.Path, "docs")
        projDBPath := filepath.Join(projDocsPath, ".atd_index.db")

        // Search in this project
        results, err := SemanticSearchInProject(opts.Query, projDBPath, opts.Limit, opts.Scope, project.Name)
        if err != nil {
            // Fallback to grep if index missing
            results, err = GrepSearchInProject(opts.Query, projDocsPath, project.Name)
        }
        allResults = append(allResults, results...)
    }

    // Sort by similarity and limit
    sort.Slice(allResults, func(i, j int) bool {
        return allResults[i].Similarity > allResults[j].Similarity
    })

    if len(allResults) > opts.Limit {
        allResults = allResults[:opts.Limit]
    }

    return allResults, nil
}

// New SearchResult with project info
type SearchResult struct {
    FilePath   string  `json:"file_path"`
    ChunkText  string  `json:"chunk_text"`
    Similarity float64 `json:"similarity"`
    Project    string  `json:"project"` // NEW: Which project this belongs to
}
```

**New helper functions:**
```go
func SemanticSearchInProject(query, dbPath string, limit int, scope, projectName string) ([]SearchResult, error) {
    // Similar to existing SemanticSearch, but tags results with project name
    // ...
}

func GrepSearchInProject(keyword, docsPath, projectName string) ([]SearchResult, error) {
    // Similar to existing GrepSearch, but tags results with project name
    // ...
}
```

---

#### 1.2 Update Assemble with Workspace Support

**File:** `atd/pkg/exploration/assemble.go`

**Current `AssembleOptions`:**
```go
type AssembleOptions struct {
    Starts         string
    Intent         string
    Length         string
    Structured     bool
    AsJSON         bool
    OnlyParents    bool
    OnlyDependents bool
    DocsDir        string
}
```

**New `AssembleOptions`:**
```go
type AssembleOptions struct {
    Starts         string
    Intent         string
    Length         string
    Structured     bool
    AsJSON         bool
    OnlyParents    bool
    OnlyDependents bool
    DocsDir        string
    Workspace      bool    // NEW: Allow cross-project assembly
}
```

**Changes to `Assemble()`:**
```go
func Assemble(opts AssembleOptions) (string, error) {
    if opts.Starts == "" {
        return "", fmt.Errorf("starts parameter is required")
    }

    startIDs := strings.Split(opts.Starts, ",")
    for i := range startIDs {
        startIDs[i] = strings.TrimSpace(startIDs[i])
    }

    graph := &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}

    // Determine how to load atoms
    if opts.Workspace && config.ActiveConfig.Workspace != nil {
        // Load atoms from all projects in workspace
        if err := CrawlWorkspaceDocs(graph); err != nil {
            return "", err
        }
    } else {
        // Single-project mode (existing behavior)
        if err := CrawlDocs(opts.DocsDir, graph); err != nil {
            return "", err
        }
    }

    // Rest of existing logic for building relationships and assembly
    // ...
}

// NEW: Crawl all project docs in workspace
func CrawlWorkspaceDocs(graph *DependencyGraph) error {
    workspace := config.ActiveConfig.Workspace

    for _, project := range workspace.Projects {
        projDocsPath := filepath.Join(workspace.LoadedFrom, project.Path, "docs")

        // Tag atoms with project info
        if err := CrawlDocsWithTag(projDocsPath, graph, project.Name); err != nil {
            // Log warning but continue with other projects
            fmt.Printf("Warning: failed to crawl %s: %v\n", project.Name, err)
        }
    }

    return nil
}

// NEW: Crawl docs and tag atoms with project name
func CrawlDocsWithTag(docsDir string, graph *DependencyGraph, projectName string) error {
    return filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return nil
        }

        if info.IsDir() {
            return nil
        }

        if strings.HasSuffix(path, ".atom.md") {
            data, err := atom.Load(path)
            if err != nil {
                return nil
            }

            // Tag with project name
            if data.Metadata == nil {
                data.Metadata = make(map[string]string)
            }
            data.Metadata["project"] = projectName

            graph.Atoms[data.ID] = data
        }
        return nil
    })
}
```

---

#### 1.3 Update Atom Data Structure

**File:** `atd/pkg/atom/atom.go`

**Update `AtomData` to include project metadata:**
```go
type AtomData struct {
    ID             string            `json:"id"`
    HumanName      string            `json:"human_name"`
    Type           string            `json:"type"`
    Layer          string            `json:"layer"`
    Status         string            `json:"status"`
    Priority       string            `json:"priority"`
    Intent         string            `json:"intent"`
    Logic          string            `json:"logic"`
    Interface      string            `json:"interface"`
    Expectation    string            `json:"expectation"`
    Parents        []string          `json:"parents"`
    Dependents     []string          `json:"dependents"`
    FilePath       string            `json:"file_path"`
    HasTests       bool              `json:"has_tests"`
    Implementations []Implementation  `json:"implementations"`
    Metadata       map[string]string `json:"metadata"` // NEW: For project tagging
}

// NEW: Get project from metadata
func (a *AtomData) GetProject() string {
    if a.Metadata != nil {
        return a.Metadata["project"]
    }
    return ""
}
```

---

### Phase 2: MCP Server Updates

#### 2.1 Update MCP Tool Definitions

**File:** `atd/cmd/atd/cmd/mcp.go` (or wherever tools are registered)

**Update `atd_search` tool:**
```go
{
    Name:        "atd_search",
    Description: "Search ATD atoms by semantic similarity",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "query": map[string]any{
                "type":        "string",
                "description": "Search query",
            },
            "scope": map[string]any{
                "type":        "string",
                "description": "Search scope: 'all', 'docs', or 'code'",
                "enum":        []string{"all", "docs", "code"},
                "default":     "all",
            },
            "workspace": map[string]any{
                "type":        "boolean",
                "description": "Search across all projects in workspace (requires workspace mode)",
                "default":     false,
            },
            "projects": map[string]any{
                "type":        "array",
                "items":       map[string]any{"type": "string"},
                "description": "Specific projects to search (if workspace=true, empty means all)",
            },
        },
        "required": []string{"query"},
    },
}
```

**Update `atd_assemble` (or equivalent) tool:**
```go
{
    Name:        "atd_assemble",
    Description: "Assemble documentation from ATD atoms",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "starts": map[string]any{
                "type":        "string",
                "description": "Comma-separated atom IDs to start from",
            },
            "intent": map[string]any{
                "type":        "string",
                "description": "Intent/narrative for the assembly",
            },
            "workspace": map[string]any{
                "type":        "boolean",
                "description": "Allow cross-project atom assembly",
                "default":     false,
            },
        },
        "required": []string{"starts", "intent"},
    },
}
```

---

#### 2.2 Update MCP Tool Handlers

**Update `atd_search` handler:**
```go
func handleSearch(args map[string]any) (string, error) {
    query, _ := args["query"].(string)
    scope, _ := args["scope"].(string)
    if scope == "" {
        scope = "all"
    }

    workspace, _ := args["workspace"].(bool)
    var projects []string
    if projectsRaw, ok := args["projects"].([]any); ok {
        for _, p := range projectsRaw {
            if name, ok := p.(string); ok {
                projects = append(projects, name)
            }
        }
    }

    opts := exploration.SearchOptions{
        Query:     query,
        Scope:     scope,
        DBPath:    filepath.Join(config.DocsDir(), ".atd_index.db"),
        Limit:     10,
        Root:      config.ProjectRoot(),
        Workspace: workspace,
        Projects:  projects,
    }

    results, err := exploration.Search(opts)
    if err != nil {
        return "", err
    }

    b, _ := json.MarshalIndent(results, "", "  ")
    return string(b), nil
}
```

**Update `atd_assemble` handler:**
```go
func handleAssemble(args map[string]any) (string, error) {
    starts, _ := args["starts"].(string)
    intent, _ := args["intent"].(string)
    workspace, _ := args["workspace"].(bool)

    opts := exploration.AssembleOptions{
        Starts:     starts,
        Intent:     intent,
        Structured: true,
        AsJSON:     true,
        DocsDir:    config.DocsDir(),
        Workspace:  workspace,
    }

    result, err := exploration.Assemble(opts)
    if err != nil {
        return "", err
    }

    return result, nil
}
```

---

### Phase 3: WebUI Backend API Updates

#### 3.1 Update `/api/search` Endpoint

**File:** `atd/pkg/webui/handlers.go`

**Current:**
```go
func (s *Server) handleSearch(c *gin.Context) {
    query := c.Query("q")
    grep := c.Query("grep")

    opts := exploration.SearchOptions{
        Query:  query,
        Grep:   grep,
        DBPath: filepath.Join(config.DocsDir(), ".atd_index.db"),
        Limit:  10,
        Scope:  "all",
        Root:   config.ProjectRoot(),
    }
    // ...
}
```

**New:**
```go
func (s *Server) handleSearch(c *gin.Context) {
    query := c.Query("q")
    grep := c.Query("grep")
    workspace := c.Query("workspace") == "true"

    // Parse project filter
    var projects []string
    if projectsParam := c.Query("projects"); projectsParam != "" {
        projects = strings.Split(projectsParam, ",")
    }

    opts := exploration.SearchOptions{
        Query:     query,
        Grep:      grep,
        DBPath:    filepath.Join(config.DocsDir(), ".atd_index.db"),
        Limit:     10,
        Scope:     "all",
        Root:      config.ProjectRoot(),
        Workspace: workspace,
        Projects:  projects,
    }

    results, err := exploration.Search(opts)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed", "details": err.Error()})
        return
    }

    // Convert to atom results with project info
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    type AtomSearchResult struct {
        ID         string  `json:"id"`
        HumanName  string  `json:"human_name"`
        Type       string  `json:"type"`
        Layer      string  `json:"layer"`
        Status     string  `json:"status"`
        Intent     string  `json:"intent"`
        Similarity float64 `json:"similarity"`
        Project    string  `json:"project"` // NEW
    }

    var atomResults []AtomSearchResult
    seen := make(map[string]bool)

    for _, result := range results {
        // Extract atom ID from file path
        var atomID string
        if strings.HasSuffix(result.FilePath, ".atom.md") {
            parts := strings.Split(result.FilePath, "/")
            filename := parts[len(parts)-1]
            atomID = strings.TrimSuffix(filename, ".atom.md")
        }

        if atomID == "" || seen[atomID] {
            continue
        }

        // Look up full atom data
        if atom, exists := s.explorer.Graph.Atoms[atomID]; exists {
            seen[atomID] = true
            atomResults = append(atomResults, AtomSearchResult{
                ID:         atom.ID,
                HumanName:  atom.HumanName,
                Type:       atom.Type,
                Layer:      atom.Layer,
                Status:     atom.Status,
                Intent:     atom.Intent,
                Similarity: result.Similarity,
                Project:    result.Project, // NEW: From search result
            })
        }
    }

    c.JSON(http.StatusOK, atomResults)
}
```

---

#### 3.2 Update `/api/generate-document` Endpoint

**File:** `atd/pkg/webui/handlers.go`

**Current:**
```go
func (s *Server) handleGenerateDocument(c *gin.Context) {
    var req struct {
        Intent string   `json:"intent"`
        Starts []string `json:"starts"`
        Length string   `json:"length,omitempty"`
    }
    // ...

    opts := exploration.AssembleOptions{
        Starts:     strings.Join(req.Starts, ","),
        Intent:     req.Intent,
        Length:     req.Length,
        Structured: true,
        AsJSON:     true,
        DocsDir:    config.DocsDir(),
    }
    // ...
}
```

**New:**
```go
func (s *Server) handleGenerateDocument(c *gin.Context) {
    var req struct {
        Intent    string   `json:"intent"`
        Starts    []string `json:"starts"`
        Length    string   `json:"length,omitempty"`
        Workspace bool     `json:"workspace"` // NEW
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
        return
    }

    opts := exploration.AssembleOptions{
        Starts:     strings.Join(req.Starts, ","),
        Intent:     req.Intent,
        Length:     req.Length,
        Structured: true,
        AsJSON:     true,
        DocsDir:    config.DocsDir(),
        Workspace:  req.Workspace, // NEW
    }

    resStr, err := exploration.Assemble(opts)
    // ... rest of existing logic
}
```

---

### Phase 4: Frontend UI Updates

#### 4.1 Update Ctrl+K Search with Workspace Support

**File:** `atd/pkg/webui/static/js/search.js`

**New imports:**
```javascript
import { state, setCurrentAtom, emit, getWorkspace, isInWorkspace } from './state.js';
import { searchAtoms } from './api.js';
```

**Update `openSearchOverlay()`:**
```javascript
function openSearchOverlay() {
    if (overlay) return;

    const inWorkspace = isInWorkspace();
    const workspace = getWorkspace();

    overlay = document.createElement('div');
    overlay.className = 'search-overlay';
    // ...

    // NEW: Workspace scope toggle
    if (inWorkspace) {
        const scopeContainer = document.createElement('div');
        scopeContainer.className = 'search-scope-container';

        const scopeLabel = document.createElement('span');
        scopeLabel.className = 'search-scope-label';
        scopeLabel.textContent = 'Search scope:';

        const scopeToggle = document.createElement('div');
        scopeToggle.className = 'search-scope-toggle';

        const projectRadio = document.createElement('label');
        projectRadio.innerHTML = `
            <input type="radio" name="search-scope" value="project" checked>
            <span>${workspace?.active_project || 'Current Project'}</span>
        `;

        const workspaceRadio = document.createElement('label');
        workspaceRadio.innerHTML = `
            <input type="radio" name="search-scope" value="workspace">
            <span>All Projects</span>
        `;

        scopeToggle.appendChild(projectRadio);
        scopeToggle.appendChild(workspaceRadio);
        scopeContainer.appendChild(scopeLabel);
        scopeContainer.appendChild(scopeToggle);
        panel.appendChild(scopeContainer);
    }

    // ... rest of existing code
    panel.appendChild(input);
    panel.appendChild(results);
    // ...
}
```

**Update `performSearch()`:**
```javascript
async function performSearch(query, resultsContainer, hint) {
    try {
        // Check workspace scope selection
        const inWorkspace = isInWorkspace();
        const scopeRadio = document.querySelector('input[name="search-scope"]:checked');
        const useWorkspace = inWorkspace && scopeRadio?.value === 'workspace';

        const atoms = await searchAtoms(query, useWorkspace);
        resultsContainer.innerHTML = '';

        if (!atoms || atoms.length === 0) {
            hint.textContent = 'No results found';
            return;
        }

        hint.textContent = `${atoms.length} result${atoms.length > 1 ? 's' : ''}`;

        atoms.slice(0, 20).forEach(atom => {
            const item = document.createElement('div');
            item.className = 'search-result-item';

            const layerColors = {
                CUSTOMER: 'var(--color-customer)',
                ARCHITECTURE: 'var(--color-architecture)',
                IMPLEMENTATION: 'var(--color-implementation)',
            };

            // NEW: Add project badge if in workspace mode
            const projectBadge = atom.project && atom.project !== state.activeProject
                ? `<span class="search-result-project" title="From project: ${atom.project}">${atom.project}</span>`
                : '';

            item.innerHTML = `
                <div class="search-result-left">
                    <span class="search-result-type" style="color: ${layerColors[atom.layer] || 'var(--text-muted)'}">${atom.type}</span>
                    <span class="search-result-name">${escapeHtml(atom.human_name || atom.id)}</span>
                    ${projectBadge}
                </div>
                <div class="search-result-right">
                    <span class="search-result-layer">${atom.layer}</span>
                    <span class="search-result-status ${atom.status}">${atom.status}</span>
                </div>
            `;

            if (atom.intent) {
                const intentEl = document.createElement('div');
                intentEl.className = 'search-result-intent';
                intentEl.textContent = atom.intent;
                item.appendChild(intentEl);
            }

            item.addEventListener('click', () => {
                const fullAtom = state.atoms.find(a => a.id === atom.id);
                if (fullAtom) {
                    setCurrentAtom(fullAtom);
                    emit('navigate-to-atom', fullAtom);
                    window.history.pushState({}, '', `?atom=${atom.id}`);
                }
                closeSearchOverlay();
            });

            resultsContainer.appendChild(item);
        });
    } catch (err) {
        hint.textContent = 'Search error: ' + err.message;
    }
}
```

**Update `api.js`:**
```javascript
// @spec-link [[ui_webui_search_overlay]]
export async function searchAtoms(query, workspaceScope = false) {
    let url = `/api/search?q=${encodeURIComponent(query)}`;
    if (workspaceScope) {
        url += '&workspace=true';
    }
    const resp = await fetch(url);
    if (!resp.ok) throw new Error('Search failed');
    return resp.json();
}
```

---

#### 4.2 Update Document Generation with Workspace Toggle

**File:** `atd/pkg/webui/static/js/documents.js`

**Update `initDocuments()`:**
```javascript
export function initDocuments() {
    // ... existing initialization ...

    // NEW: Workspace toggle
    const workspaceToggleContainer = document.getElementById('doc-workspace-toggle');
    if (workspaceToggleContainer) {
        const toggle = document.createElement('label');
        toggle.className = 'workspace-doc-toggle';
        toggle.innerHTML = `
            <input type="checkbox" id="doc-workspace-cross">
            <span class="toggle-slider"></span>
            <span class="toggle-text">Allow cross-project atoms</span>
        `;
        workspaceToggleContainer.appendChild(toggle);
    }
}
```

**Update `handleGenerate()`:**
```javascript
async function handleGenerate() {
    const prompt = promptInput.value.trim();
    if (!prompt) {
        alert('Please provide a prompt or narrative request.');
        return;
    }

    const starts = Array.from(selectedAtoms);
    if (starts.length === 0) {
        alert('Please select at least one ATD to begin assembly.');
        return;
    }

    // NEW: Check workspace toggle
    const workspaceCheckbox = document.getElementById('doc-workspace-cross');
    const allowWorkspace = workspaceCheckbox?.checked || false;

    generateBtn.disabled = true;
    loadingSpinner.style.display = 'flex';

    try {
        let length = null;
        const lengthValue = lengthSelect.value.trim();
        if (lengthValue) {
            length = lengthValue;
        }

        const doc = await generateDocument(prompt, starts, length, allowWorkspace);
        setupModal.style.display = 'none';
        showDocumentViewer(doc);
    } catch(err) {
        alert('Generation failed: ' + err.message);
    } finally {
        generateBtn.disabled = false;
        loadingSpinner.style.display = 'none';
    }
}
```

**Update `handleSearch()` for document context:**
```javascript
async function handleSearch(query) {
    if (query.length < 2) {
        foundList.innerHTML = '<li style="padding: 10px; text-align: center;">Type at least 2 characters</li>';
        return;
    }

    // NEW: Check workspace scope
    const workspaceCheckbox = document.getElementById('doc-workspace-cross');
    const useWorkspace = workspaceCheckbox?.checked || false;

    foundList.innerHTML = '<li style="padding: 10px; text-align: center;">Searching...</li>';

    try {
        const atoms = await searchAtoms(query, useWorkspace);
        foundAtoms = atoms || [];
        updateFoundDisplay();
    } catch (err) {
        foundList.innerHTML = `<li style="padding: 10px; color: var(--color-red);">Search error: ${err.message}</li>`;
    }
}
```

**Update `updateFoundDisplay()` to show project badges:**
```javascript
function updateFoundDisplay() {
    foundList.innerHTML = '';

    if (!foundAtoms || foundAtoms.length === 0) {
        foundList.innerHTML = '<li style="padding: 10px; text-align: center;">No ATDs found</li>';
        document.getElementById('doc-setup-found-count').textContent = '0 found';
        return;
    }

    document.getElementById('doc-setup-found-count').textContent = `${foundAtoms.length} found`;

    const layerColors = {
        CUSTOMER: 'var(--color-customer)',
        ARCHITECTURE: 'var(--color-architecture)',
        IMPLEMENTATION: 'var(--color-implementation)',
    };

    foundAtoms.forEach(atom => {
        const li = document.createElement('li');
        li.style.padding = '8px 10px';
        li.style.borderBottom = '1px solid var(--border)';
        li.style.display = 'flex';
        li.style.alignItems = 'center';
        li.style.gap = '10px';
        li.style.cursor = 'pointer';
        li.style.transition = 'background 0.2s';

        const isSelected = selectedAtoms.has(atom.id);
        if (isSelected) {
            li.style.background = 'var(--color-selected)';
        }

        // NEW: Project badge
        const projectBadge = atom.project
            ? `<span class="atom-project-badge" style="font-size: 10px; background: var(--color-purple-light); color: var(--color-purple-dark); padding: 2px 6px; border-radius: 3px;">${atom.project}</span>`
            : '';

        const info = document.createElement('div');
        info.style.flex = '1';
        info.innerHTML = `
            <span style="color: ${layerColors[atom.layer] || 'var(--text-muted)'}; font-size: 11px; margin-right: 5px; font-weight: bold;">[${atom.layer}]</span>
            <span style="font-weight: 500;">${atom.id}</span>
            ${projectBadge}
            <div style="font-size: 11px; color: var(--text-muted); margin-top: 2px;">${atom.human_name || ''}</div>
        `;

        li.appendChild(info);
        li.addEventListener('click', () => toggleAtomSelection(atom));
        // ... rest of existing event listeners

        foundList.appendChild(li);
    });
}
```

**Update `api.js`:**
```javascript
// @spec-link [[mechanic_webui_document_generation]]
export async function generateDocument(intent, starts, length = null, workspace = false) {
    const payload = { intent, starts, workspace }; // NEW: workspace flag
    if (length !== null && length !== '') {
        payload.length = length;
    }

    const resp = await fetch('/api/generate-document', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    });
    if (!resp.ok) {
        const errorData = await resp.json().catch(() => ({error: 'Failed to generate document'}));
        throw new Error(errorData.error || 'Failed to generate document');
    }
    return resp.json();
}
```

---

#### 4.3 Update HTML for Document Generation Modal

**File:** `atd/pkg/webui/static/index.html`

**Add workspace toggle in document setup modal:**
```html
<!-- After the Document Length field -->
<div class="form-group" style="margin-top: 10px;">
    <label style="display: flex; justify-content: space-between; align-items: center;">
        <span>Workspace Scope</span>
        <span id="doc-workspace-scope-hint" style="font-size: 11px; color: var(--text-muted);">Current project only</span>
    </label>
    <div id="doc-workspace-toggle" style="margin-top: 5px;">
        <!-- Populated by JS if in workspace -->
    </div>
</div>
```

---

#### 4.4 Add CSS for Workspace UI Elements

**File:** `atd/pkg/webui/static/styles.css`

```css
/* Search scope toggle */
.search-scope-container {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border);
    font-size: 13px;
}

.search-scope-label {
    color: var(--text-muted);
}

.search-scope-toggle {
    display: flex;
    gap: 15px;
}

.search-scope-toggle label {
    display: flex;
    align-items: center;
    gap: 5px;
    cursor: pointer;
}

.search-scope-toggle input {
    cursor: pointer;
}

/* Project badge in search results */
.search-result-project {
    font-size: 10px;
    background: var(--color-purple-light);
    color: var(--color-purple-dark);
    padding: 2px 6px;
    border-radius: 3px;
    margin-left: 8px;
    font-weight: 600;
}

/* Workspace document toggle */
.workspace-doc-toggle {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
    user-select: none;
}

.workspace-doc-toggle input[type="checkbox"] {
    appearance: none;
    width: 40px;
    height: 22px;
    background: var(--border);
    border-radius: 11px;
    position: relative;
    cursor: pointer;
    transition: background 0.2s;
}

.workspace-doc-toggle input[type="checkbox"]:checked {
    background: var(--color-purple);
}

.workspace-doc-toggle input[type="checkbox"]::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    width: 18px;
    height: 18px;
    background: white;
    border-radius: 50%;
    transition: transform 0.2s;
}

.workspace-doc-toggle input[type="checkbox"]:checked::after {
    transform: translateX(18px);
}

.workspace-doc-toggle .toggle-text {
    font-size: 13px;
    color: var(--text-primary);
}

/* Atom project badge in document setup */
.atom-project-badge {
    display: inline-block;
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 3px;
    margin-left: 5px;
    font-weight: 600;
}
```

---

## Implementation Order

1. **Backend Core (Phase 1)** - Update exploration tools
   - Add `Workspace` and `Projects` to `SearchOptions`
   - Implement `WorkspaceSearch()`
   - Add `Workspace` to `AssembleOptions`
   - Implement `CrawlWorkspaceDocs()`
   - Update `AtomData` with `Metadata` field

2. **MCP Updates (Phase 2)** - Add workspace parameters to tools
   - Update `atd_search` tool schema and handler
   - Update `atd_assemble` tool schema and handler

3. **WebUI Backend (Phase 3)** - Update API handlers
   - Update `handleSearch()` to support workspace parameter
   - Update `handleGenerateDocument()` to support workspace parameter

4. **Frontend UI (Phase 4)** - Update JavaScript and HTML
   - Update `search.js` with workspace scope toggle
   - Update `documents.js` with workspace toggle
   - Update `api.js` with new parameters
   - Update `index.html` with workspace toggle UI
   - Add CSS for workspace UI elements

---

## Testing Scenarios

### Scenario 1: Ctrl+K in Non-Workspace Mode

**Setup:** Run WebUI outside workspace

**Expected:**
- No workspace scope toggle shown
- Search works as before (current project only)

---

### Scenario 2: Ctrl+K in Workspace Mode - Project Scope

**Setup:** Run WebUI in workspace, select "Current Project" scope

**Expected:**
- Workspace scope toggle shown with "Current Project" selected
- Search returns only atoms from active project
- No project badges on results

---

### Scenario 3: Ctrl+K in Workspace Mode - Workspace Scope

**Setup:** Run WebUI in workspace, select "All Projects" scope

**Expected:**
- Search returns atoms from all projects
- Cross-project results show project badge
- Clicking cross-project result opens atom (may be from inactive project)

---

### Scenario 4: Document Generation - Project Only

**Setup:** Generate document with workspace toggle OFF

**Expected:**
- Search for context only returns atoms from active project
- Assembly only includes atoms from active project's docs

---

### Scenario 5: Document Generation - Cross-Project

**Setup:** Generate document with workspace toggle ON

**Expected:**
- Search for context returns atoms from all projects
- Selected atoms show project badges
- Assembly can include atoms from multiple projects
- Generated document metadata shows cross-project sources

---

### Scenario 6: MCP Search with Workspace Flag

**Setup:** Call `atd_search` with `workspace=true`

**Expected:**
- Returns results from all projects in workspace
- Results include `project` field

---

### Scenario 7: MCP Assemble with Workspace Flag

**Setup:** Call `atd_assemble` with `workspace=true` and cross-project atom IDs

**Expected:**
- Assembles document using atoms from multiple projects
- Follows dependencies across project boundaries

---

## Open Questions

1. **Cross-Project Atom Navigation:** When clicking a cross-project search result, should we:
   - A) Just show the atom (read-only) in the UI
   - B) Auto-switch to that project
   - C) Ask the user before switching
   - **Recommendation:** Option C - show a modal asking to switch

2. **Cross-Project Dependencies in Assembly:** Should assembly follow dependencies across projects?
   - **Recommendation:** Yes, but add a `follow_cross_project` flag to `AssembleOptions` for granular control

3. **Search Result Limit:** Should workspace search limit per-project or globally?
   - **Current plan:** Global limit (easier to implement)
   - **Alternative:** Per-project limit (more balanced results)

4. **Indexing:** Does each project need its own `.atd_index.db`?
   - **Recommendation:** Yes, one per project's `docs/` directory for modularity

5. **Default Behavior:** When in workspace, what should be the default search scope?
   - **Recommendation:** Current project (safer default, user must opt-in to workspace scope)

---

## Future Enhancements

1. **Project Filter UI:** Multi-select dropdown for specific projects in search
2. **Visual Graph:** Show cross-project dependencies in dependency graph view
3. **Workspace Dashboard:** Compare atom counts, coverage, and health across projects
4. **Shared Libraries:** Special handling for atoms marked as "shared" (available to all projects)
5. **Smart Suggestion:** When selecting atoms for document, suggest related atoms from other projects

---

## Related Files

### Backend
- `atd/pkg/exploration/search.go` - Search implementation
- `atd/pkg/exploration/assemble.go` - Assembly implementation
- `atd/pkg/atom/atom.go` - Atom data structure
- `atd/config/config.go` - Workspace config

### MCP
- `atd/pkg/mcp/registry.go` - Tool registry
- `atd/cmd/atd/cmd/mcp.go` - MCP command and tool definitions

### WebUI Backend
- `atd/pkg/webui/handlers.go` - HTTP handlers

### WebUI Frontend
- `atd/pkg/webui/static/js/search.js` - Ctrl+K search UI
- `atd/pkg/webui/static/js/documents.js` - Document generation UI
- `atd/pkg/webui/static/js/api.js` - API client
- `atd/pkg/webui/static/js/state.js` - State management
- `atd/pkg/webui/static/index.html` - Main HTML
- `atd/pkg/webui/static/styles.css` - Styles
