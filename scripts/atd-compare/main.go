package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
)

// ─── Ollama ───────────────────────────────────────────────────────────────────

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
}

func queryLlama(prompt string) (string, error) {
	req := GenerateRequest{
		Model:  "llama3.2",
		Prompt: prompt,
		Stream: false,
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

// ─── Atom parser ──────────────────────────────────────────────────────────────

type AtomData struct {
	ID        string
	HumanName string
	Type      string
	Tags      string
	Parents   []string
	Intent    string
	Logic     string
}

func parseAtom(path string) (AtomData, error) {
	f, err := os.Open(path)
	if err != nil {
		return AtomData{}, err
	}
	defer f.Close()

	var data AtomData
	scanner := bufio.NewScanner(f)
	mode := "header"
	inParents := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "id:") {
			data.ID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
			inParents = false
			continue
		}
		if strings.HasPrefix(line, "human_name:") {
			data.HumanName = strings.TrimSpace(strings.TrimPrefix(line, "human_name:"))
			inParents = false
			continue
		}
		if strings.HasPrefix(line, "type:") {
			data.Type = strings.TrimSpace(strings.TrimPrefix(line, "type:"))
			inParents = false
			continue
		}
		if strings.HasPrefix(line, "tags:") {
			data.Tags = strings.TrimSpace(strings.TrimPrefix(line, "tags:"))
			inParents = false
			continue
		}
		if strings.HasPrefix(line, "parents:") {
			inParents = true
			inline := strings.TrimSpace(strings.TrimPrefix(line, "parents:"))
			if inline != "" && inline != "[]" {
				inline = strings.ReplaceAll(inline, "[", "")
				inline = strings.ReplaceAll(inline, "]", "")
				for _, p := range strings.Split(inline, ",") {
					if t := strings.TrimSpace(p); t != "" {
						data.Parents = append(data.Parents, t)
					}
				}
			}
			continue
		}
		if inParents {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") {
				entry := strings.TrimPrefix(trimmed, "- ")
				entry = strings.ReplaceAll(entry, "[", "")
				entry = strings.ReplaceAll(entry, "]", "")
				data.Parents = append(data.Parents, strings.TrimSpace(entry))
				continue
			}
			inParents = false
		}

		if mode == "header" && line == "---" {
			// skip front-matter delimiters
			continue
		}

		if strings.HasPrefix(line, "## INTENT") {
			mode = "intent"
			continue
		}
		if strings.HasPrefix(line, "## THE RULE") {
			mode = "logic"
			continue
		}
		if strings.HasPrefix(line, "##") && (mode == "intent" || mode == "logic") {
			mode = "done"
		}

		if mode == "intent" {
			data.Intent += line + "\n"
		} else if mode == "logic" {
			data.Logic += line + "\n"
		}
	}
	data.Intent = strings.TrimSpace(data.Intent)
	data.Logic = strings.TrimSpace(data.Logic)
	return data, scanner.Err()
}

// ─── Keywords extractor (simple) ─────────────────────────────────────────────

func sharedKeywords(a, b string) []string {
	freq := make(map[string]int)
	stopwords := map[string]bool{
		"the": true, "a": true, "an": true, "to": true, "of": true, "and": true,
		"is": true, "in": true, "for": true, "that": true, "this": true,
		"it": true, "be": true, "are": true, "at": true, "by": true,
	}
	tokenize := func(s string) []string {
		s = strings.ToLower(s)
		var out []string
		for _, w := range strings.Fields(s) {
			w = strings.Trim(w, ".,;:\"'()[]{}*-")
			if len(w) > 3 && !stopwords[w] {
				out = append(out, w)
			}
		}
		return out
	}
	for _, w := range tokenize(a) {
		freq[w]++
	}
	var shared []string
	seen := make(map[string]bool)
	for _, w := range tokenize(b) {
		if freq[w] > 0 && !seen[w] {
			shared = append(shared, w)
			seen[w] = true
		}
	}
	if len(shared) > 8 {
		shared = shared[:8]
	}
	return shared
}

// ─── Resolution prompt ────────────────────────────────────────────────────────

