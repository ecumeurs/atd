package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/atom"
)

func TestUpdateNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "new_atom.atom.md")

	// Create new file via atom.Update
	_, err := atom.Update(atom.UpdateOptions{
		FilePath:    filePath,
		SetArgs:     []string{"type=SERVICE", "human_name=Test Service", "id=test_service"},
		Intent:      "Test Intent",
		Logic:       "Test Logic",
		Interface:   "Test Interface",
		Expectation: "Test Expectation",
	})
	if err != nil {
		t.Fatalf("atom.Update failed: %v", err)
	}

	// Verify target file exists (it should be renamed based on ID)
	targetPath := filepath.Join(tmpDir, "service_test_service.atom.md")
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Target file not found: %v", err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, "id: service_test_service") {
		t.Errorf("Expected id: service_test_service, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "type: SERVICE") {
		t.Errorf("Expected type: SERVICE, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "status: DRAFT") {
		t.Errorf("Expected status: DRAFT, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "version: 1.0") {
		t.Errorf("Expected version: 1.0, got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "parents: []") {
		t.Errorf("Expected parents: [], got:\n%s", strContent)
	}
	if !strings.Contains(strContent, "## EXPECTATION\nTest Expectation") {
		t.Errorf("Expected EXPECTATION section, got:\n%s", strContent)
	}

	// Test newline injection
	_, err = atom.Update(atom.UpdateOptions{FilePath: targetPath, Intent: "Line1\\nLine2"})
	content, _ = os.ReadFile(targetPath)
	if !strings.Contains(string(content), "Line1\nLine2") {
		t.Errorf("Expected actual newline, got:\n%s", string(content))
	}
}

func TestUpdateNamingConvention(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "temp.atom.md")
	err := os.WriteFile(filePath, []byte("---\nid: old_id\ntype: MODULE\n---\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Update with new ID that doesn't follow convention
	_, err = atom.Update(atom.UpdateOptions{FilePath: filePath, SetArgs: []string{"id=my_new_module"}})
	if err != nil {
		t.Fatal(err)
	}

	targetPath := filepath.Join(tmpDir, "module_my_new_module.atom.md")
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Errorf("Expected renamed file %s not found", targetPath)
	} else if !strings.Contains(string(content), "id: module_my_new_module") {
		t.Errorf("Expected id: module_my_new_module, got:\n%s", string(content))
	}
}

func TestUpdateLinks(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create a file that references another
	refPath := filepath.Join(tmpDir, "ref.atom.md")
	refContent := "---\nid: ref_atom\nparents: \n  - [[module_my_new_module]]\n---\nSee [[module_my_new_module]] for details."
	os.WriteFile(refPath, []byte(refContent), 0644)

	// Create the atom to be renamed
	atomPath := filepath.Join(tmpDir, "atom.atom.md")
	atomContent := "---\nid: module_my_new_module\ntype: RULE\n---\n"
	os.WriteFile(atomPath, []byte(atomContent), 0644)

	numUpdates := atom.UpdateLinks(tmpDir, "module_my_new_module", "new_id")
	if numUpdates != 1 {
		t.Errorf("Expected 1 file updated, got %d", numUpdates)
	}

	updatedRef, _ := os.ReadFile(refPath)
	if !strings.Contains(string(updatedRef), "[[new_id]]") {
		t.Errorf("Link not updated in ref file:\n%s", string(updatedRef))
	}
	if strings.Contains(string(updatedRef), "[[module_my_new_module]]") {
		t.Error("Old link still present in ref file")
	}
}
