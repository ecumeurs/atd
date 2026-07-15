package coverage

import (
	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CheckAtomRow struct {
	AtomID    string `json:"atom_id"`
	ImplLinks int    `json:"impl_links"`
	TestLinks int    `json:"test_links"`
	Semantic  string `json:"semantic"`
	Resolution string `json:"resolution,omitempty"`
	Status    string `json:"status"`
}

type CheckSummary struct {
	Total        int `json:"total"`
	WithImpl     int `json:"with_impl"`
	WithTests    int `json:"with_tests"`
	SemanticPass int `json:"semantic_pass"`
	SemanticFail int `json:"semantic_fail"`
}

type CheckReport struct {
	Mode    string         `json:"mode"`
	Rows    []CheckAtomRow `json:"rows"`
	Summary CheckSummary   `json:"summary"`
}

type CoverageReport struct {
	Text string
}

func GenerateReport(mode, atomID, filePath, docsDir string, full, semantic bool, gitArgs []string) (*CoverageReport, error) {
	if docsDir == "" {
		docsDir = config.DocsDir()
	}

	explorer := exploration.NewExplorer(config.ProjectRoot(), docsDir)
	if err := explorer.Load(false); err != nil {
		return nil, fmt.Errorf("failed to load ATD graph: %v", err)
	}

	var atomIDs []string

	switch {
	case full:
		mode = "full"
		for id := range explorer.Graph.Atoms {
			atomIDs = append(atomIDs, id)
		}

	case atomID != "":
		mode = "atom"
		canonicalID, resolveErr := explorer.CanonicalAtomID(atomID)
		if resolveErr != nil {
			if suggestion := explorer.SuggestAtomID(atomID); suggestion != "" {
				return nil, fmt.Errorf("atom '%s' not found in workspace (did you mean '%s'?)", atomID, suggestion)
			}
			return nil, fmt.Errorf("atom '%s' not found in workspace", atomID)
		}
		atomIDs = []string{canonicalID}

	case filePath != "":
		mode = "file"
		seen := make(map[string]bool)
		for _, sl := range explorer.SpecLinks {
			if sl.FilePath == filePath && !seen[sl.AtomID] {
				atomIDs = append(atomIDs, sl.AtomID)
				seen[sl.AtomID] = true
			}
		}

	default:
		mode = "diff"
		codeFiles, atomFiles, err := diffChangedFiles(gitArgs)
		if err != nil {
			return nil, err
		}
		seen := make(map[string]bool)

		codeSet := make(map[string]bool)
		for _, f := range codeFiles {
			codeSet[f] = true
		}
		for _, sl := range explorer.SpecLinks {
			if codeSet[sl.FilePath] && !seen[sl.AtomID] {
				atomIDs = append(atomIDs, sl.AtomID)
				seen[sl.AtomID] = true
			}
		}

		for _, af := range atomFiles {
			id := strings.TrimSuffix(filepath.Base(af), ".atom.md")
			if !seen[id] {
				atomIDs = append(atomIDs, id)
				seen[id] = true
			}
		}
	}

	if len(atomIDs) == 0 {
		return &CoverageReport{Text: "No atoms found for the specified scope."}, nil
	}

	report, err := buildCoverageReport(mode, atomIDs, explorer, semantic)
	if err != nil {
		return nil, err
	}

	return &CoverageReport{Text: formatCheckReport(report)}, nil
}

func buildCoverageReport(mode string, atomIDs []string, explorer *exploration.Explorer, semantic bool) (*CheckReport, error) {
	report := &CheckReport{Mode: mode}

	implByAtom := make(map[string][]exploration.SpecLink)
	for _, sl := range explorer.SpecLinks {
		implByAtom[sl.AtomID] = append(implByAtom[sl.AtomID], sl)
	}
	testByAtom := make(map[string][]exploration.TestLink)
	for _, tl := range explorer.TestLinks {
		testByAtom[tl.AtomID] = append(testByAtom[tl.AtomID], tl)
	}

	for _, id := range atomIDs {
		implLinks := implByAtom[id]
		testLinks := testByAtom[id]

		row := CheckAtomRow{
			AtomID:    id,
			ImplLinks: len(implLinks),
			TestLinks: len(testLinks),
			Semantic:  "-",
		}

		switch {
		case row.ImplLinks == 0:
			row.Status = "NO_IMPL"
		case row.TestLinks == 0:
			row.Status = "NO_TESTS"
		default:
			row.Status = "OK"
		}

		if semantic && len(implLinks) > 0 {
			atomData, err := explorer.ResolveAtom(id)
			if err == nil {
				persona := prompt.PersonaTechLead
				if atomData.Layer == "BUSINESS" {
					persona = prompt.PersonaPM
				}

				curated := prompt.CuratedAuditAtom{
					ID:          atomData.ID,
					Type:        atomData.Type,
					Layer:       atomData.Layer,
					Intent:      atomData.Intent,
					Logic:       atomData.Logic,
					Expectation: atomData.Expectation,
				}

				sl := implLinks[0]
				snippet, _ := getSnippet(sl.FilePath, sl.Line, 30)
				auditPrompt := prompt.AuditCodeBuild(persona, curated, snippet)
				resp, queryErr := ollama.Query("code_analysis", auditPrompt, prompt.AuditCodeFormat())

				if queryErr == nil && resp != nil {
					fmt.Fprintf(os.Stderr, "[DEBUG] LLM Response: %s\n", resp.Response)
				}

				if queryErr == ollama.ErrIDEFallback {
					row.Semantic = "PENDING"
					pipeline.WritePromptFile("check_semantic_"+id, auditPrompt)
					pipeline.WriteTaskList("check --semantic --atom "+id, []pipeline.PendingTask{
						{
							PromptFile:   "check_semantic_" + id + ".prompt",
							ResultFile:   "check_semantic_" + id + ".result",
							Instruction:  "validate if code implements atom",
							OutputSchema: `{"passed": bool, "resolutionMessage": "string"}`,
						},
					})
				} else if queryErr == nil {
					var rawResult map[string]interface{}
					if json.Unmarshal([]byte(resp.Response), &rawResult) == nil {
						if msg, ok := rawResult["resolutionMessage"].(string); ok {
							row.Resolution = msg
						} else if msg, ok := rawResult["resolution_message"].(string); ok {
							row.Resolution = msg
						} else if msg, ok := rawResult["message"].(string); ok {
							row.Resolution = msg
						}

						passed := false
						if p, ok := rawResult["passed"].(bool); ok {
							passed = p
						}

						if passed {
							row.Semantic = "PASS"
						} else {
							row.Semantic = "FAIL"
							row.Status = "SEMANTIC_FAIL"
						}
					}
				}
			}
		}

		report.Rows = append(report.Rows, row)
		report.Summary.Total++
		if row.ImplLinks > 0 {
			report.Summary.WithImpl++
		}
		if row.TestLinks > 0 {
			report.Summary.WithTests++
		}
		if row.Semantic == "PASS" {
			report.Summary.SemanticPass++
		}
		if row.Semantic == "FAIL" {
			report.Summary.SemanticFail++
		}
	}

	return report, nil
}

func diffChangedFiles(gitArgs []string) (codeFiles, atomFiles []string, err error) {
	root := config.ProjectRoot()

	args := append([]string{"-C", root, "diff", "--name-only"}, gitArgs...)

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if runErr := cmd.Run(); runErr != nil {
		return nil, nil, fmt.Errorf("git diff failed: %v", runErr)
	}

	toplevel, tlErr := gitToplevel(root)

	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}

		relLine := line
		if tlErr == nil && toplevel != "" && toplevel != root {
			abs := filepath.Join(toplevel, line)
			rel, relErr := filepath.Rel(root, abs)
			if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				continue
			}
			relLine = rel
		}

		if strings.HasSuffix(relLine, ".atom.md") {
			atomFiles = append(atomFiles, relLine)
		} else {
			codeFiles = append(codeFiles, relLine)
		}
	}
	return codeFiles, atomFiles, nil
}

