package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	var docsDir string
	var targetAtom string

	flag.StringVar(&targetAtom, "target", "", "The specific Atom ID to cross-audit against its adjacent logic bounds")

	var projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsDir, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()

	if docsDir == "" {
		docsDir = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if targetAtom == "" {
		fmt.Println("Error: The -target flag is required (e.g. -target=ruler-movement-resolution)")
		os.Exit(1)
	}

	// 1. Load all ATDs
	files, err := os.ReadDir(docsDir)
	if err != nil {
		fmt.Printf("Error reading docs directory: %s\n", err)
		os.Exit(1)
	}

	atomMap := make(map[string]string)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".atom.md") {
			path := filepath.Join(docsDir, f.Name())
			content, err := os.ReadFile(path)
			if err == nil {
				id := strings.TrimSuffix(f.Name(), ".atom.md")
				atomMap[id] = string(content)
			}
		}
	}

	targetContent, ok := atomMap[targetAtom]
	if !ok {
		fmt.Printf("Error: Target atom '%s' not found in %s\n", targetAtom, docsDir)
		os.Exit(1)
	}

	// 2. Crawl relationships (Regex extracts parents, dependents, and tags)
	relatedAtoms := make(map[string]bool)
	relatedAtoms[targetAtom] = true // include self

	linkRegex := regexp.MustCompile(`\[\[(.*?)\]\]`)
	tagRegex := regexp.MustCompile(`tags:\s*\[(.*?)\]`)

	// Extract links from target
	links := linkRegex.FindAllStringSubmatch(targetContent, -1)
	for _, match := range links {
		if len(match) > 1 {
			relatedAtoms[match[1]] = true
		}
	}

	// Extract tags from target
	var targetTags []string
	tagMatch := tagRegex.FindStringSubmatch(targetContent)
	if len(tagMatch) > 1 {
		rawTags := strings.Split(tagMatch[1], ",")
		for _, t := range rawTags {
			targetTags = append(targetTags, strings.TrimSpace(t))
		}
	}

	// Find siblings sharing tags
	for id, content := range atomMap {
		if id == targetAtom {
			continue
		}

		// If this file shares a link with the target
		remoteLinks := linkRegex.FindAllStringSubmatch(content, -1)
		for _, match := range remoteLinks {
			if len(match) > 1 && match[1] == targetAtom {
				relatedAtoms[id] = true
			}
		}

		// Or if it shares tags
		rmTagMatch := tagRegex.FindStringSubmatch(content)
		if len(rmTagMatch) > 1 {
			rawTags := strings.Split(rmTagMatch[1], ",")
			for _, t := range rawTags {
				for _, tgtTag := range targetTags {
					if strings.TrimSpace(t) == tgtTag {
						relatedAtoms[id] = true
						break
					}
				}
			}
		}
	}

	// 3. Build Prompt
	prompt := "<System Objective>\n" +
		"You are the ATD Lead Architect Meta-Auditor.\n" +
		"Your absolute priority is System Congruence: verifying that all documentation rules mathematically and logically align perfectly with one another BEFORE any implementation begins.\n" +
		fmt.Sprintf("You are auditing a specific target Atom: '%s'\n", targetAtom) +
		"You must read its content and its directly related siblings (parents/dependents/shared mechanics).\n" +
		"Seek out logical contradictions, missing state resolutions, and mismatched properties strictly between the Target and its siblings.\n" +
		"Output a markdown report including a CLEAR TABLE summarizing:\n" +
		"| Sibling Atom Pair | Congruent? | Contradiction / Gap Description |\n" +
		"Do not assume code exists. Analyze strictly based on what is stated in the Markdown Rules.\n" +
		"</System Objective>\n\n" +
		"<ATD Specifications>\n"

	for id := range relatedAtoms {
		content, ok := atomMap[id]
		if ok {
			prompt += fmt.Sprintf("--- ATOM: %s.atom.md ---\n%s\n\n", id, content)
		}
	}

	prompt += "</ATD Specifications>\n"

	fmt.Println(prompt)
}
