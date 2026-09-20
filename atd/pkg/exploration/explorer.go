package exploration

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

var (
	specLinkRegex = regexp.MustCompile(`@spec-link\s+\[?\[?([a-zA-Z0-9_\-\.\:]+)\]?\]?`)
	testLinkRegex = regexp.MustCompile(`@test-link\s+\[?\[?([a-zA-Z0-9_\-\.\:]+)\]?\]?`)
)

func NewExplorer(root, docsDir string) *Explorer {
	// config.Snapshot() takes config.ActiveConfig's lock and returns a value
	// copy, unlike &config.ActiveConfig which aliases the live, mutable
	// global. See the comment in NewExplorerWithConfig for why a copy is
	// required here regardless of what the caller passes in.
	snap := config.Snapshot()
	return NewExplorerWithConfig(root, docsDir, &snap)
}

func NewExplorerWithConfig(root, docsDir string, cfg *config.Config) *Explorer {
	if cfg == nil {
		snap := config.Snapshot()
		cfg = &snap
	}
	// Take our own copy rather than keeping cfg's pointer. An Explorer
	// routinely outlives the call that constructed it -- WalkUp/WalkDown
	// (via configuredMaxDepth) read through e.Config well after construction
	// returns. If e.Config still aliased a mutable Config (the live global,
	// or a pointer some other caller keeps mutating), any later
	// config.Restore/Load from another goroutine (e.g. a t.Parallel() test's
	// cleanup) would race with that read even though the copy here executes
	// under whatever lock the caller holds. A shallow copy is enough: no
	// Explorer code writes through e.Config, and the fields it does read
	// (MaxDepth, BusinessLayerException, HierarchicalOrphanCheck,
	// SupportedExtensions) are never mutated in place -- config.Restore/Load
	// always swap in a whole new Config value rather than editing the old
	// one, so the copied map headers stay valid.
	cfgCopy := *cfg
	cfg = &cfgCopy

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
		Config:      cfg,
	}
}

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

	for _, relPath := range files {
		absPath := filepath.Join(e.ProjectRoot, relPath)
		if strings.HasSuffix(relPath, ".atom.md") {
			a, err := atom.Parse(absPath)
			if err != nil {
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

	for _, relPath := range files {
		absPath := filepath.Join(e.ProjectRoot, relPath)
		if strings.HasSuffix(relPath, ".atom.md") {
			continue
		}

		ext := filepath.Ext(relPath)
		if !e.Config.SupportedExtensions[ext] {
			continue
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		e.extractLinks(relPath, contentStr)
	}

	return nil
}

func (e *Explorer) LoadWorkspace(force bool) error {
	if e.Workspace == nil {
		return fmt.Errorf("no workspace active")
	}

	e.Graph = &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}

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

	for _, p := range e.Workspace.Projects {
		absProjPath := p.Path
		if !filepath.IsAbs(absProjPath) {
			absProjPath = filepath.Join(e.Workspace.LoadedFrom, p.Path)
		}

		codePaths := p.CodePaths
		if len(codePaths) == 0 {
			codePaths = []string{"."}
		}

		docsDir := p.DocsPath
		if docsDir == "" {
			docsDir = "docs/"
		}

		projExplorer := NewExplorer(absProjPath, "")
		projExplorer.Graph = e.Graph
		ignore := newGitignoreMatcher(absProjPath)

		for _, cp := range codePaths {
			absCP := cp
			if !filepath.IsAbs(absCP) {
				absCP = filepath.Join(absProjPath, absCP)
			}

			filepath.Walk(absCP, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				rel, err := filepath.Rel(absProjPath, path)
				if err != nil {
					return nil
				}
				if shouldSkipDiscoveredPath(rel, docsDir, ignore) {
					return nil
				}
				projExplorer.loadFileLinks(rel)
				return nil
			})
		}

		e.SpecLinks = append(e.SpecLinks, projExplorer.SpecLinks...)
		e.TestLinks = append(e.TestLinks, projExplorer.TestLinks...)
	}

	return nil
}

func (e *Explorer) loadFileLinks(relPath string) {
	absPath := filepath.Join(e.ProjectRoot, relPath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return
	}
	e.extractLinks(relPath, string(content))
}

func (e *Explorer) extractLinks(relPath string, contentStr string) {
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

				if node, err := e.ResolveAtom(targetID); err == nil {
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

				if node, err := e.ResolveAtom(targetID); err == nil {
					node.HasTests = true
				}
			}
		}
	}
}

func (e *Explorer) ListFiles() ([]string, error) {
	var files []string

	ignore := newGitignoreMatcher(e.ProjectRoot)

	err := filepath.Walk(e.ProjectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(e.ProjectRoot, path)
		if err != nil {
			return nil
		}

		// Docs are intentionally NOT excluded here: .atom.md files live
		// under the docs directory and Load() relies on ListFiles() to
		// surface them for atom parsing.
		if shouldSkipDiscoveredPath(rel, "", ignore) {
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

func (e *Explorer) SpecLinksForAtom(id string) []SpecLink {
	var out []SpecLink
	for _, sl := range e.SpecLinks {
		if sl.AtomID == id {
			out = append(out, sl)
		}
	}
	return out
}

func (e *Explorer) TestLinksForAtom(id string) []TestLink {
	var out []TestLink
	for _, tl := range e.TestLinks {
		if tl.AtomID == id {
			out = append(out, tl)
		}
	}
	return out
}
