package exploration

// @spec-link [[mechanic_atd_exploration_graph]]

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/workspace"
	"fmt"
	"os"
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
	SpecLinks   []SpecLink
	TestLinks   []TestLink
	Workspace   *workspace.Workspace
	Resolver    *workspace.Resolver
	Index       *workspace.AtomIndex
}

func NewExplorer(root, docsDir string) *Explorer {
	if root == "" {
		root = config.ProjectRoot()
	}
	if docsDir == "" {
		docsDir = config.DocsDir()
	}

	ws, _ := workspace.LoadWorkspace(root)
	var idx *workspace.AtomIndex
	var resolver *workspace.Resolver
	if ws != nil {
		idx, _ = ws.BuildIndex()
		currentProj := ws.FindProjectByCWD(root)
		projName := ""
		if currentProj != nil {
			projName = currentProj.Name
		}
		resolver = workspace.NewResolver(ws, idx, projName)
	}

	return &Explorer{
		ProjectRoot: root,
		DocsDir:     docsDir,
		Workspace:   ws,
		Index:       idx,
		Resolver:    resolver,
	}
}

// @spec-link [[mechanic_atd_exploration_graph]]
func (e *Explorer) Load(force bool) error {
	if !force && e.Graph != nil {
		return nil
	}

	e.Graph = &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	e.SpecLinks = []SpecLink{}
	e.TestLinks = []TestLink{}

	files, err := e.ListFiles()
	if err != nil {
		return err
	}

	// Support [[id]], [id], or id formats
	specLinkRegex := regexp.MustCompile(`@spec-link\s+\[?\[?([a-zA-Z0-9_\-\.\:]+)\]?\]?`)
	testLinkRegex := regexp.MustCompile(`@test-link\s+\[?\[?([a-zA-Z0-9_\-\.\:]+)\]?\]?`)

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

		contentStr := string(content)

		// Spec Links
		specIndices := specLinkRegex.FindAllStringSubmatchIndex(contentStr, -1)
		for _, idx := range specIndices {
			if len(idx) >= 4 {
				atomID := contentStr[idx[2]:idx[3]]
				
				var targetID string
				found := false
				
				if e.Resolver != nil {
					parsed, err := e.Resolver.Resolve(atomID)
					if err == nil && parsed.Location != nil {
						targetID = parsed.AtomID
						if parsed.Type == workspace.ReferenceCrossProject {
							targetID = fmt.Sprintf("%s:%s", parsed.Project, parsed.AtomID)
						}
						found = true
					}
				} else {
					if _, exists := e.Graph.Atoms[atomID]; exists {
						targetID = atomID
						found = true
					}
				}

				if found {
					lineNum := strings.Count(contentStr[:idx[0]], "\n") + 1
					location := fmt.Sprintf("%s:%d", relPath, lineNum)
					
					// We need the node in the graph
					node, err := e.ResolveAtom(targetID)
					if err == nil {
						node.Implementations = append(node.Implementations, location)
						e.SpecLinks = append(e.SpecLinks, SpecLink{
							AtomID:   targetID,
							FilePath: relPath,
							Line:     lineNum,
						})
					}
				}
			}
		}

		// Test Links
		testIndices := testLinkRegex.FindAllStringSubmatchIndex(contentStr, -1)
		for _, idx := range testIndices {
			if len(idx) >= 4 {
				atomID := contentStr[idx[2]:idx[3]]
				
				var targetID string
				found := false
				
				if e.Resolver != nil {
					parsed, err := e.Resolver.Resolve(atomID)
					if err == nil && parsed.Location != nil {
						targetID = parsed.AtomID
						if parsed.Type == workspace.ReferenceCrossProject {
							targetID = fmt.Sprintf("%s:%s", parsed.Project, parsed.AtomID)
						}
						found = true
					}
				} else {
					if _, exists := e.Graph.Atoms[atomID]; exists {
						targetID = atomID
						found = true
					}
				}

				if found {
					lineNum := strings.Count(contentStr[:idx[0]], "\n") + 1
					e.TestLinks = append(e.TestLinks, TestLink{
						AtomID:   targetID,
						TestFile: relPath,
						Line:     lineNum,
					})
					
					node, err := e.ResolveAtom(targetID)
					if err == nil {
						node.HasTests = true
					}
				}
			}
		}
	}

	return nil
}

