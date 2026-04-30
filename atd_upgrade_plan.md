# ATD Workspace Support Upgrade Plan

**Date:** 2026-04-24
**Status:** Draft
**Priority:** HIGH
**Related Issues:** ISS-091 (ATD Workspace & Multi-Project Support)

---

## Executive Summary

The ATD workspace infrastructure is complete, but core tooling lacks automatic cross-project reference handling. This plan outlines the required upgrades to make ATD fully workspace-aware.

**Current State:**
- ✅ `.atd.workspace` configuration works
- ✅ `atd workspace list/detect` works
- ✅ `atd stats --project` works
- ✅ `atd stats --workspace` works
- ✅ `atd crawl --gaps` works workspace-wide

**Missing Features:**
- ❌ `atd weave` does not auto-prefix cross-project references
- ❌ `atd verify` does not search workspace for @spec-link
- ❌ `atd recon` does not search workspace for atom validation
- ❌ `atd test-links` does not search workspace for @test-link
- ❌ VS Code extension has no workspace awareness

**Goal:** Make cross-project reference resolution automatic and transparent across all ATD tools.

---

## Architecture Overview

### Reference Resolution Strategy

```
┌─────────────────────────────────────────────────────────────────┐
│                         Reference Resolver                        │
├─────────────────────────────────────────────────────────────────┤
│  Input: [[atom_id]] or @spec-link [[atom_id]]                    │
│                                                                  │
│  1. Check current project                                       │
│     ↓ Found? → Return local path                                │
│     ↓ Not found?                                                 │
│  2. Search workspace projects                                   │
│     ↓ Found in project X? → Return [[projectX:atom_id]]         │
│     ↓ Not found?                                                 │
│  3. Return unresolved reference (for error reporting)           │
└─────────────────────────────────────────────────────────────────┘
```

### Core Components

| Component | Current State | Target State |
|-----------|---------------|--------------|
| `workspace` package | Reads config | Full workspace API |
| `resolver` | None | New: Cross-project reference resolver |
| `weave` | Local only | Workspace-aware with auto-prefix |
| `verify/recon/test-links` | Local only | Workspace-aware |
| VS Code extension | First folder only | Full workspace support |

---

## Phase 1: Core Workspace API (Foundation)

### Goal: Build a robust workspace API that all tools can use.

### 1.1 Extend Workspace Configuration

**File:** `atd/workspace/config.go`

**Current:**
```go
type WorkspaceConfig struct {
    WorkspaceName string    `json:"workspace_name"`
    WorkspaceRoot string    `json:"workspace_root"`
    Projects      []Project `json:"projects"`
}
```

**Enhanced:**
```go
type WorkspaceConfig struct {
    WorkspaceName   string                 `json:"workspace_name"`
    WorkspaceRoot   string                 `json:"workspace_root"`
    Projects        []Project              `json:"projects"`
    SharedLibraries map[string]string      `json:"shared_libraries"` // Optional
    CommonSettings  map[string]interface{} `json:"common_settings"`  // Optional
}

type Project struct {
    Name       string `json:"name"`
    Path       string `json:"path"`
    ConfigPath string `json:"config_path,omitempty"`   // Auto-detected
    DocsPath   string `json:"docs_path,omitempty"`     // From .atd
}
```

### 1.2 Workspace Loader and Cache

**New File:** `atd/workspace/loader.go`

```go
package workspace

import (
    "sync"
    "path/filepath"
    "os"
)

var (
    workspaceCache = make(map[string]*Workspace)
    cacheMutex     sync.RWMutex
)

// LoadWorkspace loads and caches a workspace configuration
func LoadWorkspace(root string) (*Workspace, error) {
    cacheMutex.RLock()
    if ws, ok := workspaceCache[root]; ok {
        cacheMutex.RUnlock()
        return ws, nil
    }
    cacheMutex.RUnlock()

    // Load .atd.workspace
    wsPath := filepath.Join(root, ".atd.workspace")
    if _, err := os.Stat(wsPath); os.IsNotExist(err) {
        return nil, ErrNoWorkspace
    }

    ws, err := parseWorkspaceConfig(wsPath)
    if err != nil {
        return nil, err
    }

    // Auto-detect project paths and configs
    for i := range ws.Projects {
        projPath := filepath.Join(root, ws.Projects[i].Path)
        configPath := filepath.Join(projPath, ".atd")

        // Load project config to get docs_path
        if config, err := LoadProjectConfig(configPath); err == nil {
            ws.Projects[i].ConfigPath = configPath
            ws.Projects[i].DocsPath = config.DocsPath
            ws.Projects[i].FullDocsPath = filepath.Join(projPath, config.DocsPath)
        }
    }

    cacheMutex.Lock()
    workspaceCache[root] = ws
    cacheMutex.Unlock()

    return ws, nil
}

// FindProjectByCWD finds the active project based on current working directory
func (ws *Workspace) FindProjectByCWD(cwd string) *Project {
    for i := range ws.Projects {
        projPath := filepath.Join(ws.WorkspaceRoot, ws.Projects[i].Path)
        if strings.HasPrefix(cwd, projPath) {
            return &ws.Projects[i]
        }
    }
    return nil
}
```

