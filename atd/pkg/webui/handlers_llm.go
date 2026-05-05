package webui

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/chat"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// @spec-link [[mechanic_webui_llm_chat_orchestration]]
func (s *Server) registerLLMRoutes(api *gin.RouterGroup) {
	// New generalized routes
	api.POST("/llm/chat", s.handleLLMChat)
	api.GET("/llm/models", s.handleLLMModels)
	api.GET("/llm/atoms", s.handleLLMAtoms)
	api.GET("/llm/atom/:id", s.handleLLMAtomDetail)
	api.POST("/llm/apply-proposal", s.handleApplyProposal)

	// Legacy routes for backward compatibility with frontend
	api.POST("/gemini/chat", s.handleLLMChat)
	api.GET("/gemini/models", s.handleLLMModels)
	api.GET("/gemini/atoms", s.handleLLMAtoms)
	api.GET("/gemini/atom/:id", s.handleLLMAtomDetail)
	api.POST("/gemini/apply-proposal", s.handleApplyProposal)
}

func (s *Server) getProviders() []chat.Provider {
	var providers []chat.Provider

	// 1. Add Gemini if configured
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey != "" {
		providers = append(providers, &chat.GeminiProvider{DefaultAPIKey: geminiKey})
	}

	// 2. Add other providers from config
	for _, p := range config.ActiveConfig.LLM.Providers {
		if p.Type == "passthrough" {
			continue
		}

		// If it's an OpenAI-compatible provider (or explicitly named minimax)
		if strings.Contains(strings.ToLower(p.Name), "minimax") || p.Type == "openai" || strings.Contains(p.BaseURL, "/v1") {
			apiKey := ""
			// Determine env var name (default to PROVIDER_NAME_API_KEY)
			envVar := strings.ToUpper(p.Name) + "_API_KEY"
			apiKey = os.Getenv(envVar)

			providers = append(providers, &chat.OpenAIProvider{
				BaseURL:       p.BaseURL,
				DefaultAPIKey: apiKey,
				ProviderName:  p.Name,
			})
		}
	}

	return providers
}

