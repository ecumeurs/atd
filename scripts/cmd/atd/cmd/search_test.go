package cmd

import (
	"testing"
)

func TestSearchValidation(t *testing.T) {
	// Clear flags
	searchCmd.Flags().Set("query", "")
	searchCmd.Flags().Set("grep", "")

	err := searchCmd.RunE(searchCmd, []string{})
	if err == nil {
		t.Error("Expected error for empty query/grep, got nil")
	}
}
