package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Atom struct to represent the ATD Frontmatter
type Atom struct {
	ID            string        `json:"id" yaml:"id"`
	HumanName     string        `json:"human_name" yaml:"human_name"`
	Type          string        `json:"type" yaml:"type"`
	Version       string        `json:"version" yaml:"version"`
	Status        string        `json:"status" yaml:"status"`
	Priority      string        `json:"priority" yaml:"priority"`
	Tags          []string      `json:"tags" yaml:"tags"`
	ParentsRaw    []interface{} `yaml:"parents"`
	DependentsRaw []interface{} `yaml:"dependents"`

	Parents    []string `json:"parents"`
	Dependents []string `json:"dependents"`

	// Enriched Data
	FilePath    string   `json:"file_path"`
	Content     string   `json:"content"`
	LinkedCodes []string `json:"linked_codes"`
	HasTests    bool     `json:"has_tests"`
	IsGreen     bool     `json:"is_green"`
}

var (
	linkRegex     = regexp.MustCompile(`\[\[(.*?)\]\]`)
	specLinkRegex = regexp.MustCompile(`@spec-link\s+\[\[(.*?)\]\]`)
)

func extractLinkStr(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		if len(val) > 0 {
			// recursively extract innermost string if it was a nested array `[[ ]]`
			return extractLinkStr(val[0])
		}
	}
	return ""
}

func cleanLinks(raw []interface{}) []string {
	var cleaned []string
	for _, item := range raw {
		l := extractLinkStr(item)
		if l == "" {
			continue
		}
		matches := linkRegex.FindStringSubmatch(l)
		if len(matches) > 1 {
			cleaned = append(cleaned, matches[1])
		} else {
			cleaned = append(cleaned, l)
		}
	}
	return cleaned
}

// ParseAtoms traverses the given directory and parses all `.atom.md` files
func ParseAtoms(dir string) (map[string]*Atom, error) {
	atoms := make(map[string]*Atom)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			atom, err := parseAtomFile(path)
			if err != nil {
				fmt.Printf("Error parsing %s: %v\n", path, err)
				return nil
			}
			atoms[atom.ID] = atom
		}
		return nil
	})

	return atoms, err
}

func parseAtomFile(path string) (*Atom, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parts := bytes.SplitN(content, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid frontmatter in %s", path)
	}

	var atom Atom
	err = yaml.Unmarshal(parts[1], &atom)
	if err != nil {
		return nil, err
	}

	atom.Parents = cleanLinks(atom.ParentsRaw)
	atom.Dependents = cleanLinks(atom.DependentsRaw)
	atom.FilePath = path
	atom.Content = string(parts[2])

	return &atom, nil
}

// FindLinkedCode traverses the project and finds @spec-link references
func FindLinkedCode(projectPath string, atoms map[string]*Atom) error {
	// Initialize empty arrays
	for _, a := range atoms {
		a.LinkedCodes = make([]string, 0)
	}

	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && (info.Name() == ".git" || info.Name() == "docs" || info.Name() == "webui") {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(info.Name())
		if ext != ".go" && ext != ".py" && ext != ".js" && ext != ".ts" { // Only check source files
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 1
		for scanner.Scan() {
			line := scanner.Text()
			matches := specLinkRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					atomID := match[1]
					if atom, exists := atoms[atomID]; exists {
						snip := fmt.Sprintf("%s:%d", path, lineNum)
						atom.LinkedCodes = append(atom.LinkedCodes, snip)
						if strings.Contains(strings.ToLower(path), "test") {
							atom.HasTests = true
						}
					}
				}
			}
			lineNum++
		}
		return nil
	})
	return err
}

// CalculateGreenStatus determines if an ATD is fully implemented, following traceability rules
func CalculateGreenStatus(atoms map[string]*Atom) {
	for _, atom := range atoms {
		// Rule: A block requires an upstream SPECIFICATION or REQUIREMENT atom to become fully "Green"
		hasSpecParent := hasAncestorType(atom, atoms, "SPECIFICATION", "REQUIREMENT")

		// Let's say green is when it has tests AND traces to a spec
		if atom.HasTests && hasSpecParent {
			atom.IsGreen = true
		} else {
			atom.IsGreen = false
		}
	}
}

func hasAncestorType(atom *Atom, all map[string]*Atom, targetTypes ...string) bool {
	visited := make(map[string]bool)
	var queue []*Atom
	queue = append(queue, atom)
	visited[atom.ID] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, t := range targetTypes {
			if curr.Type == t {
				return true
			}
		}

		for _, parentID := range curr.Parents {
			if !visited[parentID] {
				visited[parentID] = true
				if parent, exists := all[parentID]; exists {
					queue = append(queue, parent)
				}
			}
		}
	}
	return false
}
