package cmd

// Regression tests for failures/20260916_orphaned_architecture_atom_blocks_commit.md:
// the ATD Structural Integrity pre-commit hook (preCommitHookContent, see
// init.go) used to fail a commit for ANY staged .atom.md file whose current
// (working-tree/staged) content had zero parents on an ARCHITECTURE or
// IMPLEMENTATION atom -- even when that atom was ALREADY orphaned at HEAD
// and the staged change never touched `parents` at all. That ambushed
// whoever next happened to touch a long-orphaned atom for an unrelated
// reason.
//
// FIXED: the hook now only blocks the commit for an orphan condition THIS
// commit introduces (a brand-new atom with zero parents) or worsens (an
// atom that had at least one parent at HEAD but is now staged with zero).
// A pre-existing orphan (zero parents already at HEAD, untouched by this
// commit's `parents` field) prints a non-blocking warning instead of
// failing the commit.
//
// These tests exercise the hook script as literal bash against a real,
// isolated temp git repository -- the same way git itself invokes
// .git/hooks/pre-commit: no arguments, cwd at the repo root, with the
// index already staged via `git add`.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initHookTestRepo creates an isolated temp git repository (fully isolated
// from the host's global/system git config so the test is hermetic) and
// installs preCommitHookContent as its pre-commit hook.
func initHookTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
			"HOME="+dir,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
		return string(out)
	}

	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")
	run("config", "commit.gpgsign", "false")

	hooksDir := filepath.Join(dir, ".git", "hooks")
	hookPath := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(preCommitHookContent), 0755); err != nil {
		t.Fatalf("failed to write pre-commit hook: %v", err)
	}

	return dir
}

