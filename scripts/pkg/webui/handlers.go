package webui

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/exploration"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/genai"
)

var (
	atomsMap   map[string]*atom.AtomData
	atomsMutex sync.RWMutex
)

func (s *Server) registerATDRoutes(api *gin.RouterGroup) {
	api.GET("/info", s.handleInfo)
	api.GET("/tree", s.handleTree)
	api.GET("/atd/:id", s.handleAtomDetail)
	api.GET("/atd/:id/code", s.handleAtomCode)
	api.POST("/atd/:id/update", s.handleAtomUpdate)
	api.GET("/summary/:id", s.handleSummary)
	api.GET("/search", s.handleSearch)
	api.GET("/stats", s.handleStats)

	// Document generation
	api.POST("/search-document-context", s.handleSearchDocumentContext)
	api.POST("/generate-document", s.handleGenerateDocument)
	api.GET("/documents", s.handleListDocuments)
	api.GET("/documents/:id", s.handleGetDocument)

	// Gemini integration
	s.registerGeminiRoutes(api)
}

func (s *Server) refreshAtoms() error {
	docsDir := config.DocsDir()
	graph := &exploration.DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	if err := exploration.CrawlDocs(docsDir, graph); err != nil {
		return err
	}
	// Also crawl source for implementations
	if err := exploration.CrawlSrc(config.ProjectRoot(), graph); err != nil {
		return err
	}

	atomsMutex.Lock()
	atomsMap = graph.Atoms
	atomsMutex.Unlock()
	return nil
}

func (s *Server) handleInfo(c *gin.Context) {
	atomsMutex.RLock()
	count := len(atomsMap)
	atomsMutex.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"project_path": config.ProjectRoot(),
		"docs_path":    config.DocsDir(),
		"atd_count":    count,
	})
}

func (s *Server) handleTree(c *gin.Context) {
	atomsMutex.RLock()
	defer atomsMutex.RUnlock()
	
	var slice []*atom.AtomData
	for _, v := range atomsMap {
		slice = append(slice, v)
	}
	c.JSON(http.StatusOK, slice)
}

