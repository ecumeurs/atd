package cmd
// @spec-link [[mechanic_atd_weave]]

import (
	"fmt"
	"os"
	"path/filepath"
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

		text, err := runWeave(docsDir)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func runWeave(docsDir string) (string, error) {
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
		return "", err
	}

	// 2. Rewrite each file with updated dependents
	edited := 0

	for _, ref := range atomRefs {
		content, err := os.ReadFile(ref.Path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		var newLines []string
		inFrontmatter := false
		frontmatterEnd := -1
		dependentsLineIdx := -1
		dependentsEndIdx := -1

		// Find frontmatter and dependents section
		for i, line := range lines {
			if line == "---" {
				if !inFrontmatter && i == 0 {
					inFrontmatter = true
				} else if inFrontmatter {
					inFrontmatter = false
					frontmatterEnd = i
					break
				}
			}
			if inFrontmatter {
				if strings.HasPrefix(line, "dependents:") {
					dependentsLineIdx = i
					// Check if it's a multi-line list
					j := i + 1
					for j < len(lines) && strings.HasPrefix(lines[j], "  -") {
						j++
					}
					dependentsEndIdx = j - 1
				}
			}
		}

		if frontmatterEnd == -1 {
			continue
		}

		deps := parentToDependents[ref.ID]
		var newDepsLines []string
		if len(deps) == 0 {
			newDepsLines = append(newDepsLines, "dependents: []")
		} else {
			newDepsLines = append(newDepsLines, "dependents:")
			for _, dep := range deps {
				newDepsLines = append(newDepsLines, fmt.Sprintf("  - [[%s]]", dep))
			}
		}

		// Reconstruct file
		if dependentsLineIdx != -1 {
			// Replace existing dependents section
			newLines = append(newLines, lines[:dependentsLineIdx]...)
			newLines = append(newLines, newDepsLines...)
			newLines = append(newLines, lines[dependentsEndIdx+1:]...)
		} else {
			// Insert before the closing ---
			newLines = append(newLines, lines[:frontmatterEnd]...)
			newLines = append(newLines, newDepsLines...)
			newLines = append(newLines, lines[frontmatterEnd:]...)
		}

		newContent := strings.Join(newLines, "\n")
		if newContent != string(content) {
			if err := os.WriteFile(ref.Path, []byte(newContent), 0644); err != nil {
				return "", fmt.Errorf("failed to write %s: %v", ref.Path, err)
			}
			edited++
			if Verbose {
				fmt.Printf("Updated dependents for %s\n", ref.ID)
			}
		}
	}

	result := fmt.Sprintf("Link Weaving Complete. Updated %d files.", edited)
	config.Log("atd-weave", result)
	return result, nil
}

func init() {
	rootCmd.AddCommand(weaveCmd)
	weaveCmd.Flags().String("docs", "", "Override docs directory")
}
