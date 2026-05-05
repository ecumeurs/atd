package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	DefaultAPIKey string
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) getClient(ctx context.Context, reqKey string) (*genai.Client, error) {
	apiKey := reqKey
	if apiKey == "" {
		apiKey = p.DefaultAPIKey
	}
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not configured")
	}

	return genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
}

func (p *GeminiProvider) ListModels(ctx context.Context) ([]ModelInfo, error) {
	client, err := p.getClient(ctx, "")
	if err != nil {
		return nil, err
	}

	var models []ModelInfo
	page, err := client.Models.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	for {
		for _, m := range page.Items {
			models = append(models, ModelInfo{
				ID:          m.Name,
				DisplayName: m.DisplayName,
				Description: m.Description,
				Actions:     m.SupportedActions,
				Provider:    "gemini",
			})
		}
		if page.NextPageToken == "" {
			break
		}
		page, err = page.Next(ctx)
		if err != nil {
			break
		}
	}
	return models, nil
}

func (p *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	client, err := p.getClient(ctx, req.APIKey)
	if err != nil {
		return nil, err
	}

	// Build system instruction
	systemInstruction := req.System
	if len(req.AtdContext) > 0 {
		systemInstruction += "\n\n--- CURRENT ATD CONTEXT ---\nThe following atoms are currently relevant to this conversation:\n"
		for _, atd := range req.AtdContext {
			atdJSON, _ := json.Marshal(atd)
			systemInstruction += string(atdJSON) + "\n"
		}
	}

	if len(req.Actions) > 0 {
		systemInstruction += "\n\n--- USER ACTION HISTORY ---\nThe user has taken the following actions on previous proposals:\n"
		for _, action := range req.Actions {
			systemInstruction += fmt.Sprintf("- Proposal for atom '%s': %s. %s\n", action.AtomID, action.Action, action.Summary)
		}
	}

	// Build conversation
	var contents []*genai.Content
	messagesToInclude := req.Messages
	if req.OmitHistory && len(req.Messages) > 0 {
		messagesToInclude = req.Messages[len(req.Messages)-1:]
	}

	for _, msg := range messagesToInclude {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
		})
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
		},
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"message": {Type: genai.TypeString},
				"proposals": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"action":  {Type: genai.TypeString, Enum: []string{"CREATE", "UPDATE", "DELETE"}},
							"atom_id": {Type: genai.TypeString},
							"content": {
								Type: genai.TypeObject,
								Properties: map[string]*genai.Schema{
									"human_name": {Type: genai.TypeString},
									"type": {
										Type: genai.TypeString,
										Enum: []string{
											"MODULE", "SERVICE", "ENTITY", "RULE", "MECHANIC", "DOMAIN",
											"API", "UI", "DATA", "USAGE", "BUILD", "REQUIREMENT",
											"SPECIFICATION", "USECASE", "USER_STORY",
										},
									},
									"layer": {
										Type: genai.TypeString,
										Enum: []string{"BUSINESS", "ARCHITECTURE", "IMPLEMENTATION"},
									},
									"priority":            {Type: genai.TypeString},
									"tags":                {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
									"parents":             {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
									"intent":              {Type: genai.TypeString},
									"logic":               {Type: genai.TypeString},
									"technical_interface": {Type: genai.TypeString},
									"expectation":         {Type: genai.TypeString},
								},
								Required: []string{
									"human_name", "type", "intent", "logic", "expectation",
								},
							},
							"impact_summary": {Type: genai.TypeString},
						},
						Required: []string{"action", "atom_id"},
					},
				},
			},
			Required: []string{"message"},
		},
		Temperature: genai.Ptr(float32(0.7)),
	}

	model := req.Model
	if !strings.HasPrefix(model, "models/") {
		model = "models/" + model
	}

	result, err := client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, err
	}

	responseText := result.Text()
	var resp ChatResponse
	if err := json.Unmarshal([]byte(responseText), &resp); err != nil {
		resp = ChatResponse{
			Message:   responseText,
			Proposals: []Proposal{},
		}
	}

	if result.UsageMetadata != nil {
		resp.Usage = UsageRecord{
			PromptTokens:     int(result.UsageMetadata.PromptTokenCount),
			CandidatesTokens: int(result.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(result.UsageMetadata.TotalTokenCount),
		}
	}

	return &resp, nil
}
