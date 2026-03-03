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
	Type    string
	Parents []string
	Vector  []float64
}

type Config struct {
	DefaultStrictness  float64            `json:"default_strictness"`
	CollisionThreshold float64            `json:"collision_threshold"`
	TypeOverrides      map[string]float64 `json:"type_overrides"`
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

func parseAtomFile(path string) (intent string, logic string, parents []string, atomType string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", nil, "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var intentBuilder, logicBuilder strings.Builder
	mode := "none"

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "type:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				atomType = strings.TrimSpace(parts[1])
				atomType = strings.ReplaceAll(atomType, "[", "")
				atomType = strings.ReplaceAll(atomType, "]", "")
			}
			continue
		}

		if strings.HasPrefix(line, "parents:") {
			inlineVal := strings.TrimSpace(strings.TrimPrefix(line, "parents:"))
			if inlineVal != "" && inlineVal != "[]" {
				// Inline format: parents: [[id1]], [[id2]]
				inlineVal = strings.ReplaceAll(inlineVal, "[", "")
				inlineVal = strings.ReplaceAll(inlineVal, "]", "")
				inlineVal = strings.ReplaceAll(inlineVal, "\"", "")
				for _, p := range strings.Split(inlineVal, ",") {
					trimmed := strings.TrimSpace(p)
					if trimmed != "" {
						parents = append(parents, trimmed)
					}
				}
			}
			// Multi-line format: next lines are "  - [[id]]"
			// Keep reading until scanner hits a non-list line
			for scanner.Scan() {
				nextLine := scanner.Text()
				trimmed := strings.TrimSpace(nextLine)
				if strings.HasPrefix(trimmed, "- ") {
					entry := strings.TrimPrefix(trimmed, "- ")
					entry = strings.ReplaceAll(entry, "[", "")
					entry = strings.ReplaceAll(entry, "]", "")
					entry = strings.TrimSpace(entry)
					if entry != "" {
						parents = append(parents, entry)
					}
				} else {
					// Not a list item — process this line normally
					if strings.HasPrefix(nextLine, "## INTENT") {
						mode = "intent"
					} else if strings.HasPrefix(nextLine, "## THE RULE") {
						mode = "logic"
					} else if strings.HasPrefix(nextLine, "##") {
						mode = "none"
					}
					break
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

	return strings.TrimSpace(intentBuilder.String()), strings.TrimSpace(logicBuilder.String()), parents, atomType, nil
}

func queryBloatIndicator(model, role, text string, strictness float64) string {
	if strictness <= 0.0 {
		return "NO" // Auto-pass
	}

	role = fmt.Sprintf("%s\n\nStrictness Threshold (0.0 to 1.0): %.1f. At 1.0, aggressively fail broad statements. At lower thresholds, allow contextual grouping.", role, strictness)

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
	threshold := flag.Float64("threshold", -1, "Override collision similarity threshold (default: from config or 0.85)")
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

	// Load Config
	configPath := filepath.Join(binPath, ".atd_audit_config.json")
	var config Config
	configBytes, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(configBytes, &config)
	} else {
		// Fallback defaults
		config = Config{
			DefaultStrictness:  0.8,
			CollisionThreshold: 0.85,
			TypeOverrides: map[string]float64{
				"REQUIREMENT":   0.3,
				"SPECIFICATION": 0.3,
				"MODULE":        0.3,
				"DOMAIN":        0.5,
				"ENTITY":        0.5,
			},
		}
	}

	// CLI flag overrides config; -1 means "not set, use config"
	collisionThreshold := config.CollisionThreshold
	if config.CollisionThreshold == 0 {
		collisionThreshold = 0.85 // absolute fallback if config file exists but field is missing
	}
	if *threshold >= 0 {
		collisionThreshold = *threshold
	}

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

	// Schema upgrades — ignore errors if columns already exist
	db.Exec(`ALTER TABLE atom_docs_index ADD COLUMN atom_type TEXT`)
	db.Exec(`ALTER TABLE atom_docs_index ADD COLUMN bloat_result TEXT`)

	// Collision pair cache — keyed by sorted (file_a, file_b) + both mtimes
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS collision_cache (
		file_a TEXT NOT NULL,
		file_b TEXT NOT NULL,
		mtime_a INTEGER NOT NULL,
		mtime_b INTEGER NOT NULL,
		similarity REAL,
		result TEXT,
		PRIMARY KEY (file_a, file_b)
	)`)
	if err != nil {
		fmt.Printf("Error creating collision_cache table: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== ATD AUDIT PROTOCOL INITIATED ===")
	fmt.Println("Target Directory:", *docsDir)
	fmt.Println("\n--- PHASE 1: The Bloat Metric (Syntactic Validator) ---")

	intentRoleBase := `You are an Architectural Linter for a codebase following the 'Minimum Atomic Scale' rule. 
Evaluate this INTENT text for a documentation component. 
An INTENT section is 'bloated' if it describes more than one distinct feature or goal (using compound words like 'and', 'also', 'furthermore' to list separate features).`

	logicRoleBase := `You are an Architectural Linter for a codebase following the 'Minimum Atomic Scale' rule. 
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

		// If file hasn't been modified since we last indexed it, load from cache.
		if err == nil && currentMtime <= dbMtime {
			var embJSON []byte
			var dbType, cachedBloat string
			err2 := db.QueryRow("SELECT embedding, atom_type, bloat_result FROM atom_docs_index WHERE id = ?", filename).Scan(&embJSON, &dbType, &cachedBloat)
			if err2 == nil {
				_, _, parents, atomType, perr := parseAtomFile(f)
				if perr == nil {
					if dbType != "" {
						atomType = dbType
					}
					var vec []float64
					if jerr := json.Unmarshal(embJSON, &vec); jerr == nil {
						embeddings[filename] = FileMetadata{Type: atomType, Parents: parents, Vector: vec}
						filenames = append(filenames, filename)
						if cachedBloat != "" {
							fmt.Printf("Auditing: %s ... [CACHED: %s]\n", filename, cachedBloat)
						} else {
							fmt.Printf("Auditing: %s ... [CACHED: PASS]\n", filename)
						}
						continue
					}
				}
			}
		}

		intentStr, logicStr, parents, atomType, err := parseAtomFile(f)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", filename, err)
			continue
		}

		fmt.Printf("Auditing: %s [Type: %s] ... ", filename, atomType)

		strictness := config.DefaultStrictness
		if val, ok := config.TypeOverrides[atomType]; ok {
			strictness = val
		}

		var intentResult, logicResult string
		if strictness <= 0.0 {
			intentResult = "NO"
			logicResult = "NO"
		} else {
			intentResult = queryBloatIndicator(*model, intentRoleBase, intentStr, strictness)
			logicResult = queryBloatIndicator(*model, logicRoleBase, logicStr, strictness)
		}

		var bloatResult string
		if intentResult == "YES" || logicResult == "YES" {
			bloatResult = "BLOATED"
			fmt.Println("[BLOATED] This Atom contains compound rules and violates Minimum Atomic Scale.")
		} else if intentResult == "NO" && logicResult == "NO" {
			bloatResult = "PASS"
			fmt.Println("[PASS] Atom is compliant.")
		} else {
			bloatResult = "WARN"
			fmt.Printf("[WARN] LLM output ambiguous: Intent: %s | Logic: %s\n", intentResult, logicResult)
		}

		// Save the contents for Phase 2 while we opened the file
		contentBytes, _ := os.ReadFile(f)
		content := string(contentBytes)

		vec, err := getEmbedding(content)
		if err == nil {
			embeddings[filename] = FileMetadata{Type: atomType, Parents: parents, Vector: vec}
			filenames = append(filenames, filename)

			embJSON, _ := json.Marshal(vec)
			_, err = db.Exec(`
				INSERT INTO atom_docs_index (id, file_path, intent_text, logic_text, embedding, file_mtime, atom_type, bloat_result)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					file_path=excluded.file_path,
					intent_text=excluded.intent_text,
					logic_text=excluded.logic_text,
					embedding=excluded.embedding,
					file_mtime=excluded.file_mtime,
					atom_type=excluded.atom_type,
					bloat_result=excluded.bloat_result
			`, filename, f, intentStr, logicStr, embJSON, currentMtime, atomType, bloatResult)
			if err != nil {
				fmt.Printf("Error saving to DB: %v\n", err)
			}
		}
	}

	fmt.Println("\n--- PHASE 2: The Collision Map (Semantic Overlap Detection) ---")
	fmt.Println("Detecting architectural overlaps using Nomic Vector Embeddings...")
	fmt.Println("\n--- Collision Matrix Results ---")

	collisionsFound := false

	// Helper: get file mtime from disk (fast, no LLM)
	getMtime := func(fname string) int64 {
		for _, f := range files {
			if filepath.Base(f) == fname {
				if info, err := os.Stat(f); err == nil {
					return info.ModTime().Unix()
				}
			}
		}
		return 0
	}

	// Build a full parent index: ATD id -> []parent ids
	// This lets us do ancestry walks rather than single-level checks.
	parentIndex := make(map[string][]string)
	for _, fname := range filenames {
		id := strings.TrimSuffix(fname, ".atom.md")
		parentIndex[id] = embeddings[fname].Parents
	}

	// isAncestor returns true if candidateAncestor is reachable by
	// following parents[] links upward from startID (BFS, cycle-safe).
	isAncestor := func(startID, candidateAncestor string) bool {
		visited := make(map[string]bool)
		queue := []string{startID}
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			if visited[cur] {
				continue
			}
			visited[cur] = true
			for _, p := range parentIndex[cur] {
				if p == candidateAncestor {
					return true
				}
				if !visited[p] {
					queue = append(queue, p)
				}
			}
		}
		return false
	}

	for i := 0; i < len(filenames); i++ {
		for j := i + 1; j < len(filenames); j++ {
			f1 := filenames[i]
			f2 := filenames[j]

			// Sorted key for symmetry
			cacheKeyA, cacheKeyB := f1, f2
			if cacheKeyA > cacheKeyB {
				cacheKeyA, cacheKeyB = cacheKeyB, cacheKeyA
			}
			mtA := getMtime(cacheKeyA)
			mtB := getMtime(cacheKeyB)

			// Check collision cache
			var cachedSim float64
			var cachedResult string
			var cacheMtA, cacheMtB int64
			cErr := db.QueryRow(`SELECT mtime_a, mtime_b, similarity, result FROM collision_cache WHERE file_a=? AND file_b=?`,
				cacheKeyA, cacheKeyB).Scan(&cacheMtA, &cacheMtB, &cachedSim, &cachedResult)
			if cErr == nil && cacheMtA == mtA && cacheMtB == mtB {
				// Cache hit — print without recomputing
				if cachedResult == "SOUND" {
					// Silent — don't flood output with passed pairs
				} else if cachedResult == "CROSS_TYPE" {
					// Also silent
				} else if cachedResult == "MISSING_ABSTRACTION" {
					collisionsFound = true
					fmt.Printf("\n[COLLISION] %s <--> %s (Similarity: %.2f) [CACHED]\n", f1, f2, cachedSim)
					fmt.Printf("  Result: [MISSING ABSTRACTION] These Atoms share %d%% logic but lack a shared parent.\n", int(cachedSim*100))
				}
				continue
			}

			type1 := embeddings[f1].Type
			type2 := embeddings[f2].Type

			vec1 := embeddings[f1].Vector
			vec2 := embeddings[f2].Vector

			sim := cosineSimilarity(vec1, vec2)

			var pairResult string
			if sim >= collisionThreshold {
				if type1 != type2 && type1 != "" && type2 != "" {
					pairResult = "CROSS_TYPE"
					fmt.Printf("\n[CROSS-TYPE ALIGNMENT] %s (%s) <--> %s (%s) (Similarity: %.2f)\n", f1, type1, f2, type2, sim)
					fmt.Printf("  Status: IGNORED. Valid structural translation between %s and %s.\n", type1, type2)
				} else {
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

					// Derive ATD ids from filenames (strip .atom.md suffix)
					id1 := strings.TrimSuffix(f1, ".atom.md")
					id2 := strings.TrimSuffix(f2, ".atom.md")

					// Check ancestry in both directions (full BFS walk, not just immediate parent)
					isRelated := isAncestor(id1, id2) || isAncestor(id2, id1)

					fmt.Printf("\n[COLLISION] %s <--> %s (Similarity: %.2f) [Type: %s]\n", f1, f2, sim, type1)
					if isRelated {
						pairResult = "SOUND"
						fmt.Printf("  Result: Structurally Sound (Ancestor Relationship)\n")
					} else if len(shared) > 0 {
						pairResult = "SOUND"
						fmt.Printf("  Result: Structurally Sound (Shared Parent: %v)\n", shared)
					} else {
						pairResult = "MISSING_ABSTRACTION"
						collisionsFound = true
						fmt.Printf("  Result: [MISSING ABSTRACTION] These Atoms share %d%% logic but lack a shared parent. Refactor into Base Class/Interface.\n", int(sim*100))
					}
				}
			} else {
				pairResult = "SOUND"
			}

			// Upsert collision cache
			db.Exec(`INSERT INTO collision_cache (file_a, file_b, mtime_a, mtime_b, similarity, result)
				VALUES (?, ?, ?, ?, ?, ?)
				ON CONFLICT(file_a, file_b) DO UPDATE SET
					mtime_a=excluded.mtime_a, mtime_b=excluded.mtime_b,
					similarity=excluded.similarity, result=excluded.result`,
				cacheKeyA, cacheKeyB, mtA, mtB, sim, pairResult)
		}
	}

	if !collisionsFound {
		fmt.Println("\nNo critical semantic collisions detected above threshold.")
	}

	fmt.Println("\n=== AUDIT COMPLETE ===")
}
