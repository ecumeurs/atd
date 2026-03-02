package main

import (
	"bufio"
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

type FileMetadata struct {
	Parents []string
	Vector  []float64
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

func parseAtomFile(path string) (intent string, logic string, parents []string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var intentBuilder, logicBuilder strings.Builder
	mode := "none"

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "parents:") {
			parts := strings.Split(line, "parents:")
			if len(parts) > 1 {
				parentsStr := strings.TrimSpace(parts[1])
				parentsStr = strings.ReplaceAll(parentsStr, "[", "")
				parentsStr = strings.ReplaceAll(parentsStr, "]", "")
				parentsStr = strings.ReplaceAll(parentsStr, "\"", "")
				for _, p := range strings.Split(parentsStr, ",") {
					trimmed := strings.TrimSpace(p)
					if trimmed != "" {
						parents = append(parents, trimmed)
					}
				}
			}
			continue
		}

		if strings.HasPrefix(line, "## INTENT") {
			mode = "intent"
			continue
		} else if strings.HasPrefix(line, "## THE RULE") {
			mode = "logic"
			continue
		} else if strings.HasPrefix(line, "##") && mode != "none" {
			mode = "none"
		}

		if mode == "intent" {
			intentBuilder.WriteString(line + "\n")
		} else if mode == "logic" {
			logicBuilder.WriteString(line + "\n")
		}
	}

	return strings.TrimSpace(intentBuilder.String()), strings.TrimSpace(logicBuilder.String()), parents, nil
}

func queryBloatIndicator(model, role, text string) string {
	prompt := fmt.Sprintf(`%s

Respond with exactly one word: YES (if it is bloated) or NO (if it is compliant). Do not add any extra text.

TEXT TO EVALUATE:
%s`, role, text)

	resp, err := getOllamaResponse(model, prompt)
	if err != nil {
		return "ERROR"
	}

	resp = strings.TrimSpace(strings.ToUpper(resp))
	if strings.Contains(resp, "YES") {
		return "YES"
	} else if strings.Contains(resp, "NO") {
		return "NO"
	}
	return fmt.Sprintf("WARN - unparsable output: %s", resp)
}