### 1.3 Atom Registry

**New File:** `atd/workspace/registry.go`

```go
package workspace

import (
    "sync"
)

// AtomIndex maps atom IDs to their location
type AtomIndex struct {
    sync.RWMutex
    byID       map[string]*AtomLocation  // "atom_id" -> location
    byProject  map[string][]string        // "project" -> ["atom1", "atom2"]
}

type AtomLocation struct {
    Project string
    Path    string
    Exists  bool
}

// BuildIndex scans all projects and builds atom index
func (ws *Workspace) BuildIndex() (*AtomIndex, error) {
    index := &AtomIndex{
        byID:      make(map[string]*AtomLocation),
        byProject: make(map[string][]string),
    }

    for _, project := range ws.Projects {
        atoms, err := ListAtoms(project.FullDocsPath)
        if err != nil {
            return nil, err
        }

        index.byProject[project.Name] = atoms

        for _, atomID := range atoms {
            index.byID[atomID] = &AtomLocation{
                Project: project.Name,
                Path:    filepath.Join(project.FullDocsPath, atomID+".atom.md"),
                Exists:  true,
            }
        }
    }

    return index, nil
}

// FindAtom searches the workspace for an atom
func (idx *AtomIndex) FindAtom(atomID string) *AtomLocation {
    idx.RLock()
    defer idx.RUnlock()
    return idx.byID[atomID]
}

// FindAtomInProject finds an atom in a specific project
func (idx *AtomIndex) FindAtomInProject(project, atomID string) *AtomLocation {
    idx.RLock()
    defer idx.RUnlock()

    loc := idx.byID[atomID]
    if loc != nil && loc.Project == project {
        return loc
    }
    return nil
}
```

**Acceptance Criteria:**
- [ ] Workspace config loads correctly
- [ ] Project paths are auto-detected
- [ ] Atom index builds in <1 second for 300 atoms
- [ ] FindAtom() returns correct location
- [ ] Cache works across multiple calls

---

## Phase 2: Cross-Project Reference Resolver

### Goal: A unified resolver that all tools can use to find atoms across the workspace.

### 2.1 Reference Resolver

**New File:** `atd/workspace/resolver.go`

