package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// @spec-link [[module_webui]]
type Config struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	ProjectPath string `json:"project_path"`
	ATDPath     string `json:"atd_path"`
	ToolkitPath string `json:"toolkit_path"`
}

var AppConfig Config

// @spec-link [[module_webui]]
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

// @spec-link [[module_webui]]
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home := os.Getenv("HOME")
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		return os.Getenv("HOME")
	}
	return path
}
