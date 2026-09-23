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
	"sync"
)

// DefaultAuditConcurrency bounds how many files Phase 1 audits at once when
// no explicit concurrency is requested (concurrency <= 0). Phase 1's per-file
// work is dominated by up to 3 sequential Ollama HTTP round-trips (2 bloat
// judges + 1 embed), each with up to a 120s timeout -- this bounds
// simultaneous outbound HTTP calls to the LLM backend, not CPU-bound work, so
// a small fixed pool is enough to get most of the wall-clock win without
// hammering a local Ollama instance with dozens of concurrent requests.
const DefaultAuditConcurrency = 4

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

func RunFullAudit(docsDir string, threshold float64, workspace bool, concurrency int) (*AuditReport, error) {
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

	return runAudit(files, docsDir, threshold, concurrency)
}

// RunScopedAudit runs the same bloat-detection and collision-detection
// passes as RunFullAudit but scoped to exactly one atom file, instead of
// globbing every "*.atom.md" file under a docs directory. It backs the CLI's
// narrower `--atom` compliance-check mode (see cmd/atd/cmd/audit.go):
// pointing `--atom` at a single file must never fall back to a full
// directory sweep.
//
// pkg/audit has no atom-vs-code compliance comparison capability today (no
// prompt or scoring path takes a code snippet as input), so a `--code`
// argument for a true single-atom-vs-single-file comparison is intentionally
// not accepted here yet -- only atom-level scoping is provided.
func RunScopedAudit(atomPath string, threshold float64, concurrency int) (*AuditReport, error) {
	if _, err := os.Stat(atomPath); err != nil {
		return nil, fmt.Errorf("atom path not found: %v", err)
	}
	return runAudit([]string{atomPath}, filepath.Dir(atomPath), threshold, concurrency)
}

// fileAuditResult is one worker's Phase-1 output for a single file: report
// text to merge into the shared output (in original file order), and the
// state (meta/counters) runAudit needs to fold into auditMetas/ids/the
// summary counters after the worker pool drains. meta is nil whenever the
// file was skipped (stat/parse failure, malformed bloat-check response,
// etc.) -- mirroring the original sequential loop's `continue` points, which
// left such files out of auditMetas/ids entirely.
type fileAuditResult struct {
	filename   string
	text       string
	meta       *atomAuditMeta
	bloated    bool
	queryError bool
	embedError bool
}

// auditOneFile runs Phase 1's bloat-check/embedding pipeline for a single
// file. It is the unit of work scheduled by runAudit's worker pool, and its
// control flow is an exact copy of the original sequential loop body -- only
// the scheduling around it changed -- so every existing report line, error
// message, and cache-hit/miss behavior is preserved verbatim per file.
func auditOneFile(f string, store *atdstore.Store) fileAuditResult {
	var out strings.Builder
	filename := filepath.Base(f)
	res := fileAuditResult{filename: filename}

	info, statErr := os.Stat(f)
	if statErr != nil {
		out.WriteString(fmt.Sprintf("Auditing: %s ... [ERROR: %v]\n", filename, statErr))
		res.text = out.String()
		return res
	}
	mtime := info.ModTime().Unix()

	cached, err := store.GetAuditCache(filename)
	if err == nil && cached != nil && mtime <= cached.Mtime {
		data, err := atom.Parse(f)
		if err != nil {
			out.WriteString(fmt.Sprintf("Auditing: %s ... [ERROR: %v]\n", filename, err))
			res.text = out.String()
			return res
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
		res.meta = &meta
		out.WriteString(fmt.Sprintf("Auditing: %s ... [CACHED]\n", filename))
		res.text = out.String()
		return res
	}

	data, err := atom.Parse(f)
	if err != nil {
		out.WriteString(fmt.Sprintf("Auditing: %s ... [ERROR: %v]\n", filename, err))
		res.text = out.String()
		return res
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
				out.WriteString(fmt.Sprintf("  [ERROR] Failed to parse intent response: %v\n", err))
				res.text = out.String()
				return res
			}
			if !hasKey(resI.Response, "is_bloated") {
				// Syntactically valid JSON that simply omits "is_bloated"
				// unmarshals with no error at all -- bI.IsBloated would
				// silently stay its zero value (false), classifying the
				// atom PASS even though the model's response was
				// unusable. Treat a missing required key the same as
				// malformed JSON: a loud [ERROR] line, never a silent
				// PASS.
				out.WriteString(fmt.Sprintf("  [ERROR] Intent response missing required \"is_bloated\" key: %.200s\n", resI.Response))
				res.text = out.String()
				return res
			}
			if err := json.Unmarshal([]byte(resL.Response), &bL); err != nil {
				out.WriteString(fmt.Sprintf("  [ERROR] Failed to parse logic response: %v\n", err))
				res.text = out.String()
				return res
			}
			if !hasKey(resL.Response, "is_bloated") {
				out.WriteString(fmt.Sprintf("  [ERROR] Logic response missing required \"is_bloated\" key: %.200s\n", resL.Response))
				res.text = out.String()
				return res
			}

			if bI.IsBloated || bL.IsBloated {
				bloatResult = "BLOATED"
			}
		} else {
			// A generic (non-IDE-fallback) error from either query --
			// timeout, connection refused, malformed backend response,
			// etc. Falling through here would leave bloatResult at its
			// pre-set "PASS" default with zero trace in the report,
			// making a broken run indistinguishable from a genuinely
			// clean one (failures/20260917_atd_audit_docs_exits_zero_with_no_report.md).
			bloatResult = "ERROR"
			res.queryError = true
			var errMsg strings.Builder
			if errI != nil {
				errMsg.WriteString(fmt.Sprintf("intent query: %v", errI))
			}
			if errL != nil {
				if errMsg.Len() > 0 {
					errMsg.WriteString("; ")
				}
				errMsg.WriteString(fmt.Sprintf("logic query: %v", errL))
			}
			out.WriteString(fmt.Sprintf("  [ERROR] LLM bloat check failed for %s: %s\n", data.ID, errMsg.String()))
		}
	}
	if bloatResult == "BLOATED" {
		res.bloated = true
	}

	pEmbed, embResolveErr := ollama.ResolveProvider("embed")
	var emb []float32
	if embResolveErr == nil && !pEmbed.IsIDE {
		content, readErr := os.ReadFile(f)
		if readErr != nil {
			out.WriteString(fmt.Sprintf("  [ERROR] Failed to read %s for embedding: %v\n", filename, readErr))
			res.embedError = true
		} else {
			var embErr error
			emb, embErr = ollama.QueryEmbed(string(content))
			if embErr != nil {
				// Silently discarding this (as `_`) meant an atom could
				// drop out of collision detection with zero visibility
				// into why -- surface it loudly instead.
				out.WriteString(fmt.Sprintf("  [ERROR] Embedding failed for %s: %v (excluded from collision detection)\n", data.ID, embErr))
				res.embedError = true
			}
		}
	} else if embResolveErr != nil {
		out.WriteString(fmt.Sprintf("  [ERROR] Failed to resolve embed provider for %s: %v (excluded from collision detection)\n", data.ID, embResolveErr))
		res.embedError = true
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
	res.meta = &meta

	intentStr := ""
	logicStr := ""
	if bloatResult == "BLOATED" {
		intentStr = data.Intent
		logicStr = data.Logic
	}
	store.PutAuditCache(filename, emb, intentStr, logicStr, mtime)

	out.WriteString(fmt.Sprintf("Auditing: %s ... [%s]\n", filename, bloatResult))
	res.text = out.String()
	return res
}

