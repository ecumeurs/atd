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

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Search ATD atoms by metadata fields",
	Long: `Search ATD atoms by metadata fields using regex or keyword matches.
Outputs a JSON list of matching file paths.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		field, _ := cmd.Flags().GetString("field")
		search, _ := cmd.Flags().GetString("search")
		docsDir, _ := cmd.Flags().GetString("docs")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		text, err := runQuery(docsDir, field, search)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func runQuery(docsDir, field, search string) (string, error) {
	if search == "" {
		return "", fmt.Errorf("--search parameter is required")
	}

	matches, err := searchAtoms(docsDir, field, search)
	if err != nil {
		return "", err
	}

	output, _ := json.MarshalIndent(matches, "", "  ")
	return string(output), nil
}

func searchAtoms(dir string, field string, term string) ([]string, error) {
	var matches []string
	yamlRegex := regexp.MustCompile(`(?s)^---[\r\n]+(.*?)[\r\n]+---`)
	fieldRegex := regexp.MustCompile(fmt.Sprintf(`(?m)^%s:\s*\[?(.*?)\]?$`, regexp.QuoteMeta(field)))

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".atom.md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		yamlMatch := yamlRegex.FindStringSubmatch(string(content))
		if len(yamlMatch) > 1 {
			frontmatter := yamlMatch[1]
			fieldMatch := fieldRegex.FindStringSubmatch(frontmatter)
			if len(fieldMatch) > 1 {
				value := fieldMatch[1]
				if strings.Contains(strings.ToLower(value), strings.ToLower(term)) {
					matches = append(matches, path)
				}
			}
		}
		return nil
	})

	return matches, err
}

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringP("field", "f", "id", "Metadata field to search (e.g., id, human_name, status)")
	queryCmd.Flags().StringP("search", "s", "", "Keyword or regex to match")
	queryCmd.Flags().String("docs", "", "Override docs directory")
}
