package cmd

import (
	"strings"
	"testing"
)

func TestDiscoverFlags(t *testing.T) {
	cmd := discoverCmd
	if cmd.Use != "discover" {
		t.Errorf("expected use 'discover', got %s", cmd.Use)
	}

	if cmd.Flag("file") == nil {
		t.Error("expected flag 'file'")
	}
}

func TestDiscoverRun(t *testing.T) {
	discoverCmd.Flags().Set("file", "")
	err := discoverCmd.RunE(discoverCmd, []string{})
	if err == nil {
		t.Error("expected error for empty file path")
	}
	if !strings.Contains(err.Error(), "--file is required") {
		t.Errorf("expected '--file is required', got %v", err)
	}
}
