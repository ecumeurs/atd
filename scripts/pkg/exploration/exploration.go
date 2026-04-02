package exploration

// @spec-link [[mechanic_atd_exploration_graph]]

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/pkg/atom"
)

type DependencyGraph struct {
	Atoms map[string]*AtomNode `json:"atoms"`
}

type AtomNode struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	Layer           string   `json:"layer"`
	Status          string   `json:"status"`
	FilePath        string   `json:"filepath"`
	Parents         []string `json:"parents"`
	Dependents      []string `json:"dependents"`
	Implementations []string `json:"source_implementations"`
}

func CrawlDocs(dir string, graph *DependencyGraph) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip path on error
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			a, err := atom.Parse(path)
			if err != nil {
				// Fallback to minimal parse if full parse fails (e.g. malformed sections)
				id, _, _, metaErr := atom.ParseMeta(path)
				if metaErr != nil {
					return nil
				}
				if id != "" {
					graph.Atoms[id] = &AtomNode{
						ID:              id,
						FilePath:        path,
						Parents:         []string{},
						Dependents:      []string{},
						Implementations: []string{},
					}
				}
				return nil
			}

			graph.Atoms[a.ID] = &AtomNode{
				ID:              a.ID,
				Type:            a.Type,
				Layer:           a.Layer,
				Status:          a.Status,
				FilePath:        path,
				Parents:         a.Parents,
				Dependents:      a.Dependents,
				Implementations: []string{},
			}
		}
		return nil
	})
}

func CrawlSrc(dir string, graph *DependencyGraph) error {
	specLinkRegex := regexp.MustCompile(`@spec-link\s+\[?\[?([^\]\s]+)\]?\]?`)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		// Skip hidden dirs and atom files
		if strings.Contains(path, "/.") || strings.Contains(path, "vendor") || strings.HasSuffix(path, ".atom.md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			matches := specLinkRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
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

// WalkUp recursively visits parents of the given ID and calls the provided callback list on each.
// Include self is determined by the caller logic; by default this function visits the target itself if not in visited map.
func (g *DependencyGraph) WalkUp(id string, visited map[string]bool, onVisit func(string)) {
	if visited[id] {
		return
	}
	visited[id] = true
	node, ok := g.Atoms[id]
	if !ok {
		return
	}
	onVisit(id)
	for _, p := range node.Parents {
		g.WalkUp(p, visited, onVisit)
	}
}

// WalkDown recursively visits dependents of the given ID and calls the provided callback list on each.
// Include self is determined by the caller logic; by default this function visits the target itself if not in visited map.
func (g *DependencyGraph) WalkDown(id string, visited map[string]bool, onVisit func(string)) {
	if visited[id] {
		return
	}
	visited[id] = true
	node, ok := g.Atoms[id]
	if !ok {
		return
	}
	onVisit(id)
	for _, d := range node.Dependents {
		g.WalkDown(d, visited, onVisit)
	}
}
