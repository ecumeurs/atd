package ollama

import (
	"fmt"
	"testing"
	"atd-tools/config"
)

func TestResolveProvider(t *testing.T) {
	// Save original ListModels
	origListModels := ListModels
	defer func() { ListModels = origListModels }()

	// Mock ListModels
	ListModels = func(baseURL string, timeoutMs int) ([]string, error) {
		if baseURL == "http://remote" {
			return []string{"llama3.1", "nomic-embed-text"}, nil
		}
		if baseURL == "http://local" {
			return []string{"llama3.2", "qwen2.5"}, nil
		}
		if baseURL == "http://offline" {
			return nil, fmt.Errorf("offline")
		}
		return nil, nil
	}

	// Mock config
	config.ActiveConfig.LLM = config.LLMConfig{
		Providers: []config.LLMProvider{
			{Name: "remote", BaseURL: "http://remote"},
			{Name: "local", BaseURL: "http://local"},
			{Name: "ide", Type: "passthrough"},
		},
		Models: map[string]config.ModelConfig{
			"llama3.1": {Tasks: []string{"dissect"}},
			"nomic-embed-text": {Tasks: []string{"embed"}},
			"llama3.2": {Tasks: []string{"audit"}},
		},
		FallbackModel: "llama3.2",
	}

	// Case 1: Remote has desired model
	res, err := ResolveProvider("dissect")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Provider != "remote" || res.Model != "llama3.1" {
		t.Errorf("Expected remote/llama3.1, got %s/%s", res.Provider, res.Model)
	}

	// Case 2: Only local has desired model
	res, err = ResolveProvider("audit")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Provider != "local" || res.Model != "llama3.2" {
		t.Errorf("Expected local/llama3.2, got %s/%s", res.Provider, res.Model)
	}

	// Case 3: Remote unreachable, Local has desired model
	config.ActiveConfig.LLM.Providers[0].BaseURL = "http://offline"
	res, err = ResolveProvider("audit")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Provider != "local" || res.Model != "llama3.2" {
		t.Errorf("Expected local/llama3.2, got %s/%s", res.Provider, res.Model)
	}

	// Case 4: Remote offline, Local doesn't have model, Fallback to local fallback
	config.ActiveConfig.LLM.Models["nonexistent"] = config.ModelConfig{Tasks: []string{"unknown"}}
	res, err = ResolveProvider("unknown")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Provider != "local" || res.Model != "llama3.2" {
		t.Errorf("Expected local/llama3.2 (fallback), got %s/%s", res.Provider, res.Model)
	}

	// Case 5: Both offline, Fallback to IDE
	config.ActiveConfig.LLM.Providers[1].BaseURL = "http://offline"
	res, err = ResolveProvider("dissect")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !res.IsIDE || res.Provider != "ide" {
		t.Errorf("Expected IDE fallback, got %s/%s", res.Provider, res.Model)
	}
}

func TestQueryEmbed(t *testing.T) {
	// Save original ListModels
	origListModels := ListModels
	defer func() { ListModels = origListModels }()

	// Case: Remote offline, Local offline, IDE only
	ListModels = func(baseURL string, timeoutMs int) ([]string, error) {
		return nil, fmt.Errorf("offline")
	}
	config.ActiveConfig.LLM.Providers = []config.LLMProvider{
		{Name: "remote", BaseURL: "http://offline"},
		{Name: "ide", Type: "passthrough"},
	}

	_, err := QueryEmbed("hello")
	if err == nil {
		t.Error("Expected error for QueryEmbed with IDE fallback, got nil")
	}
}
