package cmd

import (
	"strings"
	"testing"
)

func TestCongruenceFlags(t *testing.T) {
	cmd := congruenceCmd
	if cmd.Use != "congruence" {
		t.Errorf("expected use 'congruence', got %s", cmd.Use)
	}

	targetFlag := cmd.Flag("target")
	if targetFlag == nil {
		t.Error("expected flag 'target'")
	}
}

func TestCongruenceRun(t *testing.T) {
	// Simple validation test
	congruenceCmd.Flags().Set("target", "")
	err := congruenceCmd.RunE(congruenceCmd, []string{})
	if err == nil {
		t.Error("expected error for empty target")
	}
	if !strings.Contains(err.Error(), "--target is required") {
		t.Errorf("expected '--target is required', got %v", err)
	}
}