func (s *Server) handleAtomDetail(c *gin.Context) {
	id := c.Param("id")
	atomsMutex.RLock()
	data, exists := atomsMap[id]
	atomsMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (s *Server) handleAtomCode(c *gin.Context) {
	id := c.Param("id")
	atomsMutex.RLock()
	node, exists := atomsMap[id]
	atomsMutex.RUnlock()

	if exists {
		c.JSON(http.StatusOK, gin.H{"linked_codes": node.Implementations})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
	}
}

func (s *Server) handleAtomUpdate(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ID        string   `json:"id"`
		HumanName string   `json:"human_name"`
		Type      string   `json:"type"`
		Status    string   `json:"status"`
		Priority  string   `json:"priority"`
		Intent    string   `json:"intent"`
		Logic     string   `json:"logic"`
		Interface string   `json:"interface"`
		Expectation string `json:"expectation"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	atomsMutex.RLock()
	node, exists := atomsMap[id]
	atomsMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
		return
	}

	opts := atom.UpdateOptions{
		FilePath:    node.FilePath,
		Intent:      req.Intent,
		Logic:       req.Logic,
		Interface:   req.Interface,
		Expectation: req.Expectation,
	}

	if req.ID != "" && req.ID != id {
		opts.SetArgs = append(opts.SetArgs, "id="+req.ID)
	}
	if req.HumanName != "" {
		opts.SetArgs = append(opts.SetArgs, "human_name="+req.HumanName)
	}
	if req.Type != "" {
		opts.SetArgs = append(opts.SetArgs, "type="+req.Type)
	}
	if req.Status != "" {
		opts.SetArgs = append(opts.SetArgs, "status="+req.Status)
	}
	if req.Priority != "" {
		opts.SetArgs = append(opts.SetArgs, "priority="+req.Priority)
	}

	res, err := atom.Update(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed", "details": err.Error()})
		return
	}

	s.refreshAtoms()
	c.JSON(http.StatusOK, gin.H{"message": res, "new_id": req.ID})
}

func (s *Server) handleSummary(c *gin.Context) {
	id := c.Param("id")
	length := c.DefaultQuery("length", "default")

	opts := exploration.AssembleOptions{
		Starts:     id,
		Intent:     "summarize",
		Length:     length,
		Structured: true,
		AsJSON:     true,
		DocsDir:    config.DocsDir(),
	}

	res, err := exploration.Assemble(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate summary", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res) // Assumes res is already JSON string if AsJSON=true
}

func (s *Server) handleSearch(c *gin.Context) {
	query := c.Query("q")
	grep := c.Query("grep")
	
	opts := exploration.SearchOptions{
		Query:  query,
		Grep:   grep,
		DBPath: filepath.Join(config.DocsDir(), ".atd_index.db"),
		Limit:  10,
		Scope:  "all",
		Root:   config.ProjectRoot(),
	}

	results, err := exploration.Search(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) handleStats(c *gin.Context) {
	atomsMutex.RLock()
	defer atomsMutex.RUnlock()

	var total, covered, tested, orphans int
	for _, node := range atomsMap {
		total++
		if len(node.Implementations) > 0 { covered++ }
		// Note: has_tests logic would require crawling tests too, which CrawlSrc should do if @test-link is used.
		// For now simple stats.
	}

	c.JSON(http.StatusOK, gin.H{
		"Total": total,
		"SpecCoverage": covered,
		"TestCoverage": tested,
		"Orphans": orphans,
	})
}

// Document Generation (Cached)
type CachedDocument struct {
	ID            string        `json:"id"`
	Intent        string        `json:"intent"`
	Timestamp     time.Time     `json:"timestamp"`
	Content       string        `json:"content"`
	InvolvedAtoms []string      `json:"atoms_involved"`
}

var (
	documentCache []CachedDocument
	cacheMutex    sync.RWMutex
	maxCacheSize  = 10
)

func (s *Server) handleSearchDocumentContext(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query required"})
		return
	}

	opts := exploration.SearchOptions{
		Query:  req.Query,
		DBPath: filepath.Join(config.DocsDir(), ".atd_index.db"),
		Limit:  5,
		Scope:  "docs",
	}

	results, err := exploration.Search(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) handleGenerateDocument(c *gin.Context) {
	var req struct {
		Intent string   `json:"intent"`
		Starts []string `json:"starts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	opts := exploration.AssembleOptions{
		Starts:     strings.Join(req.Starts, ","),
		Intent:     req.Intent,
		Structured: true,
		AsJSON:     true,
		DocsDir:    config.DocsDir(),
	}

	resStr, err := exploration.Assemble(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Generation failed", "details": err.Error()})
		return
	}

	var res exploration.AssembleJSON
	json.Unmarshal([]byte(resStr), &res)

	doc := CachedDocument{
		ID:        uuid.New().String(),
		Intent:    req.Intent,
		Timestamp: time.Now(),
		Content:   res.Content,
	}
	for _, m := range res.Metadata {
		doc.InvolvedAtoms = append(doc.InvolvedAtoms, m.ID)
	}

	cacheMutex.Lock()
	documentCache = append([]CachedDocument{doc}, documentCache...)
	if len(documentCache) > maxCacheSize {
		documentCache = documentCache[:maxCacheSize]
	}
	cacheMutex.Unlock()

	c.JSON(http.StatusOK, doc)
}

func (s *Server) handleListDocuments(c *gin.Context) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	c.JSON(http.StatusOK, documentCache)
}

