package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

func main() {
	loadConfig()

	// Parse Data
	atdDir := filepath.Join(AppConfig.ProjectPath, AppConfig.ATDPath)
	atoms, err := parser.ParseAtoms(atdDir)
	if err != nil {
		log.Printf("Warning: Failed to parse atoms: %v", err)
	} else {
		parser.FindLinkedCode(AppConfig.ProjectPath, atoms)
		parser.CalculateGreenStatus(atoms)
		fmt.Printf("Parsed %d atoms from %s\n", len(atoms), atdDir)
	}

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
				"atd_count":    len(atoms),
			})
		})

		api.GET("/tree", func(c *gin.Context) {
			// Convert mapping to slice for easy JSON response
			var slice []*parser.Atom
			for _, v := range atoms {
				slice = append(slice, v)
			}
			c.JSON(http.StatusOK, slice)
		})

		api.GET("/atd/:id", func(c *gin.Context) {
			id := c.Param("id")
			if atom, exists := atoms[id]; exists {
				c.JSON(http.StatusOK, atom)
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
			}
		})

		api.GET("/atd/:id/code", func(c *gin.Context) {
			id := c.Param("id")
			if atom, exists := atoms[id]; exists {
				c.JSON(http.StatusOK, gin.H{"linked_codes": atom.LinkedCodes})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
			}
		})

		api.GET("/atd/:id/tests", func(c *gin.Context) {
			id := c.Param("id")
			if atom, exists := atoms[id]; exists {
				// Simply returning whether tests exist for now.
				c.JSON(http.StatusOK, gin.H{
					"has_tests": atom.HasTests,
					"is_green":  atom.IsGreen,
				})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
			}
		})

		api.GET("/summary/:id", func(c *gin.Context) {
			// Stub for Ollama call
			c.JSON(http.StatusOK, gin.H{"summary": "Ollama summary will be generated here."})
		})
	}

	addr := fmt.Sprintf("%s:%d", AppConfig.Host, AppConfig.Port)
	fmt.Printf("Server starting on http://%s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
