package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
)

func TestLintValidAtom(t *testing.T) {
	tmpData := `---
id: test_atom_1
human_name: "Test Atom 1"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: []
parents: []
dependents: []
---

# Test Atom 1

## INTENT
This is the intent.

## THE RULE / LOGIC
This is the logic.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** ` + "`" + `@spec-link [[test_atom_1]]` + "`" + `

## EXPECTATION
This is the expectation.
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_atom_1.atom.md")
	if err := os.WriteFile(path, []byte(tmpData), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := runLint(tmpDir)
	if err != nil {
		t.Fatalf("Expected nil err, got %v: %s", err, out)
	}
	if out != "" {
		t.Fatalf("Expected empty output, got %s", out)
	}
}

func TestLintInvalidAtom(t *testing.T) {
	tmpData := `---
id: test_atom_invalid
type: RULE
status: DRAFT
parents:
  - [[missing_parent]]
---

# Test Atom Invalid

## INTENT
This is the intent.

## THE RULE / LOGIC
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test_atom_invalid.atom.md")
	if err := os.WriteFile(path, []byte(tmpData), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := runLint(tmpDir)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedErrors := []string{
		"Missing mandatory field: human_name",
		"Missing mandatory field: layer",
		"Missing mandatory field: priority",
		"Missing mandatory section: ## TECHNICAL INTERFACE",
		"Missing mandatory section: ## EXPECTATION",
		"Unresolved parent link: [[missing_parent]]",
	}

	for _, expected := range expectedErrors {
		if !strings.Contains(out, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, out)
		}
	}
}

func TestLintGovernanceAtomAsParent(t *testing.T) {
	contractData := `---
id: contract_test
human_name: "Test Contract"
type: CONTRACT
layer: BUSINESS
version: 1.0
status: STABLE
priority: CORE
tags: []
parents: []
dependents: []
---

# Test Contract

## INTENT
intent

## THE RULE / LOGIC
logic

## TECHNICAL INTERFACE (The Bridge)
bridge

## EXPECTATION
expectation
`
	visionData := `---
id: vision_test
human_name: "Test Vision"
type: VISION
layer: BUSINESS
version: 1.0
status: STABLE
priority: CORE
tags: []
parents: []
dependents: []
---

# Test Vision

## INTENT
intent

## THE RULE / LOGIC
logic

## TECHNICAL INTERFACE (The Bridge)
bridge

## EXPECTATION
expectation
`
	ruleData := `---
id: rule_bad_parent
human_name: "Rule With Bad Parent"
type: RULE
layer: BUSINESS
version: 1.0
status: DRAFT
priority: 3
tags: []
parents:
  - [[contract_test]]
dependents: []
---

# Rule With Bad Parent

## INTENT
intent

## THE RULE / LOGIC
logic

## TECHNICAL INTERFACE (The Bridge)
bridge

## EXPECTATION
expectation
`
	tmpDir := t.TempDir()
	for name, data := range map[string]string{
		"contract_test.atom.md":   contractData,
		"vision_test.atom.md":     visionData,
		"rule_bad_parent.atom.md": ruleData,
	} {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}

	out, err := runLint(tmpDir)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expected := "CONTRACT/VISION referenced as parent: [[contract_test]]"
	if !strings.Contains(out, expected) {
		t.Errorf("Expected output to contain '%s', got: %s", expected, out)
	}
}

// governanceAtomSource renders a CONTRACT/VISION atom with caller-supplied
// parents:/dependents: blocks, so the isolation rule (ATD.md §1.4) can be
// exercised from both link fields.
func governanceAtomSource(id, atomType, parents, dependents string) string {
	return `---
id: ` + id + `
human_name: "` + id + `"
type: ` + atomType + `
layer: BUSINESS
version: 1.0
status: STABLE
priority: CORE
tags: []
parents: ` + parents + `
dependents: ` + dependents + `
---

# ` + id + `

## INTENT
intent

## THE RULE / LOGIC
logic

## TECHNICAL INTERFACE (The Bridge)
bridge

## EXPECTATION
expectation
`
}

// ordinaryAtomSource renders a non-governance BUSINESS atom with caller-supplied
// link blocks, used as the resolvable target of a governance atom's links.
func ordinaryAtomSource(id, parents, dependents string) string {
	return `---
id: ` + id + `
human_name: "` + id + `"
type: RULE
layer: BUSINESS
version: 1.0
status: DRAFT
priority: 3
tags: []
parents: ` + parents + `
dependents: ` + dependents + `
---

# ` + id + `

## INTENT
intent

## THE RULE / LOGIC
logic

## TECHNICAL INTERFACE (The Bridge)
bridge

## EXPECTATION
expectation
`
}

func writeAtoms(t *testing.T, files map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return tmpDir
}