// LoadWorkspace loads all atoms from all projects in the workspace.
func (e *Explorer) LoadWorkspace(force bool) error {
	if e.Workspace == nil {
		return fmt.Errorf("no workspace active")
	}

	e.Graph = &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	
	// Load all atoms from the workspace index
	for atomID, loc := range e.Index.ByID {
		a, err := atom.Parse(loc.Path)
		if err != nil {
			continue
		}
		
		prefixedID := atomID
		if loc.Project != "" && loc.Project != "shared" {
			prefixedID = fmt.Sprintf("%s:%s", loc.Project, atomID)
		}
		
		e.Graph.Atoms[prefixedID] = &a
	}

	// Crawl source for each project to find links
	for _, p := range e.Workspace.Projects {
		absProjPath := p.Path
		if !filepath.IsAbs(absProjPath) {
			absProjPath = filepath.Join(e.Workspace.LoadedFrom, p.Path)
		}

		codePaths := p.CodePaths
		if len(codePaths) == 0 {
			codePaths = []string{"."}
		}
		
		// Create a temporary explorer for this project to use its context
		projExplorer := NewExplorer(absProjPath, "")
		projExplorer.Graph = e.Graph // Share the workspace graph
		
		for _, cp := range codePaths {
			absCP := cp
			if !filepath.IsAbs(absCP) {
				absCP = filepath.Join(absProjPath, absCP)
			}
			
			// Manually crawl and use ResolveAtom
			filepath.Walk(absCP, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(absProjPath, path)
				projExplorer.loadFileLinks(rel)
				return nil
			})
		}
		
		// Merge links
		e.SpecLinks = append(e.SpecLinks, projExplorer.SpecLinks...)
		e.TestLinks = append(e.TestLinks, projExplorer.TestLinks...)
	}

	return nil
}

// loadFileLinks is a helper for LoadWorkspace to process a single file for links
func (e *Explorer) loadFileLinks(relPath string) {
	absPath := filepath.Join(e.ProjectRoot, relPath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return
	}
	contentStr := string(content)

	// Use regexes from Load()
	specLinkRegex := regexp.MustCompile(`@spec-link\s+\[?\[?([a-zA-Z0-9_\-\.\:]+)\]?\]?`)
	testLinkRegex := regexp.MustCompile(`@test-link\s+\[?\[?([a-zA-Z0-9_\-\.\:]+)\]?\]?`)

	// Spec Links
	specIndices := specLinkRegex.FindAllStringSubmatchIndex(contentStr, -1)
	for _, idx := range specIndices {
		if len(idx) >= 4 {
			atomID := contentStr[idx[2]:idx[3]]
			if parsed, err := e.Resolver.Resolve(atomID); err == nil && parsed.Location != nil {
				targetID := parsed.AtomID
				if parsed.Type == workspace.ReferenceCrossProject {
					targetID = fmt.Sprintf("%s:%s", parsed.Project, parsed.AtomID)
				}
				
				lineNum := strings.Count(contentStr[:idx[0]], "\n") + 1
				location := fmt.Sprintf("%s:%d", relPath, lineNum)
				
				if node, ok := e.Graph.Atoms[targetID]; ok {
					node.Implementations = append(node.Implementations, location)
					e.SpecLinks = append(e.SpecLinks, SpecLink{
						AtomID:   targetID,
						FilePath: relPath,
						Line:     lineNum,
					})
				}
			}
		}
	}

	// Test Links
	testIndices := testLinkRegex.FindAllStringSubmatchIndex(contentStr, -1)
	for _, idx := range testIndices {
		if len(idx) >= 4 {
			atomID := contentStr[idx[2]:idx[3]]
			if parsed, err := e.Resolver.Resolve(atomID); err == nil && parsed.Location != nil {
				targetID := parsed.AtomID
				if parsed.Type == workspace.ReferenceCrossProject {
					targetID = fmt.Sprintf("%s:%s", parsed.Project, parsed.AtomID)
				}
				
				lineNum := strings.Count(contentStr[:idx[0]], "\n") + 1
				e.TestLinks = append(e.TestLinks, TestLink{
					AtomID:   targetID,
					TestFile: relPath,
					Line:     lineNum,
				})
				if node, ok := e.Graph.Atoms[targetID]; ok {
					node.HasTests = true
				}
			}
		}
	}
}

