package ollama
// @spec-link [[atd_tiered_provider]]

import (
	"atd-tools/config"
	"fmt"
	"os"
	"strings"
)

// Resolution holds the result of provider resolution.
type Resolution struct {
	BaseURL  string // Empty if IDE passthrough
	Model    string // Empty if IDE passthrough
	IsIDE    bool   // True if falling back to IDE agent
	Provider string // Provider name for logging
}

// ResolveProvider determines which provider and model to use for a given task type.
func ResolveProvider(taskType string) (Resolution, error) {
	cfg := config.ActiveConfig.LLM
	candidateModels := config.ModelForTask(taskType)

	type providerModels struct {
		provider config.LLMProvider
		models   []string
	}
	var cachedProviders []providerModels

	// 1. Try all preferred candidate models across all providers
	for _, desired := range candidateModels {
		hasTag := strings.Contains(desired, ":")

		for i, provider := range cfg.Providers {
			// IDE passthrough handled separately if no Ollama matches
			if provider.Type == "passthrough" {
				continue
			}

			// Cache models per provider to avoid repeated API calls
			var serverModels []string
			if len(cachedProviders) > i {
				serverModels = cachedProviders[i].models
			} else {
				var err error
				serverModels, err = ListModels(provider.BaseURL, provider.TimeoutMs)
				// Cache even if it's nil (error) to avoid re-trying
				cachedProviders = append(cachedProviders, providerModels{provider, serverModels})
				if err != nil {
					continue
				}
			}

			if serverModels == nil {
				continue
			}

			for _, m := range serverModels {
				matched := false
				if hasTag {
					matched = (m == desired)
				} else {
					matched = (m == desired || strings.HasPrefix(m, desired+":"))
				}

				if matched {
					res := Resolution{
						BaseURL:  provider.BaseURL,
						Model:    m,
						Provider: provider.Name,
					}
					fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=%s Provider=%s\n", taskType, res.Model, res.Provider)
					return res, nil
				}
			}
		}
	}

	// 2. Try fallback if none of the preferred models were found
	if cfg.FallbackModel != "" {
		fallbackHasTag := strings.Contains(cfg.FallbackModel, ":")
		for _, cp := range cachedProviders {
			if cp.provider.Type == "passthrough" || cp.models == nil {
				continue
			}

			for _, m := range cp.models {
				matched := false
				if fallbackHasTag {
					matched = (m == cfg.FallbackModel)
				} else {
					matched = (m == cfg.FallbackModel || strings.HasPrefix(m, cfg.FallbackModel+":"))
				}

				if matched {
					res := Resolution{
						BaseURL:  cp.provider.BaseURL,
						Model:    m,
						Provider: cp.provider.Name,
					}
					fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=%s Provider=%s (Fallback)\n", taskType, res.Model, res.Provider)
					return res, nil
				}
			}
		}
	}

	// 3. Last resort: IDE passthrough
	for _, provider := range cfg.Providers {
		if provider.Type == "passthrough" {
			res := Resolution{IsIDE: true, Provider: provider.Name}
			fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=N/A Provider=%s (IDE Fallback)\n", taskType, res.Provider)
			return res, nil
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
