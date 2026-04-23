# Issue: ATD Workspace & Multi-Project Support

**ID:** `20260423_atd_workspace_multi_project_support`
**Ref:** `ISS-091`
**Date:** 2026-04-23
**Severity:** High
**Status:** Open
**Component:** ATD Core, CLI, MCP Server
**Affects:** Monorepos, parent projects with multiple sub-projects

---

## Summary

ATD currently assumes a single `.atd` configuration per directory tree. In monorepo structures like `upsilon-hub`, a parent project contains multiple sub-projects (battleui, upsilonbattle, upsiloncli, upsilonapi), each requiring its own ATD documentation, configuration, and code paths. Currently, all atoms crowd into a single parent `docs/` directory, making it difficult to maintain project-specific documentation and verify sub-project compliance.

**Related:** `ISS-082` (Parent Directory Search) - prerequisite for this feature

---

## Problem Scenario

### Current Structure (upsilon-hub)
```
upsilon-hub/
├── .atd                    # Single config for entire hub
├── docs/                    # All atoms (407+) for ALL sub-projects
│   ├── mechanic_atd_*       # ATD framework atoms
│   ├── ui_webui_*          # WebUI atoms
│   ├── mechanic_vscode_*     # VS Code extension atoms
│   ├── ui_dashboard_*       # battleui atoms (mixed!)
│   ├── entity_player_*       # upsilonbattle atoms (mixed!)
│   └── ui_registration_*    # upsiloncli atoms (mixed!)
├── battleui/                # Sub-project 1
│   ├── src/
│   └── package.json
├── upsilonbattle/           # Sub-project 2
│   ├── upsilonbattle/
│   └── go.mod
├── upsiloncli/              # Sub-project 3
│   └── upsiloncli/
└── upsilonapi/              # Sub-project 4
    └── upsilonapi/
```

**Issues:**
1. **No Isolation**: All sub-project atoms share `docs/` directory
2. **Crowded Namespace**: 407+ atoms makes navigation difficult
3. **Cross-Contamination**: Changing battleui doc might affect upsilonbattle via broken links
4. **No Per-Project Settings**: All sub-projects share same LLM, code paths, etc.
5. **Wrong Orphans**: 52 orphans are mostly atoms pointing to code in wrong sub-projects
6. **Verification Confusion**: `atd verify` sees changes across all sub-projects simultaneously

### Desired Structure
```
upsilon-hub/                      # Parent workspace
├── .atd.workspace                 # Workspace config (optional)
├── battleui/
│   ├── .atd                       # battleui-specific config
│   ├── docs/                       # battleui atoms only
│   └── src/
├── upsilonbattle/
│   ├── .atd                       # upsilonbattle-specific config
│   ├── docs/                       # upsilonbattle atoms only
│   └── upsilonbattle/
├── upsiloncli/
│   ├── .atd                       # upsiloncli-specific config
│   └── upsiloncli/
└── upsilonapi/
    ├── .atd                       # upsilonapi-specific config
    └── upsilonapi/
```

---

## Requirements

### Core Functionality

#### R1: Workspace Detection
- Detect when ATD is invoked from a directory within a workspace
- Identify active sub-project based on CWD
- Use sub-project's `.atd` config by default

#### R2: Project Isolation
- Each sub-project maintains its own `docs/` directory
- Each sub-project can have different `code_paths` in its `.atd`
- Each sub-project can have different LLM providers/models
- Orphans are scoped to active project only

#### R3: Cross-Project Links (Optional, P1)
- Support `@spec-link [[other-project:atom_id]]` syntax for external references
- Visual indication of external links in tools
- Optional: Shared atoms (common business rules) via reference

#### R4: Workspace-Wide Operations
- `atd stats --workspace` to aggregate stats from all sub-projects
- `atd audit --workspace` to audit entire workspace
- `atd crawl --workspace` to build workspace-wide dependency graph

#### R5: Workspace Configuration (.atd.workspace)
- Define sub-project boundaries
- Shared settings (common LLM provider, shared atom libraries)
- Workspace-level policies (naming conventions, type restrictions)

---

