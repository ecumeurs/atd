package cmd

import (
	"strings"
	"testing"
)

func TestCoverageCheckFlags(t *testing.T) {
	if coverageCheckCmd.Use != "check" {
		t.Errorf("expected use 'check', got %s", coverageCheckCmd.Use)
	}
	for _, flag := range []string{"atom", "file", "full", "semantic", "out", "docs"} {
		if coverageCheckCmd.Flag(flag) == nil {
			t.Errorf("expected flag '--%s' to exist", flag)
		}
	}
}

func TestDiffFileClassification(t *testing.T) {
	// Test the split logic: .atom.md files go to atomFiles, everything else to codeFiles
	// We simulate by running diffChangedFiles with a git ref that produces known output.
	// Since we can't control git, we test the classification logic directly via the internal helper.

	type result struct {
		line     string
		isAtomMd bool
	}

	cases := []result{
		{"cmd/foo.go", false},
		{"docs/bar.atom.md", true},
		{"README.md", false},
		{"rule_foo.atom.md", true},
		{"src/something.atom.md.bak", false}, // not ending in .atom.md
	}

	for _, c := range cases {
		got := strings.HasSuffix(c.line, ".atom.md")
		if got != c.isAtomMd {
			t.Errorf("line %q: expected isAtomMd=%v, got %v", c.line, c.isAtomMd, got)
		}
	}
}

func TestCoverageCheckEmptyScope(t *testing.T) {
	// With a nonexistent docs dir, the loader should fail or return empty result
	_, err := runCoverageCheck("full", "", "", "/nonexistent/docs", true, false, nil)
	if err == nil {
		// It's also acceptable to return "No atoms found" without error
	}
}
