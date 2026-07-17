package cmd

// @spec-link [[service_atd_init]]

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
	"github.com/spf13/cobra"
)

const preCommitHookContent = `#!/bin/bash
# ATD Structural Integrity Check
# Zero-latency, LLM-free structural bouncer.
# Ensures every ARCHITECTURE or IMPLEMENTATION atom staged for commit
# declares at least one parent, preventing orphaned lower-layer atoms.
#
# Installed automatically by: atd init / atd init --upgrade

STAGED_ATOMS=$(git diff --cached --name-only --diff-filter=ACM | grep '\.atom\.md$')

if [ -z "$STAGED_ATOMS" ]; then
  exit 0
fi

FAIL=0

echo "🔍 Running ATD Structural Integrity Check..."

for file in $STAGED_ATOMS; do
  LAYER=$(grep -E "^layer:" "$file" | awk '{print $2}' | tr -d '\r')
  PARENTS_COUNT=$(awk '/^parents:/ {flag=1; next} /^[^ -]/ {flag=0} flag && /-[[:space:]]+\[\[.*\]\]/ {print}' "$file" | wc -l)

  if [[ "$LAYER" == "IMPLEMENTATION" || "$LAYER" == "ARCHITECTURE" ]]; then
    if [ "$PARENTS_COUNT" -eq 0 ]; then
      echo "❌ ERROR: Orphaned Atom Detected -> $file"
      echo "   Reason: This is an $LAYER atom but has no parents defined."
      echo "   Fix 1 : Add a parent business/design requirement -> parents: [[req_your_parent]]"
      echo "   Fix 2 : Use the escape hatch -> parents: [[req_tech_debt_backlog]]"
      echo ""
      FAIL=1
    fi
  fi
done

if [ "$FAIL" -eq 1 ]; then
  echo "🛑 ATD Check Failed. Commit aborted. Please fix the above atoms."
  exit 1
fi

# Dogfood gate (test_atd_07_26.md §3.4, WP-3): if this checkout carries the
# ATD toolkit's own Makefile with a dogfood-quick target, run the fast
# lint-ratchet subset whenever an atom is staged. This is a no-op (skipped
# silently) in any downstream project that only installed this hook via
# 'atd init'/'atd init --upgrade' and doesn't have atd/Makefile at all -- the
# hook stays generic; only ATD's own repo dogfoods itself here. Kept fast on
# purpose: it is pure file parsing plus link resolution against the already
# staged docs corpus, no binary build and no full 'check --full' walk (that
# full gate is 'make -C atd dogfood', a manual/CI step).
if [ -f "atd/Makefile" ] && grep -q '^dogfood-quick:' atd/Makefile; then
  echo "🔍 Running ATD dogfood gate (quick)..."
  if ! make -C atd dogfood-quick; then
    echo "🛑 Dogfood gate failed (new lint finding vs. atd/testdata/dogfood_baseline.txt)."
    echo "   Run 'make -C atd dogfood' for the full report, fix the atom, or if it is"
    echo "   genuinely pre-existing/accepted debt: 'make -C atd dogfood-update-baseline'."
    exit 1
  fi
fi

echo "✅ ATD Check Passed."
exit 0
`

const techDebtAtomContent = `---
id: req_tech_debt_backlog
human_name: Technical Debt Backlog
type: REQUIREMENT
layer: BUSINESS
version: 1.0
status: STABLE
priority: 3
tags: [tech-debt, escape-hatch, governance]
parents: []
dependents: []
---
# Technical Debt Backlog

## INTENT
Acts as a temporary anchor for implementation atoms created without formal business requirements.

## THE RULE / LOGIC
Any ARCHITECTURE or IMPLEMENTATION atom that cannot yet be traced to a real business requirement must declare ` + "`[[req_tech_debt_backlog]]`" + ` as its parent rather than being committed without traceability. These atoms must be groomed and re-parented to proper business atoms during scheduled tech-debt cycles.

- **This is not a free pass.** Using this anchor must be intentional and visible to reviewers.
- **Cycle responsibility:** The team must periodically query ` + "`parents: [[req_tech_debt_backlog]]`" + ` and groom those atoms toward real requirements.
- **Acceptable scenarios:** rapid prototyping, emergency fixes, infrastructure atoms with no direct user story.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** ` + "`@spec-link [[req_tech_debt_backlog]]`" + ` — do not use in source code; this is documentation-only.
- **Query:** ` + "`atd query --field parents --search req_tech_debt_backlog`" + ` to list all tech debt atoms.

## EXPECTATION (For Testing)
- No atom at ARCHITECTURE or IMPLEMENTATION layer is committed with an empty ` + "`parents`" + ` list.
- The count of atoms pointing here decreases over time, not increases.
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap a .atd config file in a directory",
	Long: `Create a default .atd configuration file in the specified directory.

