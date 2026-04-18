package cmd

import (
	"os"
	"testing"
)

func TestParseBloatedFiles(t *testing.T) {
	content := `=== ATD AUDIT PROTOCOL INITIATED ===
Phase 1: The Bloat Metric (docs count: 2)
Auditing: foo.atom.md ... [BLOATED]
Auditing: bar.atom.md ... [PASS]
`
	tmp, _ := os.CreateTemp("", "report.txt")
	defer os.Remove(tmp.Name())
	tmp.WriteString(content)
	tmp.Close()

	bloated, err := parseBloatedFiles(tmp.Name())
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(bloated) != 1 || bloated[0] != "foo.atom.md" {
		t.Errorf("Expected [foo.atom.md], got %v", bloated)
	}
}
