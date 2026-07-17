package cmd

// @spec-link [[service_atd_stats]]

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/workspace"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

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

func runStats(srcPath, docsDir string, workspaceFlag bool) (string, error) {
	// Resolve against the loaded config's project root, not the process cwd
	// (test_atd_07_26.md §2.1/§3.6, §8.3 defect #7). NewExplorer already
	// falls back to config.ProjectRoot() when srcPath is "" -- passing the
	// literal "." here instead bypassed that and anchored every stats run to
	// whatever directory the process happened to be started/invoked from
	// (an MCP server's cwd, a `go test` binary's package dir, etc.), a
	// read-only cousin of the I-1 cwd-anchoring hazard.
	explorer := exploration.NewExplorer(srcPath, docsDir)

	if workspaceFlag {
		ws, err := workspace.LoadWorkspace(".")
		if err != nil {
			return "", fmt.Errorf("failed to load workspace: %v", err)
		}
		explorer.Workspace = ws
		idx, _ := ws.BuildIndex()
		explorer.Index = idx
		explorer.Resolver = workspace.NewResolver(ws, idx, "")

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
		ByType:   make(map[string]int),
		ByStatus: make(map[string]int),
		ByLayer:  make(map[string]int),
		Project:  config.ActiveConfig.ActiveProject,
	}

	// In project (non-workspace) scope, the link crawl resolves every
	// @spec-link it encounters via ResolveAtom, which inserts the resolved
	// atom into explorer.Graph.Atoms even when it lives in another project
	// (see exploration.go ResolveAtom). Those cross-project atoms are usually
	// keyed with a "project:" prefix, unlike local atoms which have a bare
	// key -- but a bare (unqualified) @spec-link that resolves workspace-wide
	// can still land in the graph under a bare key if the resolver's
	// "current project" detection is off. To catch that case too, we also
	// check FilePath: atoms found by this project's own file crawl always
	// live under explorer.ProjectRoot (see Explorer.Load, which joins
	// ProjectRoot with each relative path it walks), while atoms pulled in
	// via the workspace atom index (ResolveAtom) carry an AtomLocation.Path
	// rooted at a *different* project's directory (see
	// workspace.AtomIndex.BuildIndex) and so fall outside it. Containment
	// (via filepath.Rel) is therefore the reliable cross-project signal,
	// independent of how the id happened to get keyed.
	//
	// This used to be a plain filepath.IsAbs(node.FilePath) check, which
	// only worked by accident: it relied on explorer.ProjectRoot itself
	// being a relative "." (defect #7 -- see the NewExplorer call above),
	// which made every local FilePath relative too and every workspace-index
	// FilePath (always absolute) look distinctly "foreign". Now that
	// ProjectRoot is resolved to config.ProjectRoot() (always absolute),
	// every local FilePath is absolute as well, so IsAbs alone can no longer
	// distinguish them -- containment relative to explorer.ProjectRoot can.
	//
	// If we don't exclude cross-project atoms here, every aggregate below
	// (total, by-type, by-status, by-layer, stable/orphan/implemented
	// counts) ends up describing a mixed population that is neither "this
	// project's atoms" nor "the workspace's atoms". Workspace scope
	// legitimately aggregates atoms from every project via
	// LoadWorkspace/the index, so no filter is applied there.
	isLocalAtom := func(id string, node *atom.AtomData) bool {
		if workspaceFlag {
			return true
		}
		if strings.Contains(id, ":") {
			return false
		}
		if node.FilePath != "" {
			rel, err := filepath.Rel(explorer.ProjectRoot, node.FilePath)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return false
			}
		}
		return true
	}

	var stableCount int
	var candidateCount int

	for id, node := range graph.Atoms {
		if !isLocalAtom(id, node) {
			continue
		}

		report.TotalAtoms++
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
			}
			// Gate on actual implementation, not merely "not an orphan":
			// IsOrphan returns false for many unimplemented atoms (BUSINESS
			// layer exception, excluded types, parents "covered by proxy"),
			// so "not an orphan" != "implemented". Without this, a STABLE
			// atom with zero linked code inflates this count past
			// ImplementedTotalCount, which is mathematically impossible for
			// honest labels.
			if isImplemented {
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
		for id, node := range graph.Atoms {
			if !isLocalAtom(id, node) {
				continue
			}
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