```go
package workspace

import (
    "strings"
)

// ReferenceType indicates if a reference is local or cross-project
type ReferenceType int

const (
    ReferenceLocal ReferenceType = iota
    ReferenceCrossProject
    ReferenceUnresolved
)

// ParsedReference represents a parsed [[atom_id]] or [[project:atom_id]]
type ParsedReference struct {
    Original   string
    Project    string // "" for local references
    AtomID     string
    Type       ReferenceType
    Location   *AtomLocation
}

// ParseReference parses a reference string
func ParseReference(ref string) *ParsedReference {
    // Remove brackets
    content := strings.Trim(ref, "[]")

    // Check for project prefix
    if strings.Contains(content, ":") {
        parts := strings.SplitN(content, ":", 2)
        return &ParsedReference{
            Original: ref,
            Project:  parts[0],
            AtomID:   parts[1],
            Type:     ReferenceCrossProject,
        }
    }

    return &ParsedReference{
        Original: ref,
        Project:  "",
        AtomID:   content,
        Type:     ReferenceLocal,
    }
}

// Resolver resolves references across the workspace
type Resolver struct {
    workspace *Workspace
    index     *AtomIndex
    currentProject string // Context for local references
}

// NewResolver creates a new resolver
func NewResolver(ws *Workspace, idx *AtomIndex, currentProject string) *Resolver {
    return &Resolver{
        workspace: ws,
        index:     idx,
        currentProject: currentProject,
    }
}

// Resolve resolves a reference to its canonical form with project prefix
func (r *Resolver) Resolve(ref string) (*ParsedReference, error) {
    parsed := ParseReference(ref)

    // Already has prefix
    if parsed.Type == ReferenceCrossProject {
        loc := r.index.FindAtom(parsed.AtomID)
        if loc == nil || loc.Project != parsed.Project {
            return nil, ErrAtomNotFound
        }
        parsed.Location = loc
        return parsed, nil
    }

    // Local reference - check current project first
    if loc := r.index.FindAtomInProject(r.currentProject, parsed.AtomID); loc != nil {
        parsed.Location = loc
        parsed.Type = ReferenceLocal
        return parsed, nil
    }

    // Not in current project - search workspace
    if loc := r.index.FindAtom(parsed.AtomID); loc != nil {
        // Found in another project - update to cross-project reference
        parsed.Project = loc.Project
        parsed.Type = ReferenceCrossProject
        parsed.Location = loc
        return parsed, nil
    }

    parsed.Type = ReferenceUnresolved
    return parsed, ErrAtomNotFound
}

// CanonicalForm returns the canonical [[project:atom_id]] format
func (r *Resolver) CanonicalForm(ref string) (string, error) {
    parsed, err := r.Resolve(ref)
    if err != nil {
        return ref, err
    }

    if parsed.Type == ReferenceLocal {
        return fmt.Sprintf("[[%s]]", parsed.AtomID), nil
    }
    return fmt.Sprintf("[[%s:%s]]", parsed.Project, parsed.AtomID), nil
}

// ShouldUpdate checks if a reference needs updating
func (r *Resolver) ShouldUpdate(ref string) (shouldUpdate bool, canonical string) {
    parsed := ParseReference(ref)

    // Already has prefix - check if correct
    if parsed.Type == ReferenceCrossProject {
        loc := r.index.FindAtom(parsed.AtomID)
        if loc == nil {
            return false, ref // Can't verify, leave as-is
        }
        if loc.Project == parsed.Project {
            return false, ref // Already correct
        }
        return true, fmt.Sprintf("[[%s:%s]]", loc.Project, parsed.AtomID)
    }

    // Local reference - check if actually cross-project
    parsedResolved, err := r.Resolve(ref)
    if err != nil {
        return false, ref // Not found, leave as-is for error reporting
    }

    if parsedResolved.Type == ReferenceCrossProject {
        return true, fmt.Sprintf("[[%s:%s]]", parsedResolved.Project, parsedResolved.AtomID)
    }

    return false, ref // Local reference is correct
}
```

**Acceptance Criteria:**
- [ ] ParseReference() handles all formats
- [ ] Resolve() finds atoms in same project
- [ ] Resolve() finds atoms in other projects
- [ ] CanonicalForm() returns correct format
- [ ] ShouldUpdate() correctly identifies outdated references

---

## Phase 3: Upgrade `atd weave`

### Goal: Make `atd weave` automatically fix cross-project references.

### 3.1 Enhanced Weave Command

**File:** `cmd/weave.go` (or equivalent)

