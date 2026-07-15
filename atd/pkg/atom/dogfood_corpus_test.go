//go:build dogfood

// S6 — corpus parse audit (test_atd_07_26.md §4 S6, §6 WP-3): the scenario
// that would have caught E4, the H3-swallow bug (report §2, incident row
// E4) — real atoms in docs/ use "### Description"-style subheadings inside
// "## TECHNICAL INTERFACE"/"## EXPECTATION", and a parser that only
// recognizes the first paragraph of a section silently starves
// `check --semantic`, `trace --summary`, and `assemble` without ever
// failing a synthetic unit test built from minimal fixtures.
//
// For every *.atom.md in the REAL docs/ corpus at the repo root:
//  1. atom.Parse yields a non-empty Intent.
//  2. If the raw file contains a "## TECHNICAL INTERFACE" H2 with at least
//     one non-blank body line before the next H2, the parsed Interface is
//     non-empty (the H3-swallow invariant).
//
// Pre-existing authoring debt is folded into the same ratcheted baseline the
// lint gate uses (atd/testdata/dogfood_baseline.txt) rather than weakening
// either assertion — see UpdateBaseline's doc comment in pkg/dogfood.
//
// Build-tagged `dogfood`: reads the live repo docs/ tree, must never run in
// the default `go test ./...` lane. Invoke via `make -C atd dogfood`.
package atom

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"atd-tools/pkg/dogfood"
)

const (
	s6EmptyIntentPrefix        = "S6_EMPTY_INTENT\t"
	s6SwallowedInterfacePrefix = "S6_SWALLOWED_INTERFACE\t"
)

var technicalInterfaceH2 = regexp.MustCompile(`^##\s+TECHNICAL INTERFACE\b`)

func TestDogfoodCorpusParseAudit(t *testing.T) {
	docsDir, err := dogfood.DocsDir()
	if err != nil {
		t.Fatalf("dogfood: %v", err)
	}
	baselinePath, err := dogfood.BaselinePath()
	if err != nil {
		t.Fatalf("dogfood: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(docsDir, "*.atom.md"))
	if err != nil {
		t.Fatalf("dogfood: failed to scan %s: %v", docsDir, err)
	}
	if len(files) == 0 {
		t.Fatalf("dogfood: no *.atom.md files found under %s — is the repo root detection broken?", docsDir)
	}

	emptyIntent := make(map[string]bool)
	swallowedInterface := make(map[string]bool)

	for _, f := range files {
		id := strings.TrimSuffix(filepath.Base(f), ".atom.md")

		a, parseErr := Parse(f)
		if parseErr != nil {
			t.Errorf("dogfood S6: atom.Parse(%s) returned an error (not an authoring-debt case — a real parse failure): %v", f, parseErr)
			continue
		}
		if a.ID != "" {
			id = a.ID
		}

		if strings.TrimSpace(a.Intent) == "" {
			emptyIntent[s6EmptyIntentPrefix+id] = true
		}

		raw, readErr := os.ReadFile(f)
		if readErr != nil {
			t.Errorf("dogfood S6: failed to read %s for raw-vs-parsed comparison: %v", f, readErr)
			continue
		}
		if fileHasNonBlankTechnicalInterface(string(raw)) && strings.TrimSpace(a.Interface) == "" {
			swallowedInterface[s6SwallowedInterfacePrefix+id] = true
		}
	}

	if os.Getenv("ATD_DOGFOOD_UPDATE") == "1" {
		if err := dogfood.UpdateBaseline(baselinePath, s6EmptyIntentPrefix, dogfood.Keys(emptyIntent)); err != nil {
			t.Fatalf("dogfood: failed to update baseline (empty intent): %v", err)
		}
		if err := dogfood.UpdateBaseline(baselinePath, s6SwallowedInterfacePrefix, dogfood.Keys(swallowedInterface)); err != nil {
			t.Fatalf("dogfood: failed to update baseline (swallowed interface): %v", err)
		}
		t.Logf("dogfood: baseline updated — %d empty-intent, %d swallowed-interface", len(emptyIntent), len(swallowedInterface))
		return
	}

	baseline, err := dogfood.LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("dogfood: failed to load baseline %s: %v", baselinePath, err)
	}

	checkCategory(t, "empty INTENT", emptyIntent, dogfood.FilterPrefix(baseline, s6EmptyIntentPrefix), s6EmptyIntentPrefix, baselinePath)
	checkCategory(t, "H3-swallowed TECHNICAL INTERFACE", swallowedInterface, dogfood.FilterPrefix(baseline, s6SwallowedInterfacePrefix), s6SwallowedInterfacePrefix, baselinePath)

	t.Logf("dogfood S6: audited %d atoms — %d empty-intent, %d swallowed-interface (baseline: %d, %d)",
		len(files), len(emptyIntent), len(swallowedInterface),
		len(dogfood.FilterPrefix(baseline, s6EmptyIntentPrefix)), len(dogfood.FilterPrefix(baseline, s6SwallowedInterfacePrefix)))
}

func checkCategory(t *testing.T, label string, current, baseline map[string]bool, prefix, baselinePath string) {
	t.Helper()
	newOnes, fixed := dogfood.Diff(current, baseline)

	if len(fixed) > 0 {
		t.Logf("dogfood S6 (%s): %d baselined finding(s) no longer reproduce (ratchet-down candidate):", label, len(fixed))
		for _, fp := range fixed {
			t.Logf("  fixed: %s", strings.TrimPrefix(fp, prefix))
		}
	}

	if len(newOnes) > 0 {
		var b strings.Builder
		fmt.Fprintf(&b, "%d NEW %s finding(s) not present in the checked-in baseline (%s):\n", len(newOnes), label, baselinePath)
		for _, fp := range newOnes {
			fmt.Fprintf(&b, "  + %s\n", strings.TrimPrefix(fp, prefix))
		}
		b.WriteString("Fix the atom, or if genuinely pre-existing/accepted debt, run\n")
		b.WriteString("'make -C atd dogfood-update-baseline' to consciously baseline it.\n")
		t.Error(b.String())
	}
}

// fileHasNonBlankTechnicalInterface reports whether raw contains a
// "## TECHNICAL INTERFACE" H2 followed by at least one non-blank line before
// the next H2 (or EOF). This mirrors what a human editing the atom would
// call "the section has content", independent of how Parse chose to bucket
// it — which is exactly the invariant the H3-swallow bug (E4) violated.
func fileHasNonBlankTechnicalInterface(raw string) bool {
	inSection := false
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimSpace(line)
		if technicalInterfaceH2.MatchString(trimmed) {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(trimmed, "## ") {
				return false
			}
			if trimmed != "" {
				return true
			}
		}
	}
	return false
}