func main() {
	threshold := flag.Float64("threshold", 0.85, "Similarity threshold for collision detection")
	model := flag.String("model", "llama3.2", "Local Ollama model to use for Phase 1")

	projectPathPtr := flag.String("project", ".", "Path to the root of the project")
	docsDirStr := flag.String("docs", "", "Path to the docs directory (default: projectPath/docs/)")
	binPathPtr := flag.String("bin", "", "Path to the ATD tools bin directory (default: projectPath/.agent/skills/atd/tools/)")
	flag.Parse()

	projectPath := *projectPathPtr
	docsStrVal := *docsDirStr
	binPath := *binPathPtr

	if docsStrVal == "" {
		docsStrVal = filepath.Join(projectPath, "docs")
	}
	if binPath == "" {
		binPath = filepath.Join(projectPath, ".agent/skills/atd/tools/")
	}
	docsDir := &docsStrVal

	files, err := filepath.Glob(filepath.Join(*docsDir, "*.atom.md"))
	if err != nil || len(files) == 0 {
		fmt.Println("No .atom.md files found in", *docsDir)
		os.Exit(1)
	}

	dbPath := filepath.Join(*docsDir, ".atd_docs_index.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Error opening database at %s: %v\n", dbPath, err)
		os.Exit(1)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS atom_docs_index (
		id TEXT PRIMARY KEY,
		file_path TEXT,
		intent_text TEXT,
		logic_text TEXT,
		embedding JSON,
		file_mtime INTEGER
	)`)
	if err != nil {
		fmt.Printf("Error creating table: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== ATD AUDIT PROTOCOL INITIATED ===")
	fmt.Println("Target Directory:", *docsDir)
	fmt.Println("\n--- PHASE 1: The Bloat Metric (Syntactic Validator) ---")

	intentRole := `You are an Architectural Linter for a codebase following the 'Minimum Atomic Scale' rule. 
Evaluate this INTENT text for a documentation component. 
An INTENT section is 'bloated' if it describes more than one distinct feature or goal (using compound words like 'and', 'also', 'furthermore' to list separate features).`

	logicRole := `You are an Architectural Linter for a codebase following the 'Minimum Atomic Scale' rule. 
Evaluate this RULE/LOGIC text for a documentation component. 
A RULE section is 'bloated' if it contains more than one completely distinct state-changing rule or calculation that could logically be split into two separate components.`

	embeddings := make(map[string]FileMetadata)
	var filenames []string

	for _, f := range files {
		filename := filepath.Base(f)

		info, err := os.Stat(f)
		if err != nil {
			fmt.Printf("Error stating %s: %v\n", filename, err)
			continue
		}

		currentMtime := info.ModTime().Unix()
		var dbMtime int64
		err = db.QueryRow("SELECT file_mtime FROM atom_docs_index WHERE id = ?", filename).Scan(&dbMtime)

		// If file hasn't been modified since we last indexed it, try to load it from DB
		// so we don't have to re-evaluate it with LLM and Nomic.
		if err == nil && currentMtime <= dbMtime {
			fmt.Printf("Skipping %s (Already indexed and unmodified)\n", filename)

			// Load from DB for collision map phase
			var embJSON []byte
			err = db.QueryRow("SELECT embedding FROM atom_docs_index WHERE id = ?", filename).Scan(&embJSON)
			if err == nil {
				// parents need to be parsed from the file anyway or stored in DB. Let's just re-parse the file quickly
				// it's very cheap compared to LLM calls.
				_, _, parents, err := parseAtomFile(f)
				if err == nil {
					var vec []float64
					if err := json.Unmarshal(embJSON, &vec); err == nil {
						embeddings[filename] = FileMetadata{Parents: parents, Vector: vec}
						filenames = append(filenames, filename)
						continue
					}
				}
			}
		}

		intentStr, logicStr, parents, err := parseAtomFile(f)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("Auditing: %s ... ", filename)

		intentResult := queryBloatIndicator(*model, intentRole, intentStr)
		logicResult := queryBloatIndicator(*model, logicRole, logicStr)

		if intentResult == "YES" || logicResult == "YES" {
			fmt.Println("[BLOATED] This Atom contains compound rules and violates Minimum Atomic Scale.")
		} else if intentResult == "NO" && logicResult == "NO" {
			fmt.Println("[PASS] Atom is compliant.")
		} else {
			fmt.Printf("[WARN] LLM output ambiguous: Intent: %s | Logic: %s\n", intentResult, logicResult)
		}

		// Save the contents for Phase 2 while we opened the file
		contentBytes, _ := os.ReadFile(f)
		content := string(contentBytes)

		vec, err := getEmbedding(content)
		if err == nil {
			embeddings[filename] = FileMetadata{Parents: parents, Vector: vec}
			filenames = append(filenames, filename)

			embJSON, _ := json.Marshal(vec)
			_, err = db.Exec(`
				INSERT INTO atom_docs_index (id, file_path, intent_text, logic_text, embedding, file_mtime)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					file_path=excluded.file_path,
					intent_text=excluded.intent_text,
					logic_text=excluded.logic_text,
					embedding=excluded.embedding,
					file_mtime=excluded.file_mtime
			`, filename, f, intentStr, logicStr, embJSON, currentMtime)
			if err != nil {
				fmt.Printf("Error saving to DB: %v\n", err)
			}
		}
	}

	fmt.Println("\n--- PHASE 2: The Collision Map (Semantic Overlap Detection) ---")
	fmt.Println("Detecting architectural overlaps using Nomic Vector Embeddings...")
	fmt.Println("\n--- Collision Matrix Results ---")

	collisionsFound := false

	for i := 0; i < len(filenames); i++ {
		for j := i + 1; j < len(filenames); j++ {
			f1 := filenames[i]
			f2 := filenames[j]

			vec1 := embeddings[f1].Vector
			vec2 := embeddings[f2].Vector

			sim := cosineSimilarity(vec1, vec2)
			if sim >= *threshold {
				collisionsFound = true
				fmt.Printf("\n[COLLISION] %s <--> %s (Similarity: %.2f)\n", f1, f2, sim)

				p1 := embeddings[f1].Parents
				p2 := embeddings[f2].Parents

				sharedIndex := make(map[string]bool)
				for _, p := range p1 {
					sharedIndex[p] = true
				}

				var shared []string
				for _, p := range p2 {
					if sharedIndex[p] {
						shared = append(shared, p)
					}
				}

				if len(shared) > 0 {
					fmt.Printf("  Result: Structurally Sound (Shared Parent: %v)\n", shared)
				} else {
					fmt.Printf("  Result: [MISSING ABSTRACTION] These Atoms share %d%% logic but lack a shared parent. Refactor into Base Class/Interface.\n", int(sim*100))
				}
			}
		}
	}

	if !collisionsFound {
		fmt.Println("\nNo critical semantic collisions detected above threshold.")
	}

	fmt.Println("\n=== AUDIT COMPLETE ===")
}
