package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTraceRecursiveTraversal(t *testing.T) {
	// 1. Setup temp environment
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	srcDir := filepath.Join(tmpDir, "src")
	err := os.MkdirAll(docsDir, 0755)
	if err != nil { t.Fatal(err) }
	err = os.MkdirAll(srcDir, 0755)
	if err != nil { t.Fatal(err) }

	// 2. Create atoms: C1 -> A1 -> I1
	// C1 (Customer)
	c1Content := `---
id: c1
human_name: "Customer 1"
layer: CUSTOMER
dependents:
  - [[a1]]
---
# C1`
	os.WriteFile(filepath.Join(docsDir, "c1.atom.md"), []byte(c1Content), 0644)

	// A1 (Architecture)
	a1Content := `---
id: a1
human_name: "Arch 1"
layer: ARCHITECTURE
parents:
  - [[c1]]
dependents:
  - [[i1]]
---
# A1`
	os.WriteFile(filepath.Join(docsDir, "a1.atom.md"), []byte(a1Content), 0644)

	// I1 (Implementation)
	i1Content := `---
id: i1
human_name: "Impl 1"
layer: IMPLEMENTATION
parents:
  - [[a1]]
---
# I1
## TECHNICAL INTERFACE
- @spec-link [[i1]]`
	os.WriteFile(filepath.Join(docsDir, "i1.atom.md"), []byte(i1Content), 0644)

	// 3. Create mock source file for I1
	srcContent := "// @spec-link [[i1]]"
	os.WriteFile(filepath.Join(srcDir, "logic.go"), []byte(srcContent), 0644)

	// 4. Run Trace on A1
	resultJSON, err := runTrace("a1", docsDir, srcDir)
	if err != nil {
		t.Fatalf("runTrace failed: %v", err)
	}

	var snap TraceSnapshot
	err = json.Unmarshal([]byte(resultJSON), &snap)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	// 5. Verify Ancestry (UP)
	foundC1 := false
	for _, p := range snap.GraphSlice.Parents {
		if p == "c1" {
			foundC1 = true
			break
		}
	}
	if !foundC1 {
		t.Errorf("Expected c1 in parents, got %v", snap.GraphSlice.Parents)
	}

	// 6. Verify Dependents (DOWN)
	foundI1 := false
	for _, d := range snap.GraphSlice.Dependents {
		if d == "i1" {
			foundI1 = true
			break
		}
	}
	if !foundI1 {
		t.Errorf("Expected i1 in dependents, got %v", snap.GraphSlice.Dependents)
	}

	// 7. Verify Health Metrics
	if !snap.HealthSummary.HasCustomerOrigin {
		t.Error("Expected atom a1 to have customer origin (c1)")
	}
	if snap.Metrics.ImplementedDependents != 1 {
		t.Errorf("Expected 1 implemented dependent (i1), got %d", snap.Metrics.ImplementedDependents)
	}
}
