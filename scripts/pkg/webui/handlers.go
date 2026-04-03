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

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		ID          string `json:"id"`
		HumanName   string `json:"human_name"`
		Type        string `json:"type"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
		Intent      string `json:"intent"`
		Logic       string `json:"logic"`
		Interface   string `json:"interface"`
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
		if len(node.Implementations) > 0 {
			covered++
		}
		// Note: has_tests logic would require crawling tests too, which CrawlSrc should do if @test-link is used.
		// For now simple stats.
	}

	c.JSON(http.StatusOK, gin.H{
		"Total":        total,
		"SpecCoverage": covered,
		"TestCoverage": tested,
		"Orphans":      orphans,
	})
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


