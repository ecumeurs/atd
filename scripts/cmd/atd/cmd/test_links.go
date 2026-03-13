package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
	"github.com/spf13/cobra"
)

type TestLink struct {
	AtomID   string `json:"atom_id"`
	TestFile string `json:"test_file"`
	Line     int    `json:"line"`
}

var testLinksCmd = &cobra.Command{
	Use:   "test-links",
	Short: "Audit @test-link tags in source files",
	Long: `Scan source files for @test-link [[ATOM_ID]] tags and 
report the mapping between atoms and their verification tests.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		srcPath, _ := cmd.Flags().GetString("src")
		targetAtom, _ := cmd.Flags().GetString("atom")
		docsDir, _ := cmd.Flags().GetString("docs")

		if srcPath == "" {
			srcPath = "."
		}
		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		links := []TestLink{}
		testLinkRegex := regexp.MustCompile(`@test-link\s+\[?\[?([^\]\s]+)\]?\]?`)

		// 1. Walk source files looking for tags
		err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || strings.Contains(path, "/.") {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			lines := strings.Split(string(content), "\n")
			for i, line := range lines {
				matches := testLinkRegex.FindAllStringSubmatch(line, -1)
				for _, match := range matches {
					if len(match) > 1 {
						links = append(links, TestLink{
							AtomID:   match[1],
							TestFile: path,
							Line:     i + 1,
						})
					}
				}
			}
			return nil
		})

		if err != nil {
			return err
		}

		// 2. Filter or expand if --atom is specified
		if targetAtom != "" {
			// Find all related atoms (parents/dependents) to show full test coverage
			related := make(map[string]bool)
			related[targetAtom] = true
			
			// We might want to crawl the hierarchy here, but for a deterministic tool,
			// just filtering for the specific atom is the primary goal.
			// Let's at least filter for the target atom links found.
			filtered := []TestLink{}
			for _, l := range links {
				if l.AtomID == targetAtom {
					filtered = append(filtered, l)
				}
			}
			links = filtered
		}

		output, _ := json.MarshalIndent(links, "", "  ")
		fmt.Println(string(output))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testLinksCmd)
	testLinksCmd.Flags().String("src", ".", "Path to source code")
	testLinksCmd.Flags().String("atom", "", "Focus on a specific Atom ID")
	testLinksCmd.Flags().String("docs", "", "Override docs directory")
}
