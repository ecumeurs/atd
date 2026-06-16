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

	config.LoadFromDir(tmpDir)
	defer func() { config.ActiveConfig.Workspace = nil }()

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