atd init must be run once per project before any other atd command.
It writes a complete default configuration including docs path, LLM provider
chain, and model assignments, and installs the ATD pre-commit hook.

Use --upgrade on an existing project to install or refresh the pre-commit hook
and ensure the tech-debt escape-hatch atom exists without touching .atd.

Examples:
  atd init                       # Initialize in current directory
  atd init --dir /path/to/proj   # Initialize in a specific directory
  atd init --docs specs/ --model deepseek-r1:7b
  atd init --force               # Overwrite an existing .atd
  atd init --upgrade             # Install hook + atom only (safe on existing projects)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		docsPath, _ := cmd.Flags().GetString("docs")
		model, _ := cmd.Flags().GetString("model")
		force, _ := cmd.Flags().GetBool("force")
		upgrade, _ := cmd.Flags().GetBool("upgrade")

		result, err := runInit(dir, docsPath, model, force, upgrade)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

// runInit creates a .atd file in dir with the provided defaults, installs the
// pre-commit hook, and ensures the tech-debt escape-hatch atom exists.
// When upgrade is true, the .atd file is left untouched.
func runInit(dir, docsPath, model string, force, upgrade bool) (string, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot determine current directory: %w", err)
		}
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("invalid directory: %w", err)
	}

	if stat, err := os.Stat(absDir); err != nil || !stat.IsDir() {
		return "", fmt.Errorf("directory does not exist: %s", absDir)
	}

	if docsPath == "" {
		docsPath = "docs/"
	}
	if model == "" {
		model = "llama3.2"
	}

	var msgs []string

	if upgrade {
		// In upgrade mode, read existing .atd to discover the real docs path.
		if err := config.LoadFromDir(absDir); err == nil && config.ActiveConfig.DocsPath != "" {
			docsPath = config.ActiveConfig.DocsPath
		}
		msgs = append(msgs, fmt.Sprintf("Mode:    upgrade (config unchanged)"))
	} else {
		atdPath := filepath.Join(absDir, ".atd")

		if _, err := os.Stat(atdPath); err == nil && !force {
			return "", fmt.Errorf(".atd already exists at %s — use --force to overwrite, or --upgrade to add governance files only", atdPath)
		}

		cfg := config.Config{
			DocsPath:                docsPath,
			DiffSimilarityThreshold: 0.85,
			BloatingFactor: config.BloatingFactorConfig{
				Default: 0.8,
				TypeOverrides: map[string]float64{
					"REQUIREMENT":   0.3,
					"SPECIFICATION": 0.3,
					"MODULE":        0.3,
					"USECASE":       0.1,
					"USER_STORY":    0.1,
					"API":           0.1,
				},
			},
			Model: model,
			Logging: config.LoggingConfig{
				LogPath: ".agent/logs/atd_trace.log",
			},
			LLM: config.LLMConfig{
				Providers: []config.LLMProvider{
					{Name: "remote", BaseURL: "http://192.168.1.10:11434", TimeoutMs: 2000},
					{Name: "local", BaseURL: "http://localhost:11434", TimeoutMs: 500},
					{Name: "ide_agent", Type: "passthrough"},
				},
				Models: map[string]config.ModelConfig{
					"qwen2.5-coder:14b": {Tasks: []string{"code_analysis"}, Priority: 50},
					"deepseek-r1:7b":    {Tasks: []string{"code_analysis", "text_analysis", "text_generation"}, Priority: 40},
					"llama3.1:8b":       {Tasks: []string{"*"}, Priority: 20},
					"llama3.2":          {Tasks: []string{"*"}, Priority: 10},
					"nomic-embed-text":  {Tasks: []string{"embedding"}, Priority: 100},
				},
				FallbackModel: model,
			},
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := os.WriteFile(atdPath, data, 0644); err != nil {
			return "", fmt.Errorf("failed to write .atd: %w", err)
		}
		msgs = append(msgs, fmt.Sprintf("Config:  %s", atdPath))

		// Create docs directory if it doesn't exist.
		docsAbs := filepath.Join(absDir, docsPath)
		if _, err := os.Stat(docsAbs); os.IsNotExist(err) {
			if mkErr := os.MkdirAll(docsAbs, 0755); mkErr != nil {
				return "", fmt.Errorf("created .atd but failed to create docs dir %s: %w", docsAbs, mkErr)
			}
			msgs = append(msgs, fmt.Sprintf("Docs:    %s (created)", docsAbs))
		} else {
			msgs = append(msgs, fmt.Sprintf("Docs:    %s", docsAbs))
		}
	}

	// Install pre-commit hook.
	hookPath, hookErr := installPreCommitHook(absDir)
	if hookErr != nil {
		msgs = append(msgs, fmt.Sprintf("Hook:    SKIPPED (%s)", hookErr))
	} else {
		msgs = append(msgs, fmt.Sprintf("Hook:    %s", hookPath))
	}

	// Ensure tech-debt escape-hatch atom.
	docsAbs := filepath.Join(absDir, docsPath)
	atomPath, atomCreated, atomErr := ensureTechDebtAtom(docsAbs)
	if atomErr != nil {
		msgs = append(msgs, fmt.Sprintf("Atom:    SKIPPED (%s)", atomErr))
	} else if atomCreated {
		msgs = append(msgs, fmt.Sprintf("Atom:    %s (created)", atomPath))
	} else {
		msgs = append(msgs, fmt.Sprintf("Atom:    %s (already exists)", atomPath))
	}

	return "Initialized ATD project.\n  " + strings.Join(msgs, "\n  "), nil
}