func (s *Server) handleGetDocument(c *gin.Context) {
	id := c.Param("id")
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	for _, doc := range documentCache {
		if doc.ID == id {
			c.JSON(http.StatusOK, doc)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
}

// --- Gemini Integration ---

func (s *Server) registerGeminiRoutes(api *gin.RouterGroup) {
	api.POST("/gemini/chat", s.handleGeminiChat)
	api.GET("/gemini/models", s.handleGeminiModels)
	api.POST("/gemini/apply-proposal", s.handleApplyProposal)
}

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

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []ChatMessage            `json:"messages"`
	Model       string                   `json:"model"`
	AtdContext  []map[string]interface{} `json:"atd_context"`
	OmitHistory bool                     `json:"omit_history"`
}

func (s *Server) handleGeminiModels(c *gin.Context) {
	// Return a static list of known-good models to avoid SDK-specific pagination complexity
	models := []map[string]interface{}{
		{"id": "gemini-3.1-flash-lite-preview", "display_name": "Gemini 1.5 Flash Lite Preview"},
		{"id": "gemini-2.0-flash", "display_name": "Gemini 2.0 Flash"},
		{"id": "gemini-2.0-pro-exp-02-05", "display_name": "Gemini 2.0 Pro Experimental"},
	}

	c.JSON(http.StatusOK, gin.H{
		"models":  models,
		"default": "gemini-3.1-flash-lite-preview",
	})
}

func (s *Server) handleGeminiChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GEMINI_API_KEY not configured"})
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Gemini client: " + err.Error()})
		return
	}

	systemInstruction := atdManifesto
	if len(req.AtdContext) > 0 {
		systemInstruction += "\n\n--- CURRENT ATD CONTEXT ---\n"
		for _, atd := range req.AtdContext {
			atdJSON, _ := json.Marshal(atd)
			systemInstruction += string(atdJSON) + "\n"
		}
	}

	var contents []*genai.Content
	messagesToInclude := req.Messages
	if req.OmitHistory && len(req.Messages) > 0 {
		messagesToInclude = req.Messages[len(req.Messages)-1:]
	}

	for _, msg := range messagesToInclude {
		role := msg.Role
		if role == "assistant" { role = "model" }
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{genai.NewPartFromText(msg.Content)},
		})
	}

	modelID := req.Model
	if modelID == "" { modelID = "gemini-3.1-flash-lite-preview" }

	generateConfig := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)},
		},
		ResponseMIMEType: "application/json",
		Temperature:      genai.Ptr(float32(0.7)),
	}

	result, err := client.Models.GenerateContent(ctx, modelID, contents, generateConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gemini API error: " + err.Error()})
		return
	}

	var geminiResp map[string]interface{}
	json.Unmarshal([]byte(result.Text()), &geminiResp)

	c.JSON(http.StatusOK, geminiResp)
}

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

	// Directly call atom.Update
	opts := atom.UpdateOptions{
		FilePath:    filepath.Join(config.DocsDir(), proposal.AtomID+".atom.md"),
	}

	if proposal.Action == "UPDATE" {
		atomsMutex.RLock()
		if node, ok := atomsMap[proposal.AtomID]; ok {
			opts.FilePath = node.FilePath
		}
		atomsMutex.RUnlock()
	}

	if v, ok := proposal.Content["human_name"].(string); ok { opts.SetArgs = append(opts.SetArgs, "human_name="+v) }
	if v, ok := proposal.Content["type"].(string); ok { opts.SetArgs = append(opts.SetArgs, "type="+v) }
	if v, ok := proposal.Content["layer"].(string); ok { opts.SetArgs = append(opts.SetArgs, "layer="+v) }
	if v, ok := proposal.Content["status"].(string); ok { opts.SetArgs = append(opts.SetArgs, "status="+v) }
	if v, ok := proposal.Content["priority"].(string); ok { opts.SetArgs = append(opts.SetArgs, "priority="+v) }
	
	if v, ok := proposal.Content["intent"].(string); ok { opts.Intent = v }
	if v, ok := proposal.Content["logic"].(string); ok { opts.Logic = v }
	if v, ok := proposal.Content["technical_interface"].(string); ok { opts.Interface = v }
	if v, ok := proposal.Content["expectation"].(string); ok { opts.Expectation = v }

	_, err := atom.Update(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply proposal", "details": err.Error()})
		return
	}

	s.refreshAtoms()
	c.JSON(http.StatusOK, gin.H{"message": "Proposal applied successfully"})
}
