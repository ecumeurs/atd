package webui

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Server represents the WebUI server instance.
type Server struct {
	Engine     *gin.Engine
	DevMode    bool
	StaticPath string
	explorer   *exploration.Explorer
	mutex      sync.RWMutex
}

// NewServer creates a new WebUI server.
func NewServer(devMode bool, staticPath string) *Server {
	return NewServerWithConfig(devMode, staticPath, &config.ActiveConfig)
}

// NewServerWithConfig creates a new WebUI server with the provided config.
func NewServerWithConfig(devMode bool, staticPath string, cfg *config.Config) *Server {
	if cfg == nil {
		cfg = &config.ActiveConfig
	}
	
	// Load .env if it exists (for GEMINI_API_KEY)
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, relying on environment variables")
	}

	r := gin.Default()
	s := &Server{
		Engine:     r,
		DevMode:    devMode,
		StaticPath: staticPath,
		explorer:   exploration.NewExplorerWithConfig("", "", cfg),
	}

	s.setupRoutes()
	if err := s.refreshAtoms(); err != nil {
		log.Printf("Warning: failed to initial refresh atoms: %v", err)
	}
	return s
}

func (s *Server) setupRoutes() {
	fs := GetFileSystem(s.DevMode, s.StaticPath)

	// Serve static files
	s.Engine.StaticFS("/static", fs)

	s.Engine.GET("/", func(c *gin.Context) {
		// Serve index.html from the filesystem or embed
		content, err := GetFileContent(fs, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load index.html: %v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	api := s.Engine.Group("/api")
	s.registerATDRoutes(api)
	s.registerLLMRoutes(api)
}

func (s *Server) Start() error {
	cfg := config.ActiveConfig
	host := cfg.WebUI.Host
	port := cfg.WebUI.Port
	if port == 0 {
		port = 8080
	}
	addr := fmt.Sprintf("%s:%d", host, port)

	displayAddr := addr
	if host == "" {
		displayAddr = fmt.Sprintf("localhost:%d", port)
	}

	fmt.Printf("WebUI server starting on http://%s (DevMode: %v)\n", displayAddr, s.DevMode)
	return s.Engine.Run(addr)
}