## Configuration Schema

### Workspace Config: `.atd.workspace`

```json
{
  "workspace_name": "upsilon-hub",
  "workspace_root": ".",
  "projects": [
    {
      "name": "battleui",
      "path": "./battleui",
      "config_path": "./battleui/.atd",
      "docs_path": "./battleui/docs",
      "code_paths": ["./battleui/src"]
    },
    {
      "name": "upsilonbattle",
      "path": "./upsilonbattle",
      "config_path": "./upsilonbattle/.atd",
      "docs_path": "./upsilonbattle/docs",
      "code_paths": ["./upsilonbattle/upsilonbattle"]
    },
    {
      "name": "upsiloncli",
      "path": "./upsiloncli",
      "config_path": "./upsiloncli/.atd",
      "docs_path": "./upsiloncli/docs",
      "code_paths": ["./upsiloncli/upsiloncli"]
    },
    {
      "name": "upsilonapi",
      "path": "./upsilonapi",
      "config_path": "./upsilonapi/.atd",
      "docs_path": "./upsilonapi/docs",
      "code_paths": ["./upsilonapi/upsilonapi"]
    }
  ],
  "shared_libraries": {
      "common_auth_rules": "./shared/docs/auth"
  },
  "common_settings": {
    "llm": {
      "providers": [
        {
          "name": "remote",
          "base_url": "http://192.168.1.10:11434"
        }
      ]
    }
  }
}
```

### Sub-Project Config: `.atd` (existing format, extended)

```json
{
  "docs_path": "docs/",
  "code_paths": ["../upsilonbattle/upsilonbattle"],
  "bloating_factor": {
    "default": 0.8
  },
  "llm": {
    "providers": [
      {
        "name": "remote",
        "base_url": "http://192.168.1.10:11434"
      }
    ]
  },
  "workspace_ref": "..",  // Reference to workspace root
  "external_projects": {      // Cross-project references
    "battleui": "../battleui",
    "upsiloncli": "../upsiloncli",
    "upsilonapi": "../upsilonapi"
  }
}
```

---

## CLI Changes

### New Command: `atd workspace`

```bash
# Initialize a workspace
atd workspace init --name "upsilon-hub"

# Add a sub-project
atd workspace add --name "battleui" --path ./battleui

# List all projects in workspace
atd workspace list

# Switch active project context
atd workspace use battleui

# Workspace-wide operations
atd stats --workspace          # Aggregate all sub-projects
atd audit --workspace         # Audit entire workspace
atd crawl --workspace         # Workspace dependency graph
```

### Modified Commands

All existing commands support `--workspace` flag for workspace-scope operations:

```bash
atd stats --project battleui           # Stats for specific project
atd verify --project upsilonbattle --file src/combat.go
atd search --workspace query "authentication"  # Search across all projects
```

### Environment-Aware Behavior

When invoked from `upsilon-hub/battleui/`:
```bash
cd upsilon-hub/battleui
atd stats              # Uses battleui/.atd (auto-detected)
atd verify              # Only sees battleui changes
atd crawl --gaps        # Only battleui orphans
```

When invoked from `upsilon-hub/`:
```bash
cd upsilon-hub
atd stats              # Error: specify --project or --workspace
atd stats --workspace  # Aggregate of all projects
atd stats --project battleui  # Specific project stats
```

---

## Atom Format Extensions

### Cross-Project Links

```markdown
---
id: api_auth_login
parents:
  - [[shared:auth_requirement]]          # External reference
  - [[battleui:login_component]]        # Cross-project reference
  - [[local:password_validation]]        # Local reference (explicit)
dependents:
  - [[upsiloncli:auth_command]]
---
```

### External Link Syntax

- `[[project_id:atom_id]]` - Link to atom in another project
- `[[shared:atom_id]]` - Link to shared library atom
- `[[local:atom_id]]` - Explicit local reference (default if no prefix)

### External Link Resolution

