package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"webui/parser"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type Config struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	ProjectPath string `json:"project_path"`
	ATDPath     string `json:"atd_path"`
	ToolkitPath string `json:"toolkit_path"`
}

var AppConfig Config
var Atoms map[string]*parser.Atom

// --- Gemini Chat Types ---

type ChatMessage struct {
	Role    string `json:"role"` // "user" or "model"
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages   []ChatMessage            `json:"messages"`    // Full conversation history
	Model      string                   `json:"model"`       // Selected Gemini model (default: gemini-2.5-flash)
	AtdContext []map[string]interface{} `json:"atd_context"` // ATD atoms to inject as context
	Actions    []ActionRecord           `json:"actions"`     // Accept/reject history
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

type GeminiResponse struct {
	Message   string           `json:"message"`
	Proposals []GeminiProposal `json:"proposals"`
}

const atdManifesto = `You are an ATD (Atomic Traceable Documentation) Specification Architect.

RULES YOU MUST FOLLOW:
0. You are a sounding board for the user. You are not here to replace the user's judgement, but to help them make better decisions. You may challenge the user's assumptions and propose alternative solutions. You are not expected to provide new/update ATD at every message. You may ask for clarifications.
1. Every atom has EXACTLY ONE state-changing rule. If an intent needs "and" or "also", split into multiple atoms.
2. Atoms have strict YAML frontmatter: id, human_name, type, layer, version, status, priority, tags, parents, dependents.
3. The hierarchy is: CUSTOMER (requirements, usecases) -> ARCHITECTURE (modules, APIs) -> IMPLEMENTATION (mechanics, builds).
4. Valid types: MODULE, SERVICE, ENTITY, RULE, MECHANIC, DOMAIN, API, UI, DATA, USAGE, BUILD, REQUIREMENT, SPECIFICATION, USECASE, USER_STORY.
5. Valid statuses: DRAFT, REVIEW, STABLE.
6. Valid layers: CUSTOMER, ARCHITECTURE, IMPLEMENTATION.
7. Each atom has 4 mandatory H2 sections: INTENT, THE RULE / LOGIC, TECHNICAL INTERFACE, EXPECTATION.
8. The INTENT must be ONE sentence, no "and" or "also".
9. Parents link upward (impl -> arch -> customer). Dependents link downward.

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
        "status": "DRAFT",
        "priority": "3",
        "tags": ["tag1"],
        "parents": ["parent_atom_id"],
        "intent": "Single sentence why this exists.",
        "logic": "The core specification.",
        "technical_interface": "API endpoints, code tags, test names.",
        "expectation": "Verifiable acceptance criteria."
      },
      "impact_summary": "Brief description of what this change means."
    }
  ]
}

When the conversation is exploratory or you need clarification, return "proposals": []. Only propose atoms when you have sufficient information and the user's intent is clear.
When proposing updates, only include fields that change in "content".
When no proposals are needed (e.g. answering a question, challenging an assumption, or asking for more details), return an empty proposals array.
Always explain your reasoning in "message" before listing proposals.`

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config.json: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&AppConfig); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}
	fmt.Printf("Loaded Config: %+v\n", AppConfig)
}

func refreshAtoms() {
	atdDir := filepath.Join(AppConfig.ProjectPath, AppConfig.ATDPath)
	newAtoms, err := parser.ParseAtoms(atdDir)
	if err != nil {
		log.Printf("Warning: Failed to parse atoms: %v", err)
		return
	}
	parser.FindLinkedCode(AppConfig.ProjectPath, newAtoms)
	parser.CalculateStatuses(newAtoms)
	Atoms = newAtoms
	fmt.Printf("Refreshed %d atoms from %s\n", len(Atoms), atdDir)
}

