package cmd

import (
	"strings"
	"testing"
)

func TestReconcileFlags(t *testing.T) {
	cmd := reconcileCmd
	if cmd.Use != "reconcile" {
		t.Errorf("expected use 'reconcile', got %s", cmd.Use)
	}

	for _, flagName := range []string{"new", "store"} {
		if cmd.Flag(flagName) == nil {
			t.Errorf("expected flag '%s'", flagName)
		}
	}
}

func TestReconcileRun(t *testing.T) {
	reconcileCmd.Flags().Set("new", "")
	reconcileCmd.Flags().Set("store", "")
	err := reconcileCmd.RunE(reconcileCmd, []string{})
	if err == nil {
		t.Error("expected error for empty paths")
	}
	if !strings.Contains(err.Error(), "--new and --store are required") {
		t.Errorf("expected '--new and --store are required', got %v", err)
	}
}
