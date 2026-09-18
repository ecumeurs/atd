package atom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/testutil"
)

// TestUpdateKnownDefect_IDDoublePrefixesAbbreviatedType pins the defect from
// failures/20260917_atd_update_double_prefixes_id_that_already_starts_with_type_abbreviation.md.
//
// KNOWN DEFECT: Update's naming-convention enforcement (pkg/atom/update.go,
// the `if newID != "" && atomType != ""` block) only checks whether newID
// already has the *exact full* type prefix (e.g. "mechanic_"). If the
// caller's id already carries a partial/abbreviated form of the type name
// that isn't the exact prefix (e.g. "mech_economy_purge" against
// type=MECHANIC), the check fails and the code blindly prepends the full
// prefix onto the *entire* existing string, instead of detecting and
// reconciling the abbreviation. This produces a non-conformant,
// double-prefixed id ("mechanic_mech_economy_purge") — itself a violation
// of the <type>_<slug> convention the normalizer exists to enforce.
//
// If this test starts failing because Update now produces a clean id
// (either by accepting "mech_" as a tolerated abbreviation, or by stripping
// it before re-prefixing to "mechanic_economy_purge"), treat that as the
// defect being fixed and relax/remove this pin.
func TestUpdateKnownDefect_IDDoublePrefixesAbbreviatedType(t *testing.T) {
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

	wantBadID := "mechanic_mech_economy_purge"
	if !strings.Contains(msg, "id="+wantBadID) {
		t.Errorf("KNOWN DEFECT expectation changed: Update no longer double-prefixes the abbreviated id — got message %q, want it to still report id=%s; if fixed intentionally (a clean id like mechanic_economy_purge), update/relax this pin", msg, wantBadID)
	}

	badPath := filepath.Join(dir, wantBadID+".atom.md")
	if _, statErr := os.Stat(badPath); statErr != nil {
		t.Errorf("KNOWN DEFECT expectation changed: expected the double-prefixed file %s to exist (pinning current buggy behavior), got stat error: %v", badPath, statErr)
	}

	cleanPath := filepath.Join(dir, "mechanic_economy_purge.atom.md")
	if _, statErr := os.Stat(cleanPath); statErr == nil {
		t.Errorf("KNOWN DEFECT expectation changed: the correctly-normalized file %s now exists — Update must have started reconciling the mech_ abbreviation instead of double-prefixing; if intentional, relax this pin", cleanPath)
	}
}

// TestUpdateKnownDefect_ReapplyingAbbreviatedIDReproducesDoublePrefix pins
// the second half of the same field report: a follow-up Update call that
// re-supplies the caller's originally-intended short id (without also
// re-supplying type — matching the report's exact repro, where type is read
// back from the file's own on-disk frontmatter, which is still MECHANIC
// from the first call) does NOT recover the intended id. It re-runs the same
// buggy prefix logic against the same input and deterministically
// regenerates the identical double-prefixed id, so the file is rewritten
// in place with no visible change and the call still reports "Success".
//
// This is not a distinct no-op code path — it is the same normalizer bug
// applied idempotently to already-bad input — but it is worth pinning
// separately because from the caller's side it presents as "my fix attempt
// silently did nothing", which is the more dangerous-looking symptom flagged
// in the field report.
//
// If this test starts failing because the second call reaches a clean id
// (mech_economy_purge or mechanic_economy_purge) instead of reproducing
// mechanic_mech_economy_purge, treat that as the defect being fixed and
// relax/remove this pin.
func TestUpdateKnownDefect_ReapplyingAbbreviatedIDReproducesDoublePrefix(t *testing.T) {
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
	badPath := filepath.Join(dir, "mechanic_mech_economy_purge.atom.md")
	if _, statErr := os.Stat(badPath); statErr != nil {
		t.Fatalf("setup: expected the double-prefixed file to exist before reproducing the follow-up defect: %v", statErr)
	}

	// Follow-up "fix" attempt: re-supply only the desired short id.
	msg, err := Update(UpdateOptions{
		FilePath: badPath,
		SetArgs:  []string{"id=mech_economy_purge"},
	})
	if err != nil {
		t.Fatalf("follow-up Update failed: %v", err)
	}

	if !strings.Contains(msg, "Success") {
		t.Errorf("KNOWN DEFECT expectation changed: follow-up Update no longer reports Success while failing to reach the requested id — got %q", msg)
	}
	if !strings.Contains(msg, "id=mechanic_mech_economy_purge") {
		t.Errorf("KNOWN DEFECT expectation changed: follow-up Update no longer reproduces the same double-prefixed id — got message %q; if the normalizer now reaches a clean id via legitimate --set id= input, this pin should be relaxed", msg)
	}

	// The file that would represent the caller's actually-intended id must
	// never appear — the "fix" attempt just regenerates the identical bad id.
	wantPath := filepath.Join(dir, "mech_economy_purge.atom.md")
	if _, statErr := os.Stat(wantPath); statErr == nil {
		t.Errorf("KNOWN DEFECT expectation changed: a file at the caller's originally-requested id %s now exists — the normalizer must have started honoring mech_ as a valid abbreviation; if intentional, relax this pin", wantPath)
	}
}

// TestUpdateKnownDefect_ExpectationNotAppendedWhenHeaderMissing pins the
// defect from
// failures/20260917_atd_update_expectation_no_ops_when_section_header_missing.md.
//
// KNOWN DEFECT: Update's body-section rewrite (pkg/atom/update.go, the
// isTargetHeader loop over the file's existing body lines) only ever
// replaces content under a `## EXPECTATION` header it finds while iterating
// the current file. If no such header exists in the file at all, that
// branch never fires and there is no fallback that appends a new
// `## EXPECTATION` header + the supplied text. Worse, the success log
// message's "| updated body sections" suffix is gated only on
// `expectationText != ""` (did the caller pass --expectation), not on
// whether a matching header was actually found and rewritten — so the CLI
// reports success and claims to have updated body sections even when the
// write was a complete no-op for that field.
//
// If this test starts failing because Update now inserts a missing
// `## EXPECTATION` header (or otherwise persists the supplied text
// somewhere in the file), treat that as the defect being fixed and
// relax/remove this pin.
func TestUpdateKnownDefect_ExpectationNotAppendedWhenHeaderMissing(t *testing.T) {
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
		t.Errorf("KNOWN DEFECT expectation changed: Update no longer reports 'updated body sections' for this call (message: %q) — if it now correctly gates that phrase on actually finding the section, this pin may need adjusting rather than removing", msg)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading file after update: %v", err)
	}
	gotStr := string(got)

	if strings.Contains(gotStr, "## EXPECTATION") {
		t.Errorf("KNOWN DEFECT expectation changed: Update now inserts a missing '## EXPECTATION' header — got file:\n%s\nif fixed intentionally, relax/remove this pin", gotStr)
	}
	if strings.Contains(gotStr, expectationText) {
		t.Errorf("KNOWN DEFECT expectation changed: Update now persists the --expectation text somewhere even without an existing header — got file:\n%s\nif fixed intentionally, relax/remove this pin", gotStr)
	}
}
