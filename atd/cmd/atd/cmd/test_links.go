package cmd
// @spec-link [[specification_atd_test_links]]

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var testLinksCmd = &cobra.Command{
	Use:   "test-links",
	Short: "Audit @test-link tags in source files",
	Long: `Scan source files for @test-link [[ATOM_ID]] tags and 
report the mapping between atoms and their verification tests.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath, _ := cmd.Flags().GetString("src")
		targetAtom, _ := cmd.Flags().GetString("atom")
		docsDir, _ := cmd.Flags().GetString("docs")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		text, err := runTestLinks(srcPath, targetAtom, docsDir)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func runTestLinks(srcPath, targetAtom, docsDir string) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), docsDir)
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	links := explorer.GetTestLinks()

	// Filter if --atom is specified
	if targetAtom != "" {
		filtered := []exploration.TestLink{}
		for _, l := range links {
			if l.AtomID == targetAtom {
				filtered = append(filtered, l)
			}
		}
		links = filtered
	}

	output, _ := json.MarshalIndent(links, "", "  ")
	return string(output), nil
}

func init() {
	rootCmd.AddCommand(testLinksCmd)
	testLinksCmd.Flags().String("src", ".", "Path to source code")
	testLinksCmd.Flags().String("atom", "", "Focus on a specific Atom ID")
	testLinksCmd.Flags().String("docs", "", "Override docs directory")
}