// @spec-link [[module_webui]]
func main() {
	// Load .env file for GEMINI_API_KEY
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, relying on environment variables")
	}

	loadConfig()
	refreshAtoms()

	r := gin.Default()

	// Serve static files from the "static" directory
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	api := r.Group("/api")
	{
		api.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"project_path": AppConfig.ProjectPath,
				"atd_path":     AppConfig.ATDPath,
				"atd_count":    len(Atoms),
			})
		})

		api.GET("/tree", func(c *gin.Context) {
			// Convert mapping to slice for easy JSON response
			var slice []*parser.Atom
			for _, v := range Atoms {
				slice = append(slice, v)
			}
			c.JSON(http.StatusOK, slice)
		})

		api.GET("/atd/:id", func(c *gin.Context) {
			id := c.Param("id")
			if atom, exists := Atoms[id]; exists {
				c.JSON(http.StatusOK, atom)
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
			}
		})

		api.GET("/atd/:id/code", func(c *gin.Context) {
			id := c.Param("id")
			if atom, exists := Atoms[id]; exists {
				c.JSON(http.StatusOK, gin.H{"linked_codes": atom.LinkedCodes})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
			}
		})

		api.GET("/atd/:id/tests", func(c *gin.Context) {
			id := c.Param("id")
			if atom, exists := Atoms[id]; exists {
				// Simply returning whether tests exist for now.
				c.JSON(http.StatusOK, gin.H{
					"has_tests": atom.HasTests,
					"is_green":  atom.IsGreen,
				})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
			}
		})

		api.POST("/bulk-update", func(c *gin.Context) {
			var req struct {
				IDs    []string `json:"ids"`
				Status string   `json:"status"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Toolkit path might be relative to webui/
			toolPath := filepath.Join(expandPath(AppConfig.ToolkitPath), "atd")

			for _, id := range req.IDs {
				atom, exists := Atoms[id]
				if !exists {
					continue
				}

				// Execute atd update --file <path> --set status=<status>
				cmd := exec.Command(toolPath, "update", "--file", atom.FilePath, "--set", "status="+req.Status)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("Failed to update atom %s: %v, output: %s", id, err, string(output))
				} else {
					log.Printf("Successfully updated atom %s to %s", id, req.Status)
				}
			}

			refreshAtoms()
			c.JSON(http.StatusOK, gin.H{"message": "Bulk update completed"})
		})

		api.GET("/summary/:id", func(c *gin.Context) {
			// Stub for Ollama call
			c.JSON(http.StatusOK, gin.H{"summary": "Ollama summary will be generated here."})
		})

		api.POST("/atd/:id/update", func(c *gin.Context) {
			id := c.Param("id")
			var req struct {
				ID        string   `json:"id"`
				HumanName string   `json:"human_name"`
				Type      string   `json:"type"`
				Status    string   `json:"status"`
				Priority  string   `json:"priority"`
				Tags      []string `json:"tags"`
				Content   string   `json:"content"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			atom, exists := Atoms[id]
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
				return
			}

			toolPath := filepath.Join(expandPath(AppConfig.ToolkitPath), "atd")

			// 1. Handle Metadata Updates (including ID/Rename)
			args := []string{"update", "--file", atom.FilePath}
			if req.ID != "" && req.ID != atom.ID {
				args = append(args, "--set", "id="+req.ID)
			}
			if req.HumanName != "" {
				args = append(args, "--set", "human_name="+req.HumanName)
			}
			if req.Type != "" {
				args = append(args, "--set", "type="+req.Type)
			}
			if req.Status != "" {
				args = append(args, "--set", "status="+req.Status)
			}
			if req.Priority != "" {
				args = append(args, "--set", "priority="+req.Priority)
			}
			if len(req.Tags) > 0 {
				args = append(args, "--set", "tags="+strings.Join(req.Tags, ","))
			}

			if len(args) > 3 { // "update", "--file", atom.FilePath are already 3
				cmd := exec.Command(toolPath, args...)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("Metadata update failed for %s: %v, output: %s", id, err, string(output))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update metadata", "details": string(output)})
					return
				}
			}

			// 2. Handle Content Update
			// After rename, file path might have changed
			newPath := atom.FilePath
			if req.ID != "" && req.ID != atom.ID {
				newPath = filepath.Join(filepath.Dir(atom.FilePath), req.ID+".atom.md")
			}

			if req.Content != "" {
				// We need to read the file, replace the section after the second ---
				fullContent, err := os.ReadFile(newPath)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file for content update"})
					return
				}

				parts := bytes.SplitN(fullContent, []byte("---"), 3)
				if len(parts) < 3 {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid file structure during content update"})
					return
				}

				// Rebuild the file: parts[0] is empty or whitespace before first ---, parts[1] is YAML, parts[2] is old content
				var buf bytes.Buffer
				buf.Write(parts[0])
				buf.WriteString("---")
				buf.Write(parts[1])
				buf.WriteString("---")
				buf.WriteString(req.Content)

				err = os.WriteFile(newPath, buf.Bytes(), 0644)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write updated content"})
					return
				}
			}

			refreshAtoms()
			c.JSON(http.StatusOK, gin.H{"message": "Atom updated successfully", "new_id": req.ID})
		})

		api.POST("/gemini/chat", handleGeminiChat)

		// @spec-link [[mechanic_webui_gemini_proxy]]
		api.GET("/gemini/models", func(c *gin.Context) {
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
				"default": "gemini-2.5-flash",
			})
		})

		// @spec-link [[mechanic_webui_gemini_proxy]]
		api.GET("/gemini/atoms", func(c *gin.Context) {
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
		})
	}

	addr := fmt.Sprintf("%s:%d", AppConfig.Host, AppConfig.Port)
	fmt.Printf("Server starting on http://%s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		return os.Getenv("HOME")
	}
	return path
}