// IsOrphan determines if an atom should be considered a "true orphan" based on type-aware rules.
func (e *Explorer) IsOrphan(node *atom.AtomData) bool {
	// 1. Only STABLE atoms are tracked as orphans in the health metrics
	if node.Status != "STABLE" {
		return false
	}

	// 2. If it has direct source code implementations, it is definitely not an orphan
	if len(node.Implementations) > 0 {
		return false
	}

	// 3. Exclude specific types that are architectural groupings or top-level documents
	if config.GetOrphanExcludedTypes()[node.Type] {
		return false
	}

	// 4. Handle Customer Layer exceptions (they are requirements, not implementation units)
	if config.ActiveConfig.CustomerLayerException && node.Layer == "BUSINESS" {
		return false
	}

	// 5. Hierarchical check: a parent atom is NOT an orphan if it has implemented children
	if config.ActiveConfig.HierarchicalOrphanCheck {
		for _, depID := range node.Dependents {
			if depNode, ok := e.Graph.Atoms[depID]; ok {
				// If a child has direct implementations, the parent is "covered by proxy"
				if len(depNode.Implementations) > 0 {
					return false
				}
				// Note: Deep recursion could be added if needed, but one-level implementation
				// detection handles the majority of false positives in ATD.
			}
		}
	}

	return true
}

