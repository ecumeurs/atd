package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"atd-tools/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap a .atd config file in a directory",
	Long: `Create a default .atd configuration file in the specified directory.

atd init must be run once per project before any other atd command.
It writes a complete default configuration including docs path, LLM provider
chain, model assignments, and logging setup.

Examples:
  atd init                       # Initialize in current directory
  atd init --dir /path/to/proj   # Initialize in a specific directory
  atd init --docs specs/ --model deepseek-r1:7b
  atd init --force               # Overwrite an existing .atd`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		docsPath, _ := cmd.Flags().GetString("docs")
		model, _ := cmd.Flags().GetString("model")
		force, _ := cmd.Flags().GetBool("force")

		result, err := runInit(dir, docsPath, model, force)
		if err != nil {
			return err
		}
		fmt.Println(result)
		return nil
	},
}

// runInit creates a .atd file in dir with the provided defaults.
// Returns a human-readable success message, or an error.
func runInit(dir, docsPath, model string, force bool) (string, error) {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("cannot determine current directory: %w", err)
		}
	}

	// Resolve to absolute path.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("invalid directory: %w", err)
	}

	if stat, err := os.Stat(absDir); err != nil || !stat.IsDir() {
		return "", fmt.Errorf("directory does not exist: %s", absDir)
	}

	atdPath := filepath.Join(absDir, ".atd")

	if _, err := os.Stat(atdPath); err == nil && !force {
		return "", fmt.Errorf(".atd already exists at %s — use --force to overwrite", atdPath)
	}

	if docsPath == "" {
		docsPath = "docs/"
	}
	if model == "" {
		model = "llama3.2"
	}

	cfg := config.ATDConfig{
		DocsPath:                docsPath,
		DiffSimilarityThreshold: 0.85,
		BloatingFactor: config.BloatingFactorConfig{
			Default: 0.8,
			TypeOverrides: map[string]float64{
				"REQUIREMENT": 0.3,
				"SPECIFICATION": 0.3,
				"MODULE":      0.3,
				"USECASE":     0.1,
				"USER_STORY":  0.1,
				"API":         0.1,
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
				"llama3.2":          {Tasks: []string{"audit_bloat", "intent_extract", "snapshot"}},
				"deepseek-r1:7b":    {Tasks: []string{"audit_code", "compare", "congruence", "reconcile", "fix_split"}},
				"qwen2.5-coder:14b": {Tasks: []string{"dissect", "recon"}},
				"nomic-embed-text":  {Tasks: []string{"embed"}},
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

	// Create docs directory if it doesn't exist.
	docsAbs := filepath.Join(absDir, docsPath)
	if _, err := os.Stat(docsAbs); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(docsAbs, 0755); mkErr != nil {
			return "", fmt.Errorf("created .atd but failed to create docs dir %s: %w", docsAbs, mkErr)
		}
		return fmt.Sprintf("Initialized ATD project.\n  Config: %s\n  Docs:   %s (created)", atdPath, docsAbs), nil
	}

	return fmt.Sprintf("Initialized ATD project.\n  Config: %s\n  Docs:   %s", atdPath, docsAbs), nil
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().String("dir", "", "Target directory (defaults to current directory)")
	initCmd.Flags().String("docs", "docs/", "Path to docs folder (relative to target dir)")
	initCmd.Flags().String("model", "llama3.2", "Default fallback model name")
	initCmd.Flags().Bool("force", false, "Overwrite an existing .atd file")
}
