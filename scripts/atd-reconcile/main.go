package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"atd-tools/config"
)

// atd-reconcile: Matches new inbound doc logic against the existing library, finding redundancies, overlaps, and conflicts.
func main() {
	var input string
	var existing string

	flag.StringVar(&input, "new", "", "Path to file containing inbound markdown edits")
	flag.StringVar(&existing, "store", "", "Path to file containing existing atom blocks")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-reconcile", "Started process")

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if input == "" || existing == "" {
		fmt.Println("Error: -new and -store parameters required.")
		os.Exit(1)
	}

	inboundContent, err1 := os.ReadFile(input)
	storeContent, err2 := os.ReadFile(existing)

	if err1 != nil || err2 != nil {
		fmt.Println("Error reading files.")
		os.Exit(1)
	}

	prompt := fmt.Sprintf(`
<System Objective>
You are an ATD Reconciler managing conflict detection. Analyze the Inbound changes against the Existing knowledge store. Return a Semantic Diff mapping identifying contradictions or overlaps: [{"proposed_id": string, "relationship": "UPDATE|CONFLICT|NEW", "change_context": string}] 
</System Objective>

<Existing Store>
%s
</Existing Store>

<Inbound Edits>
%s
</Inbound Edits>
`, string(storeContent), string(inboundContent))

	fmt.Println(prompt)
}
