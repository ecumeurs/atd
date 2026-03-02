package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type ElementType string

const (
	TypeGeneric ElementType = "structure"
)

type Status string

const (
	StatusPending    Status = "PENDING"
	StatusDocumented Status = "DOCUMENTED"
)

type RoadmapItem struct {
	FilePath string      `json:"file_path"`
	Line     int         `json:"line"`
	Type     ElementType `json:"type"`
	Name     string      `json:"name"`
	Keyword  string      `json:"keyword"`
	Status   Status      `json:"status"`
	ATDLink  string      `json:"atd_link,omitempty"`
}

type Roadmap struct {
	Items []RoadmapItem `json:"items"`
}

var (
	// Matches `class Name`, `type Name`, `interface Name`, `func Name`, `function Name`, `def Name`, `fn Name`, `struct Name`
	// Accommodates optional visibility modifiers before the keyword.
	genericRegex  = regexp.MustCompile(`^\s*(?:export\s+|public\s+|private\s+|protected\s+|static\s+|async\s+)*(class|struct|interface|func|function|def|fn|type|namespace|module|impl)\s+([a-zA-Z0-9_]+)`)
	specLinkRegex = regexp.MustCompile(`@spec-link\s+\[?\[?([^\]\s]+)\]?\]?`)
)

func main() {
	var scanDir string
	var outPath string

	flag.StringVar(&scanDir, "dir", ".", "Directory to scan")
	flag.StringVar(&outPath, "out", "roadmap.json", "Output path for the roadmap JSON")

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

	roadmap := Roadmap{
		Items: []RoadmapItem{},
	}

	validExts := map[string]bool{
		".go": true, ".py": true, ".ts": true, ".js": true,
		".rs": true, ".java": true, ".c": true, ".cpp": true,
		".h": true, ".hpp": true, ".cs": true, ".php": true,
		".rb": true, ".swift": true, ".kt": true, ".scala": true,
	}

	cmd := exec.Command("git", "ls-files", "-c", "-o", "--exclude-standard")
	cmd.Dir = scanDir
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Warning: Failed to run git ls-files. Ensure %s is a git repository. Error: %v\n", scanDir, err)
		return
	}

	fileList := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, relPath := range fileList {
		if relPath == "" {
			continue
		}

		fullPath := filepath.Join(scanDir, relPath)
		ext := filepath.Ext(fullPath)
		if validExts[ext] {
			processGenericFile(fullPath, &roadmap)
		}
	}

	data, err := json.MarshalIndent(roadmap, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding roadmap: %v\n", err)
		return
	}

	err = os.WriteFile(outPath, data, 0644)
	if err != nil {
		fmt.Printf("Error writing roadmap: %v\n", err)
		return
	}

	fmt.Printf("Generic Roadmap generated at %s with %d elements tracked.\n", outPath, len(roadmap.Items))
}

func processGenericFile(path string, roadmap *Roadmap) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	var pendingATD string

	for i, line := range lines {
		// Just a heuristic matching, we don't care about precise language grammars
		if match := specLinkRegex.FindStringSubmatch(line); len(match) > 1 {
			pendingATD = match[1]
			continue
		}

		if match := genericRegex.FindStringSubmatch(line); len(match) > 2 {
			keyword := match[1]
			name := match[2]

			// Skip common noise words or single letter vars
			if len(name) <= 2 && keyword != "fn" && keyword != "def" {
				continue
			}

			status := StatusPending
			if pendingATD != "" {
				status = StatusDocumented
			}

			roadmap.Items = append(roadmap.Items, RoadmapItem{
				FilePath: path,
				Line:     i + 1,
				Type:     TypeGeneric,
				Keyword:  keyword,
				Name:     name,
				Status:   status,
				ATDLink:  pendingATD,
			})

			pendingATD = ""
			continue
		}

		// Decay pending ATD if the line is not empty and not a comment,
		// to avoid blindly attaching it miles away.
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "/*") && !strings.HasPrefix(trimmed, "*") {
			pendingATD = ""
		}
	}
}