func gitToplevel(root string) (string, error) {
	cmd := exec.Command("git", "-C", root, "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel failed: %v", err)
	}
	return strings.TrimSpace(out.String()), nil
}

func getSnippet(filePath string, line, context int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	start := line - context
	if start < 0 {
		start = 0
	}
	end := line + context
	if end > len(lines) {
		end = len(lines)
	}

	return strings.Join(lines[start:end], "\n"), nil
}

func formatCheckReport(r *CheckReport) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# ATD Coverage Report (mode: %s)\n\n", r.Mode))

	sb.WriteString(fmt.Sprintf("%-50s %6s %6s %9s %14s\n", "Atom ID", "Impl", "Tests", "Semantic", "Status"))
	sep := strings.Repeat("-", 90) + "\n"
	sb.WriteString(sep)

	for _, row := range r.Rows {
		sb.WriteString(fmt.Sprintf("%-50s %6d %6d %9s %14s\n",
			row.AtomID, row.ImplLinks, row.TestLinks, row.Semantic, row.Status))
	}

	sb.WriteString(sep)
	sb.WriteString(fmt.Sprintf("\nSummary: %d atoms | %d with impl | %d with tests",
		r.Summary.Total, r.Summary.WithImpl, r.Summary.WithTests))
	if r.Summary.SemanticPass+r.Summary.SemanticFail > 0 {
		sb.WriteString(fmt.Sprintf(" | semantic: %d pass / %d fail",
			r.Summary.SemanticPass, r.Summary.SemanticFail))
	}
	sb.WriteString("\n")

	hasFailures := false
	for _, row := range r.Rows {
		if row.Semantic == "FAIL" && row.Resolution != "" {
			if !hasFailures {
				sb.WriteString("\n--- Semantic Failure Details ---\n")
				hasFailures = true
			}
			sb.WriteString(fmt.Sprintf("[%s]: %s\n", row.AtomID, row.Resolution))
		}
	}

	return sb.String()
}