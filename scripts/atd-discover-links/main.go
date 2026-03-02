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
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

func getEmbedding(text string) ([]float64, error) {
	reqBody := EmbeddingRequest{
		Model:  "nomic-embed-text",
		Prompt: text,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://127.0.0.1:11434/api/embeddings", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var embResp EmbeddingResponse
	if err := json.Unmarshal(bodyBytes, &embResp); err != nil {
		return nil, err
	}
	return embResp.Embedding, nil
}

func getOllamaResponse(model, prompt string) (string, error) {
	reqBody := GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://127.0.0.1:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var genResp GenerateResponse
	if err := json.Unmarshal(bodyBytes, &genResp); err != nil {
		return "", err
	}
	return genResp.Response, nil
}

func cosineSimilarity(v1, v2 []float64) float64 {
	if len(v1) == 0 || len(v2) == 0 || len(v1) != len(v2) {
		return 0.0
	}
	var dotProduct, mag1, mag2 float64
	for i := 0; i < len(v1); i++ {
		dotProduct += v1[i] * v2[i]
		mag1 += v1[i] * v1[i]
		mag2 += v2[i] * v2[i]
	}
	if mag1*mag2 == 0 {
		return 0.0
	}
	return dotProduct / (math.Sqrt(mag1) * math.Sqrt(mag2))
}

func main() {
	var file string

	flag.StringVar(&file, "file", "", "Undocumented source code file")

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

	if file == "" {
		fmt.Println("Error: -file required.")
		os.Exit(1)
	}

	code, _ := os.ReadFile(file)

	codeContent := string(code)

	// Ensure the index database exists
	dbPath := filepath.Join(docsPath, ".atd_docs_index.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("Error: ATD Vector Index not found at %s. Please run atd-audit first.\n", dbPath)
		os.Exit(1)
	}

	fmt.Println("Extracting architectural intent from source code...")
	intentPrompt := fmt.Sprintf(`Summarize the architectural and domain-level intent of this code in 2 sentences. Focus on what subsystem it belongs to and the core rules it enforces. Do not talk about specific variable names.:
%s`, codeContent)
	codeIntent, err := getOllamaResponse("llama3.2", intentPrompt)
	if err != nil {
		fmt.Printf("Failed to extract intent: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Embedding code intent for semantic search...")
	queryEmb, err := getEmbedding(codeIntent)
	if err != nil {
		fmt.Printf("Failed to embed intent: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Failed to open db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, intent_text, embedding FROM atom_docs_index`)
	if err != nil {
		fmt.Printf("Failed to query DB: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	type Match struct {
		id         string
		intentText string
		similarity float64
	}
	var matches []Match

	for rows.Next() {
		var id, intentText string
		var embJSON []byte
		if err := rows.Scan(&id, &intentText, &embJSON); err != nil {
			continue
		}

		var chunkEmb []float64
		if err := json.Unmarshal(embJSON, &chunkEmb); err != nil {
			continue
		}

		sim := cosineSimilarity(queryEmb, chunkEmb)
		matches = append(matches, Match{id, intentText, sim})
	}

	// Sort results by similarity descending
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].similarity > matches[i].similarity {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	limit := 3
	max := limit
	if len(matches) < max {
		max = len(matches)
	}

	var registryBuilder strings.Builder
	for i := 0; i < max; i++ {
		registryBuilder.WriteString(fmt.Sprintf("- [[%s]]: %s\n", matches[i].id, matches[i].intentText))
	}
	registryStr := registryBuilder.String()

	fmt.Printf("Found %d top semantic matches. Prompting LLM for final recommendation...\n\n", max)

	prompt := fmt.Sprintf(`
<System Objective>
You are an ATD Link Discoverer. Read the Target Source Code carefully and deduce which Atoms from the Known Atom Registry define this logic. Output a list of recommended IDs.
Only recommend IDs from the provided registry list if they truly map to the codebase logic.
</System Objective>

<Known Atom Registry (Top Semantic Matches)>
%s
</Known Atom Registry (Top Semantic Matches)>

<Target Source Code>
%s
</Target Source Code>
`, registryStr, codeContent)

	fmt.Println(prompt)
	// You could optionally send this prompt automatically to the LLM here instead of just printing it
}