```go
type AtomReference struct {
    ProjectID string  // "" for local, "shared" for libraries, or project name
    AtomID    string  // The actual atom ID
}

func ResolveReference(ref AtomReference) (*Atom, error) {
    if ref.ProjectID == "" || ref.ProjectID == "local" {
        return FindLocalAtom(ref.AtomID)
    }
    if ref.ProjectID == "shared" {
        workspace := LoadWorkspaceConfig()
        return FindSharedAtom(workspace.SharedLibraries, ref.AtomID)
    }
    // Cross-project reference
    project := workspace.FindProject(ref.ProjectID)
    if project == nil {
        return nil, fmt.Errorf("project '%s' not found in workspace", ref.ProjectID)
    }
    return FindAtomInProject(project.ConfigPath, ref.AtomID)
}
```

---

## MCP Server Changes

### Session-Aware Operations

MCP server must track which project context is active:

```json
{
  "method": "initialize",
  "params": {
    "workspace_root": "/home/bastien/work/upsilon-hub",
    "active_project": "battleui",
    "capabilities": {
      "workspace_operations": true
    }
  }
}
```

### Workspace Tools

New MCP tools:

| Tool | Description |
|-------|-------------|
| `atd_workspace_list` | List all projects in workspace |
| `atd_workspace_stats` | Aggregate stats from all projects |
| `atd_workspace_switch` | Switch active project context |
| `atd_workspace_search` | Search atoms across all projects |

Modified tools:

| Tool | Change |
|-------|---------|
| `atd_query` | Add `project` parameter to filter by sub-project |
| `atd_search` | Add `scope: workspace` to search all projects |
| `atd_verify` | Auto-detect project from file path |

---

## Implementation Phases

### Phase 1: Core Workspace Detection (P0 - Foundation)

**Duration**: 2-3 days

**Tasks**:
1. Implement `FindProjectRoot()` with workspace detection
2. Parse `.atd.workspace` format
3. Auto-detect active project based on CWD
4. Update `atd init` to detect workspace context

**Deliverables**:
- Workspace config parser
- Project detection logic
- Unit tests for workspace discovery

### Phase 2: Command Isolation (P0 - Core)

**Duration**: 2-3 days

**Tasks**:
1. Scope all existing commands to active project
2. Add `--project` flag to override auto-detection
3. Update error messages for workspace context
4. Update `atd stats` to show project name

**Deliverables**:
- Project-isolated command execution
- Project override mechanism
- Updated documentation

### Phase 3: Workspace Commands (P1 - UX)

**Duration**: 2 days

**Tasks**:
1. Implement `atd workspace init`
2. Implement `atd workspace add`
3. Implement `atd workspace list`
4. Implement `atd workspace use`

**Deliverables**:
- `atd workspace` command group
- Workspace management UX
- Migration helper for existing setups

### Phase 4: Cross-Project Links (P2 - Advanced)

**Duration**: 3-4 days

**Tasks**:
1. Extend atom parser for `[[project:atom]]` syntax
2. Implement reference resolver across projects
3. Update `atd weave` to handle cross-project dependencies
4. Visual indicators in UI/tools for external links

**Deliverables**:
- Cross-project link support
- Updated dependency graph
- Visual feedback for external references

### Phase 5: Workspace-Wide Operations (P1 - UX)

**Duration**: 2-3 days

**Tasks**:
1. Implement `--workspace` flag for stats
2. Implement `--workspace` flag for audit
3. Implement `--workspace` flag for crawl
4. Aggregate results from all projects

**Deliverables**:
- Workspace-wide commands
- Aggregated reports
- Workspace-level orphans detection

### Phase 6: MCP Workspace Integration (P1 - Integration)

**Duration**: 2 days

**Tasks**:
1. Add workspace-aware tools to MCP server
2. Track session project context
3. Update VS Code extension to show active project

**Deliverables**:
- MCP workspace tools
- VS Code integration
- Updated documentation

---

## Testing Strategy

### Unit Tests

```go
func TestWorkspaceDetection(t *testing.T) {
    // Test auto-detection from subdirectory
    // Test workspace config parsing
    // Test project list parsing
    // Test override via --project flag
}

func TestCrossProjectReferences(t *testing.T) {
    // Test external reference parsing
    // Test shared library resolution
    // Test invalid reference handling
}

func TestProjectIsolation(t *testing.T) {
    // Test command scoped to project
    // Test stats return project-specific data
    // Test orphans are project-scoped
}
```