func (s *Server) handleLLMModels(c *gin.Context) {
	providers := s.getProviders()

	// Add custom provider if headers are present
	customBaseURL := c.GetHeader("X-LLM-Base-URL")
	if customBaseURL != "" {
		customProvider := &chat.OpenAIProvider{
			BaseURL:       customBaseURL,
			DefaultAPIKey: c.GetHeader("X-LLM-Key"),
			ProviderName:  "custom",
		}
		providers = append([]chat.Provider{customProvider}, providers...)
	}

	var allModels []chat.ModelInfo

	ctx := context.Background()
	for _, p := range providers {
		models, err := p.ListModels(ctx)
		if err == nil {
			allModels = append(allModels, models...)
		}
	}

	// If no models found, return a default list or an error
	if len(allModels) == 0 {
		allModels = append(allModels, chat.ModelInfo{
			ID:          "gemini-3.1-flash-lite-preview",
			DisplayName: "Gemini 3.1 Flash Lite",
			Provider:    "gemini",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"models":  allModels,
		"default": "models/gemini-3.1-flash-lite-preview",
	})
}

func (s *Server) handleLLMChat(c *gin.Context) {
	var req chat.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 1. Resolve Provider
	providers := s.getProviders()
	var selectedProvider chat.Provider

	// Check for custom OpenAI base URL (frontend override)
	customBaseURL := c.GetHeader("X-LLM-Base-URL")
	if customBaseURL != "" {
		selectedProvider = &chat.OpenAIProvider{
			BaseURL:      customBaseURL,
			ProviderName: "custom",
		}
	} else {
		// If the model ID has a prefix or if we can match it to a provider
		for _, p := range providers {
			if strings.Contains(strings.ToLower(req.Model), strings.ToLower(p.Name())) {
				selectedProvider = p
				break
			}
		}

		// Fallback to Gemini if no provider matched and model is gemini
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No LLM provider configured or available"})
		return
	}

	// 2. Inject API Key from headers if present (BYOK)
	// Header format: X-LLM-Key-ProviderName (e.g. X-LLM-Key-Minimax)
	headerSuffix := strings.Title(strings.ToLower(selectedProvider.Name()))
	if key := c.GetHeader("X-LLM-Key-" + headerSuffix); key != "" {
		req.APIKey = key
	} else if key := c.GetHeader("X-LLM-Key"); key != "" {
		req.APIKey = key
	}

	// 3. Inject system prompt
	req.System = chatManifesto

	// 4. Call Provider
	resp, err := selectedProvider.Chat(context.Background(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "limit") {
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{"error": fmt.Sprintf("%s error: %v", selectedProvider.Name(), err)})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleLLMAtoms(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var results []map[string]interface{}
	if s.explorer.Graph != nil {
		for _, a := range s.explorer.Graph.Atoms {
			if query == "" || strings.Contains(strings.ToLower(a.ID), query) || strings.Contains(strings.ToLower(a.HumanName), query) {
				results = append(results, map[string]interface{}{
					"id":         a.ID,
					"human_name": a.HumanName,
					"type":       a.Type,
					"layer":      a.Layer,
					"status":     a.Status,
					"intent":     a.Intent,
				})
			}
		}
	}
	c.JSON(http.StatusOK, results)
}

func (s *Server) handleLLMAtomDetail(c *gin.Context) {
	id := c.Param("id")
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.explorer.Graph == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Graph not loaded"})
		return
	}

	a, exists := s.explorer.Graph.Atoms[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

// @spec-link [[mechanic_atd_update]]
func (s *Server) handleApplyProposal(c *gin.Context) {
	var proposal struct {
		Action  string                 `json:"action"`
		AtomID  string                 `json:"atom_id"`
		Content map[string]interface{} `json:"content"`
	}
	if err := c.ShouldBindJSON(&proposal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	opts := atom.UpdateOptions{
		FilePath: filepath.Join(config.DocsDir(), proposal.AtomID+".atom.md"),
	}

	if proposal.Action == "UPDATE" || proposal.Action == "DELETE" {
		s.mutex.RLock()
		node, ok := s.explorer.Graph.Atoms[proposal.AtomID]
		if ok {
			opts.FilePath = node.FilePath
		}
		s.mutex.RUnlock()

		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found: " + proposal.AtomID})
			return
		}
	}

	if proposal.Action == "DELETE" {
		if err := os.Remove(opts.FilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete atom file", "details": err.Error()})
			return
		}
	} else {
		// CREATE or UPDATE
		if proposal.Action == "CREATE" {
			opts.SetArgs = append(opts.SetArgs, "id="+proposal.AtomID)
		}

		if v, ok := proposal.Content["human_name"].(string); ok && v != "" {
			opts.SetArgs = append(opts.SetArgs, "human_name="+v)
		}
		if v, ok := proposal.Content["type"].(string); ok && v != "" {
			opts.SetArgs = append(opts.SetArgs, "type="+v)
		}
		if v, ok := proposal.Content["layer"].(string); ok && v != "" {
			opts.SetArgs = append(opts.SetArgs, "layer="+v)
		}
		if v, ok := proposal.Content["status"].(string); ok && v != "" {
			opts.SetArgs = append(opts.SetArgs, "status="+v)
		}
		if v, ok := proposal.Content["priority"].(string); ok && v != "" {
			opts.SetArgs = append(opts.SetArgs, "priority="+v)
		}
		if v, ok := proposal.Content["intent"].(string); ok && v != "" {
			opts.Intent = v
		}
		if v, ok := proposal.Content["logic"].(string); ok && v != "" {
			opts.Logic = v
		}
		if v, ok := proposal.Content["technical_interface"].(string); ok && v != "" {
			opts.Interface = v
		}
		if v, ok := proposal.Content["expectation"].(string); ok && v != "" {
			opts.Expectation = v
		}

		_, err := atom.Update(opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply proposal", "details": err.Error()})
			return
		}
	}

	s.refreshAtoms()
	c.JSON(http.StatusOK, gin.H{"message": "Proposal applied successfully", "atom_id": proposal.AtomID})
}

const chatManifesto = `You are an ATD (Atomic Traceable Documentation) Specification Architect.

RULES YOU MUST FOLLOW:
0. You are a sounding board for the user. You are not here to replace the user's judgement, but to help them make better decisions. You may challenge the user's assumptions and propose alternative solutions. You are not expected to provide new/update ATD at every message. You may ask for clarifications.
0.1. Your goal is also to find Underspecified Boundaries. When a user proposes an atom, look for the 'Inverse Rule' (e.g., If they define 'Login Success,' ask where the 'Account Locked' rule is). Do not just confirm their input; hunt for the missing logic that an agent would fail to guess.
1. Every atom has EXACTLY ONE state-changing rule. If an intent needs "and" or "also", split into multiple atoms.
2. The hierarchy is divided into 3 layers: BUSINESS (requirements, rules; global imperatives) -> ARCHITECTURE (modules, APIs, entities; system organization) -> IMPLEMENTATION (mechanics; technical execution).
3. Valid types: REQUIREMENT, RULE, USER_STORY, API, UI, ENTITY, MECHANIC, MODULE, DOMAIN.
4. Each atom has 4 sections: intent, logic, technical_interface, expectation.
5. The intent must be ONE sentence, no "and" or "also".

RESPONSE FORMAT:
You MUST respond with valid JSON matching this schema:
{
  "message": "Your conversational response explaining your reasoning",
  "proposals": [
    {
      "action": "CREATE" | "UPDATE" | "DELETE",
      "atom_id": "type_snake_case_name",
      "content": {
        "human_name": "Human Readable Name",
        "type": "MECHANIC",
        "layer": "IMPLEMENTATION",
        "intent": "Single sentence why this exists.",
        "logic": "The core specification.",
        "technical_interface": "API endpoints, code tags, test names.",
        "expectation": "Verifiable acceptance criteria."
        "tags": ["tag1"],
        "parents": ["parent_atom_id"],
      },
      "impact_summary": "Brief description of what this change means."
    }
  ]
}

TYPES:
Atoms are grouped into **11 consolidated types** across three functional families. The **Bloat Factor** column maps to the default "bloating_factor" per type in ".atd" config (1.0 = strictest, 0.1 = most relaxed).

| Type | Family | Typical Layer | Bloat Factor | Granularity |
|---|---|---|---|---|
| "CONTRACT" | Governance | BUSINESS | 0.1 | **Unique**; project-wide mandatory rules |
| "VISION" | Governance | BUSINESS | 0.1 | **Unique**; project-wide scope/philosophy |
| "REQUIREMENT" | Requirements | BUSINESS | 0.3 | High-level external contract or constraint |
| "USER_STORY" | Requirements | BUSINESS | 0.1 | User-facing workflow (synonym: "USECASE", "WORKFLOW") |
| "RULE" | Logic | BUSINESS / ARCHITECTURE | 0.8 | Single business constraint or boolean check |
| "DOMAIN" | Logic | BUSINESS | 0.8 | Narrative-driven context: "The Why" |
| "MECHANIC" | Logic | IMPLEMENTATION | 0.8 | One algorithm or procedural step |
| "MODULE" | Architectural | ARCHITECTURE | 0.3 | High-level grouping; broad scope is acceptable |
| "ENTITY" | Architectural | ARCHITECTURE | 0.8 | Single data structure or state model |
| "API" | Interface | ARCHITECTURE | 0.1 | One contract per atom; include sample payloads |
| "UI" | Interface | ARCHITECTURE | 0.8 | One screen or interaction flow |


When the conversation is exploratory or you need clarification, return "proposals": []. Only propose atoms when you have sufficient information and the user's intent is clear.
When no proposals are needed (e.g. answering a question, challenging an assumption, or asking for more details), return an empty proposals array.
Always explain your reasoning in "message", and then you may propose either new related ideas or remarks regarding the current topic.
`
