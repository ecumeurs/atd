package cmd
// @spec-link [[service_atd_audit]]

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/cosine"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"github.com/spf13/cobra"
	_ "github.com/mattn/go-sqlite3"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit atoms for bloat and collisions",
	Long: `Audit performs structural and semantic analysis of ATD atoms.

Phase 1 (Bloat Detection): Uses LLM to identify atoms containing compound rules.
Phase 2 (Collision Detection): Uses embeddings to find semantic overlaps between atoms.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		threshold, _ := cmd.Flags().GetFloat64("threshold")
		docsDir, _ := cmd.Flags().GetString("docs")
		workspace, _ := cmd.Flags().GetBool("workspace")

		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		if threshold <= 0 {
			threshold = config.ActiveConfig.DiffSimilarityThreshold
			if threshold <= 0 {
				threshold = 0.85
			}
		}

		return runFullAudit(docsDir, threshold, workspace)
	},
}

type atomAuditMeta struct {
	ID           string
	FilePath     string
	AtomType     string
	Parents      []string
	Embedding    []float64
	BloatResult  string
	LastModified int64
}

func runFullAudit(docsDir string, threshold float64, workspace bool) error {
	dbPath := filepath.Join(docsDir, ".atd_audit.db") // Renamed from original to avoid conflict with search index if needed
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS audit_cache (
		id TEXT PRIMARY KEY,
		file_path TEXT,
		atom_type TEXT,
		embedding BLOB,
		bloat_result TEXT,
		last_modified INTEGER
	)`)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS collision_cache (
		file_a TEXT,
		file_b TEXT,
		mtime_a INTEGER,
		mtime_b INTEGER,
		similarity REAL,
		result TEXT,
		PRIMARY KEY (file_a, file_b)
	)`)
	if err != nil {
		return fmt.Errorf("failed to create collision_cache: %v", err)
	}

	var files []string
	if workspace && config.ActiveConfig.Workspace != nil {
		for _, p := range config.ActiveConfig.Workspace.Projects {
			absProjPath := p.Path
			if !filepath.IsAbs(absProjPath) {
				absProjPath = filepath.Join(config.ActiveConfig.Workspace.LoadedFrom, p.Path)
			}
			pDocs := p.DocsPath
			if pDocs == "" {
				pDocs = "docs/"
			}
			if !filepath.IsAbs(pDocs) {
				pDocs = filepath.Join(absProjPath, pDocs)
			}

			pFiles, _ := filepath.Glob(filepath.Join(pDocs, "*.atom.md"))
			files = append(files, pFiles...)
		}
	} else {
		files, _ = filepath.Glob(filepath.Join(docsDir, "*.atom.md"))
	}

	if len(files) == 0 {
		return fmt.Errorf("no atoms found")
	}

	fmt.Println("=== ATD AUDIT PROTOCOL INITIATED ===")
	fmt.Printf("Phase 1: The Bloat Metric (docs count: %d)\n", len(files))

	auditMetas := make(map[string]atomAuditMeta)
	var ids []string

	for _, f := range files {
		filename := filepath.Base(f)
		info, _ := os.Stat(f)
		mtime := info.ModTime().Unix()

		var cached atomAuditMeta
		var embBytes []byte
		err := db.QueryRow("SELECT atom_type, embedding, bloat_result, last_modified FROM audit_cache WHERE id = ?", filename).Scan(&cached.AtomType, &embBytes, &cached.BloatResult, &cached.LastModified)

		if err == nil && mtime <= cached.LastModified {
			json.Unmarshal(embBytes, &cached.Embedding)
			// Need parents for BFS walk
			data, _ := atom.Parse(f)
			cached.ID = data.ID
			cached.FilePath = f
			cached.Parents = data.Parents
			auditMetas[filename] = cached
			ids = append(ids, filename)
			fmt.Printf("Auditing: %s ... [CACHED: %s]\n", filename, cached.BloatResult)
			continue
		}

		// Re-audit
		data, err := atom.Parse(f)
		if err != nil {
			fmt.Printf("Auditing: %s ... [ERROR: %v]\n", filename, err)
			continue
		}

		strictness := config.GetBloatingStrictness(data.Type)
		bloatResult := "PASS"
		// @spec-link [[rule_atd_atom_overrides]]
		if data.Bloating == "off" {
			bloatResult = "SKIP"
		} else if strictness > 0 {
			// Phase 1: Bloat Detection
			intentPrompt := prompt.AuditBloatBuild("Architectural Linter", data.Intent, strictness)
			logicPrompt := prompt.AuditBloatBuild("Architectural Linter", data.Logic, strictness)

			resI, errI := ollama.Query("text_analysis", intentPrompt, prompt.AuditBloatFormat())
			resL, errL := ollama.Query("text_analysis", logicPrompt, prompt.AuditBloatFormat())

			if errI == ollama.ErrIDEFallback || errL == ollama.ErrIDEFallback {
				promptName := "audit_bloat_" + data.ID
				pipeline.WritePromptFile(promptName, intentPrompt+"\n\n"+logicPrompt)
				bloatResult = "PENDING_IDE"
			} else if errI == nil && errL == nil {
				var bI, bL struct {
					IsBloated bool `json:"is_bloated"`
				}
				json.Unmarshal([]byte(resI.Response), &bI)
				json.Unmarshal([]byte(resL.Response), &bL)

				if bI.IsBloated || bL.IsBloated {
					bloatResult = "BLOATED"
				}
			}
		}

		// Embedding
		pEmbed, _ := ollama.ResolveProvider("embed")
		var emb []float64
		if !pEmbed.IsIDE {
			// We embed the combined content for Phase 2
			content, _ := os.ReadFile(f)
			emb, _ = ollama.QueryEmbed(string(content))
		}

		meta := atomAuditMeta{
			ID:           data.ID,
			FilePath:     f,
			AtomType:     data.Type,
			Parents:      data.Parents,
			Embedding:    emb,
			BloatResult:  bloatResult,
			LastModified: mtime,
		}
		auditMetas[filename] = meta
		ids = append(ids, filename)

		embBytes, _ = json.Marshal(emb)
		db.Exec(`INSERT INTO audit_cache (id, file_path, atom_type, embedding, bloat_result, last_modified) 
			VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET 
			atom_type=excluded.atom_type, embedding=excluded.embedding, 
			bloat_result=excluded.bloat_result, last_modified=excluded.last_modified`,
			filename, f, data.Type, embBytes, bloatResult, mtime)

		fmt.Printf("Auditing: %s ... [%s]\n", filename, bloatResult)
	}

	fmt.Println("\nPhase 2: The Collision Map (Semantic Overlap Detection)")
	
	// Parent Index for BFS
	parentMap := make(map[string][]string)
	for _, m := range auditMetas {
		parentMap[m.ID] = m.Parents
	}

	isAncestor := func(startID, targetID string) bool {
		visited := make(map[string]bool)
		queue := []string{startID}
		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			if visited[curr] { continue }
			visited[curr] = true
			for _, p := range parentMap[curr] {
				if p == targetID { return true }
				queue = append(queue, p)
			}
		}
		return false
	}

	collisionsFound := false
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			f1, f2 := ids[i], ids[j]
			m1, m2 := auditMetas[f1], auditMetas[f2]

			if m1.AtomType != m2.AtomType { continue }
			if len(m1.Embedding) == 0 || len(m2.Embedding) == 0 { continue }

			sim := cosine.Similarity(m1.Embedding, m2.Embedding)
			if sim < threshold { continue }

			// Potential collision. Check if structurally sound.
			isRelated := isAncestor(m1.ID, m2.ID) || isAncestor(m2.ID, m1.ID)
			
			// Check if they share an immediate parent
			sharedParent := ""
			pMap := make(map[string]bool)
			for _, p := range m1.Parents { pMap[p] = true }
			for _, p := range m2.Parents {
				if pMap[p] {
					sharedParent = p
					break
				}
			}

			if !isRelated && sharedParent == "" {
				collisionsFound = true
				fmt.Printf("\n[COLLISION] %s <--> %s (Similarity: %.2f)\n", f1, f2, sim)
				fmt.Printf("  Result: [MISSING ABSTRACTION] Atoms share %d%% logic but lack shared parent.\n", int(sim*100))
			}
		}
	}

	if !collisionsFound {
		fmt.Println("\nNo critical semantic collisions detected.")
	}

	fmt.Println("\n=== AUDIT COMPLETE ===")
	return nil
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.Flags().Float64("threshold", 0, "Similarity threshold (0.0 - 1.0)")
	auditCmd.Flags().String("docs", "", "Path to docs directory")
	auditCmd.Flags().Bool("workspace", false, "Audit all atoms in workspace")
	auditCmd.Flags().String("code", "", "Snippet path for compliance check")
	auditCmd.Flags().String("atom", "", "Atom path for compliance check")
}
