package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

// --- Gemini Chat Types ---

type ChatMessage struct {
	Role    string `json:"role"` // "user" or "model"
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []ChatMessage            `json:"messages"`     // Full conversation history
	Model       string                   `json:"model"`        // Selected Gemini model
	AtdContext  []map[string]interface{} `json:"atd_context"`  // ATD atoms to inject as context
	Actions     []ActionRecord           `json:"actions"`      // Accept/reject history
	OmitHistory bool                     `json:"omit_history"` // If true, only send the latest message
}

type ActionRecord struct {
	ProposalID string `json:"proposal_id"`
	AtomID     string `json:"atom_id"`
	Action     string `json:"action"` // "ACCEPTED" or "REJECTED"
	Summary    string `json:"summary"`
}

type GeminiProposal struct {
	Action        string                 `json:"action"` // CREATE, UPDATE, DELETE
	AtomID        string                 `json:"atom_id"`
	Content       map[string]interface{} `json:"content"`
	ImpactSummary string                 `json:"impact_summary"`
}

type UsageRecord struct {
	PromptTokens     int `json:"prompt_tokens"`
	CandidatesTokens int `json:"candidates_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type GeminiResponse struct {
	Message   string           `json:"message"`
	Proposals []GeminiProposal `json:"proposals"`
	Usage     UsageRecord      `json:"usage"`
}

// @spec-link [[mechanic_webui_gemini_chat_orchestration]]
const atdManifesto = `You are an ATD (Atomic Traceable Documentation) Specification Architect.

RULES YOU MUST FOLLOW:
0. You are a sounding board for the user. You are not here to replace the user's judgement, but to help them make better decisions. You may challenge the user's assumptions and propose alternative solutions. You are not expected to provide new/update ATD at every message. You may ask for clarifications.
1. Every atom has EXACTLY ONE state-changing rule. If an intent needs "and" or "also", split into multiple atoms.
2. Atoms have strict YAML frontmatter: id, human_name, type, layer, version, status, priority, tags, parents, dependents.
3. The hierarchy is: CUSTOMER (requirements, usecases) -> ARCHITECTURE (modules, APIs) -> IMPLEMENTATION (mechanics, builds).
4. Valid types: MODULE, SERVICE, ENTITY, RULE, MECHANIC, DOMAIN, API, UI, DATA, USAGE, BUILD, REQUIREMENT, SPECIFICATION, USECASE, USER_STORY.
5. Valid layers: CUSTOMER, ARCHITECTURE, IMPLEMENTATION.
6. Each atom has 4 mandatory sections: intent, logic, technical_interface, expectation.
7. The intent must be ONE sentence, no "and" or "also".
8. Parents link upward (impl -> arch -> customer). Dependents link downward.

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

When the conversation is exploratory or you need clarification, return "proposals": []. Only propose atoms when you have sufficient information and the user's intent is clear.
When proposing updates, only include fields that change in "content".
When no proposals are needed (e.g. answering a question, challenging an assumption, or asking for more details), return an empty proposals array.
Always explain your reasoning in "message" before listing proposals.
Always provide the content for intent and logic. Expectation should be provided if you have enough information to define it.
Your goal is also to find Underspecified Boundaries. When a user proposes an atom, look for the 'Inverse Rule' (e.g., If they define 'Login Success,' ask where the 'Account Locked' rule is). Do not just confirm their input; hunt for the missing logic that an agent would fail to guess.
`

// registerGeminiRoutes registers all Gemini-related API endpoints.
func registerGeminiRoutes(api *gin.RouterGroup) {
	api.POST("/gemini/chat", handleGeminiChat)
	api.GET("/gemini/models", handleGeminiModels)
	api.GET("/gemini/atoms", handleGeminiAtoms)
	api.GET("/gemini/atom/:id", handleGeminiAtomDetail)
	api.POST("/gemini/apply-proposal", handleApplyProposal)
}

// @spec-link [[mechanic_webui_gemini_model_list]]
func handleGeminiModels(c *gin.Context) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY not configured"})
		return
	}
	os.Setenv("GOOGLE_API_KEY", apiKey)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create client: " + err.Error()})
		return
	}

	var models []map[string]interface{}
	page, err := client.Models.List(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list models: " + err.Error()})
		return
	}
	for {
		for _, m := range page.Items {
			models = append(models, map[string]interface{}{
				"id":           m.Name,
				"display_name": m.DisplayName,
				"description":  m.Description,
				"actions":      m.SupportedActions,
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

	c.JSON(http.StatusOK, gin.H{
		"models":  models,
		"default": "models/gemini-3.1-flash-lite-preview",
	})
}

// @spec-link [[mechanic_webui_gemini_proxy]]
func handleGeminiAtoms(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	var results []map[string]interface{}
	for _, atom := range Atoms {
		if query == "" || strings.Contains(strings.ToLower(atom.ID), query) || strings.Contains(strings.ToLower(atom.HumanName), query) {
			results = append(results, map[string]interface{}{
				"id":         atom.ID,
				"human_name": atom.HumanName,
				"type":       atom.Type,
				"layer":      atom.Layer,
				"status":     atom.Status,
				"intent":     extractIntent(atom.Content),
			})
		}
	}
	c.JSON(http.StatusOK, results)
}

// @spec-link [[mechanic_webui_gemini_proxy]]
func handleGeminiAtomDetail(c *gin.Context) {
	id := c.Param("id")
	atom, exists := Atoms[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         atom.ID,
		"human_name": atom.HumanName,
		"type":       atom.Type,
		"status":     atom.Status,
		"content":    atom.Content,
		"parents":    atom.Parents,
		"dependents": atom.Dependents,
	})
}

// @spec-link [[mechanic_atd_update]]
func handleApplyProposal(c *gin.Context) {
	var proposal struct {
		Action  string                 `json:"action"`
		AtomID  string                 `json:"atom_id"`
		Content map[string]interface{} `json:"content"`
	}
	if err := c.ShouldBindJSON(&proposal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	toolPath := expandPath(AppConfig.ToolkitPath) + "/atd"
	atdDir := AppConfig.ProjectPath + "/" + AppConfig.ATDPath

	switch proposal.Action {
	case "CREATE":
		filePath := atdDir + "/" + proposal.AtomID + ".atom.md"
		args := []string{"update", "--file", filePath, "--set", "id=" + proposal.AtomID}

		if v, ok := proposal.Content["human_name"].(string); ok && v != "" {
			args = append(args, "--set", "human_name="+v)
		}
		if v, ok := proposal.Content["type"].(string); ok && v != "" {
			args = append(args, "--set", "type="+v)
		}
		if v, ok := proposal.Content["layer"].(string); ok && v != "" {
			args = append(args, "--set", "layer="+v)
		}
		if v, ok := proposal.Content["status"].(string); ok && v != "" {
			args = append(args, "--set", "status="+v)
		}
		if v, ok := proposal.Content["priority"].(string); ok && v != "" {
			args = append(args, "--set", "priority="+v)
		}
		if v, ok := proposal.Content["intent"].(string); ok && v != "" {
			args = append(args, "--intent", v)
		}
		if v, ok := proposal.Content["logic"].(string); ok && v != "" {
			args = append(args, "--logic", v)
		}
		if v, ok := proposal.Content["technical_interface"].(string); ok && v != "" {
			args = append(args, "--interface", v)
		}
		if v, ok := proposal.Content["expectation"].(string); ok && v != "" {
			args = append(args, "--expectation", v)
		}

		cmd := exec.Command(toolPath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create atom",
				"details": string(output),
			})
			return
		}

	case "UPDATE":
		atom, exists := Atoms[proposal.AtomID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found: " + proposal.AtomID})
			return
		}

		args := []string{"update", "--file", atom.FilePath}
		if v, ok := proposal.Content["human_name"].(string); ok && v != "" {
			args = append(args, "--set", "human_name="+v)
		}
		if v, ok := proposal.Content["type"].(string); ok && v != "" {
			args = append(args, "--set", "type="+v)
		}
		if v, ok := proposal.Content["status"].(string); ok && v != "" {
			args = append(args, "--set", "status="+v)
		}
		if v, ok := proposal.Content["intent"].(string); ok && v != "" {
			args = append(args, "--intent", v)
		}
		if v, ok := proposal.Content["logic"].(string); ok && v != "" {
			args = append(args, "--logic", v)
		}
		if v, ok := proposal.Content["technical_interface"].(string); ok && v != "" {
			args = append(args, "--interface", v)
		}
		if v, ok := proposal.Content["expectation"].(string); ok && v != "" {
			args = append(args, "--expectation", v)
		}

		cmd := exec.Command(toolPath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update atom",
				"details": string(output),
			})
			return
		}

	case "DELETE":
		atom, exists := Atoms[proposal.AtomID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found: " + proposal.AtomID})
			return
		}
		if err := os.Remove(atom.FilePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete atom file"})
			return
		}
	}

	refreshAtoms()
	c.JSON(http.StatusOK, gin.H{"message": "Proposal applied successfully", "atom_id": proposal.AtomID})
}

// @spec-link [[mechanic_webui_gemini_chat_orchestration]]
func handleGeminiChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY not configured"})
		return
	}

	os.Setenv("GOOGLE_API_KEY", apiKey)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Gemini client: " + err.Error()})
		return
	}

	// Build system instruction
	systemInstruction := atdManifesto

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
										Enum: []string{"CUSTOMER", "ARCHITECTURE", "IMPLEMENTATION"},
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
	if model == "" {
		model = "gemini-3.1-flash-lite-preview"
	}

	result, err := client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") || strings.Contains(err.Error(), "Quota exceeded") {
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{"error": "Gemini API error: " + err.Error()})
		return
	}

	responseText := result.Text()

	var geminiResp GeminiResponse
	if err := json.Unmarshal([]byte(responseText), &geminiResp); err != nil {
		geminiResp = GeminiResponse{
			Message:   responseText,
			Proposals: []GeminiProposal{},
		}
	}

	// @spec-link [[requirement_webui_token_transparency]]
	if result.UsageMetadata != nil {
		geminiResp.Usage = UsageRecord{
			PromptTokens:     int(result.UsageMetadata.PromptTokenCount),
			CandidatesTokens: int(result.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(result.UsageMetadata.TotalTokenCount),
		}
	}

	c.JSON(http.StatusOK, geminiResp)
}
