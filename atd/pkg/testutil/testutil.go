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
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
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

// SnapshotConfigLocked is SnapshotConfig's t.Parallel()-safe counterpart: the
// registered restore takes configMu before writing config.ActiveConfig, so it
// can never race with another parallel test's (*SB).Run critical section.
// Any t.Parallel() test that loads/mutates config.ActiveConfig outside of
// Sandbox (e.g. an ad-hoc temp-dir fixture rather than a copied testdata/
// fixture) should call this instead of SnapshotConfig.
func SnapshotConfigLocked(t testing.TB) {
	t.Helper()
	configMu.Lock()
	saved := config.Snapshot()
	configMu.Unlock()
	t.Cleanup(func() {
		configMu.Lock()
		defer configMu.Unlock()
		config.Restore(saved)
	})
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

// ─────────────────────────────────────────────────────────────────────────
// WP-2 fixture-corpus / scenario harness (test_atd_07_26.md §3.2).
//
// This extends the WP-0 primitives above (SnapshotConfig/Tripwire/Guard)
// with the Sandbox/Run/Golden/Git API the scenario catalog (§4) is built on.
// A Sandbox copies a named fixture under testdata/ into a fresh t.TempDir(),
// so every test operates on its own disposable, git-free copy — no scenario
// test may mutate the checked-in fixture, and (per incident I-1) no test
// may accidentally anchor product code's cwd-fallback config at a real
// source directory.
// ─────────────────────────────────────────────────────────────────────────

// updateGolden is the -update flag: run `go test ./... -update` to
// (re)generate golden files instead of comparing against them. Registered
// here (not in a _test.go file) so any test binary that imports testutil —
// in either Go module — gets the flag for free.
var updateGolden = flag.Bool("update", false, "regenerate golden files instead of comparing against them")

// configMu serializes access to config.ActiveConfig across concurrent
// t.Parallel() scenario tests. config.ActiveConfig is a package-level
// global (that's the whole reason WP-0 exists), so two sandboxes cannot
// safely load their own config into it at the same instant. Run() takes
// this lock for the (re)load-then-invoke critical section only; copying a
// fixture into t.TempDir() in Sandbox() itself needs no lock and is free to
// run concurrently, which is what actually makes t.Parallel() worthwhile
// here.
var configMu sync.Mutex

// SB is a sandboxed copy of a fixture project. Every path a test touches
// lives under Root; DocsDir and SrcDir are the conventional docs/ and src/
// subdirectories every fixture in testdata/ uses.
type SB struct {
	t       testing.TB
	Root    string
	DocsDir string
	SrcDir  string
}

// Result captures the output/error of one entry-point call made through
// (*SB).Run.
type Result struct {
	Output string
	Err    error
}

// Sandbox copies the named fixture (a directory under testdata/, resolved
// by walking upward from the calling test's working directory — this works
// from either Go module, whichever package's tests call it) into a fresh
// t.TempDir(), and registers cleanups that restore config.ActiveConfig and
// run the source-directory tripwire (see Guard/Tripwire above). It does
// NOT itself load config.ActiveConfig — that happens inside Run, under
// configMu, so that concurrent t.Parallel() sandboxes never race on the
// global. Call (*SB).Sub to scope into a member project of a workspace
// fixture (e.g. Sandbox(t, "fixture_workspace").Sub("zzfix_a")).
//
// Note this deliberately does NOT call the package-level SnapshotConfig
// helper: that registers a t.Cleanup that calls config.Restore directly,
// unsynchronized -- fine for WP-0's original serial tests, but a real
// (go test -race confirmed) data race against Run's mutex-protected loads
// once scenario tests run with t.Parallel(). The cleanup below takes
// configMu itself so restore-on-cleanup and load-in-Run can never overlap.
func Sandbox(t *testing.T, fixture string) *SB {
	t.Helper()

	SnapshotConfigLocked(t)
	installTripwire(t, 2)

	src, err := findTestdataFixture(fixture)
	if err != nil {
		t.Fatalf("testutil.Sandbox(%q): %v", fixture, err)
	}

	dst := t.TempDir()
	if err := copyDir(src, dst); err != nil {
		t.Fatalf("testutil.Sandbox(%q): copying fixture into sandbox: %v", fixture, err)
	}

	return &SB{
		t:       t,
		Root:    dst,
		DocsDir: filepath.Join(dst, "docs"),
		SrcDir:  filepath.Join(dst, "src"),
	}
}

// Sub returns an SB scoped to a member-project subdirectory of a workspace
// fixture (Root/rel), with DocsDir/SrcDir recomputed under it. It shares no
// mutable state with the parent SB — it is just a differently-rooted view
// of the same on-disk sandbox, so mutations through either are visible to
// both.
func (s *SB) Sub(rel string) *SB {
	root := filepath.Join(s.Root, rel)
	return &SB{
		t:       s.t,
		Root:    root,
		DocsDir: filepath.Join(root, "docs"),
		SrcDir:  filepath.Join(root, "src"),
	}
}

// Run (re)loads config.ActiveConfig from the sandbox root and invokes fn,
// capturing its output/error as a Result. The load-then-invoke pair runs
// under configMu so that concurrent t.Parallel() scenario tests — each
// with its own Sandbox — never observe a config loaded by a different
// sandbox mid-call. fn is typically a closure over one of the cmd package's
// run* entry points (the same functions the MCP handlers call), e.g.:
//
//	res := s.Run(func() (string, error) { return runLint(s.DocsDir) })
func (s *SB) Run(fn func() (string, error)) Result {
	configMu.Lock()
	defer configMu.Unlock()

	if err := config.LoadFromDir(s.Root); err != nil {
		return Result{Err: fmt.Errorf("testutil: config.LoadFromDir(%s): %w", s.Root, err)}
	}

	out, err := fn()
	return Result{Output: out, Err: err}
}

// timestampPattern matches RFC3339-ish timestamps so Golden can normalize
// them out of report text before comparing (build/report timestamps would
// otherwise make every golden file flaky).
var timestampPattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?`)

// Golden compares got (after normalization) against testdata/golden/<name>,
// failing the test on any difference. Run with -update to (re)generate the
// golden file instead of comparing — two consecutive -update runs must
// produce byte-identical files, which is exactly what the normalization
// below exists to guarantee (it strips the two things that vary run to
// run: this sandbox's absolute temp path, and any embedded timestamp).
func (s *SB) Golden(t *testing.T, name, got string) {
	t.Helper()

	normalized := normalizeGolden(got, s.Root)

	dir, err := findTestdataGoldenDir()
	if err != nil {
		t.Fatalf("testutil.Golden(%q): %v", name, err)
	}
	path := filepath.Join(dir, name)

	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("testutil.Golden(%q): creating golden dir: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(normalized), 0644); err != nil {
			t.Fatalf("testutil.Golden(%q): writing golden file: %v", name, err)
		}
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("testutil.Golden(%q): golden file missing at %s (run `go test -update ./...` to create it): %v", name, path, err)
	}
	if normalized != string(want) {
		t.Errorf("golden mismatch for %s:\n--- want (%s) ---\n%s\n--- got ---\n%s", name, path, want, normalized)
	}
}

// normalizeGolden strips the two sources of run-to-run variance from
// report-shaped output before it is compared or written as a golden file:
// the sandbox's own absolute path (a fresh t.TempDir() every run) and any
// embedded timestamp.
func normalizeGolden(s, root string) string {
	out := s
	if root != "" {
		out = strings.ReplaceAll(out, root, "<SANDBOX_ROOT>")
	}
	out = timestampPattern.ReplaceAllString(out, "<TIMESTAMP>")
	return out
}

// Git initializes a git repository inside the sandbox and commits its
// entire current contents, giving diff-mode scenarios (test_atd_07_26.md
// §4, S13) a clean HEAD to diff uncommitted changes against. Fails the
// test loudly if git is unavailable or any step errors, rather than
// silently producing a sandbox that isn't actually under version control.
func (s *SB) Git(t *testing.T) {
	t.Helper()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = s.Root
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=atd-testutil", "GIT_AUTHOR_EMAIL=atd-testutil@example.com",
			"GIT_COMMITTER_NAME=atd-testutil", "GIT_COMMITTER_EMAIL=atd-testutil@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("testutil.Git: `git %s` failed: %v\n%s", strings.Join(args, " "), err, out)
		}
	}

	run("init", "-q")
	run("-c", "user.name=atd-testutil", "-c", "user.email=atd-testutil@example.com", "add", "-A")
	run("-c", "user.name=atd-testutil", "-c", "user.email=atd-testutil@example.com", "commit", "-q", "-m", "sandbox baseline")
}

// findTestdataFixture walks upward from the current working directory (the
// tested package's own source directory, per `go test` convention — this
// is what makes the search work identically whether the calling test lives
// in the atd-tools module or the nested atd/cmd/atd module) looking for
// testdata/<fixture> as a directory. Bounded to avoid an unbounded walk if
// something odd happens with the filesystem root.
func findTestdataFixture(fixture string) (string, error) {
	return findUpward(filepath.Join("testdata", fixture))
}

// findTestdataGoldenDir locates testdata/golden the same way.
func findTestdataGoldenDir() (string, error) {
	return findUpward(filepath.Join("testdata", "golden"))
}

func findUpward(rel string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < 32; i++ {
		candidate := filepath.Join(dir, rel)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find %s by walking upward from %s", rel, mustGetwd())
}

func mustGetwd() string {
	wd, _ := os.Getwd()
	return wd
}

// copyDir recursively copies src to dst, creating dst if needed. Used to
// materialize a checked-in fixture into a disposable sandbox; file modes
// are preserved, symlinks are not supported (fixtures don't need them).
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// SortedKeys is a small shared helper scenario tests reach for repeatedly
// (e.g. iterating every atom id in a fixture's docs graph in a
// deterministic order for S3's consistency-oracle loop).
func SortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
