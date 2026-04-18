package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// captureOutput runs a cobra command and captures its stdout output.
func captureOutput(cmd *cobra.Command, args []string) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	// Reset flags
	cmd.ResetFlags()
	return buf.String(), nil
}

func TestDissectMissingFile(t *testing.T) {
	// Reset the command for testing
	dissectCmd.ResetFlags()
	dissectCmd.Flags().StringP("file", "f", "", "Target file to dissect")
	dissectCmd.Flags().Bool("llm", false, "Route through tiered LLM provider")

	err := dissectCmd.RunE(dissectCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "--file parameter is required") {
		t.Errorf("Expected missing file error, got: %v", err)
	}
}

func TestDissectNumberedLines(t *testing.T) {
	// Create a temp file with known content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	content := "line one\nline two\nline three"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Capture stdout by redirecting
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	dissectCmd.ResetFlags()
	dissectCmd.Flags().StringP("file", "f", "", "Target file")
	dissectCmd.Flags().Bool("llm", false, "LLM flag")
	dissectCmd.Flags().Set("file", tmpFile)

	err := dissectCmd.RunE(dissectCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "001:") {
		t.Errorf("Expected '001:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "002:") {
		t.Errorf("Expected '002:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "003:") {
		t.Errorf("Expected '003:' in output, got:\n%s", output)
	}
}

func TestDissectPromptContainsTags(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(tmpFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	dissectCmd.ResetFlags()
	dissectCmd.Flags().StringP("file", "f", "", "Target file")
	dissectCmd.Flags().Bool("llm", false, "LLM flag")
	dissectCmd.Flags().Set("file", tmpFile)

	err := dissectCmd.RunE(dissectCmd, []string{})
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "<System_Context>") {
		t.Errorf("Expected '<System_Context>' in prompt output:\n%s", output)
	}
	if !strings.Contains(output, "<Document>") {
		t.Errorf("Expected '<Document>' in prompt output:\n%s", output)
	}
}

func TestGenerateMissingDissect(t *testing.T) {
	generateCmd.ResetFlags()
	generateCmd.Flags().StringP("dissect", "d", "", "Path to dissect output")

	err := generateCmd.RunE(generateCmd, []string{})
	if err == nil || !strings.Contains(err.Error(), "--dissect parameter is required") {
		t.Errorf("Expected missing dissect error, got: %v", err)
	}
}

func TestGenerateNonexistentFile(t *testing.T) {
	generateCmd.ResetFlags()
	generateCmd.Flags().StringP("dissect", "d", "", "Path to dissect output")
	generateCmd.Flags().Set("dissect", "/nonexistent/path/file.txt")

	err := generateCmd.RunE(generateCmd, []string{})
	if err == nil {
		t.Error("Expected error for nonexistent dissect file")
	}
}