// installPreCommitHook writes the ATD structural check hook to .git/hooks/pre-commit.
// Returns the hook path on success, or an error if the directory is not a git repo.
func installPreCommitHook(projectDir string) (string, error) {
	gitHooksDir := filepath.Join(projectDir, ".git", "hooks")
	if _, err := os.Stat(gitHooksDir); os.IsNotExist(err) {
		return "", fmt.Errorf("no .git/hooks/ found — not a git repository")
	}
	hookPath := filepath.Join(gitHooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(preCommitHookContent), 0755); err != nil {
		return "", fmt.Errorf("failed to write hook: %w", err)
	}
	return hookPath, nil
}

// ensureTechDebtAtom creates docs/req_tech_debt_backlog.atom.md if it does not
// already exist. Returns (path, created, error).
func ensureTechDebtAtom(docsDir string) (string, bool, error) {
	atomPath := filepath.Join(docsDir, "req_tech_debt_backlog.atom.md")
	if _, err := os.Stat(atomPath); err == nil {
		return atomPath, false, nil
	}
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return "", false, fmt.Errorf("failed to create docs dir: %w", err)
	}
	if err := os.WriteFile(atomPath, []byte(techDebtAtomContent), 0644); err != nil {
		return "", false, fmt.Errorf("failed to write atom: %w", err)
	}
	return atomPath, true, nil
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("dir", "", "Target directory (defaults to current directory)")
	initCmd.Flags().String("docs", "docs/", "Path to docs folder (relative to target dir)")
	initCmd.Flags().String("model", "llama3.2", "Default fallback model name")
	initCmd.Flags().Bool("force", false, "Overwrite an existing .atd file")
	initCmd.Flags().Bool("upgrade", false, "Install hook and atom only — safe to run on existing projects")
}
