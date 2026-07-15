//go:build dogfood

// Dogfood lint ratchet (test_atd_07_26.md §3.4, §6 WP-3): runs the exact same
// `runLint` the `atd lint` command uses against the REAL docs/ corpus at the
// repo root, and fails only on findings that are NOT already recorded in the
// checked-in baseline (atd/testdata/dogfood_baseline.txt). Pre-existing
// authoring debt (missing EXPECTATION sections, unresolved fixture links,
// etc.) is baselined and never fails the gate; a freshly introduced error —
// e.g. a scratch atom with a non-canonical `type: FOO` — is not in the
// baseline and does fail it.
//
// Build-tagged `dogfood` because it reads the live repo tree instead of a
// sandbox: it must never run in the default `go test ./...` lane. Invoke via
// `make -C atd dogfood` (full gate) or `make -C atd dogfood-quick` (this test
// alone, wired into the pre-commit hook for atom-touching commits).
package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/dogfood"
)

const lintFingerprintPrefix = "LINT\t"

func TestDogfoodLintRatchet(t *testing.T) {
	repoRoot, err := dogfood.RepoRoot()
	if err != nil {
		t.Fatalf("dogfood: %v", err)
	}
	docsDir, err := dogfood.DocsDir()
	if err != nil {
		t.Fatalf("dogfood: %v", err)
	}
	baselinePath, err := dogfood.BaselinePath()
	if err != nil {
		t.Fatalf("dogfood: %v", err)
	}

	// runLint resolves parent/dependent links via config.ProjectRoot()/
	// config.DocsDir() (the *active* config), not the dir argument alone, so
	// this must load the repo's real .atd first — exactly what the `atd`
	// binary does when invoked from the repo root. Per the WP-0 sandbox
	// contract (test_atd_07_26.md §3.6, T-1), snapshot and restore
	// config.ActiveConfig so this doesn't leak into other tests sharing the
	// binary when `go test -tags dogfood ./...` runs the whole package.
	saved := config.Snapshot()
	t.Cleanup(func() { config.Restore(saved) })
	if err := config.LoadFromDir(repoRoot); err != nil {
		t.Fatalf("dogfood: failed to load repo config at %s: %v", repoRoot, err)
	}

	report, lintErr := runLint(docsDir)
	current := parseLintFingerprints(report)

	if os.Getenv("ATD_DOGFOOD_UPDATE") == "1" {
		if err := dogfood.UpdateBaseline(baselinePath, lintFingerprintPrefix, dogfood.Keys(current)); err != nil {
			t.Fatalf("dogfood: failed to update baseline: %v", err)
		}
		t.Logf("dogfood: baseline updated with %d lint fingerprint(s)", len(current))
		return
	}

	baseline, err := dogfood.LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("dogfood: failed to load baseline %s: %v", baselinePath, err)
	}
	baselineLint := dogfood.FilterPrefix(baseline, lintFingerprintPrefix)

	newOnes, fixed := dogfood.Diff(current, baselineLint)

	if len(fixed) > 0 {
		t.Logf("dogfood: %d previously-baselined lint finding(s) no longer reproduce (ratchet-down candidate; run 'make -C atd dogfood-update-baseline' to shrink the baseline):", len(fixed))
		for _, fp := range fixed {
			t.Logf("  fixed: %s", strings.TrimPrefix(fp, lintFingerprintPrefix))
		}
	}

	if len(newOnes) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%d NEW lint finding(s) not present in the checked-in baseline (%s):\n", len(newOnes), baselinePath)
		for _, fp := range newOnes {
			fmt.Fprintf(&b, "  + %s\n", strings.TrimPrefix(fp, lintFingerprintPrefix))
		}
		b.WriteString("Fix these, or if genuinely pre-existing/accepted debt, run\n")
		b.WriteString("'make -C atd dogfood-update-baseline' to consciously baseline them.\n")
		t.Error(b.String())
	}

	_ = lintErr // runLint's error is redundant with len(current) > 0; the ratchet is what gates.
}

// parseLintFingerprints turns runLint's "[atomID]\n  - msg\n..." report text
// into the "LINT\tatomID\tmessage" fingerprints the baseline stores. Kept
// deliberately dumb (line-prefix based) so it stays in lockstep with
// runLint's own rendering in lint.go without needing runLint to expose
// structured data.
func parseLintFingerprints(report string) map[string]bool {
	out := make(map[string]bool)
	var currentID string
	for _, line := range strings.Split(report, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			currentID = strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]")
			continue
		}
		if strings.HasPrefix(line, "  - ") && currentID != "" {
			msg := strings.TrimPrefix(line, "  - ")
			out[lintFingerprintPrefix+currentID+"\t"+msg] = true
		}
	}
	return out
}
