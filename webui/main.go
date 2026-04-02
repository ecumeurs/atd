package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @spec-link [[module_webui]]
func main() {
	// Load .env file for GEMINI_API_KEY
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, relying on environment variables")
	}

	loadConfig()
	refreshAtoms()

	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Register API routes
	api := r.Group("/api")
	registerATDRoutes(api)
	registerGeminiRoutes(api)

	addr := fmt.Sprintf("%s:%d", AppConfig.Host, AppConfig.Port)
	fmt.Printf("Server starting on http://%s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
