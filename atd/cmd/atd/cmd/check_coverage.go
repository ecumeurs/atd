package cmd

// @spec-link [[service_atd_check_coverage]]

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/ollama"
	"atd-tools/pkg/pipeline"
	"atd-tools/pkg/prompt"
	"github.com/spf13/cobra"
)

// CheckAtomRow is one row in the per-atom coverage report.
type CheckAtomRow struct {
	AtomID    string `json:"atom_id"`
	ImplLinks int    `json:"impl_links"`
	TestLinks int    `json:"test_links"`
	Semantic  string `json:"semantic"` // "PASS" | "FAIL" | "PENDING" | "-"
	Resolution string `json:"resolution,omitempty"`
	Status    string `json:"status"` // "OK" | "NO_IMPL" | "NO_TESTS" | "SEMANTIC_FAIL"
}

// CheckSummary aggregates coverage counts.
type CheckSummary struct {
	Total        int `json:"total"`
	WithImpl     int `json:"with_impl"`
	WithTests    int `json:"with_tests"`
	SemanticPass int `json:"semantic_pass"`
	SemanticFail int `json:"semantic_fail"`
}

// CheckReport is the full output of atd check.
type CheckReport struct {
	Mode    string         `json:"mode"`
	Rows    []CheckAtomRow `json:"rows"`
	Summary CheckSummary   `json:"summary"`
}

// runCoverageCheck is the entry point for atd check.
func runCoverageCheck(mode, atomID, filePath, docsDir string, full, semantic bool, gitArgs []string) (string, error) {
	if docsDir == "" {
		docsDir = config.DocsDir()
	}

	explorer := exploration.NewExplorer(config.ProjectRoot(), docsDir)
	if err := explorer.Load(false); err != nil {
		return "", fmt.Errorf("failed to load ATD graph: %v", err)
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
		atomIDs = []string{atomID}

	case filePath != "":
		mode = "file"
		for _, sl := range explorer.SpecLinks {
			if sl.FilePath == filePath {
				atomIDs = append(atomIDs, sl.AtomID)
			}
		}

	default:
		mode = "diff"
		codeFiles, atomFiles, err := diffChangedFiles(gitArgs)
		if err != nil {
			return "", err
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

		// Atom files changed: also include their IDs for forward coverage check
		for _, af := range atomFiles {
			id := strings.TrimSuffix(filepath.Base(af), ".atom.md")
			if !seen[id] {
				atomIDs = append(atomIDs, id)
				seen[id] = true
			}
		}
	}

	if len(atomIDs) == 0 {
		return "No atoms found for the specified scope.", nil
	}

	report, err := buildCoverageReport(mode, atomIDs, explorer, semantic)
	if err != nil {
		return "", err
	}

	return formatCheckReport(report), nil
}

// buildCoverageReport computes impl/test/semantic coverage for the given atom IDs.
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
				// Determine Persona
				persona := prompt.PersonaTechLead
				if atomData.Layer == "BUSINESS" {
					persona = prompt.PersonaPM
				}

				// Curate context
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

// diffChangedFiles runs git diff and splits results into code files and atom files.
func diffChangedFiles(gitArgs []string) (codeFiles, atomFiles []string, err error) {
	args := []string{"diff", "--name-only"}
	args = append(args, gitArgs...)

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if runErr := cmd.Run(); runErr != nil {
		return nil, nil, fmt.Errorf("git diff failed: %v", runErr)
	}

	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, ".atom.md") {
			atomFiles = append(atomFiles, line)
		} else {
			codeFiles = append(codeFiles, line)
		}
	}
	return codeFiles, atomFiles, nil
}

// formatCheckReport renders a CheckReport as a text table.
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

	// Print details for semantic failures
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

var coverageCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check ATD coverage: impl links, test links, and optional semantic compliance",
	Long: `Check ATD atom coverage for implementation and test links.

Modes:
  diff (default): git-diff driven — checks code changes and atom changes bidirectionally
  --atom <id>:    full coverage for one atom
  --file <path>:  coverage for one source file
  --full:         entire project coverage

Pass --semantic to also run LLM compliance checks per @spec-link (slow).`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		atomID, _ := cmd.Flags().GetString("atom")
		filePath, _ := cmd.Flags().GetString("file")
		docsDir, _ := cmd.Flags().GetString("docs")
		full, _ := cmd.Flags().GetBool("full")
		semantic, _ := cmd.Flags().GetBool("semantic")
		outPath, _ := cmd.Flags().GetString("out")

		mode := "diff"

		text, err := runCoverageCheck(mode, atomID, filePath, docsDir, full, semantic, args)
		if err != nil {
			return err
		}

		if outPath != "" {
			if err := os.WriteFile(outPath, []byte(text), 0644); err != nil {
				return fmt.Errorf("failed to write report to %s: %v", outPath, err)
			}
			fmt.Printf("Coverage report written to %s\n", outPath)
		} else {
			fmt.Print(text)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(coverageCheckCmd)
	coverageCheckCmd.Flags().String("atom", "", "Check full coverage for one atom ID")
	coverageCheckCmd.Flags().String("file", "", "Check coverage for a specific source file")
	coverageCheckCmd.Flags().Bool("full", false, "Check entire project coverage")
	coverageCheckCmd.Flags().Bool("semantic", false, "Run LLM compliance check per @spec-link (slow)")
	coverageCheckCmd.Flags().String("out", "", "Write report to file")
	coverageCheckCmd.Flags().String("docs", "", "Override docs directory")
}
