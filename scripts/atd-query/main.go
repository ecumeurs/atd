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

// AtomFrontmatter holds the metadata we care about filtering
type AtomFrontmatter struct {
	ID        string `json:"id"`
	HumanName string `json:"human_name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Priority  string `json:"priority"`
}

func main() {

	var searchField string
	var searchTerm string

	flag.StringVar(&searchField, "field", "id", "The header field to search within (e.g. 'id', 'human_name', 'tags')")
	flag.StringVar(&searchTerm, "search", "", "The exact keyword or regex to match")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-query", "Started process")

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if searchTerm == "" {
		fmt.Println("Error: -search parameter is required.")
		os.Exit(1)
	}

	matchedAtoms := searchAtoms(docsPath, searchField, searchTerm)

	// Output as JSON for potential piping
	output, _ := json.MarshalIndent(matchedAtoms, "", "  ")
	fmt.Println(string(output))
}

func searchAtoms(dir string, field string, term string) []string {
	var matches []string

	yamlRegex := regexp.MustCompile(`(?s)^---\n(.*?)\n---`)
	fieldRegex := regexp.MustCompile(fmt.Sprintf(`(?m)^%s:\s*\[?(.*?)\]?$`, regexp.QuoteMeta(field)))

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			// Extract YAML frontmatter
			yamlMatch := yamlRegex.FindStringSubmatch(string(content))
			if len(yamlMatch) > 1 {
				frontmatter := yamlMatch[1]

				// Extract specific field
				fieldMatch := fieldRegex.FindStringSubmatch(frontmatter)
				if len(fieldMatch) > 1 {
					value := fieldMatch[1]
					if strings.Contains(strings.ToLower(value), strings.ToLower(term)) {
						matches = append(matches, path)
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", dir, err)
	}
	return matches
}