func proposeResolution(a, b AtomData) string {
	prompt := fmt.Sprintf(`You are an ATD (Atomic Traceable Documentation) architect reviewing two semantically overlapping atoms.

Atom A — %s (%s):
Intent: %s
Logic: %s

Atom B — %s (%s):
Intent: %s
Logic: %s

These two atoms have high semantic similarity. Diagnose the relationship and propose a concrete resolution in plain text (3-5 sentences).
Choose the most appropriate category: MISSING_COMMON_PARENT, MERGE, REFACTOR, or ACCEPTABLE_SIBLING.
If MISSING_COMMON_PARENT, suggest a parent id and one-line intent for it.
If MERGE, specify which atom should absorb the other and why.
If REFACTOR, specify what specific content should move where.
If ACCEPTABLE_SIBLING, explain why the similarity is harmless.`,
		a.HumanName, a.Type, a.Intent, a.Logic,
		b.HumanName, b.Type, b.Intent, b.Logic,
	)

	resp, err := queryLlama(prompt)
	if err != nil {
		return fmt.Sprintf("_(LLM error: %v)_", err)
	}
	return strings.TrimSpace(resp)
}

// ─── Markdown report builder ──────────────────────────────────────────────────

func fenced(s string) string {
	return "```\n" + s + "\n```"
}

func buildReport(a, b AtomData, aPath, bPath, resolution string) string {
	keywords := sharedKeywords(a.Intent+" "+a.Logic, b.Intent+" "+b.Logic)
	kwStr := strings.Join(keywords, ", ")
	if kwStr == "" {
		kwStr = "_none detected_"
	}

	parentsA := "_none_"
	if len(a.Parents) > 0 {
		parentsA = "`" + strings.Join(a.Parents, "`, `") + "`"
	}
	parentsB := "_none_"
	if len(b.Parents) > 0 {
		parentsB = "`" + strings.Join(b.Parents, "`, `") + "`"
	}

	out := "# ATD Collision Report\n\n"
	out += "> **Generated by:** atd-compare  \n"
	out += "> **Status:** ⚠️ MISSING ABSTRACTION DETECTED\n\n---\n\n"

	out += "## Atom A — " + a.HumanName + "\n\n"
	out += "| Field | Value |\n|-------|-------|\n"
	out += "| **ID** | `" + a.ID + "` |\n"
	out += "| **Type** | " + a.Type + " |\n"
	out += "| **Parents** | " + parentsA + " |\n"
	out += "| **File** | `" + filepath.Base(aPath) + "` |\n\n"
	out += "### Intent\n> " + a.Intent + "\n\n"
	out += "### Rule / Logic\n" + fenced(a.Logic) + "\n\n---\n\n"

	out += "## Atom B — " + b.HumanName + "\n\n"
	out += "| Field | Value |\n|-------|-------|\n"
	out += "| **ID** | `" + b.ID + "` |\n"
	out += "| **Type** | " + b.Type + " |\n"
	out += "| **Parents** | " + parentsB + " |\n"
	out += "| **File** | `" + filepath.Base(bPath) + "` |\n\n"
	out += "### Intent\n> " + b.Intent + "\n\n"
	out += "### Rule / Logic\n" + fenced(b.Logic) + "\n\n---\n\n"

	out += "## Overlap Analysis\n\n"
	out += "| Shared Keywords |\n|----------------|\n"
	out += "| " + kwStr + " |\n\n"
	out += "> Both atoms are of the same type (**" + a.Type + "**) and share no common parent.  \n"
	out += "> This indicates a missing abstraction layer.\n\n---\n\n"

	out += "## Proposed Resolution\n\n" + resolution + "\n"

	return out
}

// ─── Main ─────────────────────────────────────────────────────────────────────

func main() {
	aPath := flag.String("a", "", "Path to first atom file")
	bPath := flag.String("b", "", "Path to second atom file")
	outPath := flag.String("out", "", "Output path for the markdown report (default: stdout)")
	flag.Parse()
	config.Load()
	config.Log("atd-compare", "Started process")


	if *aPath == "" || *bPath == "" {
		fmt.Println("Usage: atd-compare -a <atom_a.atom.md> -b <atom_b.atom.md> [-out <report.md>]")
		os.Exit(1)
	}

	atomA, err := parseAtom(*aPath)
	if err != nil {
		fmt.Printf("Error reading atom A: %v\n", err)
		os.Exit(1)
	}

	atomB, err := parseAtom(*bPath)
	if err != nil {
		fmt.Printf("Error reading atom B: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Asking llama3.2 for resolution proposal...\n")
	resolution := proposeResolution(atomA, atomB)

	report := buildReport(atomA, atomB, *aPath, *bPath, resolution)

	if *outPath == "" {
		fmt.Print(report)
	} else {
		if err := os.WriteFile(*outPath, []byte(report), 0644); err != nil {
			fmt.Printf("Error writing report: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Report written to: %s\n", *outPath)
	}
}