// @test-link [[service_atd_lint]]
func TestLintGovernanceAtomDeclaresNoLinks(t *testing.T) {
	cases := []struct {
		name     string
		files    map[string]string
		expected []string
	}{
		{
			name: "contract with parents",
			files: map[string]string{
				"contract_test.atom.md": governanceAtomSource("contract_test", "CONTRACT", "\n  - [[rule_ordinary]]", "[]"),
				"vision_test.atom.md":   governanceAtomSource("vision_test", "VISION", "[]", "[]"),
				"rule_ordinary.atom.md": ordinaryAtomSource("rule_ordinary", "[]", "[]"),
			},
			expected: []string{"CONTRACT/VISION declares parents: [[rule_ordinary]]"},
		},
		{
			name: "vision with dependents",
			files: map[string]string{
				"contract_test.atom.md": governanceAtomSource("contract_test", "CONTRACT", "[]", "[]"),
				"vision_test.atom.md":   governanceAtomSource("vision_test", "VISION", "[]", "\n  - [[rule_ordinary]]"),
				"rule_ordinary.atom.md": ordinaryAtomSource("rule_ordinary", "[]", "[]"),
			},
			expected: []string{"CONTRACT/VISION declares dependents: [[rule_ordinary]]"},
		},
		{
			name: "ordinary atom names governance as dependent",
			files: map[string]string{
				"contract_test.atom.md": governanceAtomSource("contract_test", "CONTRACT", "[]", "[]"),
				"vision_test.atom.md":   governanceAtomSource("vision_test", "VISION", "[]", "[]"),
				"rule_ordinary.atom.md": ordinaryAtomSource("rule_ordinary", "[]", "\n  - [[vision_test]]"),
			},
			expected: []string{"CONTRACT/VISION referenced as dependent: [[vision_test]]"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runLint(writeAtoms(t, tc.files))
			if err == nil {
				t.Fatalf("Expected error, got nil (output: %s)", out)
			}
			for _, expected := range tc.expected {
				if !strings.Contains(out, expected) {
					t.Errorf("Expected output to contain '%s', got: %s", expected, out)
				}
			}
		})
	}
}

// @test-link [[service_atd_lint]]
func TestLintGovernanceAtomsWithEmptyLinksPass(t *testing.T) {
	dir := writeAtoms(t, map[string]string{
		"contract_test.atom.md": governanceAtomSource("contract_test", "CONTRACT", "[]", "[]"),
		"vision_test.atom.md":   governanceAtomSource("vision_test", "VISION", "[]", "[]"),
		"rule_ordinary.atom.md": ordinaryAtomSource("rule_ordinary", "[]", "[]"),
	})

	out, err := runLint(dir)
	if err != nil {
		t.Fatalf("Expected nil err, got %v: %s", err, out)
	}
}

func TestLintUnknownProject(t *testing.T) {
	wsData := `{"workspace_name": "test_ws", "projects": [{"name": "proj1", "path": "."}]}`
	tmpData := `---
id: test_atom_invalid2
human_name: "Test Atom"
type: RULE
layer: ARCHITECTURE
version: 1.0
priority: 5
status: DRAFT
parents:
  - [[unknown_proj:missing_parent]]
---

# Test Atom Invalid

## INTENT
intent

## THE RULE / LOGIC
logic

## TECHNICAL INTERFACE (The Bridge)
bridge

## EXPECTATION
expectation
`
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, ".atd.workspace"), []byte(wsData), 0644)
	os.WriteFile(filepath.Join(tmpDir, ".atd"), []byte(`{"version": 1}`), 0644)
	path := filepath.Join(tmpDir, "test_atom_invalid2.atom.md")
	if err := os.WriteFile(path, []byte(tmpData), 0644); err != nil {
		t.Fatal(err)
	}

	saved := config.Snapshot()
	defer config.Restore(saved)
	config.LoadFromDir(tmpDir)

	out, err := runLint(tmpDir)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedErrors := []string{
		"Unresolved parent link (unknown project): [[unknown_proj:missing_parent]]",
	}

	for _, expected := range expectedErrors {
		if !strings.Contains(out, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, out)
		}
	}
}

// @test-link [[rule_atd_atom_self_sufficiency]]
func TestLintOutsideDocumentReferences(t *testing.T) {
	withLogic := func(id, parents, logic string) string {
		return strings.Replace(ordinaryAtomSource(id, parents, "[]"), "\nlogic\n", "\n"+logic+"\n", 1)
	}
	dir := writeAtoms(t, map[string]string{
		"rule_parent.atom.md": ordinaryAtomSource("rule_parent", "[]", "[]"),
		"rule_other.atom.md":  ordinaryAtomSource("rule_other", "[]", "[]"),
		"rule_leaky.atom.md": withLogic("rule_leaky", "[[rule_parent]]",
			"Builds on [[rule_parent]] but defers to [[rule_other]]; see [notes](https://example.com/n), "+
				"https://example.com/raw, `reports/why.md`, and GUIDE.md §2."),
		"rule_clean.atom.md": withLogic("rule_clean", "[[rule_parent]]",
			"Builds on [[rule_parent]]; tag `@spec-link [[rule_other]]`; endpoint `http://localhost:1/x`; writes `task_list.md`."),
	})

	out, err := runLint(dir)
	if err == nil {
		t.Fatalf("Expected lint failure, got none: %s", out)
	}
	for _, expected := range []string{
		"Wiki-link to an atom outside parents:/dependents: [[rule_other]]",
		"Markdown link to outside document: https://example.com/n",
		"URL to outside document: https://example.com/raw",
		"Citation of outside document: reports/why.md",
		"Citation of outside document: GUIDE.md §2",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("Expected output to contain %q, got: %s", expected, out)
		}
	}
	if strings.Contains(out, "outside parents:/dependents: [[rule_parent]]") {
		t.Errorf("A prose link to the atom's own parent must not be flagged, got: %s", out)
	}
	// Only rule_leaky's five references may be reported; rule_clean's parent
	// link, tag, code-span URL and artifact filename are all allowed.
	if n := strings.Count(out, "an atom must stand on its own"); n != 5 {
		t.Errorf("Expected exactly 5 self-sufficiency findings, got %d: %s", n, out)
	}
}
