# Task 07 — Deterministic Subcommands

**Depends on:** Task 01 (cobra root), Task 02 (config), Task 03 (atom parser)  
**Produces:** 8 subcommands with no LLM dependency  
**Parallelizable:** Each subcommand can be written independently.

## Context

These tools have zero LLM involvement. They operate on file I/O, regex, JSON, and Go logic. Migration is straightforward: transplant logic from the original `main.go`, remove the `--project`/`--docs`/`--bin` flags (use `config.DocsDir()` instead), register as cobra commands.

## Pattern for Each Command

```go
package cmd

import (
    "atd-tools/config"
    "github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
    Use:   "example",
    Short: "One-line description",
    Long:  `Detailed help text explaining what, why, and how.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        docsDir := config.DocsDir()
        // ... tool logic ...
        return nil
    },
}

func init() {
    rootCmd.AddCommand(exampleCmd)
    exampleCmd.Flags().StringP("flag-name", "f", "default", "description")
}
```

## 7.1 — `atd update` (from `atd-update/main.go`, 356 lines)

**File:** `cmd/update.go`  
**Flags:** `-file`, `-intent`, `-logic`, `-interface`, `-set key=value`, `--spec-link <id> <file>` (absorbs atd-legacy-wrapper)  
**Source:** `scripts/atd-update/main.go` — copy entire logic.  
**Change:** Replace `flag.Parse()` with cobra flags. Replace cwd-based path resolution with `config.ProjectRoot()`. Add `--spec-link` subfunction that prepends `// @spec-link [[id]]` to a file (from `atd-legacy-wrapper/main.go` L46).  
**Test:** Create a temp `.atom.md`, run `atd update -file <path> -set status=STABLE`, verify file modified. Test `--spec-link` prepends tag.

## 7.2 — `atd crawl` (from `atd-crawl/main.go` + `atd-report-gaps/main.go`)

**File:** `cmd/crawl.go`  
**Flags:** `--src <path>`, `--gaps`, `--docs <override>`  
**Source:** Transplant `atd-crawl` graph-building logic. When `--gaps` is set, additionally filter for STABLE atoms with no implementations (from `atd-report-gaps`).  
**Test:** Run against `upsilonbattle/` with `--src` pointing to source. Verify JSON output. With `--gaps`, verify orphaned atoms listed.

## 7.3 — `atd query` (from `atd-query/main.go`, 98 lines)

**File:** `cmd/query.go`  
**Flags:** `-field`, `-search`  
**Source:** Direct transplant. Use `config.DocsDir()` for docs path.  
**Test:** Create sample atoms, query by field `type`, verify JSON output.

## 7.4 — `atd weave` (from `atd-link-weaver/main.go`, 106 lines)

**File:** `cmd/weave.go`  
**Flags:** none (operates on docs from config)  
**Source:** Direct transplant. Reads parents, updates dependents bidirectionally.  
**Test:** Create two atoms with parent relationship, run weave, verify dependents updated.

## 7.5 — `atd verify` (from `atd-verify-diff/main.go`, 149 lines)

**File:** `cmd/verify.go`  
**Flags:** none (git-driven)  
**Source:** Direct transplant. Outputs LLM prompt to stdout (kept as stdout passthrough for IDE Agent). Uses `config.DocsDir()`.  
**Test:** Verify it runs in a git repo and produces structured output.

## 7.6 — `atd roadmap` (from `atd-roadmap-builder/main.go`, 167 lines)

**File:** `cmd/roadmap.go`  
**Flags:** `--dir`, `--out`  
**Source:** Direct transplant. Uses `config.ActiveConfig.SupportedExtensions`.  
**Test:** Run against `upsilonbattle/`, verify `roadmap.json` created with items.

## 7.7 — `atd assemble` (from `atd-assemble/main.go` + `atd-generate-snapshot/main.go`)

**File:** `cmd/assemble.go`  
**Flags:** `--starts`, `--purpose`, `--snapshot --theme`  
**Source:** Base logic from `atd-assemble`. When `--snapshot` is set, wrap assembled output in the narrative rewrite prompt from `atd-generate-snapshot`. The prompt becomes an LLM task (`snapshot`) handled by the provider chain (done in Task 11).  
**Test:** Create linked atoms, run `atd assemble --starts root_id`, verify recursive stitching.

## 7.8 — `atd test-links` (NEW — ISS-022)

**File:** `cmd/test_links.go`  
**Flags:** `--atom <id>`, `--src <path>`  
**Logic:**
1. Walk source files looking for `@test-link [[ATOM_ID]]` tags
2. If `--atom` specified, also walk ATD hierarchy (parents/dependents) to find all related tests
3. Output JSON: `[{atom_id, test_file, line}]`

**Convention:** `@test-link [[ATOM_ID]]` placed in test files alongside `@spec-link`.  
**Test:** Add `@test-link` to a test file in `upsilonbattle/`, run `atd test-links`, verify discovery.

## Acceptance Criteria

- [ ] All 8 commands register and show in `atd --help`
- [ ] Each command has `--help` with descriptive text
- [ ] `atd update -file ...` modifies atoms correctly
- [ ] `atd crawl --gaps` shows orphaned STABLE atoms
- [ ] `atd query -search "movement"` returns JSON
- [ ] `atd weave` updates dependents bidirectionally
- [ ] `atd roadmap --dir ../upsilonbattle/` generates JSON
- [ ] `go test` passes for any unit tests added
