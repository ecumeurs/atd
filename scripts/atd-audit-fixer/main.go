package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// ─── Ollama ───────────────────────────────────────────────────────────────────

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

func queryLlama(prompt string) (string, error) {
	req := GenerateRequest{
		Model:  "llama3.2",
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}
	body, _ := json.Marshal(req)
	resp, err := http.Post("http://127.0.0.1:11434/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var gen GenerateResponse
	if err := json.Unmarshal(raw, &gen); err != nil {
		return "", err
	}
	return gen.Response, nil
}

// ─── Split response schema ────────────────────────────────────────────────────

type SplitAtom struct {
	IDSuffix  string `json:"id_suffix"`
	HumanName string `json:"human_name"`
	Intent    string `json:"intent"`
	Logic     string `json:"logic"`
}

type SplitResponse struct {
	ParentLogic string      `json:"parent_logic"`
	Splits      []SplitAtom `json:"splits"`
}

// ─── Atom file builder ────────────────────────────────────────────────────────

func buildAtomContent(id, humanName, atomType, parentID, intent, logic string) string {
	return fmt.Sprintf(`---
id: %s
human_name: %s
type: %s
version: 1.0
status: DRAFT
priority: CORE
tags: []
parents: 
  - [[%s]]
dependents: []
---

# %s

## INTENT
%s

## THE RULE / LOGIC
%s

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `+"`"+`@spec-link [[%s]]`+"`"+`
`, id, humanName, atomType, parentID, humanName, intent, logic, id)
}

func buildParentContent(id, humanName, atomType, intent, logic string) string {
	return fmt.Sprintf(`---
id: %s
human_name: %s
type: MODULE
version: 1.0
status: STABLE
priority: CORE
tags: []
parents: []
dependents: []
---

# %s

## INTENT
%s

## THE RULE / LOGIC
%s

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `+"`"+`@spec-link [[%s]]`+"`"+`
`, id, humanName, humanName, intent, logic, id)
}

// ─── Audit report parser ──────────────────────────────────────────────────────

func parseBloatedFiles(reportPath string) ([]string, error) {
	f, err := os.Open(reportPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var bloated []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Matches lines like:
		//   Auditing: foo.atom.md ... [BLOATED] ...
		if strings.Contains(line, "[BLOATED]") {
			parts := strings.SplitN(line, "Auditing:", 2)
			if len(parts) == 2 {
				rest := strings.TrimSpace(parts[1])
				filename := strings.Fields(rest)[0]
				bloated = append(bloated, strings.TrimSpace(filename))
			}
		}
	}
	return bloated, scanner.Err()
}

// ─── Atom content reader ──────────────────────────────────────────────────────

func readAtomMeta(path string) (id, humanName, atomType string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "id:") {
			id = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		} else if strings.HasPrefix(line, "human_name:") {
			humanName = strings.TrimSpace(strings.TrimPrefix(line, "human_name:"))
		} else if strings.HasPrefix(line, "type:") {
			atomType = strings.TrimSpace(strings.TrimPrefix(line, "type:"))
		}
		if id != "" && humanName != "" && atomType != "" {
			break
		}
	}
	return
}

// ─── DB cache invalidation ────────────────────────────────────────────────────

