package ollama
// @spec-link [[atd_tiered_provider]]

import (
	"fmt"
	"os"
	"atd-tools/config"
)

// Resolution holds the result of provider resolution.
type Resolution struct {
	BaseURL  string // Empty if IDE passthrough
	Model    string // Empty if IDE passthrough
	IsIDE    bool   // True if falling back to IDE agent
	Provider string // Provider name for logging
}

// ResolveProvider determines which provider and model to use for a given task type.
//
// Algorithm:
// 1. Look up which model handles this task (config.ModelForTask)
// 2. For each provider in priority order:
//    a. If provider.Type == "passthrough" → return IDE fallback
//    b. Call ListModels(provider.BaseURL, provider.TimeoutMs)
//    c. If desired model is in the list → return this provider + model
//    d. If desired model NOT found but fallback_model IS found → return provider + fallback
// 3. If no provider has any model → return IDE fallback
func ResolveProvider(taskType string) (Resolution, error) {
	cfg := config.ActiveConfig.LLM
	desiredModel := config.ModelForTask(taskType)

	for _, provider := range cfg.Providers {
		// IDE passthrough is always last resort
		if provider.Type == "passthrough" {
			res := Resolution{IsIDE: true, Provider: provider.Name}
			fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=N/A Provider=%s (IDE Fallback)\n", taskType, res.Provider)
			return res, nil
		}

		models, err := ListModels(provider.BaseURL, provider.TimeoutMs)
		if err != nil {
			// Provider unreachable, try next
			continue
		}

		// Check if desired model is available
		for _, m := range models {
			if m == desiredModel {
				res := Resolution{
					BaseURL:  provider.BaseURL,
					Model:    desiredModel,
					Provider: provider.Name,
				}
				fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=%s Provider=%s\n", taskType, res.Model, res.Provider)
				return res, nil
			}
		}

		// Desired model not found, try fallback
		if cfg.FallbackModel != "" && cfg.FallbackModel != desiredModel {
			for _, m := range models {
				if m == cfg.FallbackModel {
					res := Resolution{
						BaseURL:  provider.BaseURL,
						Model:    cfg.FallbackModel,
						Provider: provider.Name,
					}
					fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=%s Provider=%s (Fallback)\n", taskType, res.Model, res.Provider)
					return res, nil
				}
			}
		}
	}

	res := Resolution{IsIDE: true, Provider: "ide_agent"}
	fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=N/A Provider=%s (Default IDE Fallback)\n", taskType, res.Provider)
	return res, nil
}

// Query resolves a provider for the given task, then calls Generate.
// If IDE fallback, returns ("", ErrIDEFallback).
func Query(taskType, prompt string, format interface{}) (*GenerateResponse, error) {
	res, err := ResolveProvider(taskType)
	if err != nil {
		return nil, err
	}
	if res.IsIDE {
		return nil, ErrIDEFallback
	}
	return Generate(res.BaseURL, res.Model, prompt, format, nil)
}

// QueryEmbed resolves a provider for "embed" task, then calls Embed.
// Embedding has NO IDE fallback — returns error if unavailable.
func QueryEmbed(text string) ([]float64, error) {
	res, err := ResolveProvider("embed")
	if err != nil {
		return nil, err
	}
	if res.IsIDE {
		return nil, fmt.Errorf("embedding requires Ollama with nomic-embed-text — no IDE fallback available")
	}
	return Embed(res.BaseURL, res.Model, text)
}

var ErrIDEFallback = fmt.Errorf("IDE_FALLBACK")
