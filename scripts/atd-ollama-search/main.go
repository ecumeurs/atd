package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"atd-tools/config"
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

func cosineSimilarity(a, b []float64) float64 {
	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

func main() {
	var dbPath string
	var query string
	var limit int

	flag.StringVar(&dbPath, "db", ".atd_index.db", "Path to SQLite database")
	flag.StringVar(&query, "query", "", "The ATD text or semantic intent to search for")
	flag.IntVar(&limit, "limit", 3, "Number of closest chunks to return")

	var docsPath, projectPath, binPath string
	flag.StringVar(&projectPath, "project", ".", "Path to the root of the project")
	flag.StringVar(&docsPath, "docs", "", "Path to the docs directory (default: projectPath/docs/)")
	flag.StringVar(&binPath, "bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()
	config.Load()
	config.Log("atd-ollama-search", "Started process")


	if docsPath == "" {
		docsPath = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}

	if query == "" {
		fmt.Println("Error: -query is required")
		os.Exit(1)
	}

	// 1. Get Query Vector
	queryEmb, err := getEmbedding(query)
	if err != nil {
		fmt.Printf("Failed to embed query: %v\n", err)
		os.Exit(1)
	}

	// 2. Open DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Failed to open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, file_path, chunk_text, embedding FROM atom_index`)
	if err != nil {
		fmt.Printf("Failed to query DB: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	// 3. Calculate Similarities
	type Result struct {
		filePath   string
		chunkText  string
		similarity float64
	}
	var results []Result

	for rows.Next() {
		var id int
		var filePath, chunkText string
		var embJSON []byte
		if err := rows.Scan(&id, &filePath, &chunkText, &embJSON); err != nil {
			continue
		}

		var chunkEmb []float64
		if err := json.Unmarshal(embJSON, &chunkEmb); err != nil {
			continue
		}

		sim := cosineSimilarity(queryEmb, chunkEmb)
		results = append(results, Result{filePath, chunkText, sim})
	}

	if err = rows.Err(); err != nil {
		fmt.Printf("Row iteration error: %v\n", err)
	}

	// Sort results by similarity descending
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].similarity > results[i].similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Print top results
	fmt.Printf("--- Top %d Semantic Matches ---\n\n", limit)
	max := limit
	if len(results) < max {
		max = len(results)
	}
	for i := 0; i < max; i++ {
		fmt.Printf("[Match %d] File: %s (Similarity: %.4f)\n%s\n\n", i+1, results[i].filePath, results[i].similarity, results[i].chunkText)
	}
}
