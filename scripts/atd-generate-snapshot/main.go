package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"atd-tools/config"
)

// atd-generate-snapshot: Extracts a cohesive document narrative for specific themes/audiences based on aggregated raw logic.
func main() {
	var theme string
	var file string

	flag.StringVar(&theme, "theme", "Executive Summary", "The target audience/format")
	flag.StringVar(&file, "file", "", "Target aggregated markdown file (often from atd-assemble)")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-generate-snapshot", "Started process")

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if file == "" {
		fmt.Println("Error: -file requirement missing.")
		os.Exit(1)
	}

	content, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	prompt := fmt.Sprintf(`
<System Objective>
You are an ATD Narrative Generator. Rewrite this fragmented logic into a cohesive, flowing document styled strictly as an: %s. Emphasize clarity and narrative over raw mathematical values.
</System Objective>

<Raw Mechanical Output>
%s
</Raw Mechanical Output>
`, theme, string(content))

	fmt.Println(prompt)
}
