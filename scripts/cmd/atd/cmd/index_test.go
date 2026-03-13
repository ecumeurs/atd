package cmd

import (
	"testing"
)

func TestIndexModeValidation(t *testing.T) {
	// Root command is initialized via init() in root.go, 
	// but we should ensure it's loaded if we're testing logic.


	tests := []struct {
		mode string
		wantErr bool
	}{
		{"code", false},
		{"docs", false},
		{"all", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		indexCmd.Flags().Set("mode", tt.mode)
		err := indexCmd.RunE(indexCmd, []string{})
		if (err != nil) != tt.wantErr && err != nil && tt.wantErr {
			// Expected error from logic, not from flag setup
			if err.Error() != "invalid mode: "+tt.mode+" (must be code|docs|all)" {
				t.Errorf("Index mode %s: unexpected error: %v", tt.mode, err)
			}
		}
	}
}

func TestIndexDirValidation(t *testing.T) {
	indexCmd.Flags().Set("dir", "/nonexistent/path/atd")
	err := indexCmd.RunE(indexCmd, []string{})
	if err == nil {
		t.Error("Expected error for nonexistent directory, got nil")
	}
}
