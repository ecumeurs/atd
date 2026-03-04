package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"atd-tools/config"
)

// atd-legacy-wrapper: Shuttles massive orphaned blocks under a single monolithic tag until a Dev refactors it.
func main() {
	var file string
	var id string

	flag.StringVar(&file, "file", "", "Target file")
	flag.StringVar(&id, "id", "", "Target atom ID")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-legacy-wrapper", "Started process")

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if file == "" || id == "" {
		fmt.Println("Usage: atd-legacy-wrapper -file main.go -id [[God_Node]]")
		os.Exit(1)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Failed to read file.")
		os.Exit(1)
	}

	updated := fmt.Sprintf("// @spec-link %s\n%s", id, string(content))
	err = os.WriteFile(file, []byte(updated), 0644)
	if err != nil {
		fmt.Println("Failed to write to file.")
		os.Exit(1)
	}

	fmt.Printf("Automated deterministic rewrite: Injected // @spec-link %s to %s\n", id, file)
}
