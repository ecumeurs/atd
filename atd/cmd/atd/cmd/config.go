package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"atd-tools/config"

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
}
