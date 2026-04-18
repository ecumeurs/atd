package cmd
// @spec-link [[service_atd_verify]]

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"atd-tools/config"
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

		text, err := runVerify(docsDir, args)
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

func runVerify(docsDir string, args []string) (string, error) {
	// 1. Get modified files from git
	gitArgs := []string{"diff", "--name-only"}
	targetRef := "" // Empty means we read from disk (working state)

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
	if len(modifiedFiles) == 1 && modifiedFiles[0] == "" {
		return "No differences found for the specified range.", nil
	}

	// 2. Extract Atom Links and identify directories with changes
	linkRegex := regexp.MustCompile(`@spec-link\s+\[\[(.*?)\]\]`)
	atomIDs := make(map[string]bool)
	changedDirs := make(map[string]bool)
	fileContents := make(map[string]string)

	for _, file := range modifiedFiles {
		if file == "" {
			continue
		}

		if strings.HasSuffix(file, ".atom.md") || strings.HasSuffix(file, ".md") {
			// Atoms don't have tests to execute on them in the verify context
			continue
		}

		var content []byte
		var err error
		if targetRef == "" {
			content, err = os.ReadFile(file)
		} else {
			showCmd := exec.Command("git", "show", targetRef+":"+file)
			content, err = showCmd.Output()
		}

		if err != nil {
			continue
		}

		fileContents[file] = string(content)
		changedDirs[filepath.Dir(file)] = true

		matches := linkRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) > 1 {
				atomIDs[match[1]] = true
			}
		}
	}

	if len(fileContents) == 0 {
		return "No source code modifications detected in the given range.", nil
	}

	if len(atomIDs) == 0 {
		return "No @spec-link tags found in the modified source files. Nothing to audit.", nil
	}

	// 3. Read the relevant Atoms
	atomContents := make(map[string]string)
	for atomID := range atomIDs {
		atomPath := filepath.Join(docsDir, atomID+".atom.md")
		content, err := os.ReadFile(atomPath)
		if err == nil {
			atomContents[atomID] = string(content)
		} else {
			atomContents[atomID] = fmt.Sprintf("[ERROR] Could not read ATD file for %s", atomID)
		}
	}

	// 4. Resolve Verify Command and Test Pattern
	dominantExt := ""
	extCounts := make(map[string]int)
	for file := range fileContents {
		ext := filepath.Ext(file)
		if ext != "" {
			extCounts[ext]++
			if dominantExt == "" || extCounts[ext] > extCounts[dominantExt] {
				dominantExt = ext
			}
		}
	}

	cmdTpl, patternTpl := config.GetVerifyDefaults(dominantExt)
	if config.ActiveConfig.Verify.Command != "" {
		cmdTpl = config.ActiveConfig.Verify.Command
	}
	if config.ActiveConfig.Verify.TestPattern != "" {
		patternTpl = config.ActiveConfig.Verify.TestPattern
	}

	if cmdTpl == "" {
		return "", fmt.Errorf("no verify command found for extension %s and no override in .atd config", dominantExt)
	}

	patterns := strings.Split(patternTpl, ",")
	for i := range patterns {
		patterns[i] = strings.TrimSpace(patterns[i])
	}

	// 5. Run native tests for the changed directories (Concurrency Limited)
	maxParallel := config.ActiveConfig.Verify.MaxParallelism
	if maxParallel <= 0 {
		maxParallel = 1
	}

	type result struct {
		dir    string
		output string
		status string
		files  string
	}

	numDirs := len(changedDirs)
	resChan := make(chan result, numDirs)
	dirChan := make(chan string, numDirs)
	var wg sync.WaitGroup

	for w := 0; w < maxParallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for dir := range dirChan {
				var firstFile string
				for f := range fileContents {
					if filepath.Dir(f) == dir {
						firstFile = f
						break
					}
				}

				res, err := executeTestCommand(cmdTpl, dir, firstFile)
				status := "PASSED"
				if err != nil {
					status = "FAILED"
				}

				// Discover test files
				testFilesContent := ""
				files, _ := os.ReadDir(dir)
				for _, f := range files {
					if f.IsDir() {
						continue
					}
					matched := false
					for _, p := range patterns {
						if ok, _ := filepath.Match(p, f.Name()); ok {
							matched = true
							break
						}
					}

					if matched {
						testPath := filepath.Join(dir, f.Name())
						var tc []byte
						if targetRef == "" {
							tc, _ = os.ReadFile(testPath)
						} else {
							showCmd := exec.Command("git", "show", targetRef+":"+testPath)
							tc, _ = showCmd.Output()
						}
						testFilesContent += fmt.Sprintf("--- Test File: %s ---\n%s\n\n", testPath, string(tc))
					}
				}

				resChan <- result{
					dir:    dir,
					output: res,
					status: status,
					files:  testFilesContent,
				}
			}
		}()
	}

	for dir := range changedDirs {
		dirChan <- dir
	}
	close(dirChan)
	wg.Wait()
	close(resChan)

	testResults := ""
	testFilesContent := ""
	for r := range resChan {
		testResults += fmt.Sprintf("=== Test Execution for %s ===\nStatus: %s\nOutput:\n%s\n\n", r.dir, r.status, r.output)
		testFilesContent += r.files
	}

	// 5. Generate LLM Prompt (Return string)
	var b strings.Builder
	b.WriteString("<System Objective>\n")
	b.WriteString("You are the ATD Lead Auditor. A developer is submitting a patch. You must evaluate the modified code and its test coverage against the strict rules defined in the Atomic Traceable Documentation (ATD).\n")
	b.WriteString("Output a markdown report including a CLEAR TABLE summarizing:\n")
	b.WriteString("| Atom ID | Rule Compliant? | Test Coverage Compliant? | Notes |\n")
	b.WriteString("Ensure you explicitly check if the test files cover all constraints mentioned in the ATD expectations.\n")
	b.WriteString("</System Objective>\n\n")

	b.WriteString("<ATD Specifications (The Rules)>\n")
	for id, content := range atomContents {
		b.WriteString(fmt.Sprintf("--- ATOM: %s ---\n%s\n\n", id, content))
	}
	b.WriteString("</ATD Specifications>\n\n")

	b.WriteString("<Modified Source Code (The Implementation)>\n")
	for file, content := range fileContents {
		b.WriteString(fmt.Sprintf("--- File: %s ---\n%s\n\n", file, content))
	}
	b.WriteString("</Modified Source Code>\n\n")

	b.WriteString("<Test Files (The Verification Specs)>\n")
	b.WriteString(testFilesContent)
	b.WriteString("</Test Files>\n\n")

	b.WriteString("<Native Test Execution Results (The Proof)>\n")
	b.WriteString(testResults)
	b.WriteString("</Native Test Execution Results>\n")

	return b.String(), nil
}

func executeTestCommand(tpl, dir, file string) (string, error) {
	t, err := template.New("cmd").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse command template: %v", err)
	}

	var buf bytes.Buffer
	data := struct {
		Dir  string
		File string
	}{
		Dir:  dir,
		File: file,
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
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	return out.String(), err
}

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().String("docs", "", "Override docs directory")
	verifyCmd.Flags().String("out", "", "Write the audit report to a file")
}

