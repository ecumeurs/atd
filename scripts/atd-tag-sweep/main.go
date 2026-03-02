package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// atd-tag-sweep: Proposes tags for orphaned logic by searching for specific matched contexts.
// In this basic implementation, it searches the folder and identifies files that seem highly related.
func main() {
	var id string
	var folder string
	var keyword string

	flag.StringVar(&id, "id", "", "Atom ID to place tags for")
	flag.StringVar(&folder, "folder", "./src", "Search base directory")
	flag.StringVar(&keyword, "keyword", "", "Keyword to sweep for")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if id == "" || keyword == "" {
		fmt.Println("Usage: atd-tag-sweep -id [[Tax_Code]] -folder ./src -keyword tax")
		os.Exit(1)
	}

	matches := 0
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			content, err := os.ReadFile(path)
			if err == nil && strings.Contains(string(content), keyword) {
				fmt.Printf("Sweep: Found potential legacy implementation for %s in %s (Keyword match: %s)\n", id, path, keyword)
				matches++
			}
		}
		return nil
	})

	fmt.Printf("Sweeping complete. Found %d potential candidates for %s.\n", matches, id)
}
