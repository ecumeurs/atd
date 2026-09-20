package exploration

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CrawlDocs(dir string, graph *DependencyGraph) error {
	return CrawlDocsWithTag(dir, graph, "")
}

func CrawlDocsWithTag(dir string, graph *DependencyGraph, projectName string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			a, err := atom.Parse(path)
			if err != nil {
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

func CrawlWorkspaceDocs(graph *DependencyGraph) error {
	snap := config.Snapshot()
	return CrawlWorkspaceDocsWithConfig(graph, &snap)
}

func CrawlWorkspaceDocsWithConfig(graph *DependencyGraph, cfg *config.Config) error {
	if cfg.Workspace == nil {
		return fmt.Errorf("no workspace active")
	}

	for _, p := range cfg.Workspace.Projects {
		absProjPath := p.Path
		if !filepath.IsAbs(absProjPath) {
			absProjPath = filepath.Join(cfg.Workspace.LoadedFrom, p.Path)
		}

		docsPath := p.DocsPath
		if docsPath == "" {
			docsPath = "docs/"
		}
		if !filepath.IsAbs(docsPath) {
			docsPath = filepath.Join(absProjPath, docsPath)
		}

		if err := CrawlDocsWithTag(docsPath, graph, p.Name); err != nil {
			fmt.Printf("Warning: failed to crawl %s: %v\n", p.Name, err)
		}
	}

	return nil
}

func CrawlSrc(dir string, graph *DependencyGraph) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if strings.Contains(path, "/.") || strings.Contains(path, "vendor") || strings.HasSuffix(path, ".atom.md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
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