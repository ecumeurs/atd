package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"atd-tools/config"
	"atd-tools/pkg/ollama"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage .atd configuration",
	Long:  `View or modify the .atd configuration file.`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current configuration as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := json.MarshalIndent(config.ActiveConfig, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	},
}

var bloatingFactorCmd = &cobra.Command{
	Use:   "bloating-factor [atom-type]",
	Short: "Get the bloating factor for a specific atom type",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		val, err := runConfigGetBloating(args[0])
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	},
}

// runConfigUpdate updates a specific field in the config.
// For now, it supports 'set-task-model' specifically for the user's need.
func runConfigUpdate(task, model string) (string, error) {
	// Re-load to get the path
	err := config.Load()
	if err != nil {
		return "", err
	}

	// P-1 scope guard (test_atd_07_26.md §3.6, §8.3 defect #7): config.Load()
	// above just re-resolved the active config from the process's cwd,
	// which is not necessarily the project this MCP server (or CLI
	// invocation) is meant to be operating on. If no real .atd was found
	// anywhere above cwd, ProjectRoot() is just the cwd fallback, not a
	// genuine project boundary -- writing a task/model assignment there
	// would create or clobber a ".atd" wherever the process happened to be
	// invoked from, exactly the write-side hazard incident I-1 (§2.1)
	// demonstrated for rename propagation. Refuse loudly instead of
	// guessing.
	if config.LoadedFromFallback() {
		return "", fmt.Errorf("atd: refusing to write task/model assignment (task=%q, model=%q): active config is cwd-fallback-anchored (no .atd found above the current directory), not a genuine project root; run atd from a directory containing a real .atd file, or switch to the intended project first (e.g. atd_workspace_use), before reassigning task models", task, model)
	}

	root := config.ProjectRoot()
	if root == "" {
		return "", fmt.Errorf("could not find .atd file to update")
	}
	configPath := filepath.Join(root, ".atd")

	// Update the local struct
	if config.ActiveConfig.LLM.Models == nil {
		config.ActiveConfig.LLM.Models = make(map[string]config.ModelConfig)
	}

	// Remove task from other models
	for m, mc := range config.ActiveConfig.LLM.Models {
		var newTasks []string
		for _, t := range mc.Tasks {
			if t != task {
				newTasks = append(newTasks, t)
			}
		}
		mc.Tasks = newTasks
		config.ActiveConfig.LLM.Models[m] = mc
	}

	// Add task to target model
	mc := config.ActiveConfig.LLM.Models[model]
	mc.Tasks = append(mc.Tasks, task)
	config.ActiveConfig.LLM.Models[model] = mc

	// Write back
	data, err := json.MarshalIndent(config.ActiveConfig, "", "  ")
	if err != nil {
		return "", err
	}
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated task '%s' to use model '%s'", task, model), nil
}

func runConfigGetBloating(atomType string) (string, error) {
	err := config.Load()
	if err != nil {
		return "", err
	}

	val := config.GetBloatingStrictness(atomType)
	return fmt.Sprintf("%f", val), nil
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(bloatingFactorCmd)
	configCmd.AddCommand(configModelCmd)
	configModelCmd.Flags().BoolP("force", "f", false, "Force re-probing of all providers, bypassing health cache")
}

var configModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Check provider connectivity and model availability",
	Long: `Validates the .atd configuration file, checks connectivity to LLM providers,
and verifies that required models are available for configured tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		output, err := runCheck(force)
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	},
}

type ProviderStatus struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Status string   `json:"status"` // "Online", "Offline", "Passthrough"
	Models []string `json:"models,omitempty"`
	Error  string   `json:"error,omitempty"`
}

type TaskResolution struct {
	Task       string   `json:"task"`
	Candidates []string `json:"candidates"`
	Resolved   string   `json:"resolved_model"`
	Provider   string   `json:"provider"`
	Status     string   `json:"status"` // "Ready", "Missing", "IDE Fallback"
}

type EnvReport struct {
	Providers []ProviderStatus `json:"providers"`
	Tasks     []TaskResolution `json:"tasks"`
}

func runCheck(force bool) (string, error) {
	report := EnvReport{}
	cfg := config.ActiveConfig.LLM

	// 1. Check Providers
	for _, p := range cfg.Providers {
		status := ProviderStatus{
			Name: p.Name,
			URL:  p.BaseURL,
		}

		if p.Type == "passthrough" {
			status.Status = "Passthrough"
		} else {
			// Bypass cache if force is true, otherwise ListModels is direct but we could check ResolveProviderEx later
			models, err := ollama.ListModels(p.BaseURL, p.TimeoutMs)
			if err != nil {
				status.Status = "Offline"
				status.Error = err.Error()
			} else {
				status.Status = "Online"
				status.Models = models
			}
		}
		report.Providers = append(report.Providers, status)
	}

	// 2. Resolve Tasks
	// Simplified categories
	tasks := []string{"code_analysis", "text_analysis", "text_generation", "embedding"}

	for _, t := range tasks {
		res := TaskResolution{
			Task:       t,
			Candidates: config.ModelForTask(t),
		}

		// Use the actual resolver logic to ensure consistency and cache usage
		resolvedRes, _ := ollama.ResolveProviderWithConfig(t, nil, force)

		if resolvedRes.Provider != "" {
			res.Resolved = resolvedRes.Model
			res.Provider = resolvedRes.Provider
			if strings.Contains(resolvedRes.Model, "Fallback") || resolvedRes.IsIDE {
				res.Status = "Ready (Fallback)"
				if resolvedRes.IsIDE {
					res.Status = "IDE Fallback"
				}
			} else {
				res.Status = "Ready"
			}
		} else {
			res.Status = "Missing"
		}

		report.Tasks = append(report.Tasks, res)
	}

	// Format output
	var sb strings.Builder
	sb.WriteString("ATD Model Configuration & Connectivity\n")
	sb.WriteString("======================================\n\n")

	sb.WriteString("Providers:\n")
	for _, p := range report.Providers {
		indicator := "[ ]"
		if p.Status == "Online" {
			indicator = "[✓]"
		} else if p.Status == "Offline" {
			indicator = "[✗]"
		} else {
			indicator = "[-]"
		}
		sb.WriteString(fmt.Sprintf("%s %-12s %-30s %s\n", indicator, p.Name, p.URL, p.Status))
		if p.Error != "" {
			sb.WriteString(fmt.Sprintf("    Error: %s\n", p.Error))
		}
		if len(p.Models) > 0 {
			sb.WriteString(fmt.Sprintf("    Available Models: %s\n", strings.Join(p.Models, ", ")))
		}
	}

	sb.WriteString("\nTask Resolution (Simplified):\n")
	for _, t := range report.Tasks {
		statusIcon := "[✓]"
		if t.Status == "Missing" {
			statusIcon = "[✗]"
		} else if strings.Contains(t.Status, "Fallback") {
			statusIcon = "[!]"
		}

		modelInfo := t.Resolved
		if modelInfo == "" {
			modelInfo = "N/A"
		}

		sb.WriteString(fmt.Sprintf("%s %-15s -> %-20s (Provider: %-10s) %s\n",
			statusIcon, t.Task, modelInfo, t.Provider, t.Status))
	}

	return sb.String(), nil
}
