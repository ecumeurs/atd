package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
)

func main() {

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-link-weaver", "Started process")

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	// 1. Discover all IDs and their declared parents
	idRegex := regexp.MustCompile(`(?m)^id:\s*([a-zA-Z0-9_\-]+)`)
	parentsRegex := regexp.MustCompile(`(?m)^parents:\s*\[(.*?)\]`)
	linkExtractRegex := regexp.MustCompile(`\[\[([a-zA-Z0-9_\-]+)\]\]`)

	type AtomData struct {
		Path    string
		ID      string
		Parents []string
	}

	atoms := []AtomData{}
	parentToDependents := make(map[string][]string)

	filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			content, _ := os.ReadFile(path)
			data := AtomData{Path: path}

			idMatch := idRegex.FindStringSubmatch(string(content))
			if len(idMatch) > 1 {
				data.ID = idMatch[1]
			} else {
				return nil // Skip if no ID
			}

			parentMatch := parentsRegex.FindStringSubmatch(string(content))
			if len(parentMatch) > 1 {
				parentLinks := linkExtractRegex.FindAllStringSubmatch(parentMatch[1], -1)
				for _, match := range parentLinks {
					if len(match) > 1 {
						data.Parents = append(data.Parents, match[1])
						// Keep track of dependents
						parentToDependents[match[1]] = append(parentToDependents[match[1]], data.ID)
					}
				}
			}
			atoms = append(atoms, data)
		}
		return nil
	})

	// 2. Rewrite each file with updated dependents
	dependentsRegex := regexp.MustCompile(`(?m)^dependents:\s*\[.*?\]`)

	edited := 0
	for _, atom := range atoms {
		content, err := os.ReadFile(atom.Path)
		if err != nil {
			continue
		}

		deps := parentToDependents[atom.ID]
		newDepsLine := "dependents: []"
		if len(deps) > 0 {
			formattedDeps := []string{}
			for _, dep := range deps {
				formattedDeps = append(formattedDeps, fmt.Sprintf("[[%s]]", dep))
			}
			newDepsLine = fmt.Sprintf("dependents: [%s]", strings.Join(formattedDeps, ", "))
		}

		// Replace the dependents line
		if dependentsRegex.Match(content) {
			newContent := dependentsRegex.ReplaceAllString(string(content), newDepsLine)
			if newContent != string(content) {
				os.WriteFile(atom.Path, []byte(newContent), 0644)
				edited++
				fmt.Printf("Updated dependents for %s\n", atom.ID)
			}
		}
	}

	fmt.Printf("Link Weaving Complete. Updated %d files.\n", edited)
}
