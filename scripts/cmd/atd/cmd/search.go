package cmd
// @spec-link [[service_atd_search]]

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/cosine"
	"atd-tools/pkg/ollama"
	"github.com/spf13/cobra"
	_ "github.com/mattn/go-sqlite3"
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

		if query == "" && keyword == "" {
			return fmt.Errorf("either --query (semantic) or --grep (keyword) must be specified")
		}

		if keyword != "" {
			return runGrepSearch(keyword)
		}

		if dbPath == "" {
			dbPath = filepath.Join(config.DocsDir(), ".atd_index.db")
		}

		return runSemanticSearch(query, dbPath, limit, scope)
	},
}

func runGrepSearch(keyword string) error {
	root := config.ProjectRoot()
	matches := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || strings.Contains(path, "/.git/") || strings.Contains(path, "/.atd") {
			return nil
		}

		ext := filepath.Ext(path)
		if !config.ActiveConfig.SupportedExtensions[ext] && !strings.HasSuffix(path, ".atom.md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if strings.Contains(string(content), keyword) {
			rel, _ := filepath.Rel(root, path)
			fmt.Printf("Grep: Found match in %s\n", rel)
			matches++
		}
		return nil
	})

	fmt.Printf("Sweeping complete. Found %d matches for '%s'.\n", matches, keyword)
	return err
}

func runSemanticSearch(query, dbPath string, limit int, scope string) error {
	// 1. Get Query Vector
	queryEmb, err := ollama.QueryEmbed(query)
	if err != nil {
		return fmt.Errorf("failed to embed query: %v", err)
	}

	// 2. Open DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT file_path, chunk_text, embedding FROM atom_index`)
	if err != nil {
		return fmt.Errorf("failed to query DB: %v (Did you run 'atd index'?)", err)
	}
	defer rows.Close()

	type Result struct {
		filePath   string
		chunkText  string
		similarity float64
	}
	var results []Result

	for rows.Next() {
		var filePath, chunkText string
		var embJSON []byte
		if err := rows.Scan(&filePath, &chunkText, &embJSON); err != nil {
			continue
		}

		// Filter by scope
		isAtom := strings.HasSuffix(filePath, ".atom.md")
		switch scope {
		case "code":
			if isAtom { continue }
		case "docs":
			if !isAtom { continue }
		}

		var chunkEmb []float64
		if err := json.Unmarshal(embJSON, &chunkEmb); err != nil {
			continue
		}

		sim := cosine.Similarity(queryEmb, chunkEmb)
		results = append(results, Result{filePath, chunkText, sim})
	}

	// Sort results by similarity descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].similarity > results[j].similarity
	})

	// Print top results
	fmt.Printf("--- Top %d Semantic Matches ---\n\n", limit)
	max := limit
	if len(results) < max {
		max = len(results)
	}
	for i := 0; i < max; i++ {
		fmt.Printf("[Match %d] File: %s (Similarity: %.4f)\n%s\n\n", i+1, results[i].filePath, results[i].similarity, results[i].chunkText)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("query", "q", "", "Semantic search query")
	searchCmd.Flags().String("db", "", "Path to SQLite database")
	searchCmd.Flags().IntP("limit", "l", 5, "Number of results to return")
	searchCmd.Flags().StringP("grep", "g", "", "Literal keyword search (grep mode)")
	searchCmd.Flags().String("scope", "all", "Search scope: code|docs|all")
}