```go
package cmd

import (
    "github.com/atd/cli/workspace"
)

func runWeave(cmd *cobra.Command, args []string) error {
    // Load workspace if available
    ws, err := workspace.LoadWorkspace(".")
    isWorkspace := err == nil

    var idx *workspace.AtomIndex
    var resolver *workspace.Resolver

    if isWorkspace {
        idx, err = ws.BuildIndex()
        if err != nil {
            return err
        }
        resolver = workspace.NewResolver(ws, idx, getActiveProject(ws))
    }

    // Process each atom file
    for _, atomPath := range atomFiles {
        atom, err := loadAtom(atomPath)
        if err != nil {
            return err
        }

        // Build dependency graph
        if isWorkspace {
            updated, err := weaveWithWorkspace(atom, resolver)
            if err != nil {
                return err
            }
            if updated {
                atom.Save()
            }
        } else {
            weaveLocal(atom)
        }
    }

    return nil
}

func weaveWithWorkspace(atom *Atom, resolver *workspace.Resolver) (bool, error) {
    updated := false

    // Process parents - add missing dependents
    for _, parentRef := range atom.Parents {
        parentID := extractAtomID(parentRef)
        parent, err := resolver.Resolve(parentRef)
        if err != nil {
            log.Printf("Warning: Parent %s not found for %s", parentRef, atom.ID)
            continue
        }

        // Load parent atom
        parentAtom, err := loadAtom(parent.Location.Path)
        if err != nil {
            return false, err
        }

        // Check if this atom is in parent's dependents
        selfRef := fmt.Sprintf("[[%s]]", atom.ID)
        if resolver.ShouldUpdateReference(parentAtom.Dependents, selfRef) {
            // Add with correct project prefix if needed
            canonical, _ := resolver.CanonicalForm(selfRef)
            parentAtom.Dependents = append(parentAtom.Dependents, canonical)
            parentAtom.Save()
            updated = true
        }
    }

    // Fix cross-project references in this atom
    atom.Parents = fixReferences(atom.Parents, resolver)
    atom.Dependents = fixReferences(atom.Dependents, resolver)

    return updated, nil
}

func fixReferences(refs []string, resolver *workspace.Resolver) []string {
    fixed := make([]string, 0, len(refs))

    for _, ref := range refs {
        shouldUpdate, canonical := resolver.ShouldUpdate(ref)
        if shouldUpdate {
            log.Printf("Updating reference: %s -> %s", ref, canonical)
            fixed = append(fixed, canonical)
        } else {
            fixed = append(fixed, ref)
        }
    }

    return fixed
}
```

### 3.2 Command Line Interface

```bash
# Weave with automatic cross-project reference fixing
atd weave                     # Auto-detects workspace
atd weave --project api       # Weave specific project
atd weave --workspace         # Weave all projects
atd weave --dry-run          # Show what would change
atd weave --verbose          # Show all reference updates
```

**Acceptance Criteria:**
- [ ] `atd weave` detects workspace automatically
- [ ] Cross-project references are auto-prefixed
- [ ] `--dry-run` shows changes without modifying
- [ ] `--workspace` processes all projects
- [ ] Verbose mode logs all reference updates

---

## Phase 4: Upgrade Verification Tools

### Goal: Make `atd verify`, `atd recon`, and `atd test-links` workspace-aware.

### 4.1 Enhanced `atd verify`

**File:** `cmd/verify.go`

```go
func runVerify(cmd *cobra.Command, args []string) error {
    // Load workspace
    ws, err := workspace.LoadWorkspace(".")
    isWorkspace := err == nil

    var idx *workspace.AtomIndex
    var resolver *workspace.Resolver

    if isWorkspace {
        idx, _ = ws.BuildIndex()
        resolver = workspace.NewResolver(ws, idx, getActiveProject(ws))
    }

    // Get modified files
    files := getModifiedFiles(args)

    // Verify each file
    for _, file := range files {
        links := extractSpecLinks(file)

        for _, link := range links {
            var atomID string
            var found bool

            if isWorkspace {
                parsed, err := resolver.Resolve(link)
                found = (err == nil && parsed.Location != nil)
                if found {
                    atomID = fmt.Sprintf("%s:%s", parsed.Project, parsed.AtomID)
                }
            } else {
                atomID = link
                found = atomExists(atomID)
            }

            if !found {
                reportError(file, line, "@spec-link [[%s]] not found", link)
            }
        }
    }
}
```

### 4.2 Enhanced `atd recon`

**File:** `cmd/recon.go`

```go
func runRecon(cmd *cobra.Command, args []string) error {
    atomID := args[0]

    // Load workspace
    ws, err := workspace.LoadWorkspace(".")
    isWorkspace := err == nil

    var idx *workspace.AtomIndex
    var resolver *workspace.Resolver

    if isWorkspace {
        idx, _ = ws.BuildIndex()
        resolver = workspace.NewResolver(ws, idx, getActiveProject(ws))
    }

    // Find the atom
    var atomPath string
    if isWorkspace {
        parsed, err := resolver.Resolve(fmt.Sprintf("[[%s]]", atomID))
        if err != nil {
            return fmt.Errorf("atom not found in workspace: %s", atomID)
        }
        atomPath = parsed.Location.Path
    } else {
        atomPath = findLocalAtom(atomID)
    }

    // Load and validate atom
    atom, err := loadAtom(atomPath)
    if err != nil {
        return err
    }

    // Check implementation
    return validateImplementation(atom, resolver)
}
```

### 4.3 Enhanced `atd test-links`

**File:** `cmd/test-links.go`

