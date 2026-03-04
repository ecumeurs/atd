package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"atd-tools/config"
)

// atd-recon: The Semantic Archaeology Engine. Helps developers validate if existing unmapped source code is an implementation of a specific Atom constraint.
func main() {
	var atomPath string
	var candidatePath string

	flag.StringVar(&atomPath, "atom", "", "Path to the target atom file")
	flag.StringVar(&candidatePath, "candidate", "", "Path to the candidate source code file")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-recon", "Started process")

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if atomPath == "" || candidatePath == "" {
		fmt.Println("Error: -atom and -candidate parameters are required.")
		os.Exit(1)
	}

	atomContent, err1 := os.ReadFile(atomPath)
	candidateContent, err2 := os.ReadFile(candidatePath)

	if err1 != nil || err2 != nil {
		fmt.Println("Error reading input files.")
		os.Exit(1)
	}

	prompt := fmt.Sprintf(`
<System Objective>
You are the ATD Recon Engine. Validate if the given candidate code logically acts as an implementation of the Target Atom even though it lacks the spec-link tag.
Output JSON: {"Confidence": int, "Mismatches": string}
</System Objective>

<Target Atom Intent>
%s
</Target Atom Intent>

<Candidate Search Hit>
%s
</Candidate Search Hit>
`, string(atomContent), string(candidateContent))

	fmt.Println(prompt)
}
