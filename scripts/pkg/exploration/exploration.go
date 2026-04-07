package exploration

// @spec-link [[mechanic_atd_exploration_graph]]

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// @spec-link [[mechanic_atd_exploration_graph]]
// Explorer manages the ATD dependency graph and source code links with in-memory caching.
type Explorer struct {
	ProjectRoot string
	DocsDir     string
	Graph       *DependencyGraph
	TestLinks   []TestLink
}

func NewExplorer(root, docsDir string) *Explorer {
	if root == "" {
		root = config.ProjectRoot()
	}
	if docsDir == "" {
		docsDir = config.DocsDir()
	}
	return &Explorer{
		ProjectRoot: root,
		DocsDir:     docsDir,
	}
}

// @spec-link [[mechanic_atd_exploration_graph]]
func (e *Explorer) Load(force bool) error {
	if !force && e.Graph != nil {
		return nil
	}

	e.Graph = &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	e.TestLinks = []TestLink{}

	files, err := e.ListFiles()
	if err != nil {
		return err
	}

	specLinkRegex := regexp.MustCompile(`@spec-link\s+\[?\[?([^\]\s]+)\]?\]?`)
	testLinkRegex := regexp.MustCompile(`@test-link\s+\[?\[?([^\]\s]+)\]?\]?`)

	// 1. Process Atoms
	for _, relPath := range files {
		absPath := filepath.Join(e.ProjectRoot, relPath)
		if strings.HasSuffix(relPath, ".atom.md") {
			a, err := atom.Parse(absPath)
			if err != nil {
				// Fallback to minimal parse
				id, _, _, metaErr := atom.ParseMeta(absPath)
				if metaErr == nil && id != "" {
					e.Graph.Atoms[id] = &atom.AtomData{
						ID:       id,
						FilePath: absPath,
					}
				}
				continue
			}
			node := a
			e.Graph.Atoms[a.ID] = &node
		}
	}

	// 2. Process Source for links
	for _, relPath := range files {
		absPath := filepath.Join(e.ProjectRoot, relPath)
		if strings.HasSuffix(relPath, ".atom.md") {
			continue
		}

		// Check extensions
		ext := filepath.Ext(relPath)
		if !config.ActiveConfig.SupportedExtensions[ext] {
			continue
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			// Spec Links
			specMatches := specLinkRegex.FindAllStringSubmatch(line, -1)
			for _, match := range specMatches {
				if len(match) > 1 {
					atomID := match[1]
					if node, exists := e.Graph.Atoms[atomID]; exists {
						location := fmt.Sprintf("%s:%d", relPath, i+1)
						node.Implementations = append(node.Implementations, location)
					}
				}
			}

			// Test Links
			testMatches := testLinkRegex.FindAllStringSubmatch(line, -1)
			for _, match := range testMatches {
				if len(match) > 1 {
					atomID := match[1]
					e.TestLinks = append(e.TestLinks, TestLink{
						AtomID:   atomID,
						TestFile: relPath,
						Line:     i + 1,
					})
					if node, exists := e.Graph.Atoms[atomID]; exists {
						node.HasTests = true
					}
				}
			}
		}
	}

	return nil
}

