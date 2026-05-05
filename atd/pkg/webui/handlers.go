package webui

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/ollama"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	api.POST("/atd/weave", s.handleWeave)
	api.GET("/health", s.handleHealth)
	api.GET("/heatmap", s.handleHeatMap)

	// Workspace support
	api.GET("/workspace/info", s.handleWorkspaceInfo)
	api.POST("/workspace/switch", s.handleWorkspaceSwitch)

	// Document generation
	api.POST("/search-document-context", s.handleSearchDocumentContext)
	api.POST("/generate-document", s.handleGenerateDocument)
	api.GET("/documents", s.handleListDocuments)
	api.GET("/documents/:id", s.handleGetDocument)

}

func (s *Server) refreshAtoms() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.explorer.Load(true)
}

func (s *Server) handleInfo(c *gin.Context) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	count := 0
	if s.explorer.Graph != nil {
		count = len(s.explorer.Graph.Atoms)
	}

	resp := gin.H{
		"project_path": config.ProjectRoot(),
		"docs_path":    config.DocsDir(),
		"atd_count":    count,
	}

	if config.ActiveConfig.Workspace != nil {
		resp["workspace"] = gin.H{
			"in_workspace":   true,
			"workspace_name": config.ActiveConfig.Workspace.WorkspaceName,
			"active_project": config.ActiveConfig.ActiveProject,
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleWorkspaceInfo(c *gin.Context) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if config.ActiveConfig.Workspace == nil {
		c.JSON(http.StatusOK, gin.H{"in_workspace": false})
		return
	}

	ws := config.ActiveConfig.Workspace
	projects := []gin.H{}
	for _, p := range ws.Projects {
		projects = append(projects, gin.H{
			"name":      p.Name,
			"path":      p.Path,
			"is_active": p.Name == config.ActiveConfig.ActiveProject,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"in_workspace":   true,
		"workspace_name": ws.WorkspaceName,
		"workspace_root": ws.WorkspaceRoot,
		"active_project": config.ActiveConfig.ActiveProject,
		"projects":       projects,
	})
}

func (s *Server) handleWorkspaceSwitch(c *gin.Context) {
	var req struct {
		Project string `json:"project"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project name required"})
		return
	}

	if err := config.SetProject(req.Project); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Refresh atoms for the new project
	if err := s.refreshAtoms(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh atoms", "details": err.Error()})
		return
	}

	// Return updated workspace info
	s.handleWorkspaceInfo(c)
}

func (s *Server) handleTree(c *gin.Context) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var slice []*atom.AtomData
	if s.explorer.Graph != nil {
		for _, v := range s.explorer.Graph.Atoms {
			slice = append(slice, v)
		}
	}
	c.JSON(http.StatusOK, slice)
}

func (s *Server) handleAtomDetail(c *gin.Context) {
	id := c.Param("id")
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.explorer.Graph == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Graph not loaded"})
		return
	}

	data, exists := s.explorer.Graph.Atoms[id]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (s *Server) handleAtomCode(c *gin.Context) {
	id := c.Param("id")
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.explorer.Graph == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Graph not loaded"})
		return
	}

	node, exists := s.explorer.Graph.Atoms[id]
	if exists {
		c.JSON(http.StatusOK, gin.H{"linked_codes": node.Implementations})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
	}
}

func (s *Server) handleAtomUpdate(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ID          string `json:"id"`
		HumanName   string `json:"human_name"`
		Type        string `json:"type"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		Intent      string `json:"intent"`
		Logic       string `json:"logic"`
		Interface   string `json:"interface"`
		Expectation string `json:"expectation"`
		Layer       string `json:"layer"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.mutex.RLock()
	node, exists := s.explorer.Graph.Atoms[id]
	s.mutex.RUnlock()

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
	if req.Layer != "" {
		opts.SetArgs = append(opts.SetArgs, "layer="+req.Layer)
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
	workspace := c.Query("workspace") == "true"
	var projects []string
	if p := c.Query("projects"); p != "" {
		projects = strings.Split(p, ",")
	}

	opts := exploration.SearchOptions{
		Query:     query,
		Grep:      grep,
		DBPath:    filepath.Join(config.DocsDir(), ".atd_index.db"),
		Limit:     10,
		Scope:     "all",
		Root:      config.ProjectRoot(),
		Workspace: workspace,
		Projects:  projects,
	}

	results, err := exploration.Search(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed", "details": err.Error()})
		return
	}

	// Convert search results to atom objects for frontend compatibility
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	type AtomSearchResult struct {
		ID         string  `json:"id"`
		HumanName  string  `json:"human_name"`
		Type       string  `json:"type"`
		Layer      string  `json:"layer"`
		Status     string  `json:"status"`
		Intent     string  `json:"intent"`
		Similarity float64 `json:"similarity"`
		Project    string  `json:"project"` // NEW
	}

	var atomResults []AtomSearchResult
	seen := make(map[string]bool)

	for _, result := range results {
		// Extract atom ID from file path if it's an atom file
		var atomID string
		if strings.HasSuffix(result.FilePath, ".atom.md") {
			// Try to extract atom ID from file path
			parts := strings.Split(result.FilePath, "/")
			filename := parts[len(parts)-1]
			atomID = strings.TrimSuffix(filename, ".atom.md")
		}

		if atomID == "" || seen[atomID] {
			continue
		}

		// Look up full atom data from the graph
		if atom, exists := s.explorer.Graph.Atoms[atomID]; exists {
			seen[atomID] = true
			atomResults = append(atomResults, AtomSearchResult{
				ID:         atom.ID,
				HumanName:  atom.HumanName,
				Type:       atom.Type,
				Layer:      atom.Layer,
				Status:     atom.Status,
				Intent:     atom.Intent,
				Similarity: result.Similarity,
				Project:    result.Project, // NEW
			})
		}
	}

	c.JSON(http.StatusOK, atomResults)
}

func (s *Server) handleStats(c *gin.Context) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var total, covered, tested, orphans int
	var byLayer = map[string]int{"BUSINESS": 0, "ARCHITECTURE": 0, "IMPLEMENTATION": 0, "UNKNOWN": 0}
	var byStatus = map[string]int{"DRAFT": 0, "REVIEW": 0, "STABLE": 0, "UNKNOWN": 0}

	if s.explorer.Graph != nil {
		for _, node := range s.explorer.Graph.Atoms {
			total++
			if len(node.Implementations) > 0 {
				covered++
			}
			if node.HasTests {
				tested++
			}
			// Orphans: non-BUSINESS atoms with no parents
			if node.Layer != "BUSINESS" && len(node.Parents) == 0 {
				orphans++
			}
			if node.Layer != "" {
				byLayer[node.Layer]++
			} else {
				byLayer["UNKNOWN"]++
			}
			if node.Status != "" {
				byStatus[node.Status]++
			} else {
				byStatus["UNKNOWN"]++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"Total":        total,
		"SpecCoverage": covered,
		"TestCoverage": tested,
		"Orphans":      orphans,
		"ByLayer":      byLayer,
		"ByStatus":     byStatus,
	})
}

// @spec-link [[api_webui_health_check]]
func (s *Server) handleHealth(c *gin.Context) {
	health := gin.H{
		"providers": []gin.H{},
		"tasks":     map[string]string{},
	}

	cfg := config.ActiveConfig.LLM

	// Check each provider
	for _, provider := range cfg.Providers {
		if provider.Type == "passthrough" {
			health["providers"] = append(health["providers"].([]gin.H), gin.H{
				"name":   provider.Name,
				"type":   "passthrough",
				"status": "available",
			})
			continue
		}

		// Check provider availability
		models, err := ollama.ListModels(provider.BaseURL, provider.TimeoutMs)
		status := "available"
		if err != nil {
			status = "unreachable"
		}

		health["providers"] = append(health["providers"].([]gin.H), gin.H{
			"name":    provider.Name,
			"type":    "ollama",
			"base_url": provider.BaseURL,
			"status":  status,
			"models":  models,
		})
	}

	// Check task availability
	tasks := map[string]string{
		"embed":                        "",
		"assemble":                     "",
		"assemble_layer_BUSINESS":      "",
		"assemble_layer_ARCHITECTURE":  "",
		"assemble_layer_IMPLEMENTATION": "",
		"assemble_final":               "",
		"audit_bloat":                  "",
		"intent_extract":               "",
		"snapshot":                     "",
		"dissect":                      "",
		"recon":                        "",
		"audit_code":                   "",
		"compare":                      "",
		"congruence":                   "",
		"reconcile":                    "",
		"fix_split":                    "",
	}

	for taskName := range tasks {
		res, err := ollama.ResolveProvider(taskName)
		if err != nil {
			tasks[taskName] = "error: " + err.Error()
		} else if res.IsIDE {
			tasks[taskName] = "ide_fallback"
		} else {
			tasks[taskName] = res.Provider + ":" + res.Model
		}
	}

	health["tasks"] = tasks
	c.JSON(http.StatusOK, health)
}

// Document Generation (Cached)
type CachedDocument struct {
	ID            string    `json:"id"`
	Intent        string    `json:"intent"`
	Timestamp     time.Time `json:"timestamp"`
	Content       string    `json:"content"`
	InvolvedAtoms []string  `json:"atoms_involved"`
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
		Intent    string   `json:"intent"`
		Starts    []string `json:"starts"`
		Length    string   `json:"length,omitempty"`
		Workspace bool     `json:"workspace"` // NEW
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	opts := exploration.AssembleOptions{
		Starts:     strings.Join(req.Starts, ","),
		Intent:     req.Intent,
		Length:     req.Length,
		Structured: true,
		AsJSON:     true,
		DocsDir:    config.DocsDir(),
		Workspace:  req.Workspace, // NEW
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


// @spec-link [[mechanic_webui_atd_weave_handler]]
func (s *Server) handleWeave(c *gin.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	res, err := s.explorer.Weave()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Weaving failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": res})
}

func (s *Server) handleHeatMap(c *gin.Context) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.explorer.Graph == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Graph not loaded"})
		return
	}

	results := make(map[string]*exploration.HeatMapResult)
	for id := range s.explorer.Graph.Atoms {
		res, err := s.explorer.GetHeatMapResult(id)
		if err == nil {
			results[id] = res
		}
	}

	c.JSON(http.StatusOK, results)
}
