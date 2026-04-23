package cmd
// @spec-link [[service_atd_stats]]

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type StatsReport struct {
	TotalAtoms             int            `json:"total_atoms"`
	ByType                 map[string]int `json:"by_type"`
	ByStatus               map[string]int `json:"by_status"`
	ByLayer                map[string]int `json:"by_layer"`
	CoverageRatio          float64        `json:"coverage_ratio"`
	CandidateCoverageRatio float64        `json:"candidate_coverage_ratio"`
	OrphanCount            int            `json:"orphan_count"`
	ImplementedStableCount int            `json:"implemented_stable_count"`
	ImplementedTotalCount  int            `json:"implemented_total_count"`
	Project                string         `json:"project,omitempty"`
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Generate quantitative documentation health metrics",
	Long:  `Generate quantitative documentation health metrics including total atoms, breakdown by type, status, domain, coverage ratio, and orphan count.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath, _ := cmd.Flags().GetString("src")
		docsDir, _ := cmd.Flags().GetString("docs")
		workspace, _ := cmd.Flags().GetBool("workspace")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		out, err := runStats(srcPath, docsDir, workspace)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func runStats(srcPath, docsDir string, workspace bool) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), docsDir)
	if workspace {
		if err := explorer.LoadWorkspace(false); err != nil {
			return "", err
		}
	} else {
		if err := explorer.Load(false); err != nil {
			return "", err
		}
	}

	graph := explorer.GetGraph()

	report := &StatsReport{
		TotalAtoms: len(graph.Atoms),
		ByType:     make(map[string]int),
		ByStatus:   make(map[string]int),
		ByLayer:    make(map[string]int),
		Project:    config.ActiveConfig.ActiveProject,
	}

	var stableCount int
	var candidateCount int

	for _, node := range graph.Atoms {
		report.ByType[node.Type]++
		report.ByStatus[node.Status]++

		if node.Layer != "" {
			report.ByLayer[node.Layer]++
		} else {
			report.ByLayer["<unspecified>"]++
		}

		isImplemented := len(node.Implementations) > 0
		if isImplemented {
			report.ImplementedTotalCount++
		}

		if node.Status == "STABLE" {
			stableCount++
			if explorer.IsOrphan(node) {
				report.OrphanCount++
			} else {
				report.ImplementedStableCount++
			}
		}
		
		if node.Status == "STABLE" || node.Status == "REVIEW" {
			candidateCount++
		}
	}

	if stableCount > 0 {
		report.CoverageRatio = float64(report.ImplementedStableCount) / float64(stableCount)
	}
	
	if candidateCount > 0 {
		implementedCandidate := 0
		for _, node := range graph.Atoms {
			if (node.Status == "STABLE" || node.Status == "REVIEW") && len(node.Implementations) > 0 {
				implementedCandidate++
			}
		}
		report.CandidateCoverageRatio = float64(implementedCandidate) / float64(candidateCount)
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
	statsCmd.Flags().Bool("workspace", false, "Aggregate stats from all projects in workspace")
}
