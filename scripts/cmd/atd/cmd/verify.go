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

	"atd-tools/config"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Audit modified code against ATD specifications",
	Long: `Identify modified files via git, extract linked atoms, run native tests,
and generate a comprehensive audit prompt for an LLM to verify compliance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		text, err := runVerify(docsDir)
		if err != nil {
			return err
		}
		fmt.Print(text)
		return nil
	},
}

func runVerify(docsDir string) (string, error) {
	// 1. Get modified files from git
	gitCmd := exec.Command("git", "diff", "--name-only")
	var out bytes.Buffer
	gitCmd.Stdout = &out
	if err := gitCmd.Run(); err != nil {
		return "", fmt.Errorf("error running git diff. Ensure you are in a git repository: %v", err)
	}

	modifiedFiles := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(modifiedFiles) == 1 && modifiedFiles[0] == "" {
		return "No tracked modified files found.", nil
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

		content, err := os.ReadFile(file)
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

	if len(atomIDs) == 0 {
		return "No @spec-link tags found in modified files. Nothing to audit.", nil
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

	// 4. Run native tests for the changed directories
	testResults := ""
	testFilesContent := ""
	for dir := range changedDirs {
		cmdTest := exec.Command("go", "test", "-v", "./"+dir)
		var testOut bytes.Buffer
		cmdTest.Stdout = &testOut
		cmdTest.Stderr = &testOut
		err := cmdTest.Run()

		status := "PASSED"
		if err != nil {
			status = "FAILED"
		}
		testResults += fmt.Sprintf("=== Test Execution for %s ===\nStatus: %s\nOutput:\n%s\n\n", dir, status, testOut.String())

		files, _ := os.ReadDir(dir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), "_test.go") {
				testPath := filepath.Join(dir, f.Name())
				tc, _ := os.ReadFile(testPath)
				testFilesContent += fmt.Sprintf("--- Test File: %s ---\n%s\n\n", testPath, string(tc))
			}
		}
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

func init() {
	rootCmd.AddCommand(verifyCmd)
	verifyCmd.Flags().String("docs", "", "Override docs directory")
}
