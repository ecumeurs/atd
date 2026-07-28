package ollama

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"atd-tools/config"
)

func TestResolveProvider(t *testing.T) {
	// Save original ListModels
	origListModels := ListModels
	defer func() { ListModels = origListModels }()

	savedConfig := config.Snapshot()
	defer config.Restore(savedConfig)

	// Mock ListModels
	ListModels = func(baseURL string, timeoutMs int) ([]string, error) {
		if baseURL == "http://remote" {
			// Remote has tagged versions
			return []string{"llama3.2:latest", "deepseek-r1:7b", "nomic-embed-text:latest"}, nil
		}
		if baseURL == "http://local" {
			// Local has mixed versions
			return []string{"llama3.2:3b", "qwen2.5:14b"}, nil
		}
		return nil, fmt.Errorf("offline")
	}

	// Mock config
	config.ActiveConfig.LLM = config.LLMConfig{
		Providers: []config.LLMProvider{
			{Name: "remote", BaseURL: "http://remote"},
			{Name: "local", BaseURL: "http://local"},
			{Name: "ide", Type: "passthrough"},
		},
		Models: map[string]config.ModelConfig{
			"llama3.2":       {Tasks: []string{"audit"}, Priority: 0},
			"deepseek-r1:7b": {Tasks: []string{"audit"}, Priority: 10},
			"qwen2.5":        {Tasks: []string{"audit_code"}},
			"nomic-embed-text": {Tasks: []string{"embed"}},
		},
		FallbackModel: "llama3.2",
	}

	// Case 1: Priority match in same provider
	// Both llama3.2 and deepseek-r1:7b handle "audit". 
	// deepseek-r1:7b has higher priority (10 vs 0).
	// remote has both (llama3.2:latest matches "llama3.2").
	res, err := ResolveProvider("audit")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Provider != "remote" || res.Model != "deepseek-r1:7b" {
		t.Errorf("Expected remote/deepseek-r1:7b (higher priority), got %s/%s", res.Provider, res.Model)
	}

	// Case 2: Version prefix match
	// config has "qwen2.5", server has "qwen2.5:14b". Should match.
	res, err = ResolveProvider("audit_code")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Provider != "local" || res.Model != "qwen2.5:14b" {
		t.Errorf("Expected local/qwen2.5:14b (prefix match), got %s/%s", res.Provider, res.Model)
	}

	// Case 3: Exact match with tag
	// config has "deepseek-r1:7b". If server only had "deepseek-r1:3b", it shouldn't match.
	// (Already covered by Case 1 where it matched remote/deepseek-r1:7b)

	// Case 4: No match, fallback to IDE
	config.ActiveConfig.LLM.Providers[0].BaseURL = "http://offline"
	config.ActiveConfig.LLM.Providers[1].BaseURL = "http://offline"
	res, err = ResolveProvider("audit_code")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !res.IsIDE || res.Provider != "ide" {
		t.Errorf("Expected IDE fallback, got %s/%s", res.Provider, res.Model)
	}
}

// TestResolveProviderScreamsOnNonLegitModel covers the case where a provider
// is genuinely reachable but the requested task's configured model name(s)
// don't exist in its catalog at all -- e.g. a typo in llm.models, or a
// model that was configured but never pulled. That's a config error, not a
// transient outage, and is easy to mistake for one once resolution quietly
// glides into IDE fallback -- ResolveProviderWithConfig must log a loud
// [LLM WARNING] distinguishing the two instead of failing silently.
func TestResolveProviderScreamsOnNonLegitModel(t *testing.T) {
	origListModels := ListModels
	defer func() { ListModels = origListModels }()

	savedConfig := config.Snapshot()
	defer config.Restore(savedConfig)

	ListModels = func(baseURL string, timeoutMs int) ([]string, error) {
		// Provider is online, but its catalog never contains the
		// configured model name.
		return []string{"llama3.2:latest"}, nil
	}

	config.ActiveConfig.LLM = config.LLMConfig{
		Providers: []config.LLMProvider{
			{Name: "remote", BaseURL: "http://remote"},
			{Name: "ide", Type: "passthrough"},
		},
		Models: map[string]config.ModelConfig{
			"deepseek-r1:7b": {Tasks: []string{"text_analysis"}},
		},
		FallbackModel: "also-never-pulled",
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()

	res, resolveErr := ResolveProvider("text_analysis")

	w.Close()
	os.Stderr = origStderr
	captured, _ := io.ReadAll(r)

	if resolveErr != nil {
		t.Errorf("Unexpected error: %v", resolveErr)
	}
	if !res.IsIDE {
		t.Errorf("Expected IDE fallback when no reachable provider has the model, got %+v", res)
	}
	if !strings.Contains(string(captured), "[LLM WARNING]") {
		t.Errorf("Expected a loud [LLM WARNING] about the missing model, got stderr: %s", captured)
	}
	if !strings.Contains(string(captured), "deepseek-r1:7b") || !strings.Contains(string(captured), "also-never-pulled") {
		t.Errorf("Expected the warning to name the missing candidate and fallback models, got stderr: %s", captured)
	}
}

func TestQueryEmbed(t *testing.T) {
	// Save original ListModels
	origListModels := ListModels
	defer func() { ListModels = origListModels }()

	savedConfig := config.Snapshot()
	defer config.Restore(savedConfig)

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
