package llmservice

import (
	"atd-tools/config"
	"atd-tools/pkg/chat"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Service struct {
	cfg       *config.Config
	providers []chat.Provider
}

func NewService(cfg *config.Config) *Service {
	if cfg == nil {
		cfg = &config.ActiveConfig
	}
	return &Service{
		cfg:       cfg,
		providers: make([]chat.Provider, 0),
	}
}

func (s *Service) getProviders() []chat.Provider {
	if len(s.providers) > 0 {
		return s.providers
	}

	var providers []chat.Provider

	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey != "" {
		providers = append(providers, &chat.GeminiProvider{DefaultAPIKey: geminiKey})
	}

	for _, p := range s.cfg.LLM.Providers {
		if p.Type == "passthrough" {
			continue
		}

		if strings.Contains(strings.ToLower(p.Name), "minimax") || p.Type == "openai" || strings.Contains(p.BaseURL, "/v1") {
			apiKey := ""
			envVar := strings.ToUpper(p.Name) + "_API_KEY"
			apiKey = os.Getenv(envVar)

			providers = append(providers, &chat.OpenAIProvider{
				BaseURL:       p.BaseURL,
				DefaultAPIKey: apiKey,
				ProviderName:  p.Name,
			})
		}
	}

	s.providers = providers
	return providers
}

func (s *Service) GetModels(ctx context.Context, customBaseURL, customKey string) ([]chat.ModelInfo, error) {
	providers := s.getProviders()

	if customBaseURL != "" {
		customProvider := &chat.OpenAIProvider{
			BaseURL:       customBaseURL,
			DefaultAPIKey: customKey,
			ProviderName:  "custom",
		}
		providers = append([]chat.Provider{customProvider}, providers...)
	}

	var allModels []chat.ModelInfo
	for _, p := range providers {
		models, err := p.ListModels(ctx)
		if err == nil {
			allModels = append(allModels, models...)
		}
	}

	if len(allModels) == 0 {
		allModels = append(allModels, chat.ModelInfo{
			ID:          "gemini-3.1-flash-lite-preview",
			DisplayName: "Gemini 3.1 Flash Lite",
			Provider:    "gemini",
		})
	}

	return allModels, nil
}

func (s *Service) Chat(ctx context.Context, req chat.ChatRequest, headers map[string]string) (*chat.ChatResponse, error) {
	providers := s.getProviders()
	var selectedProvider chat.Provider

	customBaseURL := headers["X-LLM-Base-URL"]
	if customBaseURL != "" {
		selectedProvider = &chat.OpenAIProvider{
			BaseURL:      customBaseURL,
			ProviderName: "custom",
		}
	} else {
		for _, p := range providers {
			if strings.Contains(strings.ToLower(req.Model), strings.ToLower(p.Name())) {
				selectedProvider = p
				break
			}
		}

		if selectedProvider == nil && strings.Contains(strings.ToLower(req.Model), "gemini") {
			for _, p := range providers {
				if p.Name() == "gemini" {
					selectedProvider = p
					break
				}
			}
		}

		if selectedProvider == nil && len(providers) > 0 {
			selectedProvider = providers[0]
		}
	}

	if selectedProvider == nil {
		return nil, fmt.Errorf("no LLM provider configured or available")
	}

	headerSuffix := strings.Title(strings.ToLower(selectedProvider.Name()))
	if key := headers["X-LLM-Key-"+headerSuffix]; key != "" {
		req.APIKey = key
	} else if key := headers["X-LLM-Key"]; key != "" {
		req.APIKey = key
	}

	req.System = GetChatManifesto()

	resp, err := selectedProvider.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("%s error: %w", selectedProvider.Name(), err)
	}

	return resp, nil
}

func ExtractHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	headers["X-LLM-Base-URL"] = c.GetHeader("X-LLM-Base-URL")
	headers["X-LLM-Key"] = c.GetHeader("X-LLM-Key")
	for k, v := range c.Request.Header {
		if strings.HasPrefix(k, "X-LLM-Key-") {
			headers[k] = v[0]
		}
	}
	return headers
}