// runAudit performs the shared bloat-detection (Phase 1) and
// collision-detection (Phase 2) analysis over an explicit file list. dbDir
// selects where the audit cache (.atd_audit.db) is stored. concurrency bounds
// how many files Phase 1 processes in parallel; concurrency <= 0 falls back
// to DefaultAuditConcurrency.
func runAudit(files []string, dbDir string, threshold float64, concurrency int) (*AuditReport, error) {
	if concurrency <= 0 {
		concurrency = DefaultAuditConcurrency
	}
	var output strings.Builder
	dbPath := filepath.Join(dbDir, ".atd_audit.db")
	store, err := atdstore.NewStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open store: %v", err)
	}
	defer store.Close()

	if len(files) == 0 {
		return nil, fmt.Errorf("no atoms found")
	}

	output.WriteString("=== ATD AUDIT PROTOCOL INITIATED ===\n")
	output.WriteString(fmt.Sprintf("Phase 1: The Bloat Metric (docs count: %d)\n", len(files)))

	auditMetas := make(map[string]atomAuditMeta)
	var ids []string
	var bloatedCount, collisionCount, queryErrorCount, embedErrorCount int

	// Phase 1 runs across a bounded worker pool (semaphore + WaitGroup) so up
	// to `concurrency` files are in flight at once against Ollama. Each
	// worker writes its report text into results[i] -- its own slot, keyed
	// by the file's original index in `files` -- rather than the shared
	// `output` builder (strings.Builder is not safe for concurrent writes,
	// and even if it were, completion order across goroutines is
	// nondeterministic). Results are appended to `output` in original file
	// order after every worker has finished, so report text stays
	// byte-for-byte identical regardless of concurrency or scheduling.
	results := make([]fileAuditResult, len(files))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for i, f := range files {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, f string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = auditOneFile(f, store)
		}(i, f)
	}
	wg.Wait()

	// Merge step: single-threaded, so auditMetas/ids/counters need no lock of
	// their own even though the work that produced each result ran
	// concurrently.
	for _, r := range results {
		output.WriteString(r.text)
		if r.meta != nil {
			auditMetas[r.filename] = *r.meta
			ids = append(ids, r.filename)
		}
		if r.bloated {
			bloatedCount++
		}
		if r.queryError {
			queryErrorCount++
		}
		if r.embedError {
			embedErrorCount++
		}
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
				collisionCount++
				output.WriteString(fmt.Sprintf("\n[COLLISION] %s <--> %s (Similarity: %.2f)\n", f1, f2, sim))
				output.WriteString(fmt.Sprintf("  Result: [MISSING ABSTRACTION] Atoms share %d%% logic but lack shared parent.\n", int(sim*100)))
			}
		}
	}

	if !collisionsFound {
		output.WriteString("\nNo critical semantic collisions detected.\n")
	}

	// Guaranteed final line: every run, success or not, must state what
	// happened so a clean 0-findings pass is never indistinguishable from a
	// silently-broken one (failures/20260917_atd_audit_docs_exits_zero_with_no_report.md,
	// failures/20260916_atd_audit_workspace_no_return.md).
	totalErrors := queryErrorCount + embedErrorCount
	output.WriteString(fmt.Sprintf("\nSummary: %d atom(s) scanned, %d bloated, %d collision(s), %d LLM error(s) (%d bloat-check, %d embedding).\n",
		len(ids), bloatedCount, collisionCount, totalErrors, queryErrorCount, embedErrorCount))

	output.WriteString("\n=== AUDIT COMPLETE ===\n")
	return &AuditReport{Text: output.String()}, nil
}