```go
func runTestLinks(cmd *cobra.Command, args []string) error {
    // Load workspace
    ws, err := workspace.LoadWorkspace(".")
    isWorkspace := err == nil

    var idx *workspace.AtomIndex
    var resolver *workspace.Resolver

    if isWorkspace {
        idx, _ = ws.BuildIndex()
        resolver = workspace.NewResolver(ws, idx, getActiveProject(ws))
    }

    // Scan all code files
    for _, file := range codeFiles {
        links := extractTestLinks(file)

        for _, link := range links {
            var found bool

            if isWorkspace {
                parsed, err := resolver.Resolve(link)
                found = (err == nil && parsed.Location != nil)
            } else {
                found = atomExists(link)
            }

            if !found {
                reportError(file, line, "@test-link [[%s]] not found", link)
            }
        }
    }
}
```

**Acceptance Criteria:**
- [ ] `atd verify` finds @spec-link targets in any project
- [ ] `atd recon` validates atoms across projects
- [ ] `atd test-links` finds @test-link targets in any project
- [ ] All tools accept `[[atom_id]]` or `[[project:atom_id]]` format
- [ ] Error messages include project name for cross-project references

---

## Phase 5: Upgrade VS Code Extension

### Goal: Full workspace support in VS Code extension.

### 5.1 Workspace Detection

**File:** `extension/extension.js`

```javascript
// Detect workspace configuration
async function loadWorkspaceConfig() {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
        return null;
    }

    const workspaceRoot = workspaceFolder.uri.fsPath;

    // Check for .atd.workspace
    const workspaceConfigPath = path.join(workspaceRoot, '.atd.workspace');
    if (!fs.existsSync(workspaceConfigPath)) {
        // Fall back to single-project .atd
        return loadSingleProjectConfig(workspaceRoot);
    }

    // Load workspace config
    const workspaceConfig = JSON.parse(fs.readFileSync(workspaceConfigPath, 'utf8'));

    // Load all project configs
    const projects = [];
    for (const proj of workspaceConfig.projects) {
        const projPath = path.join(workspaceRoot, proj.path);
        const atdFile = path.join(projPath, '.atd');

        if (fs.existsSync(atdFile)) {
            const config = JSON.parse(fs.readFileSync(atdFile, 'utf8'));
            projects.push({
                name: proj.name,
                path: projPath,
                docsPath: path.join(projPath, config.docs_path || 'docs'),
                config
            });
        }
    }

    return {
        isWorkspace: true,
        root: workspaceRoot,
        config: workspaceConfig,
        projects
    };
}
```

### 5.2 Active Project Switching

**File:** `extension/extension.js`

```javascript
let activeProject = null;
let workspaceConfig = null;

// Add command to switch projects
context.subscriptions.push(
    vscode.commands.registerCommand('atd.switchProject', async () => {
        if (!workspaceConfig || !workspaceConfig.isWorkspace) {
            vscode.window.showInformationMessage('Not in a workspace');
            return;
        }

        const projectNames = workspaceConfig.projects.map(p => p.name);
        const selected = await vscode.window.showQuickPick(projectNames, {
            placeHolder: 'Select active project'
        });

        if (selected) {
            activeProject = workspaceConfig.projects.find(p => p.name === selected);
            docsPath = activeProject.docsPath;
            vscode.window.showInformationMessage(`Switched to ${selected}`);

            // Update ATD Explorer
            refreshAtomExplorer();
        }
    })
);
```

### 5.3 Cross-Project Reference Resolution

**File:** `extension/extension.js`

```javascript
// Resolve atom ID across workspace
function resolveAtomID(atomID) {
    if (!workspaceConfig || !workspaceConfig.isWorkspace) {
        // Single project - check local
        return atomExists(atomID) ? atomID : null;
    }

    // Parse reference
    const parts = atomID.split(':');
    let projectName = null;
    let actualID = atomID;

    if (parts.length === 2) {
        projectName = parts[0];
        actualID = parts[1];
    }

    // If project specified, check that project
    if (projectName) {
        const project = workspaceConfig.projects.find(p => p.name === projectName);
        if (project && atomExistsInProject(actualID, project)) {
            return { atomID: actualID, project, path: getAtomPath(actualID, project) };
        }
        return null;
    }

    // Search all projects
    for (const project of workspaceConfig.projects) {
        if (atomExistsInProject(actualID, project)) {
            // Return with project prefix
            return {
                atomID: `${project.name}:${actualID}`,
                project,
                path: getAtomPath(actualID, project)
            };
        }
    }

    return null;
}

// Go to definition for atom references
const provider = vscode.languages.registerDefinitionProvider(['markdown', 'go', 'javascript'], {
    provideDefinition(document, position, token) {
        const range = document.getWordRangeAtPosition(position, /\[\[[a-z_:]+\]\]/);
        if (!range) return;

        const text = document.getText(range);
        const atomID = text.replace(/[\[\]]/g, '');

        const resolved = resolveAtomID(atomID);
        if (!resolved) return null;

        const uri = vscode.Uri.file(resolved.path);
        const doc = vscode.workspace.openTextDocument(uri);

        return new vscode.Location(uri, new vscode.Position(0, 0));
    }
});
```

