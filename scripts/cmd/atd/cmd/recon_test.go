package cmd

import (
	"strings"
	"testing"
)

func TestReconFlags(t *testing.T) {
	cmd := reconCmd
	if cmd.Use != "recon" {
		t.Errorf("expected use 'recon', got %s", cmd.Use)
	}

	for _, flagName := range []string{"atom", "candidate"} {
		if cmd.Flag(flagName) == nil {
			t.Errorf("expected flag '%s'", flagName)
		}
	}
}

func TestReconRun(t *testing.T) {
	reconCmd.Flags().Set("atom", "")
	reconCmd.Flags().Set("candidate", "")
	err := reconCmd.RunE(reconCmd, []string{})
	if err == nil {
		t.Error("expected error for empty paths")
	}
	if !strings.Contains(err.Error(), "--atom and --candidate are required") {
		t.Errorf("expected '--atom and --candidate are required', got %v", err)
	}
}
