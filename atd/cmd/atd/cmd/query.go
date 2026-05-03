package cmd

// @spec-link [[service_atd_query]]

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Search ATD atoms by metadata fields",
	Long: `Search ATD atoms by metadata fields using regex or keyword matches.
Outputs a JSON list of matching atoms.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		field, _ := cmd.Flags().GetString("field")
		search, _ := cmd.Flags().GetString("search")
		pathsOnly, _ := cmd.Flags().GetBool("paths-only")

		text, err := runQuery(field, search, pathsOnly)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func runQuery(field, search string, pathsOnly bool) (string, error) {
	if search == "" {
		return "", fmt.Errorf("--search parameter is required")
	}

	explorer := exploration.NewExplorer(config.ProjectRoot(), config.DocsDir())
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	matches := explorer.Query(field, search)

	if pathsOnly {
		var paths []string
		for _, m := range matches {
			paths = append(paths, m.FilePath)
		}
		output, _ := json.MarshalIndent(paths, "", "  ")
		return string(output), nil
	}

	output, _ := json.MarshalIndent(matches, "", "  ")
	return string(output), nil
}

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringP("field", "f", "", "Metadata field to search (e.g., id, human_name, status). Omit to search all fields.")
	queryCmd.Flags().StringP("search", "s", "", "Keyword to match")
	queryCmd.Flags().BoolP("paths-only", "P", false, "Return only a list of file paths")
}
