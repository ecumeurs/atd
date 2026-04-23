package cmd
// @spec-link [[mechanic_atd_weave]]

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"fmt"

	"github.com/spf13/cobra"
)

var weaveCmd = &cobra.Command{
	Use:   "weave",
	Short: "Bi-directionally link ATD atoms based on parent declarations",
	Long: `Crawl all ATD atoms to discover parent relationships and 
automatically update the 'dependents' field in each atom file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		workspace, _ := cmd.Flags().GetBool("workspace")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		text, err := runWeave(docsDir, workspace)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func runWeave(docsDir string, workspace bool) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), docsDir)
	if workspace {
		if err := explorer.LoadWorkspace(false); err != nil {
			return "", err
		}
	}
	return explorer.Weave()
}

func init() {
	rootCmd.AddCommand(weaveCmd)
	weaveCmd.Flags().String("docs", "", "Override docs directory")
	weaveCmd.Flags().Bool("workspace", false, "Weave entire workspace")
}
