package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"

	_ "github.com/mattn/go-sqlite3"
)

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func getEmbedding(text string) ([]float64, error) {
	reqBody := EmbeddingRequest{
		Model:  "nomic-embed-text",
		Prompt: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://localhost:11434/api/embeddings", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var embResp EmbeddingResponse
	if err := json.Unmarshal(bodyBytes, &embResp); err != nil {
		return nil, err
	}

	return embResp.Embedding, nil
}

// naiveChunker splits a file by top-level functions/structs using double newlines
func naiveChunker(content, filename string) []string {
	var chunks []string

	// Very simple naive split for MVP: split by double newline
	rawChunks := regexp.MustCompile(`\n\s*\n`).Split(content, -1)

	for i, c := range rawChunks {
		c = strings.TrimSpace(c)
		if len(c) > 20 { // skip tiny arbitrary blocks
			chunks = append(chunks, fmt.Sprintf("File: %s\nBlock %d:\n%s", filename, i, c))
		}
	}
	return chunks
}

func main() {
	var targetDir string
	var dbPath string
	flag.StringVar(&targetDir, "dir", ".", "Directory to crawl and index")
	flag.StringVar(&dbPath, "db", ".atd_index.db", "Path to SQLite database")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()

	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	// Init DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Failed to open db: %v\n", err)
		os.Exit(1)
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
		fmt.Printf("Failed to create table: %v\n", err)
		os.Exit(1)
	}

	// Read existing modification times
	rows, err := db.Query("SELECT file_path, MAX(last_modified) FROM atom_index GROUP BY file_path")
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("Failed to query existing index: %v\n", err)
	}

	fileMods := make(map[string]int64)
	if rows != nil {
		for rows.Next() {
			var path string
			var mtime int64
			if err := rows.Scan(&path, &mtime); err == nil {
				fileMods[path] = mtime
			}
		}
		rows.Close()
	}

	fmt.Printf("Crawling %s (Respecting .gitignore via git ls-files)...\n", targetDir)

	// Use Git to get tracked and untracked (non-ignored) files
	cmd := exec.Command("git", "ls-files", "-c", "-o", "--exclude-standard")
	cmd.Dir = targetDir
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error running git ls-files (is this a git repo?): %v\n", err)
		os.Exit(1)
	}

	filesToIndex := strings.Split(strings.TrimSpace(string(out)), "\n")

	// Create a worker pool to process chunks concurrently
	numWorkers := 10 // Max concurrent requests to Ollama
	type chunkJob struct {
		path      string
		chunkText string
		modTime   int64
	}

	jobs := make(chan chunkJob, 100)
	var wg sync.WaitGroup
	var indexedCount int32

	// Start workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				emb, err := getEmbedding(job.chunkText)
				if err != nil {
					fmt.Printf("Warning: Failed to embed chunk from %s (Error: %v)\n", job.path, err)
					continue
				}

				// Convert []float64 to JSON bytes for storage in SQLite BLOB
				embBytes, _ := json.Marshal(emb)

				// Re-acquire DB handle via transaction or direct insert inside mutex if using memory db,
				// but SQLite handles concurrent inserts fine with Wal.
				// To be safe against "database is locked" in SQLite, handle retries or just execute.
				_, err = db.Exec(`INSERT INTO atom_index (file_path, chunk_text, embedding, last_modified) VALUES (?, ?, ?, ?)`, job.path, job.chunkText, embBytes, job.modTime)
				if err != nil {
					fmt.Printf("Failed to insert into DB: %v\n", err)
				} else {
					atomic.AddInt32(&indexedCount, 1)
				}
			}
		}()
	}

	for _, relPath := range filesToIndex {
		if relPath == "" || !strings.HasSuffix(relPath, ".go") {
			continue
		}

		path := filepath.Join(targetDir, relPath)
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}

		modTime := info.ModTime().Unix()

		if lastMod, exists := fileMods[path]; exists && lastMod == modTime {
			// File hasn't changed, skip entirely
			fmt.Printf("Skipped (Unchanged): %s\n", path)
			continue
		}

		// File changed or is new, purge old chunks first
		db.Exec(`DELETE FROM atom_index WHERE file_path = ?`, path)

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		chunks := naiveChunker(string(contentBytes), path)
		for _, chunkText := range chunks {
			jobs <- chunkJob{path: path, chunkText: chunkText, modTime: modTime}
		}
		fmt.Printf("Queued %d chunks for %s\n", len(chunks), path)
	}

	close(jobs)
	wg.Wait()

	fmt.Printf("Indexing complete. Inserted %d total chunks.\n", indexedCount)
}
