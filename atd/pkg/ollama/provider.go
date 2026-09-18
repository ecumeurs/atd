package ollama
// @spec-link [[service_atd_tiered_provider]]

import (
	"atd-tools/config"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Resolution holds the result of provider resolution.
type Resolution struct {
	BaseURL  string // Empty if IDE passthrough
	Model    string // Empty if IDE passthrough
	IsIDE    bool   // True if falling back to IDE agent
	Provider string // Provider name for logging
}

type providerCacheEntry struct {
	lastChecked time.Time
	isOffline   bool
	models      []string
}

var (
	cacheMu    sync.Mutex
	globalCache = make(map[string]providerCacheEntry)
)

// ResolveProvider determines which provider and model to use for a given task type.
func ResolveProvider(taskType string) (Resolution, error) {
	return ResolveProviderWithConfig(taskType, &config.ActiveConfig, false)
}

// ResolveProviderWithConfig is like ResolveProvider but accepts a config parameter.
func ResolveProviderWithConfig(taskType string, cfg *config.Config, force bool) (Resolution, error) {
	if cfg == nil {
		cfg = &config.ActiveConfig
	}
	
	llmConfig := cfg.LLM
	candidateModels := config.ModelForTask(taskType)

	type providerModels struct {
		provider config.LLMProvider
		models   []string
		offline  bool
	}
	var providersToTry []providerModels

	// 1. Gather all providers and their available models (with caching)
	for _, provider := range llmConfig.Providers {
		if provider.Type == "passthrough" {
			continue
		}

		cacheMu.Lock()
		entry, found := globalCache[provider.BaseURL]
		cacheMu.Unlock()

		var serverModels []string
		var offline bool
		var err error

		healthTTL := time.Duration(llmConfig.HealthTTLs) * time.Millisecond
		modelTTL := time.Duration(llmConfig.ModelTTLs) * time.Millisecond

		useCache := found && !force
		if useCache {
			if entry.isOffline && time.Since(entry.lastChecked) < healthTTL {
				offline = true
			} else if !entry.isOffline && time.Since(entry.lastChecked) < modelTTL {
				serverModels = entry.models
			}
		}

		if !useCache || (!offline && serverModels == nil) {
			serverModels, err = ListModels(provider.BaseURL, provider.TimeoutMs)
			
			cacheMu.Lock()
			if err != nil {
				globalCache[provider.BaseURL] = providerCacheEntry{
					lastChecked: time.Now(),
					isOffline:   true,
				}
				offline = true
			} else {
				globalCache[provider.BaseURL] = providerCacheEntry{
					lastChecked: time.Now(),
					isOffline:   false,
					models:      serverModels,
				}
			}
			cacheMu.Unlock()
		}

		providersToTry = append(providersToTry, providerModels{provider, serverModels, offline})
	}

	// 2. Try all preferred candidate models across all reachable providers
	for _, desired := range candidateModels {
		hasTag := strings.Contains(desired, ":")

		for _, pt := range providersToTry {
			if pt.offline || pt.models == nil {
				continue
			}

			for _, m := range pt.models {
				matched := false
				if hasTag {
					matched = (m == desired)
				} else {
					matched = (m == desired || strings.HasPrefix(m, desired+":"))
				}

				if matched {
					res := Resolution{
						BaseURL:  pt.provider.BaseURL,
						Model:    m,
						Provider: pt.provider.Name,
					}
					fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=%s Provider=%s\n", taskType, res.Model, res.Provider)
					return res, nil
				}
			}
		}
	}

	// 3. Try fallback
	if llmConfig.FallbackModel != "" {
		fallbackHasTag := strings.Contains(llmConfig.FallbackModel, ":")
		for _, pt := range providersToTry {
			if pt.offline || pt.models == nil {
				continue
			}

			for _, m := range pt.models {
				matched := false
				if fallbackHasTag {
					matched = (m == llmConfig.FallbackModel)
				} else {
					matched = (m == llmConfig.FallbackModel || strings.HasPrefix(m, llmConfig.FallbackModel+":"))
				}

				if matched {
					res := Resolution{
						BaseURL:  pt.provider.BaseURL,
						Model:    m,
						Provider: pt.provider.Name,
					}
					fmt.Fprintf(os.Stderr, "[LLM] Task=%s Model=%s Provider=%s (Fallback)\n", taskType, res.Model, res.Provider)
					return res, nil
				}
			}
		}
	}

	// 4. Nothing matched. If at least one provider actually answered its
	// /api/tags call, the configured model name(s) simply don't exist on
	// any provider we could reach -- almost certainly a typo or a model
	// that was never pulled, not a transient outage. That failure mode is
	// easy to mistake for "provider unreachable" once step 5 quietly
	// glides into IDE fallback, so scream about it loudly here instead.
	var reachedProviders []string
	for _, pt := range providersToTry {
		if !pt.offline {
			reachedProviders = append(reachedProviders, fmt.Sprintf("%s(models=%v)", pt.provider.Name, pt.models))
		}
	}
	if len(reachedProviders) > 0 {
		var wanted []string
		wanted = append(wanted, candidateModels...)
		if llmConfig.FallbackModel != "" {
			wanted = append(wanted, llmConfig.FallbackModel)
		}
		if len(wanted) > 0 {
			fmt.Fprintf(os.Stderr, "[LLM WARNING] Task=%s: none of the configured model(s) %v exist on any reachable provider (%s) -- check for a typo in llm.models/fallback_model or a model that was never pulled\n", taskType, wanted, strings.Join(reachedProviders, ", "))
		}
	}

	// 5. Last resort: IDE passthrough
	for _, provider := range llmConfig.Providers {
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
	return Generate(res.BaseURL, res.Model, prompt, format, nil, config.GetGenerateTimeoutMs())
}

// QueryEmbed resolves a provider for "embed" task, then calls Embed.
// Embedding has NO IDE fallback — returns error if unavailable.
func QueryEmbed(text string) ([]float32, error) {
	res, err := ResolveProvider("embed")
	if err != nil {
		return nil, err
	}
	if res.IsIDE {
		return nil, fmt.Errorf("embedding requires Ollama with nomic-embed-text — no IDE fallback available")
	}
	return Embed(res.BaseURL, res.Model, text, config.GetGenerateTimeoutMs())
}

var ErrIDEFallback = fmt.Errorf("IDE_FALLBACK")