func extractIntent(content string) string {
	lines := strings.Split(content, "\n")
	inIntent := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## INTENT") {
			inIntent = true
			continue
		}
		if inIntent && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if inIntent && trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// @spec-link [[mechanic_webui_gemini_proxy]]
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

	// Set the env var the SDK expects
	os.Setenv("GOOGLE_API_KEY", apiKey)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Gemini client: " + err.Error()})
		return
	}

	// Build the system instruction with manifesto
	systemInstruction := atdManifesto

	// Append ATD context if provided
	if len(req.AtdContext) > 0 {
		systemInstruction += "\n\n--- CURRENT ATD CONTEXT ---\nThe following atoms are currently relevant to this conversation:\n"
		for _, atd := range req.AtdContext {
			atdJSON, _ := json.Marshal(atd)
			systemInstruction += string(atdJSON) + "\n"
		}
	}

	// Append action history if provided
	if len(req.Actions) > 0 {
		systemInstruction += "\n\n--- USER ACTION HISTORY ---\nThe user has taken the following actions on previous proposals:\n"
		for _, action := range req.Actions {
			systemInstruction += fmt.Sprintf("- Proposal for atom '%s': %s. %s\n", action.AtomID, action.Action, action.Summary)
		}
	}

	// Build conversation parts from history
	var contents []*genai.Content
	for _, msg := range req.Messages {
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role: role,
			Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
		})
	}

	// Call Gemini API
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
							"action":         {Type: genai.TypeString, Enum: []string{"CREATE", "UPDATE", "DELETE"}},
							"atom_id":        {Type: genai.TypeString},
							"content":        {Type: genai.TypeObject},
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

	// Use selected model or default
	model := req.Model
	if model == "" {
		model = "gemini-2.5-flash"
	}

	result, err := client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gemini API error: " + err.Error()})
		return
	}

	// Extract the text response
	responseText := result.Text()

	// Parse the structured JSON response
	var geminiResp GeminiResponse
	if err := json.Unmarshal([]byte(responseText), &geminiResp); err != nil {
		// If parsing fails, return raw text as message
		geminiResp = GeminiResponse{
			Message:   responseText,
			Proposals: []GeminiProposal{},
		}
	}

	c.JSON(http.StatusOK, geminiResp)
}
