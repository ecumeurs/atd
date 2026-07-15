// Package testutil holds the sandboxing primitives every ATD test that
// touches config.ActiveConfig or the filesystem should use. It exists
// because of incident I-1 (test_atd_07_26.md §2.1): a test that left
// config.ActiveConfig loaded (via the cwd fallback) leaked that global state
// into whichever test ran next in the same binary, and a downstream product
// feature (atom.UpdateLinks' rename-propagation walk) trusted that leaked
// config enough to rewrite the test suite's own source files — one such
// mutation was even committed to main unnoticed.
//
// This package is deliberately small: it is the WP-0 "config restore +
// tripwire" primitives only. The fuller Sandbox/Run/Golden/Git fixture-project
// API described in test_atd_07_26.md §3.2 is WP-2's job and should be built
// as an extension of SnapshotConfig/Tripwire below, not a replacement.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"atd-tools/config"
)

// SnapshotConfig saves config.ActiveConfig and registers a t.Cleanup that
// restores it once the test (and any subtests) finish. Every test that calls
// config.Load / config.LoadFromDir / config.LoadFromDirLegacy, or that
// mutates config.ActiveConfig fields directly, must call this first —
// otherwise the mutated config leaks into later tests in the same binary
// (see test_atd_07_26.md §3.6, T-1).
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    testutil.SnapshotConfig(t)
//	    config.LoadFromDir(t.TempDir())
//	    // ...
//	}
func SnapshotConfig(t testing.TB) {
	t.Helper()
	saved := config.Snapshot()
	t.Cleanup(func() { config.Restore(saved) })
}

// Tripwire records the mtime of every file under the calling test's own
// package source directory and registers a t.Cleanup that fails the test if
// any of those files were added, removed, or modified while the test ran.
// This is the local, best-effort half of the W-1 sandbox contract
// (test_atd_07_26.md §3.6): it is exactly what would have turned incident I-1
// (a test rewriting its own fixture file) into an immediate, loud test
// failure instead of a phantom bug that took a day to root-cause.
//
// Tripwire must be called directly from the test function (not from a nested
// helper) so it can locate the test's own source file via runtime.Caller.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    testutil.Tripwire(t)
//	    // ... exercise code that must never touch this package's own files ...
//	}
func Tripwire(t testing.TB) {
	t.Helper()
	installTripwire(t, 2)
}

// Guard is the common case: SnapshotConfig plus Tripwire in one call. New
// tests that load or mutate config.ActiveConfig should generally just call
// Guard(t) once at the top of the test.
func Guard(t testing.TB) {
	t.Helper()
	SnapshotConfig(t)
	installTripwire(t, 2)
}

// installTripwire does the actual work for Tripwire/Guard. skip is the
// runtime.Caller depth needed to reach the original test function (2 when
// called directly from Tripwire or Guard, which are themselves called
// directly from the test).
func installTripwire(t testing.TB, skip int) {
	t.Helper()

	_, file, _, ok := runtime.Caller(skip)
	if !ok {
		t.Fatal("testutil: could not determine caller source file for tripwire")
		return
	}
	dir := filepath.Dir(file)

	before := snapshotMtimes(dir)
	t.Cleanup(func() {
		for _, msg := range tripwireViolations(dir, before) {
			t.Error(msg)
		}
	})
}

// tripwireViolations re-snapshots dir and returns one human-readable message
// per file that was added, removed, or had its mtime change relative to
// before. Split out from installTripwire's t.Cleanup so the diff logic is
// unit-testable on its own (see testutil_test.go) without needing a real
// test failure to propagate through t.Run.
func tripwireViolations(dir string, before map[string]time.Time) []string {
	after := snapshotMtimes(dir)
	var violations []string

	for path, mt := range before {
		newMt, stillExists := after[path]
		if !stillExists {
			violations = append(violations, fmt.Sprintf("testutil tripwire: file under package source dir %s disappeared during test: %s", dir, path))
			continue
		}
		if !newMt.Equal(mt) {
			violations = append(violations, fmt.Sprintf("testutil tripwire: file under package source dir %s was modified during test (mtime %s -> %s): %s — a test must never mutate its own package source; see test_atd_07_26.md §2.1 (incident I-1)", dir, mt, newMt, path))
		}
	}
	for path := range after {
		if _, existedBefore := before[path]; !existedBefore {
			violations = append(violations, fmt.Sprintf("testutil tripwire: new file appeared under package source dir %s during test: %s", dir, path))
		}
	}
	return violations
}

// snapshotMtimes walks dir (non-recursively into subpackages is fine since
// filepath.Walk does recurse, but package source dirs are flat in practice)
// and returns a path -> mtime map for every regular file found. Errors
// walking are ignored: a best-effort tripwire is better than a fragile one
// that panics test setup.
func snapshotMtimes(dir string) map[string]time.Time {
	out := make(map[string]time.Time)
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		out[path] = info.ModTime()
		return nil
	})
	return out
}
