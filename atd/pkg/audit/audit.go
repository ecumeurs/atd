package audit

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/cosine"
	"atd-tools/pkg/ollama"
	atdstore "atd-tools/pkg/store"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AuditReport struct {
	Text string
}

// hasKey reports whether the raw JSON object in resp declares key at its top
// level. Go's encoding/json does not error when a struct's tagged field is
// absent from the source object -- it just leaves the field at its zero
// value -- so a typed Unmarshal alone can't distinguish "the model said
// false" from "the model never answered". Probing into a generic map first
// closes that gap for the "is_bloated" parse below.
func hasKey(resp, key string) bool {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(resp), &probe); err != nil {
		return false
	}
	_, ok := probe[key]
	return ok
}

type atomAuditMeta struct {
	ID           string
	FilePath     string
	AtomType     string
	Parents      []string
	Embedding    []float32
	BloatResult  string
	LastModified int64
}

func RunFullAudit(docsDir string, threshold float64, workspace bool) (*AuditReport, error) {
	var output strings.Builder
	dbPath := filepath.Join(docsDir, ".atd_audit.db")
	store, err := atdstore.NewStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open store: %v", err)
	}
	defer store.Close()

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
		return nil, fmt.Errorf("no atoms found")
	}

	output.WriteString("=== ATD AUDIT PROTOCOL INITIATED ===\n")
	output.WriteString(fmt.Sprintf("Phase 1: The Bloat Metric (docs count: %d)\n", len(files)))

	auditMetas := make(map[string]atomAuditMeta)
	var ids []string

	for _, f := range files {
		filename := filepath.Base(f)
		info, _ := os.Stat(f)
		mtime := info.ModTime().Unix()

		cached, err := store.GetAuditCache(filename)
		if err == nil && cached != nil && mtime <= cached.Mtime {
			data, err := atom.Parse(f)
			if err != nil {
				output.WriteString(fmt.Sprintf("Auditing: %s ... [ERROR: %v]\n", filename, err))
				continue
			}
			atomType := data.Type
			if atomType == "" {
				atomType = "UNKNOWN"
			}
			meta := atomAuditMeta{
				ID:           data.ID,
				FilePath:     f,
				AtomType:     atomType,
				Parents:      data.Parents,
				Embedding:    cached.Embedding,
				BloatResult:  fmt.Sprintf("CACHED: intent=%s logic=%s", cached.Intent, cached.Logic),
				LastModified: cached.Mtime,
			}
			auditMetas[filename] = meta
			ids = append(ids, filename)
			output.WriteString(fmt.Sprintf("Auditing: %s ... [CACHED]\n", filename))
			continue
		}

		data, err := atom.Parse(f)
		if err != nil {
			output.WriteString(fmt.Sprintf("Auditing: %s ... [ERROR: %v]\n", filename, err))
			continue
		}

		atomType := data.Type
		if atomType == "" {
			atomType = "UNKNOWN"
		}

		strictness := config.GetBloatingStrictness(atomType)
		bloatResult := "PASS"
		if data.Bloating == "off" {
			bloatResult = "SKIP"
		} else if strictness > 0 {
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
				if err := json.Unmarshal([]byte(resI.Response), &bI); err != nil {
					output.WriteString(fmt.Sprintf("  [ERROR] Failed to parse intent response: %v\n", err))
					continue
				}
				if !hasKey(resI.Response, "is_bloated") {
					// Syntactically valid JSON that simply omits "is_bloated"
					// unmarshals with no error at all -- bI.IsBloated would
					// silently stay its zero value (false), classifying the
					// atom PASS even though the model's response was
					// unusable. Treat a missing required key the same as
					// malformed JSON: a loud [ERROR] line, never a silent
					// PASS.
					output.WriteString(fmt.Sprintf("  [ERROR] Intent response missing required \"is_bloated\" key: %.200s\n", resI.Response))
					continue
				}
				if err := json.Unmarshal([]byte(resL.Response), &bL); err != nil {
					output.WriteString(fmt.Sprintf("  [ERROR] Failed to parse logic response: %v\n", err))
					continue
				}
				if !hasKey(resL.Response, "is_bloated") {
					output.WriteString(fmt.Sprintf("  [ERROR] Logic response missing required \"is_bloated\" key: %.200s\n", resL.Response))
					continue
				}

				if bI.IsBloated || bL.IsBloated {
					bloatResult = "BLOATED"
				}
			}
		}

		pEmbed, _ := ollama.ResolveProvider("embed")
		var emb []float32
		if !pEmbed.IsIDE {
			content, _ := os.ReadFile(f)
			emb, _ = ollama.QueryEmbed(string(content))
		}

		meta := atomAuditMeta{
			ID:           data.ID,
			FilePath:     f,
			AtomType:     atomType,
			Parents:      data.Parents,
			Embedding:    emb,
			BloatResult:  bloatResult,
			LastModified: mtime,
		}
		auditMetas[filename] = meta
		ids = append(ids, filename)

		intentStr := ""
		logicStr := ""
		if bloatResult == "BLOATED" {
			intentStr = data.Intent
			logicStr = data.Logic
		}
		store.PutAuditCache(filename, emb, intentStr, logicStr, mtime)

		output.WriteString(fmt.Sprintf("Auditing: %s ... [%s]\n", filename, bloatResult))
	}

	output.WriteString("\nPhase 2: The Collision Map (Semantic Overlap Detection)\n")

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

			isRelated := isAncestor(m1.ID, m2.ID) || isAncestor(m2.ID, m1.ID)

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
				output.WriteString(fmt.Sprintf("\n[COLLISION] %s <--> %s (Similarity: %.2f)\n", f1, f2, sim))
				output.WriteString(fmt.Sprintf("  Result: [MISSING ABSTRACTION] Atoms share %d%% logic but lack shared parent.\n", int(sim*100)))
			}
		}
	}

	if !collisionsFound {
		output.WriteString("\nNo critical semantic collisions detected.\n")
	}

	output.WriteString("\n=== AUDIT COMPLETE ===\n")
	return &AuditReport{Text: output.String()}, nil
}