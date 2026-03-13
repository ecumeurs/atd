package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"github.com/spf13/cobra"
)

type DependencyGraph struct {
	Atoms map[string]*AtomNode `json:"atoms"`
}

type AtomNode struct {
	ID              string   `json:"id"`
	Status          string   `json:"status"`
	Parents         []string `json:"parents"`
	Dependents      []string `json:"dependents"`
	Implementations []string `json:"source_implementations"`
}

type GapReport struct {
	OrphanedAtoms []string `json:"orphaned_stable_atoms"`
}

var crawlCmd = &cobra.Command{
	Use:   "crawl",
	Short: "Crawl ATD documents and source code to build a dependency graph",
	Long: `Crawl ATD documents and source code to build a dependency graph.
If --gaps is provided, identifies STABLE atoms with no implementation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath, _ := cmd.Flags().GetString("src")
		gaps, _ := cmd.Flags().GetBool("gaps")
		docsDir, _ := cmd.Flags().GetString("docs")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		graph := &DependencyGraph{
			Atoms: make(map[string]*AtomNode),
		}

		if err := crawlDocs(docsDir, graph); err != nil {
			return err
		}

		if srcPath != "" {
			if err := crawlSrc(srcPath, graph); err != nil {
				return err
			}
		}

		if gaps {
			report := GapReport{
				OrphanedAtoms: []string{},
			}
			for id, node := range graph.Atoms {
				if node.Status == "STABLE" && len(node.Implementations) == 0 {
					report.OrphanedAtoms = append(report.OrphanedAtoms, id)
				}
			}
			output, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(output))
		} else {
			output, _ := json.MarshalIndent(graph, "", "  ")
			fmt.Println(string(output))
		}

		return nil
	},
}

func crawlDocs(dir string, graph *DependencyGraph) error {
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
						Parents:         []string{},
						Dependents:      []string{},
						Implementations: []string{},
					}
				}
				return nil
			}

			graph.Atoms[a.ID] = &AtomNode{
				ID:              a.ID,
				Status:          a.Status,
				Parents:         a.Parents,
				Dependents:      a.Dependents,
				Implementations: []string{},
			}
		}
		return nil
	})
}

func crawlSrc(dir string, graph *DependencyGraph) error {
	specLinkRegex := regexp.MustCompile(`@spec-link\s+\[?\[?([^\]\s]+)\]?\]?`)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		
		// Skip hidden dirs
		if strings.Contains(path, "/.") {
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

func init() {
	rootCmd.AddCommand(crawlCmd)
	crawlCmd.Flags().String("src", "", "Path to the source code directory")
	crawlCmd.Flags().Bool("gaps", false, "Identify orphaned STABLE atoms")
	crawlCmd.Flags().String("docs", "", "Override docs directory")
}
