package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"webui/parser"

	"github.com/gin-gonic/gin"
)

type atomInfo struct {
	ID     string `json:"id"`
	Layer  string `json:"layer"`
	Type   string `json:"type"`
	Intent string `json:"intent"`
}

// registerATDRoutes registers all ATD-related API endpoints.
// @spec-link [[api_webui_health_stats]]
func registerATDRoutes(api *gin.RouterGroup) {
	api.GET("/info", handleInfo)
	api.GET("/tree", handleTree)
	api.GET("/atd/:id", handleAtomDetail)
	api.GET("/atd/:id/code", handleAtomCode)
	api.GET("/atd/:id/tests", handleAtomTests)
	api.POST("/atd/:id/update", handleAtomUpdate)
	api.POST("/bulk-update", handleBulkUpdate)
	api.GET("/summary/:id", handleSummary)
	api.GET("/search", handleSearch)
}

// @spec-link [[api_webui_health_stats]]
func handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"project_path": AppConfig.ProjectPath,
		"atd_path":     AppConfig.ATDPath,
		"atd_count":    len(Atoms),
	})
}

// @spec-link [[api_webui_health_stats]]
func handleTree(c *gin.Context) {
	var slice []*parser.Atom
	for _, v := range Atoms {
		slice = append(slice, v)
	}
	c.JSON(http.StatusOK, slice)
}

// @spec-link [[api_webui_health_stats]]
func handleAtomDetail(c *gin.Context) {
	id := c.Param("id")
	if atom, exists := Atoms[id]; exists {
		c.JSON(http.StatusOK, atom)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
	}
}

// @spec-link [[api_webui_health_stats]]
func handleAtomCode(c *gin.Context) {
	id := c.Param("id")
	if atom, exists := Atoms[id]; exists {
		c.JSON(http.StatusOK, gin.H{"linked_codes": atom.LinkedCodes})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
	}
}

// @spec-link [[api_webui_health_stats]]
func handleAtomTests(c *gin.Context) {
	id := c.Param("id")
	if atom, exists := Atoms[id]; exists {
		c.JSON(http.StatusOK, gin.H{
			"has_tests": atom.HasTests,
			"is_green":  atom.IsGreen,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
	}
}

// @spec-link [[mechanic_atd_update]]
func handleAtomUpdate(c *gin.Context) {
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

	if len(args) > 3 {
		cmd := exec.Command(toolPath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Metadata update failed for %s: %v, output: %s", id, err, string(output))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update metadata", "details": string(output)})
			return
		}
	}

	// 2. Handle Content Update
	newPath := atom.FilePath
	if req.ID != "" && req.ID != atom.ID {
		newPath = filepath.Join(filepath.Dir(atom.FilePath), req.ID+".atom.md")
	}

	if req.Content != "" {
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
}

// @spec-link [[mechanic_atd_update]]
func handleBulkUpdate(c *gin.Context) {
	var req struct {
		IDs    []string `json:"ids"`
		Status string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	toolPath := filepath.Join(expandPath(AppConfig.ToolkitPath), "atd")

	for _, id := range req.IDs {
		atom, exists := Atoms[id]
		if !exists {
			continue
		}
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
}

// @spec-link [[mechanic_webui_summary_aggregation]]
func handleSummary(c *gin.Context) {
	id := c.Param("id")
	length := c.DefaultQuery("length", "default")
	toolPath := filepath.Join(expandPath(AppConfig.ToolkitPath), "atd")

	cmd := exec.Command(toolPath, "assemble", "--starts", id, "--intent", "summarize", "--length", length, "--json", "--structured")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to run atd assemble: %v, output: %s", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate summary", "details": string(output)})
		return
	}

	var assembleResp struct {
		Content  string `json:"content"`
		Metadata []struct {
			ID       string `json:"id"`
			Layer    string `json:"layer"`
			Type     string `json:"type"`
			Filepath string `json:"filepath"`
			Intent   string `json:"intent"`
		} `json:"metadata"`
	}

	if err := json.Unmarshal(output, &assembleResp); err != nil {
		log.Printf("Failed to parse JSON from atd assemble: %v, output: %s", err, string(output))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse summary output"})
		return
	}

	for i := range assembleResp.Metadata {
		meta := &assembleResp.Metadata[i]
		if atom, ok := Atoms[meta.ID]; ok {
			meta.Intent = extractIntent(atom.Content)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"summary":        assembleResp.Content,
		"atoms_involved": assembleResp.Metadata,
		"llm_used":       true,
	})
}

// @spec-link [[ui_webui_search_overlay]]
func handleSearch(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	var results []map[string]interface{}

	for _, atom := range Atoms {
		if query == "" {
			continue // Don't return everything on empty query
		}

		// Search by ID, human_name, and content
		idMatch := strings.Contains(strings.ToLower(atom.ID), query)
		nameMatch := strings.Contains(strings.ToLower(atom.HumanName), query)
		contentMatch := strings.Contains(strings.ToLower(atom.Content), query)
		typeMatch := strings.Contains(strings.ToLower(atom.Type), query)
		tagMatch := false
		for _, tag := range atom.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				tagMatch = true
				break
			}
		}

		if idMatch || nameMatch || contentMatch || typeMatch || tagMatch {
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
