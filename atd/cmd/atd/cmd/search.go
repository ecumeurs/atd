package cmd
// @spec-link [[service_atd_search]]

import (
	"fmt"
	"path/filepath"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/ollama"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search semantically or via keyword",
	Long: `Search the indexed codebase semantically using Nomic embeddings,
or perform a literal keyword search across files.

Semantic search requires a previously built index (via 'atd index').
Grep mode performs a direct search on the filesystem.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query, _ := cmd.Flags().GetString("query")
		keyword, _ := cmd.Flags().GetString("grep")

		// Pre-check: Ensure embedding model is available if using semantic search (no IDE fallback)
		if query != "" && keyword == "" {
			res, err := ollama.ResolveProvider("embed")
			if err != nil {
				return fmt.Errorf("error resolving embedding provider: %v", err)
			}
			if res.IsIDE {
				return fmt.Errorf("semantic search requires a local or remote Ollama provider with the embedding model (e.g. nomic-embed-text) installed. No embedding model was found")
			}
		}

		dbPath, _ := cmd.Flags().GetString("db")
		limit, _ := cmd.Flags().GetInt("limit")
		scope, _ := cmd.Flags().GetString("scope")
		pathsOnly, _ := cmd.Flags().GetBool("paths-only")

		if query == "" && keyword == "" {
			return fmt.Errorf("either --query (semantic) or --grep (keyword) must be specified")
		}

		if keyword != "" {
			out, err := runGrepSearch(keyword, pathsOnly)
			if err != nil {
				return err
			}
			fmt.Println(out)
			return nil
		}

		if dbPath == "" {
			dbPath = config.IndexDBPath(config.DocsDir())
		}

		out, err := runSemanticSearch(query, dbPath, limit, scope, pathsOnly)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func runSemanticSearch(query, dbPath string, limit int, scope string, pathsOnly bool) (string, error) {
	var output strings.Builder
	opts := exploration.SearchOptions{
		Query:  query,
		DBPath: dbPath,
		Limit:  limit,
		Scope:  scope,
	}
	results, err := exploration.Search(opts)
	if err != nil {
		return "", err
	}

	if pathsOnly {
		uniquePaths := make(map[string]bool)
		for _, res := range results {
			abs := res.FilePath
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(config.ProjectRoot(), abs)
			}
			uniquePaths[abs] = true
		}
		for path := range uniquePaths {
			output.WriteString(path)
			output.WriteString("\n")
		}
		return output.String(), nil
	}

	output.WriteString(fmt.Sprintf("--- Top %d Semantic Matches ---\n\n", limit))
	if len(results) == 0 {
		output.WriteString("(0 results)\n")
		return output.String(), nil
	}
	for i, res := range results {
		output.WriteString(fmt.Sprintf("[Match %d] File: %s (Similarity: %.4f)\n%s\n\n", i+1, res.FilePath, res.Similarity, res.ChunkText))
	}
	return output.String(), nil
}

func runGrepSearch(keyword string, pathsOnly bool) (string, error) {
	var output strings.Builder
	opts := exploration.SearchOptions{
		Grep: keyword,
		Root: config.ProjectRoot(),
	}
	results, err := exploration.Search(opts)
	if err != nil {
		return "", err
	}

	if pathsOnly {
		uniquePaths := make(map[string]bool)
		for _, res := range results {
			abs := res.FilePath
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(config.ProjectRoot(), abs)
			}
			uniquePaths[abs] = true
		}
		for path := range uniquePaths {
			output.WriteString(path)
			output.WriteString("\n")
		}
		return output.String(), nil
	}

	for _, res := range results {
		output.WriteString(fmt.Sprintf("Grep: Found match in %s\n", res.FilePath))
	}
	output.WriteString(fmt.Sprintf("Sweeping complete. Found %d matches for '%s'.\n", len(results), keyword))
	return output.String(), nil
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("query", "q", "", "Semantic search query")
	searchCmd.Flags().String("db", "", "Path to SQLite database")
	searchCmd.Flags().IntP("limit", "l", 5, "Number of results to return")
	searchCmd.Flags().StringP("grep", "g", "", "Literal keyword search (grep mode)")
	searchCmd.Flags().String("scope", "all", "Search scope: code|docs|all")
	searchCmd.Flags().BoolP("paths-only", "P", false, "Return only a list of file paths")
}
