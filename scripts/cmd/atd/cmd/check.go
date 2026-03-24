package cmd

import (
	"fmt"
	"sort"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/ollama"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check configuration and model availability",
	Long: `Validates the .atd configuration file, checks connectivity to LLM providers,
and verifies that required models are available for configured tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := runCheck()
		if err != nil {
			return err
		}
		fmt.Println(output)
		return nil
	},
}

type ProviderStatus struct {
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	Status    string   `json:"status"` // "Online", "Offline", "Passthrough"
	Models    []string `json:"models,omitempty"`
	Error     string   `json:"error,omitempty"`
}

type TaskResolution struct {
	Task      string   `json:"task"`
	Candidates []string `json:"candidates"`
	Resolved  string   `json:"resolved_model"`
	Provider  string   `json:"provider"`
	Status    string   `json:"status"` // "Ready", "Missing", "IDE Fallback"
}

type CheckReport struct {
	Providers []ProviderStatus `json:"providers"`
	Tasks     []TaskResolution `json:"tasks"`
}

func runCheck() (string, error) {
	report := CheckReport{}
	cfg := config.ActiveConfig.LLM

	// 1. Check Providers
	providerModels := make(map[string][]string)
	for _, p := range cfg.Providers {
		status := ProviderStatus{
			Name: p.Name,
			URL:  p.BaseURL,
		}

		if p.Type == "passthrough" {
			status.Status = "Passthrough"
		} else {
			models, err := ollama.ListModels(p.BaseURL, p.TimeoutMs)
			if err != nil {
				status.Status = "Offline"
				status.Error = err.Error()
			} else {
				status.Status = "Online"
				status.Models = models
				providerModels[p.Name] = models
			}
		}
		report.Providers = append(report.Providers, status)
	}

	// 2. Resolve Tasks
	// Collect all unique tasks from config
	taskSet := make(map[string]bool)
	for _, mc := range cfg.Models {
		for _, t := range mc.Tasks {
			taskSet[t] = true
		}
	}
	// Add mandatory tasks if not present
	taskSet["embed"] = true

	var tasks []string
	for t := range taskSet {
		tasks = append(tasks, t)
	}
	sort.Strings(tasks)

	for _, t := range tasks {
		res := TaskResolution{
			Task:       t,
			Candidates: config.ModelForTask(t),
		}

		// Simulate ResolveProvider logic
		resolved := false
		
		// Try candidates
		for _, desired := range res.Candidates {
			hasTag := strings.Contains(desired, ":")
			for _, p := range cfg.Providers {
				if p.Type == "passthrough" {
					continue
				}
				
				models := providerModels[p.Name]
				for _, m := range models {
					matched := false
					if hasTag {
						matched = (m == desired)
					} else {
						matched = (m == desired || strings.HasPrefix(m, desired+":"))
					}
					
					if matched {
						res.Resolved = m
						res.Provider = p.Name
						res.Status = "Ready"
						resolved = true
						break
					}
				}
				if resolved { break }
			}
			if resolved { break }
		}

		// Try Fallback
		if !resolved && cfg.FallbackModel != "" {
			fallback := cfg.FallbackModel
			hasTag := strings.Contains(fallback, ":")
			for _, p := range cfg.Providers {
				if p.Type == "passthrough" {
					continue
				}
				models := providerModels[p.Name]
				for _, m := range models {
					matched := false
					if hasTag {
						matched = (m == fallback)
					} else {
						matched = (m == fallback || strings.HasPrefix(m, fallback+":"))
					}
					if matched {
						res.Resolved = m
						res.Provider = p.Name
						res.Status = "Ready (Fallback)"
						resolved = true
						break
					}
				}
				if resolved { break }
			}
		}

		// IDE Fallback
		if !resolved {
			for _, p := range cfg.Providers {
				if p.Type == "passthrough" {
					res.Provider = p.Name
					res.Status = "IDE Fallback"
					resolved = true
					break
				}
			}
		}

		if !resolved {
			res.Status = "Missing"
		}

		report.Tasks = append(report.Tasks, res)
	}

	// Format output
	var sb strings.Builder
	sb.WriteString("ATD Configuration Check\n")
	sb.WriteString("========================\n\n")
	
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
	
	sb.WriteString("\nTask Resolution:\n")
	sort.Slice(report.Tasks, func(i, j int) bool {
		return report.Tasks[i].Task < report.Tasks[j].Task
	})
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

func init() {
	rootCmd.AddCommand(checkCmd)
}
