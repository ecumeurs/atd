package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"atd-tools/config"
	"atd-tools/pkg/ollama"
	"github.com/spf13/cobra"
	_ "github.com/mattn/go-sqlite3"
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
			dbPath = filepath.Join(config.DocsDir(), ".atd_index.db")
		}
		return runIndex(targetDir, dbPath, mode)
	},
}

// runIndex builds a semantic vector index for the given directory.
// mode must be "code", "docs", or "all".
func runIndex(targetDir, dbPath, mode string) error {
	// Pre-check: Ensure embedding model is available (no IDE fallback)
	res, err := ollama.ResolveProvider("embed")
	if err != nil {
		return fmt.Errorf("error resolving embedding provider: %v", err)
	}
	if res.IsIDE {
		return fmt.Errorf("indexing requires a local or remote Ollama provider with the embedding model (e.g. nomic-embed-text) installed. No embedding model was found")
	}

	// Ensure docs dir exists for default db path
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	// Init DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS atom_index (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_path TEXT,
			chunk_text TEXT,
			embedding BLOB,
			last_modified INTEGER
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Read existing modification times
	fileMods := make(map[string]int64)
	rows, err := db.Query("SELECT file_path, MAX(last_modified) FROM atom_index GROUP BY file_path")
	if err == nil {
		for rows.Next() {
			var path string
			var mtime int64
			if err := rows.Scan(&path, &mtime); err == nil {
				fileMods[path] = mtime
			}
		}
		rows.Close()
	}

	fmt.Printf("Crawling %s (Respecting .gitignore)...\n", targetDir)

	gitCmd := exec.Command("git", "ls-files", "-c", "-o", "--exclude-standard")
	gitCmd.Dir = targetDir
	out, err := gitCmd.Output()
	if err != nil {
		return fmt.Errorf("error running git ls-files: %v", err)
	}

	filesToIndex := strings.Split(strings.TrimSpace(string(out)), "\n")

	type chunkJob struct {
		path      string
		chunkText string
		modTime   int64
	}

	jobs := make(chan chunkJob, 100)
	var wg sync.WaitGroup
	var indexedCount int32
	numWorkers := 10

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				emb, err := ollama.QueryEmbed(job.chunkText)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to embed chunk from %s (Error: %v)\n", job.path, err)
					continue
				}

				embBytes, _ := json.Marshal(emb)
				_, err = db.Exec(`INSERT INTO atom_index (file_path, chunk_text, embedding, last_modified) VALUES (?, ?, ?, ?)`, job.path, job.chunkText, embBytes, job.modTime)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to insert into DB: %v\n", err)
				} else {
					atomic.AddInt32(&indexedCount, 1)
				}
			}
		}()
	}

	var skipCount int
	var fileCount int
	for _, relPath := range filesToIndex {
		if relPath == "" {
			continue
		}

		ext := filepath.Ext(relPath)
		isAtom := strings.HasSuffix(relPath, ".atom.md")

		shouldIndex := false
		switch mode {
		case "code":
			shouldIndex = config.ActiveConfig.SupportedExtensions[ext] && !isAtom
		case "docs":
			shouldIndex = isAtom
		case "all":
			shouldIndex = config.ActiveConfig.SupportedExtensions[ext] || isAtom
		default:
			close(jobs)
			wg.Wait()
			return fmt.Errorf("invalid mode: %s (must be code|docs|all)", mode)
		}

		if !shouldIndex {
			continue
		}

		path := filepath.Join(targetDir, relPath)
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}

		modTime := info.ModTime().Unix()
		if lastMod, exists := fileMods[path]; exists && lastMod == modTime {
			skipCount++
			continue
		}

		// Purge old chunks
		db.Exec(`DELETE FROM atom_index WHERE file_path = ?`, path)

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var chunks []string
		if isAtom {
			// Split by ## sections
			raw := strings.Split(string(content), "## ")
			for _, c := range raw {
				c = strings.TrimSpace(c)
				if c != "" {
					chunks = append(chunks, "File: "+relPath+"\n## "+c)
				}
			}
		} else {
			// Split by double newline
			raw := strings.Split(string(content), "\n\n")
			for _, c := range raw {
				c = strings.TrimSpace(c)
				if len(c) > 20 {
					chunks = append(chunks, "File: "+relPath+"\n"+c)
				}
			}
		}

		fileCount++
		for _, chunkText := range chunks {
			jobs <- chunkJob{path: path, chunkText: chunkText, modTime: modTime}
		}
	}

	close(jobs)
	wg.Wait()

	fmt.Printf("Indexed %d chunks across %d files (%d skipped unchanged)\n", indexedCount, fileCount, skipCount)
	config.Log("atd-index", fmt.Sprintf("Indexed %d chunks", indexedCount))
	return nil
}

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.Flags().StringP("dir", "d", ".", "Directory to crawl and index")
	indexCmd.Flags().String("db", "", "Path to SQLite database")
	indexCmd.Flags().String("mode", "code", "Index mode: code|docs|all")
}
