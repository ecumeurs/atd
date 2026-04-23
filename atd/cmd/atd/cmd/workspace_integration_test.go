package cmd

import (
	"atd-tools/config"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceCommandsIntegration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "atd-ws-integration-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	oldCwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldCwd)

	// 1. Test 'workspace init'
	rootCmd.SetArgs([]string{"workspace", "init", "--name", "my-test-ws"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("workspace init failed: %v", err)
	}

	if _, err := os.Stat(".atd.workspace"); os.IsNotExist(err) {
		t.Fatal(".atd.workspace was not created")
	}

	data, _ := os.ReadFile(".atd.workspace")
	var ws config.WorkspaceConfig
	json.Unmarshal(data, &ws)
	if ws.WorkspaceName != "my-test-ws" {
		t.Errorf("Expected name 'my-test-ws', got '%s'", ws.WorkspaceName)
	}

	// 2. Test 'workspace add'
	projDir := filepath.Join(tmpDir, "proj-1")
	os.MkdirAll(projDir, 0755)

	rootCmd.SetArgs([]string{"workspace", "add", "--name", "proj-1", "--path", "./proj-1"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("workspace add failed: %v", err)
	}

	data, _ = os.ReadFile(".atd.workspace")
	json.Unmarshal(data, &ws)
	if len(ws.Projects) != 1 || ws.Projects[0].Name != "proj-1" {
		t.Errorf("Project was not added correctly")
	}

	// 3. Test 'workspace list'
	rootCmd.SetArgs([]string{"workspace", "list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("workspace list failed: %v", err)
	}

	// 4. Test --project flag integration
	// Create .atd in proj-1
	os.WriteFile(filepath.Join(projDir, ".atd"), []byte(`{"docs_path": "custom-docs/"}`), 0644)

	// Run stats with --project from ws root
	rootCmd.SetArgs([]string{"stats", "--project", "proj-1"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("stats --project failed: %v", err)
	}
	
	if config.ActiveConfig.ActiveProject != "proj-1" {
		t.Errorf("Expected active project 'proj-1', got '%s'", config.ActiveConfig.ActiveProject)
	}
	if config.ActiveConfig.DocsPath != "custom-docs/" {
		t.Errorf("Expected docs path 'custom-docs/', got '%s'", config.ActiveConfig.DocsPath)
	}
}
