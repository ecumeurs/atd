package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"github.com/spf13/cobra"
)

var weaveCmd = &cobra.Command{
	Use:   "weave",
	Short: "Bi-directionally link ATD atoms based on parent declarations",
	Long: `Crawl all ATD atoms to discover parent relationships and 
automatically update the 'dependents' field in each atom file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		type AtomRef struct {
			Path string
			ID   string
		}

		atomRefs := []AtomRef{}
		parentToDependents := make(map[string][]string)

		// 1. Discover all IDs and their declared parents
		err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".atom.md") {
				return nil
			}

			a, err := atom.Parse(path)
			if err != nil {
				return nil
			}

			if a.ID == "" {
				return nil
			}

			atomRefs = append(atomRefs, AtomRef{Path: path, ID: a.ID})
			for _, p := range a.Parents {
				parentToDependents[p] = append(parentToDependents[p], a.ID)
			}
			return nil
		})

		if err != nil {
			return err
		}

		// 2. Rewrite each file with updated dependents
		// We use regex to specifically replace only the dependents line to minimize diff churn
		dependentsRegex := regexp.MustCompile(`(?m)^dependents:\s*\[.*?\]`)
		edited := 0

		for _, ref := range atomRefs {
			content, err := os.ReadFile(ref.Path)
			if err != nil {
				continue
			}

			deps := parentToDependents[ref.ID]
			newDepsLine := "dependents: []"
			if len(deps) > 0 {
				formattedDeps := []string{}
				for _, dep := range deps {
					formattedDeps = append(formattedDeps, fmt.Sprintf("[[%s]]", dep))
				}
				newDepsLine = fmt.Sprintf("dependents: [%s]", strings.Join(formattedDeps, ", "))
			}

			if dependentsRegex.Match(content) {
				newContent := dependentsRegex.ReplaceAllString(string(content), newDepsLine)
				if newContent != string(content) {
					if err := os.WriteFile(ref.Path, []byte(newContent), 0644); err != nil {
						return fmt.Errorf("failed to write %s: %v", ref.Path, err)
					}
					edited++
					if Verbose {
						fmt.Printf("Updated dependents for %s\n", ref.ID)
					}
				}
			}
		}

		fmt.Printf("Link Weaving Complete. Updated %d files.\n", edited)
		config.Log("atd-weave", fmt.Sprintf("Updated %d files", edited))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(weaveCmd)
	weaveCmd.Flags().String("docs", "", "Override docs directory")
}