// gitIn runs a git command in dir with the same hermetic environment as
// initHookTestRepo, returning combined output and the *exec.ExitError (or
// nil) so callers can assert on both exit status and printed hook output.
func gitIn(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"HOME="+dir,
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func writeHookFixtureAtom(t *testing.T, dir, relPath, id, layer string, parentLines []string) {
	t.Helper()
	var parents string
	if len(parentLines) == 0 {
		parents = "parents: []\n"
	} else {
		parents = "parents:\n"
		for _, p := range parentLines {
			parents += "  - " + p + "\n"
		}
	}
	content := "---\n" +
		"id: " + id + "\n" +
		"type: " + layer + "_ATOM\n" +
		"layer: " + layer + "\n" +
		"status: DRAFT\n" +
		parents +
		"dependents: []\n" +
		"---\n\n" +
		"## INTENT\nFixture atom " + id + ".\n\n" +
		"## THE RULE / LOGIC\nFixture logic for " + id + ".\n\n" +
		"## EXPECTATION\nn/a\n"

	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// Case 1 (the field report's exact repro): a commit touches only an
// already-orphaned ARCHITECTURE atom's unrelated field, never `parents`.
// The hook must let the commit through (exit 0), while still printing a
// visible, non-blocking warning about the pre-existing debt.
func TestPreCommitHook_PreExistingOrphan_UnrelatedFieldChange_DoesNotBlock(t *testing.T) {
	dir := t.TempDir()

	// Seed history WITHOUT the hook installed, so a long-orphaned atom can
	// exist at HEAD -- exactly as it would in a real repo where the orphan
	// predates the hook (or was committed before this check existed).
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
			"HOME="+dir,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")
	run("config", "commit.gpgsign", "false")

	writeHookFixtureAtom(t, dir, "docs/api_laravel_gateway.atom.md", "api_laravel_gateway", "ARCHITECTURE", nil)
	run("add", "docs/api_laravel_gateway.atom.md")
	run("commit", "-q", "-m", "initial state: pre-existing orphan")

	// Now install the hook -- as if `atd init` (or `--upgrade`) ran after
	// this orphan already existed in history.
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte(preCommitHookContent), 0755); err != nil {
		t.Fatalf("failed to write pre-commit hook: %v", err)
	}

	// Stage an unrelated change: bump status, never touch `parents`.
	writeHookFixtureAtom(t, dir, "docs/api_laravel_gateway.atom.md", "api_laravel_gateway", "ARCHITECTURE", nil)
	content, err := os.ReadFile(filepath.Join(dir, "docs/api_laravel_gateway.atom.md"))
	if err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(string(content), "status: DRAFT", "status: STABLE", 1)
	if err := os.WriteFile(filepath.Join(dir, "docs/api_laravel_gateway.atom.md"), []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", "docs/api_laravel_gateway.atom.md")

	out, err := gitIn(t, dir, "commit", "-q", "-m", "docs: unrelated status bump")
	if err != nil {
		t.Fatalf("expected commit to succeed for pre-existing orphan touched on an unrelated field, but it failed: %v\noutput:\n%s", err, out)
	}
	if !strings.Contains(out, "pre-existing orphan") {
		t.Errorf("expected hook output to mention the pre-existing orphan as a non-blocking warning, got:\n%s", out)
	}
}

// Case 2 (regression guard): a commit newly adds an orphaned
// ARCHITECTURE/IMPLEMENTATION atom (no parents). The hook must still block
// it -- the fix must not over-relax and let brand-new orphans through.
func TestPreCommitHook_NewlyAddedOrphan_StillBlocks(t *testing.T) {
	dir := initHookTestRepo(t)

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# repo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := gitIn(t, dir, "add", "README.md"); err != nil {
		t.Fatal(err)
	}
	if out, err := gitIn(t, dir, "commit", "-q", "-m", "initial commit"); err != nil {
		t.Fatalf("initial commit failed: %v\n%s", err, out)
	}

	writeHookFixtureAtom(t, dir, "docs/arch_new.atom.md", "arch_new", "ARCHITECTURE", nil)
	if _, err := gitIn(t, dir, "add", "docs/arch_new.atom.md"); err != nil {
		t.Fatal(err)
	}

	out, err := gitIn(t, dir, "commit", "-q", "-m", "docs: add new orphaned architecture atom")
	if err == nil {
		t.Fatalf("expected commit to be blocked for a newly added orphaned atom, but it succeeded\noutput:\n%s", out)
	}
	if !strings.Contains(out, "Orphaned Atom Detected") {
		t.Errorf("expected hook output to report an orphaned atom, got:\n%s", out)
	}
	if !strings.Contains(out, "newly added") {
		t.Errorf("expected hook output to identify this as a newly added atom, got:\n%s", out)
	}
}

// Case 3 (regression guard): a commit removes the last remaining parent
// from a previously-non-orphaned ARCHITECTURE/IMPLEMENTATION atom. The hook
// must still block it -- this commit is the one introducing the
// regression, so it correctly stays a blocking condition.
func TestPreCommitHook_RemovesLastParent_StillBlocks(t *testing.T) {
	dir := initHookTestRepo(t)

	writeHookFixtureAtom(t, dir, "docs/arch_parented.atom.md", "arch_parented", "ARCHITECTURE", []string{"[[req_something]]"})
	if _, err := gitIn(t, dir, "add", "docs/arch_parented.atom.md"); err != nil {
		t.Fatal(err)
	}
	if out, err := gitIn(t, dir, "commit", "-q", "-m", "initial commit: atom with a parent"); err != nil {
		t.Fatalf("initial commit failed: %v\n%s", err, out)
	}

	// Now strip its only parent.
	writeHookFixtureAtom(t, dir, "docs/arch_parented.atom.md", "arch_parented", "ARCHITECTURE", nil)
	if _, err := gitIn(t, dir, "add", "docs/arch_parented.atom.md"); err != nil {
		t.Fatal(err)
	}

	out, err := gitIn(t, dir, "commit", "-q", "-m", "docs: accidentally drop the only parent")
	if err == nil {
		t.Fatalf("expected commit to be blocked for removing the last parent, but it succeeded\noutput:\n%s", out)
	}
	if !strings.Contains(out, "Orphaned Atom Detected") {
		t.Errorf("expected hook output to report an orphaned atom, got:\n%s", out)
	}
	if !strings.Contains(out, "last parent was just removed") {
		t.Errorf("expected hook output to identify this as a last-parent removal regression, got:\n%s", out)
	}
}

// Case 4 (optional coverage): a fresh repo with no prior HEAD (the very
// first commit ever) that adds an orphaned atom. HEAD_EXISTS is false in
// this case, and the hook must still treat the atom as "newly added" (not
// special-case away the very first commit) and block it.
func TestPreCommitHook_FirstCommitEverWithOrphan_StillBlocks(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
			"HOME="+dir,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")
	run("config", "commit.gpgsign", "false")

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte(preCommitHookContent), 0755); err != nil {
		t.Fatalf("failed to write pre-commit hook: %v", err)
	}

	writeHookFixtureAtom(t, dir, "docs/arch_first.atom.md", "arch_first", "ARCHITECTURE", nil)
	run("add", "docs/arch_first.atom.md")

	out, err := gitIn(t, dir, "commit", "-q", "-m", "very first commit: orphaned atom")
	if err == nil {
		t.Fatalf("expected the very first commit to be blocked for an orphaned atom, but it succeeded\noutput:\n%s", out)
	}
	if !strings.Contains(out, "Orphaned Atom Detected") {
		t.Errorf("expected hook output to report an orphaned atom, got:\n%s", out)
	}
}
