package atom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/testutil"
)

// TestUpdateIDAbbreviatedTypePrefixIsNotDoublePrefixed is a fixed-defect
// regression test for
// failures/20260917_atd_update_double_prefixes_id_that_already_starts_with_type_abbreviation.md.
//
// Previously, Update's naming-convention enforcement (pkg/atom/update.go,
// the `if newID != "" && atomType != ""` block) only checked whether newID
// already had the *exact full* type prefix (e.g. "mechanic_"). If the
// caller's id already carried a partial/abbreviated form of the type name
// that wasn't the exact prefix (e.g. "mech_economy_purge" against
// type=MECHANIC), the check failed and the code blindly prepended the full
// prefix onto the *entire* existing string, producing a non-conformant,
// double-prefixed id ("mechanic_mech_economy_purge").
//
// Fixed: Update now detects when an id's leading snake_case segment is
// itself a case-insensitive prefix of the full type name (e.g. "mech" of
// "mechanic") and treats it as an intentional, already-applied abbreviation
// — leaving the id as given instead of prepending the full prefix on top of
// it. This avoids inventing a new id that would orphan already-written
// @spec-link/@test-link tags (the real incident harm: 16 dangling tags from
// one bad rename).
func TestUpdateIDAbbreviatedTypePrefixIsNotDoublePrefixed(t *testing.T) {
	testutil.Guard(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "mech_economy_purge.atom.md")

	msg, err := Update(UpdateOptions{
		FilePath: path,
		SetArgs: []string{
			"id=mech_economy_purge",
			"type=MECHANIC",
			"layer=IMPLEMENTATION",
			"status=DRAFT",
		},
		Intent: "Purges economy state.",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	wantID := "mech_economy_purge"
	if !strings.Contains(msg, "id="+wantID) {
		t.Errorf("Update should leave the abbreviated id untouched — got message %q, want id=%s", msg, wantID)
	}

	cleanPath := filepath.Join(dir, wantID+".atom.md")
	if _, statErr := os.Stat(cleanPath); statErr != nil {
		t.Errorf("expected the untouched file %s to exist, got stat error: %v", cleanPath, statErr)
	}

	badPath := filepath.Join(dir, "mechanic_mech_economy_purge.atom.md")
	if _, statErr := os.Stat(badPath); statErr == nil {
		t.Errorf("the double-prefixed file %s must not exist", badPath)
	}
}

// TestUpdateReapplyingAbbreviatedIDIsIdempotent is a fixed-defect regression
// test for the second half of the same field report: a follow-up Update
// call that re-supplies the caller's originally-intended short id (without
// also re-supplying type — matching the report's exact repro, where type is
// read back from the file's own on-disk frontmatter) must recover/keep the
// intended id rather than mangling it further.
//
// Previously this re-ran the same buggy prefix logic against already-bad
// input and deterministically regenerated the identical double-prefixed id,
// so the file was rewritten in place with no visible change while still
// reporting "Success" — a "my fix attempt silently did nothing" failure
// mode.
//
// Fixed: since the first Update call no longer double-prefixes, the file
// already sits at the clean abbreviated id; a follow-up call re-supplying
// the same id is a true no-op/idempotent rename that keeps that id.
func TestUpdateReapplyingAbbreviatedIDIsIdempotent(t *testing.T) {
	testutil.Guard(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "mech_economy_purge.atom.md")

	_, err := Update(UpdateOptions{
		FilePath: path,
		SetArgs:  []string{"id=mech_economy_purge", "type=MECHANIC", "layer=IMPLEMENTATION", "status=DRAFT"},
	})
	if err != nil {
		t.Fatalf("first Update failed: %v", err)
	}
	wantPath := filepath.Join(dir, "mech_economy_purge.atom.md")
	if _, statErr := os.Stat(wantPath); statErr != nil {
		t.Fatalf("setup: expected the clean-id file to exist after the first Update: %v", statErr)
	}

	// Follow-up call re-supplying the same id (type read back from frontmatter).
	msg, err := Update(UpdateOptions{
		FilePath: wantPath,
		SetArgs:  []string{"id=mech_economy_purge"},
	})
	if err != nil {
		t.Fatalf("follow-up Update failed: %v", err)
	}

	if !strings.Contains(msg, "Success") {
		t.Errorf("follow-up Update should report Success — got %q", msg)
	}
	if !strings.Contains(msg, "id=mech_economy_purge") {
		t.Errorf("follow-up Update should keep reporting the clean id — got message %q", msg)
	}

	badPath := filepath.Join(dir, "mechanic_mech_economy_purge.atom.md")
	if _, statErr := os.Stat(badPath); statErr == nil {
		t.Errorf("the double-prefixed file %s must not exist", badPath)
	}
}

// TestUpdateExpectationAppendedWhenHeaderMissing is a fixed-defect
// regression test for
// failures/20260917_atd_update_expectation_no_ops_when_section_header_missing.md.
//
// Previously, Update's body-section rewrite (pkg/atom/update.go, the
// isTargetHeader loop over the file's existing body lines) only ever
// replaced content under a `## EXPECTATION` header it found while iterating
// the current file. If no such header existed in the file at all, that
// branch never fired and there was no fallback that appended a new
// `## EXPECTATION` header + the supplied text. Worse, the success log
// message's "| updated body sections" suffix was gated only on
// `expectationText != ""` (did the caller pass --expectation), not on
// whether a matching header was actually found and rewritten — so the CLI
// reported success and claimed to have updated body sections even when the
// write was a complete no-op for that field.
//
// Fixed: when a supplied section's header (INTENT / THE RULE / LOGIC /
// TECHNICAL INTERFACE / EXPECTATION) has no match anywhere in the file, a
// new `## <SECTION>` header is appended with the supplied text, and the
// success message's "| updated body sections" suffix is now gated on a
// section actually being found-and-updated or freshly appended.
func TestUpdateExpectationAppendedWhenHeaderMissing(t *testing.T) {
	testutil.Guard(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "ui_registration.atom.md")

	original := "---\n" +
		"id: ui_registration\n" +
		"human_name: Registration Form\n" +
		"type: UI\n" +
		"layer: ARCHITECTURE\n" +
		"version: 1.0\n" +
		"status: STABLE\n" +
		"priority: CORE\n" +
		"tags: []\n" +
		"parents:\n" +
		"  - [[req_registration]]\n" +
		"dependents: []\n" +
		"---\n" +
		"\n" +
		"# Registration Form\n" +
		"\n" +
		"## INTENT\n" +
		"Let a new user create an account.\n" +
		"\n" +
		"## THE RULE / LOGIC\n" +
		"Account name, email, password, and confirmation are mandatory.\n" +
		"\n" +
		"## TECHNICAL INTERFACE\n" +
		"- **Code Tag:** `@spec-link [[ui_registration]]`\n"

	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	expectationText := "- Submitting with only the mandatory fields succeeds."
	msg, err := Update(UpdateOptions{
		FilePath:    path,
		Expectation: expectationText,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if !strings.Contains(msg, "updated body sections") {
		t.Errorf("Update should report 'updated body sections' after appending the missing header — got message %q", msg)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file after update: %v", err)
	}
	gotStr := string(got)

	if !strings.Contains(gotStr, "## EXPECTATION") {
		t.Errorf("Update should insert a missing '## EXPECTATION' header — got file:\n%s", gotStr)
	}
	if !strings.Contains(gotStr, expectationText) {
		t.Errorf("Update should persist the --expectation text — got file:\n%s", gotStr)
	}
}
