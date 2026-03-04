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

// atd-assemble: Stitches atoms together based on recursive links, creating temporary Markdown for human reviewers.
// It starts at given Atom IDs, gathers their content and dependent content, and optionally generates an LLM prompt.
func main() {
	var starts string
	var purpose string

	flag.StringVar(&starts, "starts", "", "Comma-separated list of Root Atom IDs to assemble from")
	flag.StringVar(&purpose, "purpose", "", "Purpose of the assembly to orient the document layout and content")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-assemble", "Started process")

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if starts == "" {
		fmt.Println("Error: -starts parameter is required.")
		os.Exit(1)
	}

	startIDs := strings.Split(starts, ",")
	for i := range startIDs {
		startIDs[i] = strings.TrimSpace(startIDs[i])
	}

	type Atom struct {
		Content    string
		Dependents []string
	}

	atoms := make(map[string]*Atom)
	idRegex := regexp.MustCompile(`(?m)^id:\s*\[?\[?([^\]\s]+)\]?\]?`)
	depListRegex := regexp.MustCompile(`(?sm)^dependents:\s*[\r\n]+(.*?)(?:^[a-z_]+:|$)`)
	itemRegex := regexp.MustCompile(`-\s*\[?\[?([^\]\s]+)\]?\]?`)

	filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			content, _ := os.ReadFile(path)
			contentStr := string(content)
			if match := idRegex.FindStringSubmatch(contentStr); len(match) > 1 {
				id := match[1]
				atom := &Atom{Content: contentStr}
				if dMatch := depListRegex.FindStringSubmatch(contentStr); len(dMatch) > 1 {
					items := itemRegex.FindAllStringSubmatch(dMatch[1], -1)
					for _, it := range items {
						atom.Dependents = append(atom.Dependents, it[1])
					}
				}
				atoms[id] = atom
			}
		}
		return nil
	})

	visited := make(map[string]bool)
	var gather func(id string) string
	gather = func(id string) string {
		id = strings.TrimSpace(strings.Trim(id, "[]"))
		if visited[id] {
			return ""
		}
		visited[id] = true
		atom, ok := atoms[id]
		if !ok {
			return ""
		}
		res := atom.Content + "\n\n"
		for _, dep := range atom.Dependents {
			res += gather(dep)
		}
		return res
	}

	var assembledRaw string
	for _, startID := range startIDs {
		assembledRaw += gather(startID)
	}

	if purpose == "" {
		fmt.Println(assembledRaw)
		return
	}

	prompt := fmt.Sprintf(`
<System Objective>
You are an ATD Assembler. Rewrite the following disjointed fragments into flowing paragraphs. Do not alter rules or logic.
Your purpose for assembling this document is: %s
Keep this purpose in mind to orient the document layout and content appropriately.
</System Objective>

<Raw Fragments>
%s
</Raw Fragments>
`, purpose, assembledRaw)

	fmt.Println(prompt)
}
