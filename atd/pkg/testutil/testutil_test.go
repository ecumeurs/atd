package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"atd-tools/config"
)

// TestSnapshotConfigRestoresActiveConfig proves the config-restore half of
// W-1: after the test (and its t.Cleanup) finishes, config.ActiveConfig must
// be back to what it was before, regardless of what the test did to it.
func TestSnapshotConfigRestoresActiveConfig(t *testing.T) {
	config.ActiveConfig = config.Config{}
	before := config.ProjectRoot()

	// Drive SnapshotConfig from a subtest so its t.Cleanup fires (subtest
	// cleanups run when the subtest returns, before the parent continues),
	// then assert from the parent that the restore actually happened.
	t.Run("mutates", func(t *testing.T) {
		SnapshotConfig(t)
		if err := config.LoadFromDir(t.TempDir()); err != nil {
			t.Fatalf("LoadFromDir failed: %v", err)
		}
		if config.ProjectRoot() == before {
			t.Fatal("test setup invalid: LoadFromDir did not change ProjectRoot")
		}
	})

	if config.ProjectRoot() != before {
		t.Errorf("expected config.ProjectRoot() restored to %q after subtest cleanup, got %q", before, config.ProjectRoot())
	}
}

// TestTripwireCleanRun proves the tripwire does not false-positive when the
// test doesn't touch its own package source directory.
func TestTripwireCleanRun(t *testing.T) {
	t.Run("clean", func(t *testing.T) {
		Tripwire(t)
		// Touch only a temp dir — never this package's own source.
		f := filepath.Join(t.TempDir(), "scratch.txt")
		if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}
	})
}

// The remaining tests exercise tripwireViolations directly — the pure diff
// logic behind the t.Cleanup that Tripwire/Guard install — against a
// throwaway directory. This pins the exact incident I-1 detection behavior
// (modified / added / removed files under a package source dir) without
// needing a real test failure to propagate through t.Run, which Go's testing
// package does not let a parent test observe and suppress.

func TestTripwireViolationsDetectsModification(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "victim.go")
	if err := os.WriteFile(f, []byte("package x\n"), 0644); err != nil {
		t.Fatal(err)
	}

	before := snapshotMtimes(dir)

	// Force a detectable mtime change (some filesystems have coarse mtime
	// resolution, so bump it explicitly rather than relying on wall-clock
	// drift between two fast writes).
	newContent := []byte("package x\n// mutated\n")
	if err := os.WriteFile(f, newContent, 0644); err != nil {
		t.Fatal(err)
	}
	newTime := before[f].Add(2 * time.Second)
	if err := os.Chtimes(f, newTime, newTime); err != nil {
		t.Fatal(err)
	}

	violations := tripwireViolations(dir, before)
	if len(violations) != 1 {
		t.Fatalf("expected exactly 1 violation for the modified file, got %d: %v", len(violations), violations)
	}
}

func TestTripwireViolationsDetectsNewFile(t *testing.T) {
	dir := t.TempDir()
	before := snapshotMtimes(dir) // empty package dir

	f := filepath.Join(dir, "new_file.go")
	if err := os.WriteFile(f, []byte("package x\n"), 0644); err != nil {
		t.Fatal(err)
	}

	violations := tripwireViolations(dir, before)
	if len(violations) != 1 {
		t.Fatalf("expected exactly 1 violation for the new file, got %d: %v", len(violations), violations)
	}
}

func TestTripwireViolationsDetectsRemovedFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "doomed.go")
	if err := os.WriteFile(f, []byte("package x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	before := snapshotMtimes(dir)

	if err := os.Remove(f); err != nil {
		t.Fatal(err)
	}

	violations := tripwireViolations(dir, before)
	if len(violations) != 1 {
		t.Fatalf("expected exactly 1 violation for the removed file, got %d: %v", len(violations), violations)
	}
}

func TestTripwireViolationsCleanIsEmpty(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "untouched.go")
	if err := os.WriteFile(f, []byte("package x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	before := snapshotMtimes(dir)

	violations := tripwireViolations(dir, before)
	if len(violations) != 0 {
		t.Fatalf("expected no violations for an untouched dir, got: %v", violations)
	}
}
