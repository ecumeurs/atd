package cmd

// @spec-link [[service_atd_check_coverage]]

import (
	"fmt"
	"os"
	"atd-tools/pkg/coverage"
	"github.com/spf13/cobra"
)

var coverageCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check ATD coverage: impl links, test links, and optional semantic compliance",
	Long: `Check ATD atom coverage for implementation and test links.

Modes:
  diff (default): git-diff driven — checks code changes and atom changes bidirectionally
  --atom <id>:    full coverage for one atom
  --file <path>:  coverage for one source file
  --full:         entire project coverage

Pass --semantic to also run LLM compliance checks per @spec-link (slow).`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		atomID, _ := cmd.Flags().GetString("atom")
		filePath, _ := cmd.Flags().GetString("file")
		docsDir, _ := cmd.Flags().GetString("docs")
		full, _ := cmd.Flags().GetBool("full")
		semantic, _ := cmd.Flags().GetBool("semantic")
		outPath, _ := cmd.Flags().GetString("out")

		mode := "diff"

		report, err := coverage.GenerateReport(mode, atomID, filePath, docsDir, full, semantic, args)
		if err != nil {
			return err
		}

		if outPath != "" {
			if err := os.WriteFile(outPath, []byte(report.Text), 0644); err != nil {
				return fmt.Errorf("failed to write report to %s: %v", outPath, err)
			}
			fmt.Printf("Coverage report written to %s\n", outPath)
		} else {
			fmt.Print(report.Text)
		}
		return nil
	},
}

// runCoverageCheck is the shared entry point for the MCP handlers (atd_check,
// atd_test_links); it returns the same report text the CLI prints.
func runCoverageCheck(mode, atomID, filePath, docsDir string, full, semantic bool, gitArgs []string) (string, error) {
	report, err := coverage.GenerateReport(mode, atomID, filePath, docsDir, full, semantic, gitArgs)
	if err != nil {
		return "", err
	}
	return report.Text, nil
}

func init() {
	rootCmd.AddCommand(coverageCheckCmd)
	coverageCheckCmd.Flags().String("atom", "", "Check full coverage for one atom ID")
	coverageCheckCmd.Flags().String("file", "", "Check coverage for a specific source file")
	coverageCheckCmd.Flags().Bool("full", false, "Check entire project coverage")
	coverageCheckCmd.Flags().Bool("semantic", false, "Run LLM compliance check per @spec-link (slow)")
	coverageCheckCmd.Flags().String("out", "", "Write report to file")
	coverageCheckCmd.Flags().String("docs", "", "Override docs directory")
}