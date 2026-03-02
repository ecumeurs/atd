package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type DependencyGraph struct {
	Atoms map[string]AtomNode `json:"atoms"`
}

type AtomNode struct {
	ID              string   `json:"id"`
	Status          string   `json:"status"`
	Implementations []string `json:"source_implementations"`
}

type GapReport struct {
	OrphanedAtoms []string `json:"orphaned_stable_atoms"`
}

func main() {
	var graphPath string
	flag.StringVar(&graphPath, "graph", "", "Path to the JSON output from atd-crawl")

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

	if graphPath == "" {
		fmt.Println("Error: -graph parameter is required. (Provide the output of atd-crawl)")
		os.Exit(1)
	}

	content, err := os.ReadFile(graphPath)
	if err != nil {
		fmt.Printf("Failed to read graph file: %v\n", err)
		os.Exit(1)
	}

	var graph DependencyGraph
	if err := json.Unmarshal(content, &graph); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	report := GapReport{
		OrphanedAtoms: []string{},
	}

	for id, node := range graph.Atoms {
		if node.Status == "STABLE" && len(node.Implementations) == 0 {
			report.OrphanedAtoms = append(report.OrphanedAtoms, id)
		}
	}

	output, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(output))
}
