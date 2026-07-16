package atom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/testutil"
)

// writeAtomFile writes a minimal but complete .atom.md file with the given
// id/status/layer under t.TempDir() and returns its path.
func writeAtomFile(t *testing.T, id, status, layer string) string {
	t.Helper()
	content := `---
id: ` + id + `
human_name: Guard Test Atom
type: REQUIREMENT
layer: ` + layer + `
version: 1.0
status: ` + status + `
priority: CORE
tags: []
parents: []
dependents: []
---

# Guard Test Atom

## INTENT
Original intent.

## THE RULE / LOGIC
Original logic.

## TECHNICAL INTERFACE
Original interface.

## EXPECTATION
Original expectation.
`
	path := filepath.Join(t.TempDir(), id+".atom.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing fixture atom: %v", err)
	}
	return path
}

// TestUpdateForceGuard pins the STABLE+BUSINESS governance guard in Update
// (pkg/atom/update.go:94-108, test_atd_07_26.md §3.1 item 3): modifying an
// existing atom whose CURRENT on-disk state is STABLE+BUSINESS must be
// refused unless Force is set; every other combination of status/layer is
// unaffected.
func TestUpdateForceGuard(t *testing.T) {
	t.Parallel()

	t.Run("DRAFT BUSINESS accepted without force", func(t *testing.T) {
		t.Parallel()
		testutil.Guard(t)
		path := writeAtomFile(t, "req_draft_biz", "DRAFT", "BUSINESS")

		_, err := Update(UpdateOptions{
			FilePath: path,
			Logic:    "Updated logic.",
		})
		if err != nil {
			t.Fatalf("expected DRAFT+BUSINESS update to succeed, got error: %v", err)
		}
		data, perr := Parse(path)
		if perr != nil {
			t.Fatalf("Parse after update failed: %v", perr)
		}
		if data.Logic != "Updated logic." {
			t.Errorf("expected logic to be updated, got %q", data.Logic)
		}
	})

	t.Run("STABLE BUSINESS refused without force", func(t *testing.T) {
		t.Parallel()
		testutil.Guard(t)
		path := writeAtomFile(t, "req_stable_biz", "STABLE", "BUSINESS")

		_, err := Update(UpdateOptions{
			FilePath: path,
			Logic:    "Attempted change.",
		})
		if err == nil {
			t.Fatal("expected STABLE+BUSINESS update without Force to be refused, got nil error")
		}
		if !strings.Contains(err.Error(), "req_stable_biz") {
			t.Errorf("expected error to name the atom id, got: %v", err)
		}
		if !strings.Contains(strings.ToLower(err.Error()), "force") {
			t.Errorf("expected error to mention --force/force:true override, got: %v", err)
		}

		// The file on disk must be untouched — a refused update is a no-op.
		data, perr := Parse(path)
		if perr != nil {
			t.Fatalf("Parse after refused update failed: %v", perr)
		}
		if data.Logic != "Original logic." {
			t.Errorf("expected file untouched by refused update, got Logic=%q", data.Logic)
		}
	})

	t.Run("STABLE BUSINESS accepted with Force true", func(t *testing.T) {
		t.Parallel()
		testutil.Guard(t)
		path := writeAtomFile(t, "req_stable_biz_forced", "STABLE", "BUSINESS")

		_, err := Update(UpdateOptions{
			FilePath: path,
			Logic:    "Forced change.",
			Force:    true,
		})
		if err != nil {
			t.Fatalf("expected Force:true to override the STABLE+BUSINESS guard, got error: %v", err)
		}
		data, perr := Parse(path)
		if perr != nil {
			t.Fatalf("Parse after forced update failed: %v", perr)
		}
		if data.Logic != "Forced change." {
			t.Errorf("expected logic to be updated under Force, got %q", data.Logic)
		}
	})

	t.Run("STABLE non-BUSINESS layer accepted without force", func(t *testing.T) {
		t.Parallel()
		testutil.Guard(t)
		path := writeAtomFile(t, "mech_stable_impl", "STABLE", "IMPLEMENTATION")

		_, err := Update(UpdateOptions{
			FilePath: path,
			Logic:    "Updated impl logic.",
		})
		if err != nil {
			t.Fatalf("expected STABLE+IMPLEMENTATION (non-BUSINESS) to be unaffected by the guard, got error: %v", err)
		}
	})

	t.Run("REVIEW BUSINESS accepted without force", func(t *testing.T) {
		t.Parallel()
		testutil.Guard(t)
		path := writeAtomFile(t, "req_review_biz", "REVIEW", "BUSINESS")

		_, err := Update(UpdateOptions{
			FilePath: path,
			Logic:    "Updated review logic.",
		})
		if err != nil {
			t.Fatalf("expected REVIEW+BUSINESS (non-STABLE) to be unaffected by the guard, got error: %v", err)
		}
	})

	t.Run("new file creation ignores force guard entirely", func(t *testing.T) {
		t.Parallel()
		testutil.Guard(t)
		path := filepath.Join(t.TempDir(), "req_brand_new.atom.md")

		// type is deliberately omitted from SetArgs: setting it would trigger
		// Update's separate naming-convention enforcement (new id must be
		// prefixed with the lowercased type), which is unrelated to the
		// force guard this test targets and would rename req_brand_new to
		// requirement_req_brand_new before this test could observe it.
		_, err := Update(UpdateOptions{
			FilePath: path,
			SetArgs:  []string{"id=req_brand_new", "layer=BUSINESS", "status=STABLE"},
			Intent:   "Brand new atom, created directly as STABLE+BUSINESS.",
		})
		if err != nil {
			t.Fatalf("expected creation of a nonexistent file to bypass the guard (fileExists=false), got error: %v", err)
		}
		if _, statErr := os.Stat(path); statErr != nil {
			t.Errorf("expected new atom file to be created at %s: %v", path, statErr)
		}
	})

	t.Run("guard is case-insensitive on status and layer", func(t *testing.T) {
		t.Parallel()
		testutil.Guard(t)
		path := writeAtomFile(t, "req_lowercase_biz", "stable", "business")

		_, err := Update(UpdateOptions{
			FilePath: path,
			Logic:    "Attempted change.",
		})
		if err == nil {
			t.Fatal("expected lowercase 'stable'/'business' to still trip the guard (EqualFold), got nil error")
		}
	})
}
