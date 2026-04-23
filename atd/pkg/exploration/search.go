package exploration

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
	_ "github.com/mattn/go-sqlite3"
)

type SearchResult struct {
	FilePath   string  `json:"file_path"`
	ChunkText  string  `json:"chunk_text"`
	Similarity float64 `json:"similarity"`
}

type SearchOptions struct {
	Query   string
	Grep    string
	DBPath  string
	Limit   int
	Scope   string
	Root    string
}

func Search(opts SearchOptions) ([]SearchResult, error) {
	if opts.Grep != "" {
		return GrepSearch(opts.Grep, opts.Root)
	}
	if opts.Query != "" {
		results, err := SemanticSearch(opts.Query, opts.DBPath, opts.Limit, opts.Scope)
		if err != nil {
			// Fallback to grep search if semantic search fails (e.g. LLM provider offline)
			return GrepSearch(opts.Query, opts.Root)
		}
		return results, nil
	}
	return nil, fmt.Errorf("either query or grep must be specified")
}

func GrepSearch(keyword, root string) ([]SearchResult, error) {
	var results []SearchResult
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if info.IsDir() || strings.Contains(path, "/.git/") || strings.Contains(path, "/.atd") {
			return nil
		}

		ext := filepath.Ext(path)
		if !config.ActiveConfig.SupportedExtensions[ext] && !strings.HasSuffix(path, ".atom.md") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil { return nil }

		if strings.Contains(string(content), keyword) {
			rel, _ := filepath.Rel(root, path)
			results = append(results, SearchResult{
				FilePath:  rel,
				ChunkText: "Keyword match found.", // Grep mode doesn't provide chunks easily here
			})
		}
		return nil
	})
	return results, err
}

func SemanticSearch(query, dbPath string, limit int, scope string) ([]SearchResult, error) {
	queryEmb, err := ollama.QueryEmbed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT file_path, chunk_text, embedding FROM atom_index`)
	if err != nil {
		return nil, fmt.Errorf("failed to query DB: %v (Did you run 'atd index'?)", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var filePath, chunkText string
		var embJSON []byte
		if err := rows.Scan(&filePath, &chunkText, &embJSON); err != nil {
			continue
		}

		isAtom := strings.HasSuffix(filePath, ".atom.md")
		switch scope {
		case "code": if isAtom { continue }
		case "docs": if !isAtom { continue }
		}

		var chunkEmb []float64
		if err := json.Unmarshal(embJSON, &chunkEmb); err != nil {
			continue
		}

		sim := cosine.Similarity(queryEmb, chunkEmb)
		results = append(results, SearchResult{filePath, chunkText, sim})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	max := limit
	if len(results) < max {
		max = len(results)
	}
	return results[:max], nil
}