### 5.4 Enhanced Sidebar

**File:** `extension/extension.js`

```javascript
// Show all projects in sidebar
function refreshAtomExplorer() {
    if (!workspaceConfig) return;

    const treeDataProvider = new AtomTreeDataProvider(workspaceConfig, activeProject);
    vscode.window.createTreeView('atd-explorer', {
        treeDataProvider,
        showCollapseAll: true
    });
}

class AtomTreeDataProvider {
    constructor(workspace, activeProject) {
        this.workspace = workspace;
        this.activeProject = activeProject;
    }

    getChildren(element) {
        if (!element) {
            // Top level: show projects
            if (this.workspace.isWorkspace) {
                return this.workspace.projects.map(p => ({
                    label: p.name,
                    collapsibleState: p.name === this.activeProject?.name ?
                        vscode.TreeItemCollapsibleState.Expanded :
                        vscode.TreeItemCollapsibleState.Collapsed,
                    project: p
                }));
            }
            // Single project: show atoms
            return this.getAtoms(this.workspace.docsPath);
        }

        // Project expanded: show its atoms
        if (element.project) {
            return this.getAtoms(element.project.docsPath);
        }

        return [];
    }

    getAtoms(docsPath) {
        const atoms = glob.sync('*.atom.md', { cwd: docsPath });
        return atoms.map(atom => ({
            label: atom.replace('.atom.md', ''),
            collapsibleState: vscode.TreeItemCollapsibleState.None
        }));
    }
}
```

**Acceptance Criteria:**
- [ ] Extension detects `.atd.workspace`
- [ ] Sidebar shows all projects
- [ ] User can switch active project
- [ ] Go-to-definition works across projects
- [ ] Hover shows atom info from any project
- [ ] ATD graph includes all projects

---

## Phase 6: Testing & Validation

### 6.1 Unit Tests

**File:** `workspace/resolver_test.go`

```go
func TestResolver_LocalReference(t *testing.T) {
    ws := createTestWorkspace()
    idx, _ := ws.BuildIndex()
    resolver := NewResolver(ws, idx, "upsilonapi")

    parsed, err := resolver.Resolve("[[api_auth_login]]")
    assert.NoError(t, err)
    assert.Equal(t, ReferenceLocal, parsed.Type)
    assert.Equal(t, "api_auth_login", parsed.AtomID)
}

func TestResolver_CrossProjectReference(t *testing.T) {
    ws := createTestWorkspace()
    idx, _ := ws.BuildIndex()
    resolver := NewResolver(ws, idx, "upsilonapi")

    // Resolve reference to atom in upsiloncli
    parsed, err := resolver.Resolve("[[uc_player_login]]")
    assert.NoError(t, err)
    assert.Equal(t, ReferenceCrossProject, parsed.Type)
    assert.Equal(t, "upsiloncli", parsed.Project)
    assert.Equal(t, "uc_player_login", parsed.AtomID)
}

func TestCanonicalForm(t *testing.T) {
    ws := createTestWorkspace()
    idx, _ := ws.BuildIndex()
    resolver := NewResolver(ws, idx, "upsilonapi")

    canonical, err := resolver.CanonicalForm("[[uc_player_login]]")
    assert.NoError(t, err)
    assert.Equal(t, "[[upsiloncli:uc_player_login]]", canonical)
}

func TestShouldUpdate(t *testing.T) {
    ws := createTestWorkspace()
    idx, _ := ws.BuildIndex()
    resolver := NewResolver(ws, idx, "upsilonapi")

    shouldUpdate, canonical := resolver.ShouldUpdate("[[uc_player_login]]")
    assert.True(t, shouldUpdate)
    assert.Equal(t, "[[upsiloncli:uc_player_login]]", canonical)
}
```

