# Issue: ATD Configuration Parent Directory Search

**ID:** `20260418_atd_config_parent_directory_search`
**Ref:** `ISS-082`
**Date:** 2026-04-18
**Severity:** High
**Status:** Open
**Component:** `scripts/cmd/atd/config/config.go`
**Affects:** ATD usability in subdirectories, nested project configurations

---

## Summary

ATD CLI tool only searches for `.atd` configuration file in the current working directory (CWD). When invoked from a subdirectory without a local `.atd` file, it fails to start. The tool should search upward through parent directories until it finds a `.atd` file or reaches the filesystem root, similar to other developer tools like `git`, `npm`, etc.

---

## Technical Description

### Background
ATD CLI loads configuration from `.atd` file to determine documentation paths, code paths, LLM providers, and other project-specific settings. This enables different projects or subdirectories to have their own ATD configurations without global settings.

### The Problem Scenario
1. **CWD-Only Search**: Current implementation only checks `./.atd` in current directory
2. **Subdirectory Invocation**: When run from `upsilon-hub/battleui/` with no local `.atd`, tool fails
3. **No Fallback**: No upward directory search for configuration files
4. **Root Fallback Missing**: When reaching `/` without finding `.atd`, should fail gracefully
5. **Multi-Config Support**: Cannot handle scenarios where multiple `.atd` files exist in directory tree

### Where This Pattern Exists Today
- `scripts/cmd/atd/config/config.go`: Configuration loading logic likely checks CWD only
- `scripts/cmd/atd/cmd/root.go` (or similar): Root command initialization
- ATD CLI entry point: Main command setup before Cobra commands

### Expected Behavior vs Actual Behavior

**Expected Behavior:**
```bash
# Should work from any directory:
cd upsilon-hub/battleui && atd stats  # Should find .atd in upsilon-hub/
cd upsilon-hub/docs && atd stats       # Should find .atd in upsilon-hub/  
cd /home/bastien/work/skill && atd stats # Should find .atd in work/skill/
```

**Current Behavior:**
```bash
# Only works when .atd is in CWD:
cd upsilon-hub && atd stats           # ✅ Works
cd upsilon-hub/battleui && atd stats  # ❌ Fails: No .atd found
```

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Occurs anytime ATD is invoked from subdirectory) |
| Impact if triggered | High (Tool completely unusable from subdirectories) |
| Detectability | High (Clear error message: "No .atd configuration found") |
| Current mitigant | Users must `cd` to project root or create local `.atd` files |

---

## Recommended Fix

**Short term:** Implement parent directory search for `.atd` configuration:
- Start in CWD, search for `.atd` file
- If not found, move to parent directory
- Repeat until `.atd` found or root (`/`) reached
- If root reached without finding `.atd`, fail with clear error message

**Medium term:** Add multi-configuration support:
- Track all `.atd` files found during directory search
- Provide configuration priority (most recent/highest in tree wins)
- Support configuration layering for nested project structures
- Add `--config` flag to explicitly specify configuration file path

**Long term:** Implement workspace awareness:
- Detect workspace root patterns (git worktree, VS Code workspace, etc.)
- Automatically determine project boundaries
- Provide workspace-scoped configuration search
- Support multiple active workspaces

---

## Implementation Details

### Algorithm for Parent Directory Search

```go
func FindATDConfig() (string, error) {
    cwd, err := os.Getwd()
    if err != nil {
        return "", fmt.Errorf("failed to get working directory: %v", err)
    }
    
    dir := cwd
    var configFiles []string
    
    for {
        // Check for .atd file in current directory
        configPath := filepath.Join(dir, ".atd")
        if _, err := os.Stat(configPath); err == nil {
            if !os.IsDir(configPath) {
                configFiles = append(configFiles, configPath)
            }
        }
        
        // Stop at root or filesystem boundary
        if dir == "/" {
            break
        }
        
        parent := filepath.Dir(dir)
        if parent == dir {
            // We've reached the root without finding .atd
            break
        }
        
        dir = parent
    }
    
    // Determine which config to use
    if len(configFiles) == 0 {
        return "", fmt.Errorf("no .atd configuration found in directory tree up to %s", cwd)
    }
    
    // Use highest/most recent config if multiple found
    return configFiles[0], nil
}
```

### Configuration Priority Logic

```go
type ConfigPriority struct {
    Path     string
    Depth    int     // How many levels from CWD
    Modified time.Time
}

func SelectBestConfig(configFiles []string) string {
    if len(configFiles) == 0 {
        return ""
    }
    
    var configs []ConfigPriority
    cwd, _ := os.Getwd()
    
    for _, configPath := range configFiles {
        absPath, _ := filepath.Abs(configPath)
        relPath, _ := filepath.Rel(cwd, absPath)
        depth := strings.Count(relPath, string(os.PathSeparator))
        
        info, _ := os.Stat(absPath)
        configs = append(configs, ConfigPriority{
            Path:     absPath,
            Depth:    depth,
            Modified: info.ModTime(),
        })
    }
    
    // Sort by depth (closest first), then by modification time
    sort.Slice(configs, func(i, j int) bool {
        if configs[i].Depth != configs[j].Depth {
            return configs[i].Depth < configs[j].Depth
        }
        return configs[i].Modified.After(configs[j].Modified)
    })
    
    return configs[0].Path
}
```

### User-Facing Error Messages

