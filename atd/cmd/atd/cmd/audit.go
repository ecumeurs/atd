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
Phase 2 (Collision Detection): Uses embeddings to find semantic overlaps between atoms.

Audit does not compare an atom against a piece of code. For that, use
"atd map --atom <atom> --file <code>" instead.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		threshold, _ := cmd.Flags().GetFloat64("threshold")
		docsDir, _ := cmd.Flags().GetString("docs")
		workspace, _ := cmd.Flags().GetBool("workspace")
		atomPath, _ := cmd.Flags().GetString("atom")
		codePath, _ := cmd.Flags().GetString("code")
		concurrency, _ := cmd.Flags().GetInt("concurrency")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		if threshold <= 0 {
			threshold = config.ActiveConfig.DiffSimilarityThreshold
			if threshold <= 0 {
				threshold = 0.85
			}
		}

		report, err := dispatchAudit(docsDir, threshold, workspace, atomPath, codePath, concurrency)
		if err != nil {
			return err
		}
		fmt.Println(report.Text)
		return nil
	},
}

// dispatchAudit is the single place that decides which of pkg/audit's entry
// points to call based on --atom/--code (CLI) or atom/code (MCP atd_audit):
// atom set + code empty scopes to that one file (RunScopedAudit); atom set +
// code set is rejected, since pkg/audit has no atom-vs-code compliance
// comparison capability -- that's "atd map --atom <atom> --file <code>"
// (MCP: atd_recon), not audit; neither set sweeps the whole docs directory
// (RunFullAudit), unchanged from before. Shared by the CLI RunE and the MCP
// handler so the dispatch logic and the --atom+--code error message only
// exist once.
func dispatchAudit(docsDir string, threshold float64, workspace bool, atomPath, codePath string, concurrency int) (*audit.AuditReport, error) {
	if atomPath != "" && codePath != "" {
		return nil, fmt.Errorf("audit does not compare an atom against code; use \"atd map --atom %s --file %s\" for a single-atom-vs-single-file compliance check instead", atomPath, codePath)
	}
	if atomPath != "" {
		return audit.RunScopedAudit(atomPath, threshold, concurrency)
	}
	return audit.RunFullAudit(docsDir, threshold, workspace, concurrency)
}

// runFullAudit is the shared entry point for the atd_audit MCP handler.
func runFullAudit(docsDir string, threshold float64, workspace bool, atomPath, codePath string, concurrency int) (string, error) {
	report, err := dispatchAudit(docsDir, threshold, workspace, atomPath, codePath, concurrency)
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
	auditCmd.Flags().String("code", "", "Not supported by audit (no atom-vs-code compliance check here); combining with --atom is an error -- use \"atd map --atom <atom> --file <code>\" instead")
	auditCmd.Flags().String("atom", "", "Scope the audit's bloat/collision analysis to a single atom file instead of sweeping --docs")
	auditCmd.Flags().Int("concurrency", 0, "Max files audited in parallel during bloat detection (0 = default)")
}