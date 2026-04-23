package cmd

import (
	"atd-tools/pkg/mcp"
	"testing"
)

func TestRegisterMCPWorkspaceTools(t *testing.T) {
	registry := mcp.NewRegistry()
	RegisterMCPTools(registry)

	tools := registry.List()
	
	foundList := false
	foundUse := false
	foundWSStats := false
	
	for _, tool := range tools {
		if tool.Name == "atd_workspace_list" {
			foundList = true
		}
		if tool.Name == "atd_workspace_use" {
			foundUse = true
		}
		if tool.Name == "atd_workspace_stats" {
			foundWSStats = true
		}
	}

	if !foundList {
		t.Error("atd_workspace_list tool not registered")
	}
	if !foundUse {
		t.Error("atd_workspace_use tool not registered")
	}
	if !foundWSStats {
		t.Error("atd_workspace_stats tool not registered")
	}
}
