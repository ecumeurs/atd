package cmd
// @spec-link [[service_atd_audit]]

import (
	"fmt"
	"atd-tools/config"
	"atd-tools/pkg/audit"
	"github.com/spf13/cobra"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit atoms for bloat and collisions",
	Long: `Audit performs structural and semantic analysis of ATD atoms.

Phase 1 (Bloat Detection): Uses LLM to identify atoms containing compound rules.
Phase 2 (Collision Detection): Uses embeddings to find semantic overlaps between atoms.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		threshold, _ := cmd.Flags().GetFloat64("threshold")
		docsDir, _ := cmd.Flags().GetString("docs")
		workspace, _ := cmd.Flags().GetBool("workspace")
		atomPath, _ := cmd.Flags().GetString("atom")
		codePath, _ := cmd.Flags().GetString("code")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		if threshold <= 0 {
			threshold = config.ActiveConfig.DiffSimilarityThreshold
			if threshold <= 0 {
				threshold = 0.85
			}
		}

		var report *audit.AuditReport
		var err error
		if atomPath != "" {
			// Narrower single-atom compliance check: scope to exactly this
			// atom file instead of sweeping every atom under --docs.
			if codePath != "" {
				fmt.Printf("Note: --code %q is not yet wired to a code-vs-atom compliance comparison; scoping audit to --atom only.\n", codePath)
			}
			report, err = audit.RunScopedAudit(atomPath, threshold)
		} else {
			report, err = audit.RunFullAudit(docsDir, threshold, workspace)
		}
		if err != nil {
			return err
		}
		fmt.Println(report.Text)
		return nil
	},
}

// runFullAudit is the shared entry point for the atd_audit MCP handler.
func runFullAudit(docsDir string, threshold float64, workspace bool) (string, error) {
	report, err := audit.RunFullAudit(docsDir, threshold, workspace)
	if err != nil {
		return "", err
	}
	return report.Text, nil
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.Flags().Float64("threshold", 0, "Similarity threshold (0.0 - 1.0)")
	auditCmd.Flags().String("docs", "", "Path to docs directory")
	auditCmd.Flags().Bool("workspace", false, "Audit all atoms in workspace")
	auditCmd.Flags().String("code", "", "Snippet path for compliance check")
	auditCmd.Flags().String("atom", "", "Atom path for compliance check")
}