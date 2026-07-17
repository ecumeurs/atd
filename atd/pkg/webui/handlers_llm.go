package webui

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/chat"
	"atd-tools/pkg/llmservice"
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func (s *Server) registerLLMRoutes(api *gin.RouterGroup) {
	api.POST("/llm/chat", s.handleLLMChat)
	api.GET("/llm/models", s.handleLLMModels)
	api.GET("/llm/atoms", s.handleLLMAtoms)
	api.GET("/llm/atom/:id", s.handleLLMAtomDetail)
	api.POST("/llm/apply-proposal", s.handleApplyProposal)

	api.POST("/gemini/chat", s.handleLLMChat)
	api.GET("/gemini/models", s.handleLLMModels)
	api.GET("/gemini/atoms", s.handleLLMAtoms)
	api.GET("/gemini/atom/:id", s.handleLLMAtomDetail)
	api.POST("/gemini/apply-proposal", s.handleApplyProposal)
}

func (s *Server) handleLLMModels(c *gin.Context) {
	llmService := llmservice.NewService(nil)
	models, err := llmService.GetModels(context.Background(), c.GetHeader("X-LLM-Base-URL"), c.GetHeader("X-LLM-Key"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"models":  models,
		"default": "models/gemini-3.1-flash-lite-preview",
	})
}

func (s *Server) handleLLMChat(c *gin.Context) {
	var req chat.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	llmService := llmservice.NewService(nil)
	headers := llmservice.ExtractHeaders(c)

	resp, err := llmService.Chat(context.Background(), req, headers)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "limit") {
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{"error": err.Error()})
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

// @spec-link [[service_atd_update]]
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
