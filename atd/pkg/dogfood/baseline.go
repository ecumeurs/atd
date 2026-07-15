// Package dogfood holds the shared plumbing for WP-3's dogfood gate
// (test_atd_07_26.md §3.4, §6): running ATD's own tools (`atd lint`, the S6
// corpus parse audit) against the REAL docs/ corpus at the repo root and
// ratcheting the result against a checked-in baseline, so pre-existing
// authoring debt never fails the gate but any newly introduced finding does.
//
// This package is imported from `dogfood`-build-tagged tests in BOTH Go
// modules (atd-tools' pkg/atom and the nested cmd/atd's cmd package), which
// is why the repo-root/docs-dir discovery below anchors itself to this
// source file's own fixed location (via runtime.Caller) rather than to the
// process's cwd or either module's go.mod: a `go test` binary's cwd is its
// package directory, which differs between the two callers, but this file's
// path relative to the repo root never does.
package dogfood

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// RepoRoot walks up from this source file's own directory until it finds the
// checked-out repo root: the directory containing both a docs/ folder and a
// .atd config file (the same two markers `atd` itself uses to recognize a
// project). Returns an error if no such ancestor exists, e.g. if this file
// were vendored somewhere without the rest of the tree.
func RepoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("dogfood: could not determine this source file's location")
	}
	dir := filepath.Dir(file)
	for {
		if looksLikeRepoRoot(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("dogfood: could not locate repo root (docs/ + .atd) above %s", file)
		}
		dir = parent
	}
}

func looksLikeRepoRoot(dir string) bool {
	docsInfo, err := os.Stat(filepath.Join(dir, "docs"))
	if err != nil || !docsInfo.IsDir() {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, ".atd"))
	return err == nil
}

// DocsDir returns the real docs/ corpus directory at the repo root — the
// dogfood gate's subject, as opposed to any fixture or sandboxed docs dir
// used by the rest of the suite.
func DocsDir() (string, error) {
	root, err := RepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "docs"), nil
}

// BaselinePath returns the checked-in ratchet baseline file for the dogfood
// gate.
func BaselinePath() (string, error) {
	root, err := RepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "atd", "testdata", "dogfood_baseline.txt"), nil
}

// LoadBaseline reads the checked-in fingerprint baseline: one normalized
// finding per line, blank lines and "#"-comments ignored. A missing file
// reads as an empty baseline (i.e. every current finding is a regression
// until the gate is run once and the file is created).
func LoadBaseline(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[line] = true
	}
	return out, nil
}

// FilterPrefix returns the subset of fingerprints in m that start with
// prefix. Each dogfood test owns one fingerprint category (e.g. "LINT\t" or
// "S6_EMPTY_INTENT\t") and must only compare/update its own slice of the
// shared baseline file.
func FilterPrefix(m map[string]bool, prefix string) map[string]bool {
	out := make(map[string]bool)
	for k := range m {
		if strings.HasPrefix(k, prefix) {
			out[k] = true
		}
	}
	return out
}

// Diff reports which fingerprints in current are new relative to baseline
// (regressions — must fail the gate) and which baseline fingerprints no
// longer reproduce (fixed debt — a ratchet-down candidate, never required,
// never automatic).
func Diff(current, baseline map[string]bool) (newOnes, fixed []string) {
	for fp := range current {
		if !baseline[fp] {
			newOnes = append(newOnes, fp)
		}
	}
	for fp := range baseline {
		if !current[fp] {
			fixed = append(fixed, fp)
		}
	}
	sort.Strings(newOnes)
	sort.Strings(fixed)
	return
}

// Keys returns the sorted keys of a fingerprint set.
func Keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// UpdateBaseline rewrites the baseline file, replacing every line carrying
// the given prefix with currentForPrefix and leaving every other category's
// lines untouched. This is the ratchet's only sanctioned write path: the
// baseline for one category always ends up mirroring current reality after
// this call, whether that means fewer lines (debt fixed) or more (new debt
// consciously accepted). It never runs implicitly — only via
// ATD_DOGFOOD_UPDATE=1 (see `make dogfood-update-baseline`), so accepting new
// debt is always a deliberate, reviewable action, never a silent side effect
// of a normal test run.
func UpdateBaseline(path, prefix string, currentForPrefix []string) error {
	existing, err := LoadBaseline(path)
	if err != nil {
		return err
	}
	kept := make(map[string]bool)
	for fp := range existing {
		if !strings.HasPrefix(fp, prefix) {
			kept[fp] = true
		}
	}
	for _, fp := range currentForPrefix {
		kept[fp] = true
	}
	all := Keys(kept)

	var sb strings.Builder
	sb.WriteString("# ATD dogfood gate ratchet baseline (test_atd_07_26.md §3.4, WP-3).\n")
	sb.WriteString("#\n")
	sb.WriteString("# One normalized finding per line: PREFIX<TAB>id<TAB>detail. Pre-existing\n")
	sb.WriteString("# authoring debt lives here so it never fails `make dogfood`; any finding NOT\n")
	sb.WriteString("# in this file is a regression and fails the gate. The ratchet only loosens\n")
	sb.WriteString("# (a line disappears) when the underlying debt is actually fixed, and only\n")
	sb.WriteString("# tightens deliberately via `make dogfood-update-baseline`\n")
	sb.WriteString("# (ATD_DOGFOOD_UPDATE=1) — never hand-edit this file to hide a new failure.\n")
	for _, fp := range all {
		sb.WriteString(fp)
		sb.WriteString("\n")
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}
