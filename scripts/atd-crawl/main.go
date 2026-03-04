package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
)

type DependencyGraph struct {
	Atoms map[string]*AtomNode `json:"atoms"`
}

type AtomNode struct {
	ID              string   `json:"id"`
	Status          string   `json:"status"`
	Parents         []string `json:"parents"`
	Dependents      []string `json:"dependents"`
	Implementations []string `json:"source_implementations"`
}

func main() {

	var srcPath string

	flag.StringVar(&srcPath, "src", "./src", "Path to the source code directory")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-crawl", "Started process")

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	graph := &DependencyGraph{
		Atoms: make(map[string]*AtomNode),
	}

	crawlDocs(docsPath, graph)
	crawlSrc(srcPath, graph)

	output, _ := json.MarshalIndent(graph, "", "  ")
	fmt.Println(string(output))
}

func crawlDocs(dir string, graph *DependencyGraph) {
	yamlRegex := regexp.MustCompile(`(?s)^---[\r\n]+(.*?)[\r\n]+---`)
	idRegex := regexp.MustCompile(`(?m)^id:\s*\[?\[?([^\]\s]+)\]?\]?`)
	statusRegex := regexp.MustCompile(`(?m)^status:\s*\[?([^\]\r\n]+)\]?`)

	// Complex regex for yaml lists
	parentListRegex := regexp.MustCompile(`(?sm)^parents:\s*[\r\n]+(.*?)(?:^[a-z_]+:|$)`)
	depListRegex := regexp.MustCompile(`(?sm)^dependents:\s*[\r\n]+(.*?)(?:^[a-z_]+:|$)`)
	itemRegex := regexp.MustCompile(`-\s*\[?\[?([^\]\s]+)\]?\]?`)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			content, _ := os.ReadFile(path)
			yamlMatch := yamlRegex.FindStringSubmatch(string(content))

			if len(yamlMatch) > 1 {
				fm := yamlMatch[1]
				idMatch := idRegex.FindStringSubmatch(fm)

				if len(idMatch) > 1 {
					id := idMatch[1]
					node := &AtomNode{
						ID:              id,
						Parents:         []string{},
						Dependents:      []string{},
						Implementations: []string{},
					}

					if statMatch := statusRegex.FindStringSubmatch(fm); len(statMatch) > 1 {
						node.Status = statMatch[1]
					}

					// Extract parents
					if pMatch := parentListRegex.FindStringSubmatch(fm); len(pMatch) > 1 {
						items := itemRegex.FindAllStringSubmatch(pMatch[1], -1)
						for _, it := range items {
							node.Parents = append(node.Parents, it[1])
						}
					}
					// Extract dependents
					if dMatch := depListRegex.FindStringSubmatch(fm); len(dMatch) > 1 {
						items := itemRegex.FindAllStringSubmatch(dMatch[1], -1)
						for _, it := range items {
							node.Dependents = append(node.Dependents, it[1])
						}
					}

					graph.Atoms[id] = node
				}
			}
		}
		return nil
	})
}

func crawlSrc(dir string, graph *DependencyGraph) {
	specLinkRegex := regexp.MustCompile(`@spec-link\s+\[?\[?([^\]\s]+)\]?\]?`)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			lines := strings.Split(string(content), "\n")
			for i, line := range lines {
				if match := specLinkRegex.FindStringSubmatch(line); len(match) > 1 {
					atomID := match[1]
					if node, exists := graph.Atoms[atomID]; exists {
						location := fmt.Sprintf("%s:%d", path, i+1)
						node.Implementations = append(node.Implementations, location)
					}
				}
			}
		}
		return nil
	})
}
