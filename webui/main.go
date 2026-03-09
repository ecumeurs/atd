package main

import (
	"bytes"
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

func main() {
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
			toolPath := filepath.Join(AppConfig.ToolkitPath, "atd-update")

			for _, id := range req.IDs {
				atom, exists := Atoms[id]
				if !exists {
					continue
				}

				// Execute atd-update -file <path> -set status=<status>
				cmd := exec.Command(toolPath, "-file", atom.FilePath, "-set", "status="+req.Status)
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

			toolPath := filepath.Join(AppConfig.ToolkitPath, "atd-update")

			// 1. Handle Metadata Updates (including ID/Rename)
			args := []string{"-file", atom.FilePath}
			if req.ID != "" && req.ID != atom.ID {
				args = append(args, "-set", "id="+req.ID)
			}
			if req.HumanName != "" {
				args = append(args, "-set", "human_name="+req.HumanName)
			}
			if req.Type != "" {
				args = append(args, "-set", "type="+req.Type)
			}
			if req.Status != "" {
				args = append(args, "-set", "status="+req.Status)
			}
			if req.Priority != "" {
				args = append(args, "-set", "priority="+req.Priority)
			}
			if len(req.Tags) > 0 {
				args = append(args, "-set", "tags="+strings.Join(req.Tags, ","))
			}

			if len(args) > 2 {
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
	}

	addr := fmt.Sprintf("%s:%d", AppConfig.Host, AppConfig.Port)
	fmt.Printf("Server starting on http://%s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