### Integration Tests

```bash
# Test 1: Basic workspace setup
mkdir -p test-workspace/{proj1,proj2}
cd test-workspace
atd workspace init --name "test"
atd workspace add --name "proj1" --path ./proj1
atd workspace add --name "proj2" --path ./proj2
atd workspace list  # Should show both projects

# Test 2: Project isolation
cd test-workspace/proj1
echo '{"docs_path":"docs/"}' > .atd
mkdir docs
atd stats  # Should work, use proj1 config

# Test 3: Workspace-wide stats
cd test-workspace
atd stats --workspace  # Should aggregate both projects

# Test 4: Cross-project links
cd test-workspace/proj1
cat > docs/shared.atom.md <<EOF
---
id: shared_rule
external_projects:
  proj2: "../proj2"
---
EOF
```

### Manual Testing Scenarios

1. **Monorepo Migration**: Set up workspace for upsilon-hub, migrate atoms
2. **Multi-Developer Workflow**: Two developers work in different sub-projects
3. **Shared Library**: Define common auth rules, reference from multiple projects
4. **IDE Integration**: Use VS Code extension with workspace-aware features

---

## Migration Guide

### From Single-Config to Workspace

```bash
# Step 1: Create workspace
cd upsilon-hub
atd workspace init --name "upsilon-hub"

# Step 2: For each sub-project, create its config and migrate atoms
atd workspace add --name "battleui" --path ./battleui
mkdir battleui/docs
mv docs/ui_webui_* battleui/docs/
mv docs/api_webui_* battleui/docs/
echo '{"docs_path":"docs/"}' > battleui/.atd

# Repeat for other sub-projects...
atd workspace add --name "upsilonbattle" --path ./upsilonbattle
mkdir upsilonbattle/docs
mv docs/entity_player_* upsilonbattle/docs/
mv docs/mech_character_* upsilonbattle/docs/
echo '{"docs_path":"docs/"}' > upsilonbattle/.atd

# Step 3: Update ATD framework docs to remain at workspace level (optional)
# or keep in a separate "atd-tooling" project

# Step 4: Verify
atd workspace list
atd workspace use battleui
atd stats
```

### Atom Reference Updates

After migration, update cross-project links:

```markdown
# Before
---
parents:
  - [[webui_login_flow]]
---

# After
---
parents:
  - [[battleui:webui_login_flow]]
---
```

---

## Benefits

### For Developers

- **Clear Ownership**: Each sub-project has its own documentation
- **Faster Verification**: `atd verify` only sees relevant changes
- **Isolated Context**: No accidental cross-project contamination
- **Flexible Settings**: Different LLM providers per project if needed

### For Architects

- **Workspace Visibility**: `atd stats --workspace` shows overall health
- **Shared Standards**: Workspace config enforces naming conventions
- **Cross-Project Analysis**: Dependency graph spans entire hub

### For CI/CD

- **Project-Specific Checks**: Each sub-project can run its own ATD verification
- **Workspace-Wide Gates**: CI can also run workspace-level audits
- **Incremental Adoption**: Migrate sub-projects one at a time

---

## Open Questions

1. **Shared Libraries**: Should atoms be physically copied or referenced via symlink?
2. **Version Compatibility**: What if sub-projects have incompatible ATD versions?
3. **Deployment**: How does `atd serve` handle workspace context for MCP?
4. **Nested Workspaces**: Support workspaces within workspaces (rare but possible)?
5. **Migration Automation**: Should we provide automated atom migration scripts?

---

## References

- [ISS-082: Parent Directory Search](./ISS-082_20260418_atd_config_parent_directory_search.md) - Prerequisite
- [ATD Manual](../ATD.md) - Base ATD specification
- [Mono-repo Best Practices](https://github.com/goldberghoni/dotfiles-starter-kit) - Industry patterns
- [Bazel Workspace Concepts](https://bazel.build/concepts/build-ref#workspace) - Similar patterns