### 6.2 Integration Tests

**File:** `test/workspace_integration_test.sh`

```bash
#!/bin/bash
set -e

TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"

# Create test workspace
mkdir -p project_a/docs project_b/docs

cat > .atd.workspace <<EOF
{
  "workspace_name": "test-workspace",
  "projects": [
    {"name": "project_a", "path": "project_a"},
    {"name": "project_b", "path": "project_b"}
  ]
}
EOF

# Create atom in project_a
cat > project_a/docs/atom_a.atom.md <<EOF
---
id: atom_a
parents:
  - [[atom_b]]
---
# Atom A
EOF

# Create atom in project_b
cat > project_b/docs/atom_b.atom.md <<EOF
---
id: atom_b
dependents: []
---
# Atom B
EOF

# Run weave
atd weave --workspace

# Check that atom_a now has project prefix
grep -q "project_b:atom_b" project_a/docs/atom_a.atom.md || {
    echo "FAIL: atom_a should have project_b:atom_b reference"
    exit 1
}

# Check that atom_b now lists atom_a as dependent
grep -q "project_a:atom_a" project_b/docs/atom_b.atom.md || {
    echo "FAIL: atom_b should have project_a:atom_a as dependent"
    exit 1
}

echo "PASS: weave auto-fixed cross-project references"

# Cleanup
cd -
rm -rf "$TEST_DIR"
```

### 6.3 Manual Testing Checklist

- [ ] Create workspace with 3+ projects
- [ ] Add cross-project references manually
- [ ] Run `atd weave --workspace` - verify prefixes added
- [ ] Run `atd verify` - verify no errors for cross-project @spec-link
- [ ] Run `atd recon cross:project:atom` - verify finds atom
- [ ] Open VS Code in workspace - verify all projects shown
- [ ] Use go-to-definition on cross-project reference - verify works
- [ ] Switch active project - verify sidebar updates

---

## Implementation Timeline

| Phase | Tasks | Duration | Dependencies |
|-------|-------|----------|--------------|
| **Phase 1** | Workspace API, Loader, Registry | 2-3 days | None |
| **Phase 2** | Reference Resolver | 2-3 days | Phase 1 |
| **Phase 3** | Upgrade atd weave | 2-3 days | Phase 2 |
| **Phase 4** | Upgrade verify/recon/test-links | 3-4 days | Phase 2 |
| **Phase 5** | Upgrade VS Code extension | 2-3 days | Phase 2 |
| **Phase 6** | Testing & Validation | 2-3 days | All phases |
| **Total** | | 13-19 days | |

**Suggested Order:**
1. Week 1: Phase 1-2 (Foundation)
2. Week 2: Phase 3-4 (CLI Tools)
3. Week 3: Phase 5-6 (Extension + Testing)

---

## Success Criteria

The upgrade is complete when:

1. ✅ `atd weave` automatically adds project prefixes to cross-project references
2. ✅ `atd verify` finds @spec-link targets in any project
3. ✅ `atd recon` validates atoms across projects
4. ✅ `atd test-links` finds @test-link targets in any project
5. ✅ VS Code extension shows all projects in sidebar
6. ✅ VS Code go-to-definition works across projects
7. ✅ All tools accept both `[[atom_id]]` and `[[project:atom_id]]` formats
8. ✅ Upsilon-hub workspace works without manual reference fixes

---

## Open Questions

1. **Reference Format:** Should `atd weave` remove redundant project prefixes for local references? (e.g., `[[project_a:atom_a]]` → `[[atom_a]]` when in project_a): YES

2. **Performance:** For large workspaces (1000+ atoms), should the index be persisted to disk?: NO

3. **Shared Libraries:** How should the `shared_libraries` config be handled? Separate special project, or just symlinks?

4. **Circular Dependencies:** How should `atd weave` handle circular dependencies across projects? 

5. **CI/CD:** How should CI verify that all cross-project references are correctly prefixed?

---

## References

- [ISS-091: ATD Workspace & Multi-Project Support](/home/bastien/work/skill/issues/ISS-091_20260423_atd_workspace_multi_project_support.md)
- [Migration Investigation Result](/home/bastien/work/skill/migration_result.md)
- [Upsilon-Hub Workspace](/home/bastien/work/skill/upsilon-hub/) - Test workspace for validation
