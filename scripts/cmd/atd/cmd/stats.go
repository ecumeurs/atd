package cmd
// @spec-link [[atd_stats]]

import (
	"encoding/json"
	"fmt"

	"atd-tools/config"
	"github.com/spf13/cobra"
)

type StatsReport struct {
	TotalAtoms    int            `json:"total_atoms"`
	ByType        map[string]int `json:"by_type"`
	ByStatus      map[string]int `json:"by_status"`
	ByLayer       map[string]int `json:"by_layer"`
	CoverageRatio float64        `json:"coverage_ratio"`
	OrphanCount   int            `json:"orphan_count"`
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Generate quantitative documentation health metrics",
	Long:  `Generate quantitative documentation health metrics including total atoms, breakdown by type, status, domain, coverage ratio, and orphan count.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath, _ := cmd.Flags().GetString("src")
		docsDir, _ := cmd.Flags().GetString("docs")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		out, err := runStats(srcPath, docsDir)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func runStats(srcPath, docsDir string) (string, error) {
	graph := &DependencyGraph{
		Atoms: make(map[string]*AtomNode),
	}

	if err := crawlDocs(docsDir, graph); err != nil {
		return "", err
	}

	if srcPath != "" {
		if err := crawlSrc(srcPath, graph); err != nil {
			return "", err
		}
	}

	report := &StatsReport{
		TotalAtoms: len(graph.Atoms),
		ByType:     make(map[string]int),
		ByStatus:   make(map[string]int),
		ByLayer:    make(map[string]int),
	}

	var stableCount int
	var implementedCount int

	for _, node := range graph.Atoms {
		report.ByType[node.Type]++
		report.ByStatus[node.Status]++

		if node.Layer != "" {
			report.ByLayer[node.Layer]++
		} else {
			report.ByLayer["<unspecified>"]++
		}

		if node.Status == "STABLE" {
			stableCount++
			if len(node.Implementations) == 0 {
				report.OrphanCount++
			} else {
				implementedCount++
			}
		}
	}

	if stableCount > 0 {
		report.CoverageRatio = float64(implementedCount) / float64(stableCount)
	}

	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().String("src", "", "Path to the source code directory to calculate coverage")
	statsCmd.Flags().String("docs", "", "Override docs directory")
}
