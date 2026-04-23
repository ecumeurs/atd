package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWorkspaceConfig(t *testing.T) {
	// Create a temp workspace structure
	tmp, err := os.MkdirTemp("", "atd-test-ws-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	projA := filepath.Join(tmp, "proj-a")
	os.MkdirAll(projA, 0755)

	wsContent := `{"workspace_name": "test-ws", "projects": [{"name": "proj-a", "path": "./proj-a"}]}`
	os.WriteFile(filepath.Join(tmp, ".atd.workspace"), []byte(wsContent), 0644)

	// Test loading from root
	ws, err := LoadWorkspaceConfig(tmp)
	if err != nil {
		t.Fatalf("Failed to load workspace from root: %v", err)
	}
	if ws.WorkspaceName != "test-ws" {
		t.Errorf("Expected workspace name 'test-ws', got '%s'", ws.WorkspaceName)
	}

	// Test loading from subproject
	ws, err = LoadWorkspaceConfig(projA)
	if err != nil {
		t.Fatalf("Failed to load workspace from subproject: %v", err)
	}
	if ws.WorkspaceName != "test-ws" {
		t.Errorf("Expected workspace name 'test-ws', got '%s'", ws.WorkspaceName)
	}
}

func TestLoadInWorkspace(t *testing.T) {
	tmp, err := os.MkdirTemp("", "atd-test-load-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	projA := filepath.Join(tmp, "proj-a")
	os.MkdirAll(projA, 0755)

	wsContent := `{"workspace_name": "test-ws", "projects": [{"name": "proj-a", "path": "./proj-a"}]}`
	os.WriteFile(filepath.Join(tmp, ".atd.workspace"), []byte(wsContent), 0644)

	projAConfig := `{"docs_path": "custom-docs/"}`
	os.WriteFile(filepath.Join(projA, ".atd"), []byte(projAConfig), 0644)

	// Change CWD to projA
	oldCwd, _ := os.Getwd()
	os.Chdir(projA)
	defer os.Chdir(oldCwd)

	err = Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if ActiveConfig.Workspace == nil {
		t.Error("Workspace should be loaded")
	}
	if ActiveConfig.ActiveProject != "proj-a" {
		t.Errorf("Expected active project 'proj-a', got '%s'", ActiveConfig.ActiveProject)
	}
	if ActiveConfig.DocsPath != "custom-docs/" {
		t.Errorf("Expected docs path 'custom-docs/', got '%s'", ActiveConfig.DocsPath)
	}
}
