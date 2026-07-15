package cmd
// @spec-link [[service_atd_index]]

import (
	"fmt"
	"atd-tools/config"
	"atd-tools/pkg/indexer"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Build a semantic vector index",
	Long: `Build a semantic vector index of source code and/or ATD documents.

Uses nomic-embed-text to generate embeddings stored in SQLite.
Files unchanged since last indexing are automatically skipped.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := cmd.Flags().GetString("dir")
		dbPath, _ := cmd.Flags().GetString("db")
		mode, _ := cmd.Flags().GetString("mode")

		if targetDir == "" {
			targetDir = "."
		}
		if dbPath == "" {
			dbPath = config.IndexDBPath(config.DocsDir())
		}
		
		report, err := indexer.Index(targetDir, dbPath, mode)
		if err != nil {
			return err
		}
		fmt.Println(report.Text)
		return nil
	},
}

// runIndex is the shared entry point for the atd_index MCP handler.
func runIndex(targetDir, dbPath, mode string) (string, error) {
	report, err := indexer.Index(targetDir, dbPath, mode)
	if err != nil {
		return "", err
	}
	return report.Text, nil
}

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.Flags().StringP("dir", "d", ".", "Directory to crawl and index")
	indexCmd.Flags().String("db", "", "Path to SQLite database")
	indexCmd.Flags().String("mode", "code", "Index mode: code|docs|all")
}