func invalidateCache(db *sql.DB, filename string) {
	db.Exec("DELETE FROM atom_docs_index WHERE id = ?", filename)
	// Also wipe any collision pairs involving this file
	db.Exec("DELETE FROM collision_cache WHERE file_a = ? OR file_b = ?", filename, filename)
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	auditReport := flag.String("audit", "", "Path to atd-audit output file (contains [BLOATED] lines)")
	docsDir := flag.String("docs", "", "Path to the docs directory")
	dbPath := flag.String("db", "", "Path to the SQLite cache DB (default: docs/.atd_docs_index.db)")
	dryRun := flag.Bool("dry-run", false, "Print proposed changes without writing files")
	flag.Parse()

	if *auditReport == "" || *docsDir == "" {
		fmt.Println("Usage: atd-audit-fixer -audit <report.txt> -docs <docs_dir> [-db <db>] [-dry-run]")
		os.Exit(1)
	}

	resolvedDB := *dbPath
	if resolvedDB == "" {
		resolvedDB = filepath.Join(*docsDir, ".atd_docs_index.db")
	}

	db, err := sql.Open("sqlite3", resolvedDB)
	if err != nil {
		fmt.Printf("Warning: could not open DB at %s: %v — cache invalidation skipped.\n", resolvedDB, err)
		db = nil
	}
	if db != nil {
		defer db.Close()
	}

	bloated, err := parseBloatedFiles(*auditReport)
	if err != nil {
		fmt.Printf("Error reading audit report: %v\n", err)
		os.Exit(1)
	}

	if len(bloated) == 0 {
		fmt.Println("✓ No bloated atoms found in audit report. Nothing to fix.")
		return
	}

	fmt.Printf("Found %d bloated atom(s) to fix.\n\n", len(bloated))

	for _, filename := range bloated {
		atomPath := filepath.Join(*docsDir, filename)
		content, err := os.ReadFile(atomPath)
		if err != nil {
			fmt.Printf("[SKIP] Cannot read %s: %v\n", filename, err)
			continue
		}

		baseID, humanName, atomType, err := readAtomMeta(atomPath)
		if err != nil || baseID == "" {
			fmt.Printf("[SKIP] Cannot parse metadata from %s\n", filename)
			continue
		}

		fmt.Printf("── Processing: %s (%s) ──\n", filename, baseID)

		prompt := fmt.Sprintf(`You are an ATD (Atomic Traceable Documentation) architect.
The following ATD atom VIOLATES the "Minimum Atomic Scale" rule — its RULE/LOGIC section describes more than one distinct state-changing rule.

Your task: split it into N focused child atoms, each with EXACTLY ONE rule in its logic section.
The original atom will be rewritten as a MODULE parent that aggregates its children.

Rules:
- Each split atom must have a single, non-compound intent sentence
- Each split atom must have no more than one distinct rule in its logic section
- Output ONLY valid JSON. No markdown. No explanation text.

JSON schema:
{
  "parent_logic": "<one-sentence summary of what the children share>",
  "splits": [
    {"id_suffix": "<short_snake_case>", "human_name": "<Human Name>", "intent": "<single purpose sentence>", "logic": "<single rule text>"}
  ]
}

Atom to split:
%s`, string(content))

		fmt.Printf("  → Asking llama3.2 to propose split...\n")
		rawResp, err := queryLlama(prompt)
		if err != nil {
			fmt.Printf("  [ERROR] LLM call failed: %v\n", err)
			continue
		}

		var splitResp SplitResponse
		if err := json.Unmarshal([]byte(rawResp), &splitResp); err != nil {
			fmt.Printf("  [ERROR] Could not parse LLM JSON response: %v\nRaw: %s\n", err, rawResp)
			continue
		}

		if len(splitResp.Splits) == 0 {
			fmt.Printf("  [WARN] LLM returned no splits for %s\n", filename)
			continue
		}

		fmt.Printf("  → %d split(s) proposed:\n", len(splitResp.Splits))
		for _, s := range splitResp.Splits {
			newID := baseID + "_" + s.IDSuffix
			newFilename := newID + ".atom.md"
			newPath := filepath.Join(*docsDir, newFilename)
			fmt.Printf("    • %s  →  %s\n", s.HumanName, newFilename)

			if !*dryRun {
				atomContent := buildAtomContent(newID, s.HumanName, atomType, baseID, s.Intent, s.Logic)
				if err := os.WriteFile(newPath, []byte(atomContent), 0644); err != nil {
					fmt.Printf("    [ERROR] Could not write %s: %v\n", newPath, err)
				}
			}
		}

		// Rewrite original as MODULE parent
		parentIntent := fmt.Sprintf("To aggregate the constituent rules of %s.", humanName)
		parentLogic := splitResp.ParentLogic
		if !*dryRun {
			parentContent := buildParentContent(baseID, humanName, atomType, parentIntent, parentLogic)
			if err := os.WriteFile(atomPath, []byte(parentContent), 0644); err != nil {
				fmt.Printf("  [ERROR] Could not rewrite %s: %v\n", atomPath, err)
			} else {
				fmt.Printf("  ✓ Rewrote %s as MODULE parent\n", filename)
			}
			// Invalidate DB cache
			if db != nil {
				invalidateCache(db, filename)
				fmt.Printf("  ✓ Cache invalidated for %s\n", filename)
			}
		} else {
			fmt.Printf("  [DRY-RUN] Would rewrite %s as MODULE parent\n", filename)
		}
		fmt.Println()
	}

	fmt.Println("Done. Re-run atd-audit to verify.")
}
