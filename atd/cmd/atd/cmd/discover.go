package cmd
// @spec-link [[service_atd_discover]]

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
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"github.com/spf13/cobra"
	_ "github.com/mattn/go-sqlite3"
)

// runDiscover extracts architectural intent from a source file, searches the
// ATD index, and recommends @spec-link atom tags to apply.
func runDiscover(filePath, docsDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("--file is required")
	}
	if docsDir == "" {
		docsDir = config.DocsDir()
	}

	code, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read code file: %v", err)
	}
	codeContent := string(code)

	// Step 1: Extract intent
	fmt.Println("Extracting architectural intent from source code...")
	intentPrompt := prompt.IntentExtractBuild(codeContent)

	var codeIntent string
	resp, err := ollama.Query("intent_extract", intentPrompt, prompt.IntentExtractFormat())
	if err == ollama.ErrIDEFallback {
		codeIntent = "IDE_FALLBACK_PENDING"
	} else if err != nil {
		return "", fmt.Errorf("failed to extract intent: %v", err)
	} else {
		var res struct {
			Intent string `json:"intent"`
		}
		json.Unmarshal([]byte(resp.Response), &res)
		codeIntent = res.Intent
	}

	// Step 2: Semantic Search (Requires local Ollama/Embedding)
	var matches []struct {
		id         string
		text       string
		similarity float64
	}

	if codeIntent != "IDE_FALLBACK_PENDING" {
		fmt.Println("Embedding code intent for semantic search...")
		queryEmb, err := ollama.QueryEmbed(codeIntent)
		if err != nil {
			return "", fmt.Errorf("failed to embed intent: %v", err)
		}

		dbPath := filepath.Join(docsDir, ".atd_index.db")
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			return "", fmt.Errorf("failed to open search index: %v", err)
		}
		defer db.Close()

		rows, err := db.Query(`SELECT file_path, chunk_text, embedding FROM atom_index WHERE file_path LIKE '%.atom.md'`)
		if err != nil {
			return "", fmt.Errorf("failed to query index: %v (Did you run 'atd index'?)", err)
		}
		defer rows.Close()

		for rows.Next() {
			var fpath, chunkText string
			var embJSON []byte
			if err := rows.Scan(&fpath, &chunkText, &embJSON); err != nil {
				continue
			}

			var chunkEmb []float64
			if err := json.Unmarshal(embJSON, &chunkEmb); err != nil {
				continue
			}

			sim := cosine.Similarity(queryEmb, chunkEmb)
			matches = append(matches, struct {
				id         string
				text       string
				similarity float64
			}{
				id:         strings.TrimSuffix(filepath.Base(fpath), ".atom.md"),
				text:       chunkText,
				similarity: sim,
			})
		}

		sort.Slice(matches, func(i, j int) bool {
			return matches[i].similarity > matches[j].similarity
		})

		if len(matches) > 3 {
			matches = matches[:3]
		}
	}

	// Step 3: Recommendation Prompt
	var registryBuilder strings.Builder
	for _, m := range matches {
		registryBuilder.WriteString(fmt.Sprintf("- [[%s]]: %s\n", m.id, m.text))
	}
	registryStr := registryBuilder.String()

	requestPrompt := prompt.DiscoverLinksBuild(registryStr, codeContent)

	// Handle IDE Fallback for the whole chain
	if codeIntent == "IDE_FALLBACK_PENDING" {
		taskList, _ := pipeline.WriteTaskList("discover --file "+filePath, []pipeline.PendingTask{
			{
				PromptFile:   "discover_intent.prompt",
				ResultFile:   "discover_intent.result",
				Instruction:  "extract architectural intent from code",
				OutputSchema: "string (2 sentences)",
			},
			{
				PromptFile:   "discover_recommendations.prompt",
				ResultFile:   "discover_recommendations.result",
				Instruction:  "recommend atom links once intent is known. NOTE: Manual lookup of related atoms in docs/ required.",
				OutputSchema: "list of atom IDs",
			},
		})
		pipeline.WritePromptFile("discover_intent.prompt", intentPrompt)
		pipeline.WritePromptFile("discover_recommendations.prompt", requestPrompt)
		return fmt.Sprintf("Task delegated to IDE Agent (Multi-step): %s", taskList), nil
	}

	fmt.Printf("Found %d top semantic matches. Prompting LLM for final recommendation...\n\n", len(matches))
	respRec, err := ollama.Query("intent_extract", requestPrompt, prompt.DiscoverLinksFormat())
	if err == ollama.ErrIDEFallback {
		taskList, _ := pipeline.WriteTaskList("discover --file "+filePath, []pipeline.PendingTask{
			{
				PromptFile:   "discover_recommendations.prompt",
				ResultFile:   "discover_recommendations.result",
				Instruction:  "recommend atom links based on semantic matches",
				OutputSchema: "list of atom IDs",
			},
		})
		pipeline.WritePromptFile("discover_recommendations.prompt", requestPrompt)
		return fmt.Sprintf("Recommendation task delegated to IDE Agent: %s", taskList), nil
	}
	if err != nil {
		return "", fmt.Errorf("recommendation query failed: %v", err)
	}

	var rec struct {
		Recommendations []string `json:"recommendations"`
		Rationale      string   `json:"rationale"`
	}
	json.Unmarshal([]byte(respRec.Response), &rec)

	var b strings.Builder
	b.WriteString("### Recommended Atom Links\n")
	for _, r := range rec.Recommendations {
		b.WriteString(fmt.Sprintf("- [[%s]]\n", r))
	}
	b.WriteString("\n### Rationale\n")
	b.WriteString(rec.Rationale)

	return b.String(), nil
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover atom links for undocumented source code",
	Long:  `Extracts architectural intent from code, searches the ATD index, and recommends matching atoms.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		docsDir, _ := cmd.Flags().GetString("docs")
		out, err := runDiscover(filePath, docsDir)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)
	discoverCmd.Flags().String("file", "", "Undocumented source code file")
	discoverCmd.Flags().String("docs", "", "Path to docs directory")
}