**No Config Found:**
```
Error: No .atd configuration found in directory tree up to /home/bastien/work/skill
Searched 6 directories: 
  /home/bastien/work/skill/upsilon-hub
  /home/bastien/work/skill/upsilon-hub/docs
  /home/bastien/work/skill/upsilon-hub/upsilonbattle
  /home/bastien/work/skill/upsilon-hub/battleui
  /home/bastien/work/skill/atd
  /home/bastien/work/skill

To fix this:
  1. Create a .atd file in your project root
  2. Or use 'atd --config <path>' to specify configuration location
  3. Run 'atd init' to create a default configuration file
```

**Root Reached:**
```
Error: Reached filesystem root '/' without finding .atd configuration
ATD requires a .atd configuration file to operate.

To fix this:
  1. Create a .atd file in your project directory
  2. Run ATD commands from a directory containing .atd
```

---

## Integration with Existing Issues

### Related Issues
- **ISS-071** (ATD Indexing System Failure): Improved directory search may help with code path discovery
- **ISS-072** (ATD Orphan Detection): Better config resolution may improve file discovery
- **ISS-078** (CLAUDE.md Context Mismatch): Improved configuration behavior documented in CLAUDE.md

### Configuration Examples

**Scenario 1: Single Project**
```
/home/bastien/work/skill/
├── .atd                    ← Main config
├── atd/                     ← ATD tools  
└── upsilon-hub/
    └── .atd              ← Project-specific config
```
*Result: Uses most specific config (upsilon-hub/.atd)

**Scenario 2: Nested Workspaces**
```
/home/bastien/work/skill/
├── .atd                    ← Root workspace config
├── projects/
│   ├── upsilon-hub/
│   │   └── .atd        ← Project config
│   └── other-project/
│       └── .atd            ← Other project config
└── atd/                     ← ATD tools
```
*Result: Uses closest config, supports nested projects

**Scenario 3: Git Worktree**
```
~/workspaces/
├── upsilon-hub/
│   ├── .atd              ← Project config  
│   ├── upsilonbattle/
│   ├── battleui/
│   └── upsiloncli/
└── atd-tools/
    └── .atd                  ← ATD config
```
*Result: Respects git worktree structure, uses project-specific config

---

## Use Cases

### Use Case 1: Development from Subdirectory
```bash
cd /home/bastien/work/skill/upsilon-hub/battleui
atd stats  # Should find upsilon-hub/.atd
```
**Expected:** Success (finds parent config)

### Use Case 2: Testing from Documentation Directory
```bash
cd /home/bastien/work/skill/upsilon-hub/docs
atd audit  # Should find upsilon-hub/.atd
```
**Expected:** Success (finds parent config)

### Use Case 3: Global Installation Usage
```bash
cd /home/bastien/work/skill
atd verify  # Should find .atd in work/skill
```
**Expected:** Success (uses local config)

### Use Case 4: Explicit Config Path
```bash
atd --config /custom/path/config.atd stats
```
**Expected:** Uses specified config, overrides search behavior

---

## Testing Requirements

### Test Cases for Validation

1. **Basic Subdirectory Search**
   - Navigate to subdirectory, run ATD command
   - Verify parent config is found and used
   - **Expected:** No error, correct paths used

2. **Deep Nested Directory Search**
   - Navigate multiple levels deep, run ATD command
   - Verify search traverses upward correctly
   - **Expected:** Config found despite deep nesting

3. **No Config Scenario**
   - Create empty directory without `.atd`, navigate there
   - Verify clear error message about missing config
   - **Expected:** Graceful failure, helpful error message

4. **Root Boundary Scenario**
   - Navigate to directory tree with no `.atd`, attempt to reach root
   - Verify proper behavior at filesystem boundary
   - **Expected:** Fails gracefully at root, doesn't infinite loop

5. **Multi-Config Priority**
   - Create directory tree with multiple `.atd` files
   - Navigate to various levels, verify correct priority
   - **Expected:** Uses closest/most recent config

6. **Explicit Config Override**
   - Use `--config` flag with custom path
   - Verify specified config is used regardless of location
   - **Expected:** Uses specified config, ignores search

---

## Benefits

### User Experience
- **Intuitive Behavior**: Matches expectations from other CLI tools (`git`, `npm`, etc.)
- **Flexible Usage**: Works from any directory without `cd` to project root
- **Clear Errors**: Helpful error messages guide users to solutions
- **Workspace Support**: Handles nested project structures naturally

### Developer Experience
- **Subdirectory Development**: Can work in isolated subdirectories
- **Multiple Projects**: Easy switching between different projects
- **Configuration Isolation**: Project-specific settings don't interfere
- **Quick Testing**: Run commands from test directories without setup

### System Robustness
- **No Global Dependencies**: Doesn't rely on global configuration files
- **Filesystem Safe**: Proper boundary checking prevents infinite loops
- **Backwards Compatible**: Existing `.atd` in CWD still works
- **Config Priority**: Predictable behavior with multiple configs

---

## References

- [Current Config Implementation](file:///home/bastien/work/skill/atd/pkg/config/config.go)
- [CLI Entry Point](file:///home/bastien/work/skill/atd/cmd/atd/cmd/root.go)
- [Configuration Documentation](file:///home/bastien/work/skill/ATD.md#17-configuration)
- [Similar Tools Pattern](file:///home/bastien/work/skill/scripts/cmd/atd/cmd/config.go)