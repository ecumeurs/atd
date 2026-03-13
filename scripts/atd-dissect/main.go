package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
)

func main() {
	var target string

	flag.StringVar(&target, "file", "", "Target unstructured document to dissect into Atoms")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-dissect", "Started process")

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

	lines := strings.Split(string(content), "\n")
	var numberedContent strings.Builder
	for i, line := range lines {
		numberedContent.WriteString(fmt.Sprintf("%03d: %s\n", i+1, line))
	}

	prompt := fmt.Sprintf(`
<System_Context>
You are an ATD Architect. The target document below has line numbers prepended (e.g., 001:). 
Identify atomic boundaries where a single architectural responsibility starts and ends.
</System_Context>

<Instruction>
1. Map each Atom to its exact line_range [start, end].
2. Identify the 'responsibility' as a deterministic skill definition.
3. Response must strictly follow the JSON schema.
</Instruction>

<Document>
%s
</Document>
`, numberedContent.String())

	fmt.Println(prompt)
}