func (e *Explorer) ListFiles() ([]string, error) {
	// Try git ls-files first
	cmd := exec.Command("git", "ls-files", "-c", "-o", "--exclude-standard")
	cmd.Dir = e.ProjectRoot
	out, err := cmd.Output()
	if err == nil {
		return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
	}

	// Fallback to manual walk
	var files []string
	err = filepath.Walk(e.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(e.ProjectRoot, path)
		if err != nil {
			return nil
		}
		// Basic ignores for fallback
		if strings.Contains(rel, "/.") || strings.Contains(rel, "vendor/") || strings.Contains(rel, "node_modules/") {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	return files, err
}

func (e *Explorer) GetGraph() *DependencyGraph {
	return e.Graph
}

func (e *Explorer) GetTestLinks() []TestLink {
	return e.TestLinks
}

func (e *Explorer) Trace(targetID string) (*TraceSnapshot, error) {
	target, exists := e.Graph.Atoms[targetID]
	if !exists {
		return nil, fmt.Errorf("atom not found: %s", targetID)
	}

	snap := &TraceSnapshot{
		TargetID: targetID,
		Layer:    target.Layer,
		GraphSlice: TraceGraphSlice{
			Parents:    []string{},
			Dependents: []string{},
			CodeLinks:  []string{},
			TestLinks:  []string{},
		},
		Warnings: []string{},
	}

	// 1. Walk UP
	ancestryComplete := true
	visitedUp := make(map[string]bool)
	e.Graph.WalkUp(targetID, visitedUp, func(id string) {
		if id != targetID {
			snap.GraphSlice.Parents = append(snap.GraphSlice.Parents, id)
			node := e.Graph.Atoms[id]
			if node.Status != "STABLE" {
				ancestryComplete = false
			}
		}
	})

	// 2. Walk DOWN
	visitedDown := make(map[string]bool)
	e.Graph.WalkDown(targetID, visitedDown, func(id string) {
		if id != targetID {
			snap.GraphSlice.Dependents = append(snap.GraphSlice.Dependents, id)
		}
	})

	// 3. Collect code/test links
	codeFiles := make(map[string]bool)
	testFiles := make(map[string]bool)
	uniqueCodeForTargetOrDescendants := make(map[string]bool)

	// Helper to strip line number
	getFile := func(impl string) string {
		parts := strings.Split(impl, ":")
		return parts[0]
	}

	processNode := func(id string, node *atom.AtomData) {
		if len(node.Implementations) > 0 {
			for _, impl := range node.Implementations {
				f := getFile(impl)
				codeFiles[f] = true
				uniqueCodeForTargetOrDescendants[f] = true
			}
		}
		// Test links from e.TestLinks
		for _, tl := range e.TestLinks {
			if tl.AtomID == id {
				testFiles[tl.TestFile] = true
				uniqueCodeForTargetOrDescendants[tl.TestFile] = true
			}
		}
	}

	processNode(targetID, target)
	for _, id := range snap.GraphSlice.Dependents {
		if node, ok := e.Graph.Atoms[id]; ok {
			processNode(id, node)
		}
	}

	// 4. Layers & Warnings
	foundArchAnc := false
	foundCustAnc := false
	for _, p := range snap.GraphSlice.Parents {
		if node, ok := e.Graph.Atoms[p]; ok {
			layer := node.Layer
			if layer == "ARCHITECTURE" {
				foundArchAnc = true
			}
			if layer == "CUSTOMER" {
				foundCustAnc = true
			}
		}
	}

	foundArchDesc := false
	foundImplDesc := false
	for _, d := range snap.GraphSlice.Dependents {
		if node, ok := e.Graph.Atoms[d]; ok {
			layer := node.Layer
			if layer == "ARCHITECTURE" {
				foundArchDesc = true
			}
			if layer == "IMPLEMENTATION" {
				foundImplDesc = true
			}
		}
	}

	switch target.Layer {
	case "CUSTOMER":
		if !foundArchDesc { snap.Warnings = append(snap.Warnings, "Customer atom has no Architecture dependents") }
		if !foundImplDesc { snap.Warnings = append(snap.Warnings, "Customer atom has no Implementation dependents") }
	case "ARCHITECTURE":
		if !foundImplDesc { snap.Warnings = append(snap.Warnings, "Architecture atom has no Implementation dependents") }
		if !foundCustAnc { snap.Warnings = append(snap.Warnings, "Architecture atom has no Customer origin") }
	case "IMPLEMENTATION":
		if !foundArchAnc { snap.Warnings = append(snap.Warnings, "Implementation atom has no Architecture origin") }
		if len(uniqueCodeForTargetOrDescendants) == 0 { snap.Warnings = append(snap.Warnings, "Implementation atom has no linked code") }
	}

	for f := range codeFiles {
		snap.GraphSlice.CodeLinks = append(snap.GraphSlice.CodeLinks, f)
	}
	for f := range testFiles {
		snap.GraphSlice.TestLinks = append(snap.GraphSlice.TestLinks, f)
	}

	// 5. Health Summary
	totalPool := len(snap.GraphSlice.Dependents)
	implementedCount := 0
	testedCount := 0

	checkHealth := func(id string, node *atom.AtomData) (bool, bool) {
		impl := len(node.Implementations) > 0
		test := false
		for _, tl := range e.TestLinks {
			if tl.AtomID == id {
				test = true
				break
			}
		}
		return impl, test
	}

	if target.Layer == "IMPLEMENTATION" {
		totalPool++
		impl, test := checkHealth(targetID, target)
		if impl {
			implementedCount++
			if test { testedCount++ }
		}
	}

	for _, id := range snap.GraphSlice.Dependents {
		if node, ok := e.Graph.Atoms[id]; ok {
			impl, test := checkHealth(id, node)
			if impl {
				implementedCount++
				if test { testedCount++ }
			}
		}
	}

	snap.HealthSummary.AncestryComplete = ancestryComplete
	snap.HealthSummary.HasCustomerOrigin = foundCustAnc || target.Layer == "CUSTOMER"
	if totalPool > 0 {
		snap.HealthSummary.ImplementationRate = float64(implementedCount) / float64(totalPool)
	}
	if implementedCount > 0 {
		snap.HealthSummary.TestCoverageRate = float64(testedCount) / float64(implementedCount)
	}

	snap.Metrics.TotalDependents = len(snap.GraphSlice.Dependents)
	snap.Metrics.ImplementedDependents = implementedCount
	snap.Metrics.TotalCodeFiles = len(codeFiles)
	snap.Metrics.TotalTests = len(testFiles)

	return snap, nil
}

// @spec-link [[service_atd_query]]
func (e *Explorer) Query(field, search string) []*atom.AtomData {
	var matches []*atom.AtomData
	search = strings.ToLower(search)

	for _, a := range e.Graph.Atoms {
		match := false
		if field == "" {
			// Smart Search: ID, HumanName, and Tags
			if strings.Contains(strings.ToLower(a.ID), search) ||
				strings.Contains(strings.ToLower(a.HumanName), search) {
				match = true
			} else {
				for _, t := range a.Tags {
					if strings.Contains(strings.ToLower(t), search) {
						match = true
						break
					}
				}
			}
		} else {
			// Targeted Field Search
			var val string
			switch strings.ToLower(field) {
			case "id":
				val = a.ID
			case "human_name", "name":
				val = a.HumanName
			case "status":
				val = a.Status
			case "layer":
				val = a.Layer
			case "type":
				val = a.Type
			case "tags":
				val = strings.Join(a.Tags, ",")
			}
			if strings.Contains(strings.ToLower(val), search) {
				match = true
			}
		}

		if match {
			matches = append(matches, a)
		}
	}
	return matches
}

type DependencyGraph struct {
	Atoms map[string]*atom.AtomData `json:"atoms"`
}

type TestLink struct {
	AtomID   string `json:"atom_id"`
	TestFile string `json:"test_file"`
	Line     int    `json:"line"`
}

type TraceSnapshot struct {
	TargetID      string          `json:"target_id"`
	Layer         string          `json:"layer"`
	HealthSummary HealthSummary   `json:"health_summary"`
	Metrics       TraceMetrics    `json:"metrics"`
	GraphSlice    TraceGraphSlice `json:"graph_slice"`
	Warnings      []string        `json:"warnings"`
}

type HealthSummary struct {
	AncestryComplete   bool    `json:"ancestry_complete"`
	HasCustomerOrigin  bool    `json:"has_customer_origin"`
	ImplementationRate float64 `json:"implementation_rate"`
	TestCoverageRate   float64 `json:"test_coverage_rate"`
}

type TraceMetrics struct {
	TotalDependents       int `json:"total_dependents"`
	ImplementedDependents int `json:"implemented_dependents"`
	TotalCodeFiles        int `json:"total_code_files"`
	TotalTests            int `json:"total_tests"`
}

type TraceGraphSlice struct {
	Parents    []string `json:"parents"`
	Dependents []string `json:"dependents"`
	CodeLinks  []string `json:"code_links"`
	TestLinks  []string `json:"test_links"`
}

// @spec-link [[mechanic_atd_exploration_graph]]
func CrawlDocs(dir string, graph *DependencyGraph) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip path on error
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			a, err := atom.Parse(path)
			if err != nil {
				// Fallback to minimal parse
				id, _, _, metaErr := atom.ParseMeta(path)
				if metaErr != nil {
					return nil
				}
				if id != "" {
					graph.Atoms[id] = &atom.AtomData{
						ID:       id,
						FilePath: path,
					}
				}
				return nil
			}

			node := a
			graph.Atoms[a.ID] = &node
		}
		return nil
	})
}

// @spec-link [[mechanic_atd_exploration_graph]]
func CrawlSrc(dir string, graph *DependencyGraph) error {
	specLinkRegex := regexp.MustCompile(`@spec-link\s+\[?\[?([^\]\s]+)\]?\]?`)
	testLinkRegex := regexp.MustCompile(`@test-link\s+\[?\[?([^\]\s]+)\]?\]?`)

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
			// Check Spec Links
			specMatches := specLinkRegex.FindAllStringSubmatch(line, -1)
			for _, match := range specMatches {
				if len(match) > 1 {
					atomID := match[1]
					if node, exists := graph.Atoms[atomID]; exists {
						location := fmt.Sprintf("%s:%d", path, i+1)
						node.Implementations = append(node.Implementations, location)
					}
				}
			}

			// Check Test Links
			testMatches := testLinkRegex.FindAllStringSubmatch(line, -1)
			for _, match := range testMatches {
				if len(match) > 1 {
					atomID := match[1]
					if node, exists := graph.Atoms[atomID]; exists {
						node.HasTests = true
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
