package cmd

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var heatmapCmd = &cobra.Command{
	Use:   "heatmap",
	Short: "Visualize documentation and code heat maps",
	Long:  `Identify documentation hotspots, isolated atoms, and frequently modified areas.`,
}

var heatmapAtomCmd = &cobra.Command{
	Use:   "atom [file|id]",
	Short: "Display heat metrics for a specific atom",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := runHeatmapAtom(args[0])
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

var heatmapCodeCmd = &cobra.Command{
	Use:   "code [file]",
	Short: "Display heat metrics for a specific code file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := runHeatmapCode(args[0])
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

var heatmapProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Display project-wide heat map summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		layer, _ := cmd.Flags().GetString("layer")
		out, err := runHeatmapProject(layer)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

var heatmapReportCmd = &cobra.Command{
	Use:   "report",
	Short: "Export full heat map data to JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		outPath, _ := cmd.Flags().GetString("output")
		out, err := runHeatmapReport(outPath)
		if err != nil {
			return err
		}
		if outPath == "" {
			fmt.Println(out)
		}
		return nil
	},
}

func runHeatmapAtom(target string) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), config.DocsDir())
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	atomID := target
	if strings.HasSuffix(target, ".atom.md") {
		found := false
		for id, a := range explorer.Graph.Atoms {
			if strings.HasSuffix(a.FilePath, target) {
				atomID = id
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("atom file not found in graph: %s", target)
		}
	}

	res, err := explorer.GetHeatMapResult(atomID)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Atom:  %s\n", res.Atom))
	sb.WriteString(fmt.Sprintf("Layer: %s\n", res.Layer))
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString(fmt.Sprintf("Dependency Heat: %s %s\n", heatEmoji(res.DependencyState), strings.ToUpper(string(res.DependencyState))))
	sb.WriteString(fmt.Sprintf("                 Parents: %d, Dependents: %d\n", res.Metrics.Parents, res.Metrics.Dependents))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Code Heat:       %s %s\n", heatEmoji(res.CodeState), strings.ToUpper(string(res.CodeState))))
	sb.WriteString(fmt.Sprintf("                 Code Files Linked: %d\n", res.Metrics.CodeFilesLinked))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Update Heat:     %s %s\n", heatEmoji(res.UpdateState), strings.ToUpper(string(res.UpdateState))))
	sb.WriteString(fmt.Sprintf("                 Recent Updates: %d, Last: %s\n", res.Metrics.RecentUpdates, res.Metrics.LastUpdated))
	
	if len(res.Recommendations) > 0 {
		sb.WriteString("\nRecommendations:\n")
		for _, rec := range res.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}
	return sb.String(), nil
}

func runHeatmapCode(path string) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), config.DocsDir())
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	heat := explorer.GetCodeFileHeat(path)
	count := 0
	for _, sl := range explorer.SpecLinks {
		if sl.FilePath == path {
			count++
		}
	}

	return fmt.Sprintf("Code File: %s\nState:     %s %s\nMetrics:   %d atoms linked\n", 
		path, heatEmoji(heat), strings.ToUpper(string(heat)), count), nil
}

func runHeatmapProject(layer string) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), config.DocsDir())
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Project Heat Map (%s)\n", strings.ToUpper(layer)))
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString(fmt.Sprintf("%-30s | %-10s | %-10s\n", "Atom ID", "State", "Metrics"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	ids := make([]string, 0, len(explorer.Graph.Atoms))
	for id := range explorer.Graph.Atoms {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		res, _ := explorer.GetHeatMapResult(id)
		state := exploration.HeatCold
		metrics := ""

		switch layer {
		case "dependency":
			state = res.DependencyState
			metrics = fmt.Sprintf("P:%d D:%d", res.Metrics.Parents, res.Metrics.Dependents)
		case "code":
			state = res.CodeState
			metrics = fmt.Sprintf("L:%d", res.Metrics.CodeFilesLinked)
		case "updates":
			state = res.UpdateState
			metrics = fmt.Sprintf("U:%d", res.Metrics.RecentUpdates)
		default:
			state = res.DependencyState
			if heatRank(res.CodeState) > heatRank(state) { state = res.CodeState }
			if heatRank(res.UpdateState) > heatRank(state) { state = res.UpdateState }
			metrics = fmt.Sprintf("P:%d D:%d L:%d U:%d", res.Metrics.Parents, res.Metrics.Dependents, res.Metrics.CodeFilesLinked, res.Metrics.RecentUpdates)
		}

		sb.WriteString(fmt.Sprintf("%-30s | %-10s | %-10s\n", id, state, metrics))
	}
	return sb.String(), nil
}

func runHeatmapReport(outPath string) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), config.DocsDir())
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	results := make([]*exploration.HeatMapResult, 0, len(explorer.Graph.Atoms))
	ids := make([]string, 0, len(explorer.Graph.Atoms))
	for id := range explorer.Graph.Atoms {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		res, _ := explorer.GetHeatMapResult(id)
		results = append(results, res)
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}

	if outPath != "" {
		err := os.WriteFile(outPath, data, 0644)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Report written to %s", outPath), nil
	}
	return string(data), nil
}

func init() {
	rootCmd.AddCommand(heatmapCmd)
	heatmapCmd.AddCommand(heatmapAtomCmd)
	heatmapCmd.AddCommand(heatmapCodeCmd)
	heatmapCmd.AddCommand(heatmapProjectCmd)
	heatmapCmd.AddCommand(heatmapReportCmd)

	heatmapProjectCmd.Flags().String("layer", "all", "Heat layer: dependency, code, updates, all")
	heatmapReportCmd.Flags().String("output", "", "Output file path (default: stdout)")
}

func heatEmoji(state exploration.HeatState) string {

	switch state {
	case exploration.HeatCold:
		return "❄️ "
	case exploration.HeatOptimal:
		return "✅"
	case exploration.HeatWarm:
		return "⚡"
	case exploration.HeatHot:
		return "🔴"
	case exploration.HeatStable:
		return "⚪"
	default:
		return "❓"
	}
}

func heatRank(state exploration.HeatState) int {
	switch state {
	case exploration.HeatCold, exploration.HeatStable:
		return 0
	case exploration.HeatOptimal:
		return 1
	case exploration.HeatWarm:
		return 2
	case exploration.HeatHot:
		return 3
	default:
		return -1
	}
}
