package llmservice

import (
	"atd-tools/config"
	"atd-tools/pkg/chat"
	"fmt"
	"os"
	"strings"
)

type ProviderType string

const (
	ProviderOllama ProviderType = "ollama"
	ProviderOpenAI ProviderType = "openai"
	ProviderGemini ProviderType = "gemini"
)

func (s *Service) ResolveProvider(t ProviderType) (chat.Provider, error) {
	for _, p := range s.providers {
		if strings.EqualFold(string(t), p.Name()) {
			return p, nil
		}
	}

	switch t {
	case ProviderOllama:
		return nil, fmt.Errorf("ollama provider not supported in llmservice, use pkg/ollama directly")
	case ProviderOpenAI:
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY not set")
		}
		return &chat.OpenAIProvider{BaseURL: "https://api.openai.com/v1", DefaultAPIKey: apiKey, ProviderName: "openai"}, nil
	case ProviderGemini:
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY not set")
		}
		return &chat.GeminiProvider{DefaultAPIKey: apiKey}, nil
	}

	return nil, fmt.Errorf("unknown provider type: %s", t)
}

func (s *Service) AddProvider(provider chat.Provider) {
	s.providers = append(s.providers, provider)
}

func GetOpenAIProviderFromConfig(p config.LLMProvider) *chat.OpenAIProvider {
	apiKey := ""
	envVar := strings.ToUpper(p.Name) + "_API_KEY"
	apiKey = os.Getenv(envVar)

	return &chat.OpenAIProvider{
		BaseURL:       p.BaseURL,
		DefaultAPIKey: apiKey,
		ProviderName:  p.Name,
	}
}