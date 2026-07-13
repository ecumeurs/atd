package exploration

import (
	"database/sql"
	"encoding/json"
	"errors"
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

// ErrIndexMissing indicates the semantic index for a project has not been
// built (db file absent), is missing its schema (no atom_index table), or
// is present but empty (0 rows). Callers must surface this loudly instead
// of silently falling back to a grep search, since a missing index is a
// setup problem the user can fix by running 'atd index'.
var ErrIndexMissing = errors.New("no semantic index found for this project — run 'atd index' first")

func isNoSuchTableErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no such table")
}

type SearchResult struct {
	FilePath   string  `json:"file_path"`
	ChunkText  string  `json:"chunk_text"`
	Similarity float64 `json:"similarity"`
	Project    string  `json:"project"` // NEW: Which project this belongs to
}

type SearchOptions struct {
	Query     string
	Grep      string
	DBPath    string
	Limit     int
	Scope     string
	Root      string
	Workspace bool     // NEW: Search across workspace projects
	Projects  []string // NEW: Specific projects to search (empty = all)
}

func Search(opts SearchOptions) ([]SearchResult, error) {
	// If workspace mode is enabled and we're in a workspace
	if opts.Workspace && config.ActiveConfig.Workspace != nil {
		return WorkspaceSearch(opts)
	}

	if opts.Grep != "" {
		return GrepSearch(opts.Grep, opts.Root, "")
	}
	if opts.Query != "" {
		results, err := SemanticSearch(opts.Query, opts.DBPath, opts.Limit, opts.Scope, "")
		if err != nil {
			if errors.Is(err, ErrIndexMissing) {
				// Don't silently grep: a missing index is a setup problem,
				// not a transient failure — surface it so the caller knows
				// to run 'atd index'.
				return nil, err
			}
			// Provider/embedding failure (e.g. Ollama offline): it's fine to
			// fall back to a literal grep, but say so instead of returning
			// a quiet, misleading result.
			fmt.Printf("Notice: semantic search unavailable (%v) — falling back to keyword search\n\n", err)
			return GrepSearch(opts.Query, opts.Root, "")
		}
		return results, nil
	}
	return nil, fmt.Errorf("either query or grep must be specified")
}

func WorkspaceSearch(opts SearchOptions) ([]SearchResult, error) {
	var allResults []SearchResult
	workspace := config.ActiveConfig.Workspace

	// Determine which projects to search
	var projectsToSearch []config.ProjectConfig
	if len(opts.Projects) > 0 {
		// Filter to specified projects
		for _, p := range workspace.Projects {
			for _, name := range opts.Projects {
				if p.Name == name {
					projectsToSearch = append(projectsToSearch, p)
					break
				}
			}
		}
	} else {
		// Search all projects
		projectsToSearch = workspace.Projects
	}

	// Search each project's docs directory
	for _, project := range projectsToSearch {
		absProjPath := project.Path
		if !filepath.IsAbs(absProjPath) {
			absProjPath = filepath.Join(workspace.LoadedFrom, project.Path)
		}

		projDocsPath := project.DocsPath
		if projDocsPath == "" {
			projDocsPath = filepath.Join(absProjPath, "docs")
		} else if !filepath.IsAbs(projDocsPath) {
			projDocsPath = filepath.Join(absProjPath, projDocsPath)
		}

		projDBPath := config.IndexDBPath(projDocsPath)

		// Search in this project
		results, err := SemanticSearch(opts.Query, projDBPath, opts.Limit, opts.Scope, project.Name)
		if err != nil {
			// A workspace search spans many projects; not every project is
			// guaranteed to have a semantic index built, so we degrade to
			// grep per-project rather than aborting the whole search — but
			// we still say so instead of silently returning nothing.
			fmt.Printf("Notice: semantic search unavailable for project %q (%v) — falling back to keyword search\n\n", project.Name, err)
			results, _ = GrepSearch(opts.Query, projDocsPath, project.Name)
		}
		allResults = append(allResults, results...)
	}

	// Sort by similarity and limit
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Similarity > allResults[j].Similarity
	})

	if len(allResults) > opts.Limit && opts.Limit > 0 {
		allResults = allResults[:opts.Limit]
	}

	return allResults, nil
}

func GrepSearch(keyword, root, projectName string) ([]SearchResult, error) {
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
				Project:   projectName,
			})
		}
		return nil
	})
	return results, err
}

func SemanticSearch(query, dbPath string, limit int, scope, projectName string) ([]SearchResult, error) {
	// Check the index exists BEFORE opening it — sql.Open+Query against a
	// missing path silently creates a 0-byte SQLite file as a side effect,
	// and doing so inside the project's working tree is exactly the
	// artifact-leak this function must avoid. This also lets us short-circuit
	// with a clear "run atd index" message without spending an embedding call.
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrIndexMissing
		}
		return nil, fmt.Errorf("failed to access index db %s: %v", dbPath, err)
	}

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
		if isNoSuchTableErr(err) {
			return nil, ErrIndexMissing
		}
		return nil, fmt.Errorf("failed to query DB: %v (Did you run 'atd index'?)", err)
	}
	defer rows.Close()

	var totalRows int
	var results []SearchResult
	for rows.Next() {
		totalRows++
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
		results = append(results, SearchResult{filePath, chunkText, sim, projectName})
	}

	if totalRows == 0 {
		// The table exists but nothing was ever indexed (or everything was
		// purged) — treat this the same as a missing index rather than
		// returning a quiet empty result.
		return nil, ErrIndexMissing
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
