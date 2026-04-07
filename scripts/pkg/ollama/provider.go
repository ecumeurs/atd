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
	return ResolveProviderEx(taskType, false)
}

// ResolveProviderEx is like ResolveProvider but allows forcing a refresh of the cache.
func ResolveProviderEx(taskType string, force bool) (Resolution, error) {
	cfg := config.ActiveConfig.LLM
	candidateModels := config.ModelForTask(taskType)

	type providerModels struct {
		provider config.LLMProvider
		models   []string
		offline  bool
	}
	var providersToTry []providerModels

	// 1. Gather all providers and their available models (with caching)
	for _, provider := range cfg.Providers {
		if provider.Type == "passthrough" {
			continue
		}

		cacheMu.Lock()
		entry, found := globalCache[provider.BaseURL]
		cacheMu.Unlock()

		var serverModels []string
		var offline bool
		var err error

		healthTTL := time.Duration(cfg.HealthTTLMs) * time.Millisecond
		modelTTL := time.Duration(cfg.ModelTTLMs) * time.Millisecond

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
	if cfg.FallbackModel != "" {
		fallbackHasTag := strings.Contains(cfg.FallbackModel, ":")
		for _, pt := range providersToTry {
			if pt.offline || pt.models == nil {
				continue
			}

			for _, m := range pt.models {
				matched := false
				if fallbackHasTag {
					matched = (m == cfg.FallbackModel)
				} else {
					matched = (m == cfg.FallbackModel || strings.HasPrefix(m, cfg.FallbackModel+":"))
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

	// 4. Last resort: IDE passthrough
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
