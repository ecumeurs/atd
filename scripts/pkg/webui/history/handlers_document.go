package main

// @spec-link [[mechanic_webui_document_generation]]

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CachedDocument struct {
	ID            string        `json:"id"`
	Intent        string        `json:"intent"`
	Timestamp     time.Time     `json:"timestamp"`
	Content       string        `json:"content"`
	InvolvedAtoms []interface{} `json:"atoms_involved"`
}

var (
	documentCache []CachedDocument
	cacheMutex    sync.RWMutex
	maxCacheSize  = 5
)

// handleSearchDocumentContext queries ATD for the top 5 matching atoms
// @spec-link [[mechanic_webui_document_generation]]
func handleSearchDocumentContext(c *gin.Context) {
	var req struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request or empty query"})
		return
	}

	toolPath := filepath.Join(expandPath(AppConfig.ToolkitPath), "atd")
	cmd := exec.Command(toolPath, "search", "--query", req.Query, "--limit", "5", "--scope", "docs")
	output, err := cmd.CombinedOutput()

	if err != nil && string(output) == "" {
		log.Printf("Failed to run atd search: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search atoms", "details": err.Error()})
		return
	}

	outStr := string(output)
	var foundIDs []string
	lines := strings.Split(outStr, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "[Match ") {
			// Find File: ... (Similarity:
			startIdx := strings.Index(line, "File: ")
			endIdx := strings.Index(line, " (Similarity:")
			if startIdx != -1 && endIdx != -1 {
				filePath := line[startIdx+6 : endIdx]
				base := filepath.Base(filePath)
				if strings.HasSuffix(base, ".atom.md") {
					id := strings.TrimSuffix(base, ".atom.md")
					foundIDs = append(foundIDs, id)
				}
			}
		}
	}

	var results []interface{}
	for _, id := range foundIDs {
		if atom, exists := Atoms[id]; exists {
			results = append(results, atom)
		}
	}

	c.JSON(http.StatusOK, results)
}

// handleGenerateDocument processes the generation using `atd assemble`
// @spec-link [[mechanic_webui_document_generation]]
func handleGenerateDocument(c *gin.Context) {
	var req struct {
		Intent string   `json:"intent"`
		Starts []string `json:"starts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	if len(req.Starts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one start ID must be provided"})
		return
	}

	startsRaw := strings.Join(req.Starts, ",")
	toolPath := filepath.Join(expandPath(AppConfig.ToolkitPath), "atd")
	cmd := exec.Command(toolPath, "assemble", "--starts", startsRaw, "--intent", "document: "+req.Intent, "--structured", "--json")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Assemble error: %v. Output: %s", err, string(out))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Assemble failed", "details": string(out)})
		return
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(out, &raw); err != nil {
		log.Printf("Assemble parse error: %v. Output was: %s", err, string(out))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse assemble output"})
		return
	}

	// We extract content and atoms involved from the structured JSON
	contentRaw, _ := raw["content"].(string)

	// Create cache item
	doc := CachedDocument{
		ID:        uuid.New().String(),
		Intent:    req.Intent,
		Timestamp: time.Now(),
		Content:   contentRaw,
	}

	// Map involved atoms
	if involved, ok := raw["involved_atoms"].([]interface{}); ok {
		for _, nameRaw := range involved {
			name := nameRaw.(string)
			doc.InvolvedAtoms = append(doc.InvolvedAtoms, name)
		}
	}

	cacheMutex.Lock()
	documentCache = append([]CachedDocument{doc}, documentCache...) // prepend
	if len(documentCache) > maxCacheSize {
		documentCache = documentCache[:maxCacheSize]
	}
	cacheMutex.Unlock()

	c.JSON(http.StatusOK, doc)
}

// @spec-link [[mechanic_webui_document_generation]]
func handleListDocuments(c *gin.Context) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	// Return everything except the full 'content' to save bandwidth
	var summaries []map[string]interface{}
	for _, doc := range documentCache {
		summaries = append(summaries, map[string]interface{}{
			"id":             doc.ID,
			"intent":         doc.Intent,
			"timestamp":      doc.Timestamp,
			"atoms_involved": doc.InvolvedAtoms,
		})
	}
	c.JSON(http.StatusOK, summaries)
}

// @spec-link [[mechanic_webui_document_generation]]
func handleGetDocument(c *gin.Context) {
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
