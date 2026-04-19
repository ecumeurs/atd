package cmd
// @spec-link [[service_atd_verify]]

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/exploration"
	"github.com/spf13/cobra"
)



var verifyCmd = &cobra.Command{
	Use:   "verify [base] [target]",
	Short: "Audit modified code against ATD specifications",
	Long: `Identify modified files via git, extract linked atoms, run native tests,
and generate a comprehensive audit prompt for an LLM to verify compliance.

Usecases:
  1. Local Audit:  atd verify (audits uncommitted changes)
  2. Host Audit:   atd verify HEAD~5 (audits current changes since 5 commits ago)
  3. CI Audit:     atd verify HEAD~1 HEAD (audits evolution between two points)`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		outPath, _ := cmd.Flags().GetString("out")
		full, _ := cmd.Flags().GetBool("full")
		targetFile, _ := cmd.Flags().GetString("file")
		targetLine, _ := cmd.Flags().GetInt("line")

		config.ActiveConfig.DocsDirOverride = docsDir
		
		if full && outPath == "" {
			outPath = "recap.md"
		}

		text, err := runVerify(docsDir, args, full, targetFile, targetLine)
		if err != nil {
			return err
		}

		if outPath != "" {
			err := os.WriteFile(outPath, []byte(text), 0644)
			if err != nil {
				return fmt.Errorf("failed to write audit report to %s: %v", outPath, err)
			}
			fmt.Printf("Audit report written to %s\n", outPath)
		} else {
			fmt.Print(text)
		}
		return nil
	},
}

