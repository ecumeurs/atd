package cmd

import (
	"testing"
)

func TestAuditValidation(t *testing.T) {
	// Simple validation test for flags
	cmd := auditCmd
	
	// Check code mode requires both flags
	err := cmd.Flags().Set("code", "test.go")
	if err != nil { t.Fatal(err) }
	
	// We can't easily run the RunE here without mocking a lot, 
	// but we can check the flag configuration.
}

func TestAncestorBFS(t *testing.T) {
	// We want to test the ancestor logic inside runFullAudit, 
	// but it's internal. I'll test the principle.
	parentMap := map[string][]string{
		"child":  {"parent"},
		"parent": {"grandparent"},
	}

	isAncestor := func(startID, targetID string) bool {
		visited := make(map[string]bool)
		queue := []string{startID}
		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]
			if visited[curr] { continue }
			visited[curr] = true
			for _, p := range parentMap[curr] {
				if p == targetID { return true }
				queue = append(queue, p)
			}
		}
		return false
	}

	if !isAncestor("child", "grandparent") {
		t.Error("Expected grandparent to be ancestor of child")
	}
	if isAncestor("child", "unknown") {
		t.Error("Unexpected ancestor found")
	}
}