func (e *Explorer) ListFiles() ([]string, error) {
	var files []string
	
	// We use filepath.Walk to ensure we catch files in nested git repositories (like submodules or sub-repos)
	// which 'git ls-files' might skip depending on where it is run.
	err := filepath.Walk(e.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		rel, err := filepath.Rel(e.ProjectRoot, path)
		if err != nil {
			return nil
		}
		
		// Basic ignores
		if strings.Contains(rel, "/.") || 
		   strings.Contains(rel, "vendor/") || 
		   strings.Contains(rel, "node_modules/") ||
		   strings.Contains(rel, "dist/") ||
		   strings.Contains(rel, "build/") {
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
			if layer == "BUSINESS" {
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
	case "BUSINESS":
		if !foundArchDesc { snap.Warnings = append(snap.Warnings, "Customer atom has no Architecture dependents") }
		if !foundImplDesc { snap.Warnings = append(snap.Warnings, "Customer atom has no Implementation dependents") }
	case "ARCHITECTURE":
		// Only warn about missing Implementation dependents if atom has no direct @spec-link
		// Architecture atoms may have direct code links (90% of cases) which is valid
		if !foundImplDesc && len(target.Implementations) == 0 {
			snap.Warnings = append(snap.Warnings, "Architecture atom has no Implementation dependents or direct @spec-link")
		}
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

	// Include IMPLEMENTATION layer atoms in health calculation
	if target.Layer == "IMPLEMENTATION" {
		totalPool++
		impl, test := checkHealth(targetID, target)
		if impl {
			implementedCount++
			if test { testedCount++ }
		}
	}

	// Include ARCHITECTURE layer atoms with direct @spec-link in health calculation
	if target.Layer == "ARCHITECTURE" && len(target.Implementations) > 0 {
		totalPool++
		implementedCount++
		_, test := checkHealth(targetID, target)
		if test { testedCount++ }
	}

	for _, id := range snap.GraphSlice.Dependents {
		if node, ok := e.Graph.Atoms[id]; ok {
			// Count IMPLEMENTATION layer dependents
			if node.Layer == "IMPLEMENTATION" {
				totalPool++
				impl, test := checkHealth(id, node)
				if impl {
					implementedCount++
					if test { testedCount++ }
				}
			}
			// Count ARCHITECTURE dependents with direct @spec-link
			if node.Layer == "ARCHITECTURE" && len(node.Implementations) > 0 {
				totalPool++
				impl, test := checkHealth(id, node)
				if impl {
					implementedCount++
					if test { testedCount++ }
				}
			}
		}
	}

	snap.HealthSummary.AncestryComplete = ancestryComplete
	snap.HealthSummary.HasCustomerOrigin = foundCustAnc || target.Layer == "BUSINESS"
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

// ResolveAtom finds an atom by ID, supporting [[project:id]] syntax and workspace-wide search.
func (e *Explorer) ResolveAtom(refID string) (*atom.AtomData, error) {
	if e.Resolver != nil {
		parsed, err := e.Resolver.Resolve(refID)
		if err == nil && parsed.Location != nil {
			// If already in graph, return it
			id := parsed.AtomID
			if parsed.Type == workspace.ReferenceCrossProject {
				id = fmt.Sprintf("%s:%s", parsed.Project, parsed.AtomID)
			}
			if node, ok := e.Graph.Atoms[id]; ok {
				return node, nil
			}

			// Load from file
			a, err := atom.Parse(parsed.Location.Path)
			if err != nil {
				return nil, err
			}
			node := a
			if e.Graph != nil {
				e.Graph.Atoms[id] = &node
			}
			return &node, nil
		}
	}

	// Fallback to legacy behavior if resolver not available or failed
	if node, ok := e.Graph.Atoms[refID]; ok {
		return node, nil
	}

	return nil, fmt.Errorf("atom '%s' not found", refID)
}

type DependencyGraph struct {
	Atoms map[string]*atom.AtomData `json:"atoms"`
}

type SpecLink struct {
	AtomID   string `json:"atom_id"`
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
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
	return CrawlDocsWithTag(dir, graph, "")
}

// CrawlDocsWithTag walks a directory and adds all atoms found to the graph, tagged with a project name.
func CrawlDocsWithTag(dir string, graph *DependencyGraph, projectName string) error {
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
					a = atom.AtomData{
						ID:       id,
						FilePath: path,
					}
				} else {
					return nil
				}
			}

			// Tag with project name
			if projectName != "" {
				if a.Metadata == nil {
					a.Metadata = make(map[string]string)
				}
				a.Metadata["project"] = projectName
			}

			node := a
			graph.Atoms[a.ID] = &node
		}
		return nil
	})
}

// CrawlWorkspaceDocs crawls all projects in the workspace and gathers all atoms.
func CrawlWorkspaceDocs(graph *DependencyGraph) error {
	if config.ActiveConfig.Workspace == nil {
		return fmt.Errorf("no workspace active")
	}

	for _, p := range config.ActiveConfig.Workspace.Projects {
		absProjPath := p.Path
		if !filepath.IsAbs(absProjPath) {
			absProjPath = filepath.Join(config.ActiveConfig.Workspace.LoadedFrom, p.Path)
		}

		docsPath := p.DocsPath
		if docsPath == "" {
			docsPath = "docs/"
		}
		if !filepath.IsAbs(docsPath) {
			docsPath = filepath.Join(absProjPath, docsPath)
		}

		if err := CrawlDocsWithTag(docsPath, graph, p.Name); err != nil {
			// Log warning but continue
			fmt.Printf("Warning: failed to crawl %s: %v\n", p.Name, err)
		}
	}

	return nil
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