func runVerify(docsDir string, args []string, full bool, targetFile string, targetLine int) (string, error) {
	// 1. Initialize Explorer and Graph
	explorer := exploration.NewExplorer(config.ProjectRoot(), config.DocsDir())
	if err := explorer.Load(false); err != nil {
		return "", fmt.Errorf("failed to load ATD graph: %v", err)
	}

	var relevantLinks []exploration.SpecLink

	if targetFile != "" {
		// --- PATH A: Targeted Mode ---
		for _, l := range explorer.SpecLinks {
			if l.FilePath == targetFile {
				if targetLine == 0 || l.Line == targetLine {
					relevantLinks = append(relevantLinks, l)
				}
			}
		}
		if len(relevantLinks) == 0 {
			return fmt.Sprintf("No @spec-link tags found in %s (at line %d)", targetFile, targetLine), nil
		}
	} else if full {
		// --- PATH B: Full Project Mode ---
		relevantLinks = explorer.SpecLinks
		if len(relevantLinks) == 0 {
			return "No @spec-link tags found in the entire project.", nil
		}
	} else {
		// --- PATH C: Diff Mode (Default) ---
		gitArgs := []string{"diff", "--name-only"}
		targetRef := ""

		if len(args) == 1 {
			gitArgs = append(gitArgs, args[0])
		} else if len(args) == 2 {
			gitArgs = append(gitArgs, args[0], args[1])
			targetRef = args[1]
		}

		gitCmd := exec.Command("git", gitArgs...)
		var out bytes.Buffer
		gitCmd.Stdout = &out
		if err := gitCmd.Run(); err != nil {
			return "", fmt.Errorf("error running git diff. Ensure you are in a git repository: %v", err)
		}

		modifiedFiles := strings.Split(strings.TrimSpace(out.String()), "\n")
		modifiedMap := make(map[string]bool)
		for _, f := range modifiedFiles {
			if f != "" {
				modifiedMap[f] = true
			}
		}

		if len(modifiedMap) == 0 {
			return "No differences found for the specified range.", nil
		}

		for _, l := range explorer.SpecLinks {
			if modifiedMap[l.FilePath] {
				relevantLinks = append(relevantLinks, l)
			}
		}

		if len(relevantLinks) == 0 {
			return "No @spec-link tags found in the modified files.", nil
		}
		_ = targetRef // Keep it for now if needed for content loading
	}

	fmt.Fprintf(os.Stderr, "Phase 1: Discovery (Found %d implementation link(s))\n", len(relevantLinks))

	// 2. Statistics and Recap state
	type auditResult struct {
		Link            exploration.SpecLink
		Atom            *atom.AtomData
		Ancestry        []string
		Snippet         string
		Tests           []string
		Status          string // PASSED, FAILED, WARNING
		Finding         string
		TestOutput      string
		TestFailed      bool
		RuleCompliance  bool
		TestCompliance  bool
	}

	results := make([]auditResult, 0)
	stats := struct {
		Total   int
		Valid   int
		Invalid int
		Gaps    int
		Cached  int
	}{}

	type testResult struct {
		Output string
		Failed bool
	}
	testCache := make(map[string]testResult)

	// 3. Iterative Instance Audits
	for _, link := range relevantLinks {
		stats.Total++
		testFailed := false

		atom, exists := explorer.Graph.Atoms[link.AtomID]
		if !exists {
			continue
		}

		// Ancestry
		trace, _ := explorer.Trace(link.AtomID)
		ancestry := trace.GraphSlice.Parents

		// Surgical Snippet
		snippet, _ := getSnippet(link.FilePath, link.Line, 30)

		// Verification Proof (Surgical Test Links)
		var relevantTests []string
		for _, tl := range explorer.TestLinks {
			if tl.AtomID == link.AtomID {
				testContent, _ := os.ReadFile(tl.TestFile)
				relevantTests = append(relevantTests, fmt.Sprintf("--- Test File: %s ---\n%s\n", tl.TestFile, string(testContent)))
			}
		}

		if len(relevantTests) == 0 {
			stats.Gaps++
		}

		// Native Test Run (Whole directory still, for environmental stability)
		dir := filepath.Dir(link.FilePath)
		ext := filepath.Ext(link.FilePath)
		cmdTpl, _ := config.GetVerifyDefaults(ext)
		
		fmt.Fprintf(os.Stderr, "[%d/%d] Auditing %s at %s:%d... ", stats.Total, len(relevantLinks), link.AtomID, link.FilePath, link.Line)

		testKey := cmdTpl + ":" + dir
		var testOut string
		if cached, ok := testCache[testKey]; ok {
			testOut = cached.Output
			testFailed = cached.Failed
			stats.Cached++
			fmt.Fprintf(os.Stderr, "[CACHED TEST]\n")
		} else {
			var err error
			testOut, err = executeTestCommand(cmdTpl, dir, link.FilePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[TEST FAILED]\n")
				testFailed = true
			} else {
				fmt.Fprintf(os.Stderr, "[OK]\n")
			}
			testCache[testKey] = testResult{Output: testOut, Failed: testFailed}
		}

		results = append(results, auditResult{
			Link:     link,
			Atom:     atom,
			Ancestry: ancestry,
			Snippet:  snippet,
			Tests:    relevantTests,
			TestOutput: testOut,
			TestFailed: testFailed,
		})
	}

	// 4. Assemble the Final Recap & Multi-Prompt
	var b strings.Builder
	b.WriteString("# ATD Verification Recap\n\n")

	b.WriteString("## Statistics\n")
	b.WriteString(fmt.Sprintf("- **Internal Status:** INITIATED (CLI: %t)\n", full))
	b.WriteString(fmt.Sprintf("- **Instances Discovered:** %d\n", stats.Total))
	b.WriteString(fmt.Sprintf("- **Implementation Gaps (No Test Proof):** %d\n", stats.Gaps))
	b.WriteString(fmt.Sprintf("- **Test Optimization:** %d runs cached\n", stats.Cached))
	b.WriteString("- **Status:** PENDING LLM AUDIT\n\n")

	b.WriteString("## Discovery Table\n")
	b.WriteString("| Atom ID | File Path | Line | Coverage | Audit Status | Recommendation |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, res := range results {
		coverage := "OK"
		rec := "-"
		if len(res.Tests) == 0 {
			coverage = "⚠️ GAP"
			rec = "Add @test-link"
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %d | %s | PENDING | %s |\n", res.Atom.ID, res.Link.FilePath, res.Link.Line, coverage, rec))
	}
	b.WriteString("\n")

	b.WriteString("## Actionable Prompt\n")
	b.WriteString("Copy the content below into your LLM to perform the granular compliance audit.\n\n")
	b.WriteString("--- BUNDLE START ---\n")
	b.WriteString("<System Objective>\n")
	b.WriteString("You are the ATD Lead Auditor. You must perform a granular, instance-based compliance check for each provided @spec-link.\n")
	b.WriteString("For each instance, analyze the snippet against the rule and its ancestry (requirements context). Check if the provided tests adequately verify the logic.\n\n")
	b.WriteString("### IMPORTANT: OUTPUT FORMAT\n")
	b.WriteString("You MUST respond with a JSON array of objects, one per instance, following this schema exactly:\n")
	b.WriteString("```json\n")
	b.WriteString("[\n  {\n    \"tag\": \"file:line\",\n    \"atom_id\": \"string\",\n    \"rule_compliant\": boolean,\n    \"test_compliant\": boolean,\n    \"findings\": \"string\",\n    \"fix_suggestion\": \"string\"\n  }\n]\n")
	b.WriteString("```\n")
	b.WriteString("</System Objective>\n\n")

	for i, res := range results {
		b.WriteString(fmt.Sprintf("## INSTANCE %d: %s:%d\n", i+1, res.Link.FilePath, res.Link.Line))
		b.WriteString(fmt.Sprintf("### ATOM: %s\n", res.Link.AtomID))
		b.WriteString(fmt.Sprintf("**Ancestry Context:** %s\n\n", strings.Join(res.Ancestry, " -> ")))
		
		b.WriteString("#### THE RULE\n")
		b.WriteString(fmt.Sprintf("Intent: %s\nLogic:\n%s\n\n", res.Atom.Intent, res.Atom.Logic))

		b.WriteString("#### THE CODE (SURGICAL SNIPPET)\n")
		b.WriteString("```\n")
		b.WriteString(res.Snippet)
		b.WriteString("\n```\n\n")

		b.WriteString("#### THE VERIFICATION PROOF (TESTS)\n")
		if len(res.Tests) > 0 {
			for _, t := range res.Tests {
				b.WriteString(t)
			}
		} else {
			b.WriteString("[WARNING] No @test-link found for this atom in the project.\n")
		}

		b.WriteString("\n#### NATIVE TEST EXECUTION\n")
		b.WriteString("```\n")
		if res.TestFailed {
			b.WriteString(res.TestOutput)
		} else {
			b.WriteString("[SUCCESS] Native tests passed. Output omitted to save space.\n")
		}
		b.WriteString("\n```\n\n")
		b.WriteString("---\n\n")
	}

	return b.String(), nil
}

func getSnippet(path string, line int, context int) (string, error) {
	content, err := os.ReadFile(path)
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

func executeTestCommand(tpl, dir, file string) (string, error) {
	absDir, _ := filepath.Abs(dir)
	absFile, _ := filepath.Abs(file)
	modRoot := findGoModRoot(absDir)

	relDir := dir
	relFile := file
	if modRoot != "" {
		relDir, _ = filepath.Rel(modRoot, absDir)
		relFile, _ = filepath.Rel(modRoot, absFile)
	}

	t, err := template.New("cmd").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse command template: %v", err)
	}

	var buf bytes.Buffer
	data := struct {
		Dir  string
		File string
	}{
		Dir:  relDir,
		File: relFile,
	}

	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute command template: %v", err)
	}

	cmdStr := buf.String()
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command generated")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	if modRoot != "" {
		cmd.Dir = modRoot
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	return out.String(), err
}

func findGoModRoot(dir string) string {
	current := dir
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().String("docs", "", "Override docs directory")
	verifyCmd.Flags().String("out", "", "Write the audit report to a file")
	verifyCmd.Flags().Bool("full", false, "Perform audit on the entire project instead of diff-based")
	verifyCmd.Flags().String("file", "", "Target a specific file for verification")
	verifyCmd.Flags().Int("line", 0, "Target a specific line for verification (requires --file)")
}

