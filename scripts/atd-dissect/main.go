package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	var target string

	flag.StringVar(&target, "file", "", "Target unstructured document to dissect into Atoms")

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

	if target == "" {
		fmt.Println("Error: -file parameter is required.")
		os.Exit(1)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		fmt.Println("Error reading file.")
		os.Exit(1)
	}

	prompt := fmt.Sprintf(`
<System Objective>
You are an ATD Deconstructor mapping structural boundaries. Identify explicit conceptual shifts (Domain vs API vs Rules) and output a JSON array estimating proposed Atom boundaries: [{"proposed_id": string, "responsibility": string, "excerpt_range": string}]
</System Objective>

<Target Document>
%s
</Target Document>
`, string(content))

	fmt.Println(prompt)
}
