package cmd

// @spec-link [[mechanic_atd_weave]]

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/workspace"
	"fmt"

	"github.com/spf13/cobra"
)

var weaveCmd = &cobra.Command{
	Use:   "weave",
	Short: "Bi-directionally link ATD atoms based on parent declarations",
	Long: `Crawl all ATD atoms to discover parent relationships and
automatically update the 'dependents' field in each atom file.

When invoked from inside a workspace, weave operates across all projects
and canonicalizes cross-project references using the [[project:atom_id]]
form. The --workspace flag forces this mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		wsFlag, _ := cmd.Flags().GetBool("workspace")

		if wsFlag {
			if _, err := workspace.LoadWorkspace("."); err != nil {
				return fmt.Errorf("no workspace found but --workspace flag used")
			}
		}

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		explorer := exploration.NewExplorer(config.ProjectRoot(), docsDir)
		text, err := explorer.Weave()
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(weaveCmd)
	weaveCmd.Flags().String("docs", "", "Override docs directory")
	weaveCmd.Flags().Bool("workspace", false, "Force workspace mode (fail if no workspace)")
}
