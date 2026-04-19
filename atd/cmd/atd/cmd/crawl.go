package cmd

// @spec-link [[service_atd_crawl]]

import (
	"encoding/json"
	"fmt"

	"atd-tools/config"
	"atd-tools/pkg/exploration"

	"github.com/spf13/cobra"
)

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

		text, err := runCrawl(srcPath, docsDir, gaps)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func runCrawl(srcPath, docsDir string, gaps bool) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), docsDir)
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	graph := explorer.GetGraph()

	var output []byte
	if gaps {
		report := GapReport{
			OrphanedAtoms: []string{},
		}
		for id, node := range graph.Atoms {
			if explorer.IsOrphan(node) {
				report.OrphanedAtoms = append(report.OrphanedAtoms, id)
			}
		}
		output, _ = json.MarshalIndent(report, "", "  ")
	} else {
		output, _ = json.MarshalIndent(graph, "", "  ")
	}

	return string(output), nil
}

// func crawlDocs and crawlSrc removed

func init() {
	rootCmd.AddCommand(crawlCmd)
	crawlCmd.Flags().String("src", "", "Path to the source code directory")
	crawlCmd.Flags().Bool("gaps", false, "Identify orphaned STABLE atoms")
	crawlCmd.Flags().String("docs", "", "Override docs directory")
}
