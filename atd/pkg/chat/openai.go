package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAIProvider struct {
	BaseURL       string
	DefaultAPIKey string
	ProviderName  string // e.g. "minimax", "openai"
}

func (p *OpenAIProvider) Name() string {
	if p.ProviderName != "" {
		return p.ProviderName
	}
	return "openai"
}

type openAIChatRequest struct {
	Model          string           `json:"model"`
	Messages       []openAIMessage  `json:"messages"`
	ResponseFormat *responseFormat  `json:"response_format,omitempty"`
	Temperature    float64          `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	// For simplicity, we might not be able to list models for all providers.
	// But we can try the /v1/models endpoint.
	
	// Fallback models for known providers
	fallbacks := map[string][]ModelInfo{
		"minimax": {
			{ID: "abab6.5g-chat", DisplayName: "Minimax abab6.5g", Provider: "minimax"},
			{ID: "abab6.5s-chat", DisplayName: "Minimax abab6.5s", Provider: "minimax"},
		},
		"openai": {
			{ID: "gpt-4o", DisplayName: "GPT-4o", Provider: "openai"},
			{ID: "gpt-4o-mini", DisplayName: "GPT-4o Mini", Provider: "openai"},
		},
	}

	url := fmt.Sprintf("%s/models", strings.TrimSuffix(p.BaseURL, "/"))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		if f, ok := fallbacks[p.Name()]; ok {
			return f, nil
		}
		return nil, err
	}

	apiKey := p.DefaultAPIKey
	if apiKey == "" {
		if f, ok := fallbacks[p.Name()]; ok {
			return f, nil
		}
		return nil, fmt.Errorf("API key not configured for %s", p.Name())
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if f, ok := fallbacks[p.Name()]; ok {
			return f, nil
		}
		return nil, fmt.Errorf("failed to list models from %s: status %d", p.Name(), resp.StatusCode)
	}

	var data struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var models []ModelInfo
	for _, m := range data.Data {
		models = append(models, ModelInfo{
			ID:          m.ID,
			DisplayName: m.ID,
			Provider:    p.Name(),
		})
	}
	return models, nil
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	apiKey := req.APIKey
	if apiKey == "" {
		apiKey = p.DefaultAPIKey
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key not configured for %s", p.Name())
	}

	// Build messages
	messages := []openAIMessage{}
	
	// Add system prompt
	systemMsg := req.System
	if len(req.AtdContext) > 0 {
		systemMsg += "\n\n--- CURRENT ATD CONTEXT ---\n"
		for _, atd := range req.AtdContext {
			atdJSON, _ := json.Marshal(atd)
			systemMsg += string(atdJSON) + "\n"
		}
	}
	if len(req.Actions) > 0 {
		systemMsg += "\n\n--- USER ACTION HISTORY ---\n"
		for _, action := range req.Actions {
			systemMsg += fmt.Sprintf("- Proposal for atom '%s': %s. %s\n", action.AtomID, action.Action, action.Summary)
		}
	}
	
	messages = append(messages, openAIMessage{Role: "system", Content: systemMsg})

	// Add history
	messagesToInclude := req.Messages
	if req.OmitHistory && len(req.Messages) > 0 {
		messagesToInclude = req.Messages[len(req.Messages)-1:]
	}
	for _, m := range messagesToInclude {
		role := m.Role
		if role == "model" {
			role = "assistant"
		}
		messages = append(messages, openAIMessage{Role: role, Content: m.Content})
	}

	oaReq := openAIChatRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: 0.7,
	}

	// Some providers might support json_object
	// For Minimax, we should check if they support it. 
	// For now, we rely on the system prompt instruction.
	// oaReq.ResponseFormat = &responseFormat{Type: "json_object"}

	body, err := json.Marshal(oaReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(p.BaseURL, "/"))
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider %s error (status %d): %s", p.Name(), resp.StatusCode, string(raw))
	}

	var oaResp openAIChatResponse
	if err := json.Unmarshal(raw, &oaResp); err != nil {
		return nil, err
	}

	if oaResp.Error != nil {
		return nil, fmt.Errorf("provider %s error: %s", p.Name(), oaResp.Error.Message)
	}

	if len(oaResp.Choices) == 0 {
		return nil, fmt.Errorf("provider %s returned no choices", p.Name())
	}

	responseText := oaResp.Choices[0].Message.Content
	
	// Clean up markdown code blocks if present (common in non-schema responses)
	cleanText := responseText
	if strings.Contains(cleanText, "```json") {
		parts := strings.Split(cleanText, "```json")
		if len(parts) > 1 {
			cleanText = strings.Split(parts[1], "```")[0]
		}
	} else if strings.Contains(cleanText, "```") {
		parts := strings.Split(cleanText, "```")
		if len(parts) > 1 {
			cleanText = parts[1]
		}
	}
	cleanText = strings.TrimSpace(cleanText)

	var chatResp ChatResponse
	if err := json.Unmarshal([]byte(cleanText), &chatResp); err != nil {
		// If parsing fails, treat the whole thing as a message
		chatResp = ChatResponse{
			Message:   responseText,
			Proposals: []Proposal{},
		}
	}

	chatResp.Usage = UsageRecord{
		PromptTokens:     oaResp.Usage.PromptTokens,
		CandidatesTokens: oaResp.Usage.CompletionTokens,
		TotalTokens:      oaResp.Usage.TotalTokens,
	}

	return &chatResp, nil
}
