package atom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/testutil"
)

// TestUpdateLinksSubstringIDNotCorrupted extends the WP-0 UpdateLinks
// scope-guard coverage in update_scope_test.go with a cheap edge case that
// file didn't cover: renaming an id that is a strict text substring of
// another, unrelated atom's id must not corrupt the other atom's links.
// e.g. renaming "req_a" -> "req_a_v2" must never touch "[[req_a_extra]]",
// because both the plain-string check (oldLink = "[[req_a]]") and the regex
// (`\[\[req_a\]\]`) require the closing "]]" to immediately follow the old
// id — "req_a_extra]]" has "_extra]]" in that position, not "]]", so neither
// matcher fires. This only exercises the docs-only walk (no project-root
// scope guard needed), so it needs none of update_scope_test.go's
// config.Snapshot/Restore machinery.
func TestUpdateLinksSubstringIDNotCorrupted(t *testing.T) {
	t.Parallel()
	// UpdateLinks reads config.ActiveConfig (for the source-file propagation
	// walk's scope guard); snapshot/restore it so a leaked or concurrently
	// loaded config from another t.Parallel() test can't affect this run.
	testutil.SnapshotConfigLocked(t)

	docsDir := t.TempDir()

	targetFile := filepath.Join(docsDir, "req_a.atom.md")
	targetContent := "---\nid: req_a\nparents:\n  - [[req_root]]\n---\nSee [[req_a]] for details."
	if err := os.WriteFile(targetFile, []byte(targetContent), 0644); err != nil {
		t.Fatal(err)
	}

	// A second atom whose id has req_a as a strict prefix — the classic
	// substring-collision shape.
	supersetFile := filepath.Join(docsDir, "req_a_extra.atom.md")
	supersetContent := "---\nid: req_a_extra\nparents:\n  - [[req_a_extra]]\n---\nRefers to itself: [[req_a_extra]]."
	if err := os.WriteFile(supersetFile, []byte(supersetContent), 0644); err != nil {
		t.Fatal(err)
	}

	numUpdates := UpdateLinks(docsDir, "req_a", "req_a_v2")

	updatedTarget, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(updatedTarget), "[[req_a_v2]]") {
		t.Errorf("expected the exact-match ref to be renamed to [[req_a_v2]], got:\n%s", updatedTarget)
	}
	if strings.Contains(string(updatedTarget), "[[req_a]]") {
		t.Errorf("expected the old [[req_a]] ref to be gone, got:\n%s", updatedTarget)
	}

	untouchedSuperset, err := os.ReadFile(supersetFile)
	if err != nil {
		t.Fatal(err)
	}
	want := supersetContent
	if string(untouchedSuperset) != want {
		t.Errorf("substring collision corrupted an unrelated atom's links:\nwant:\n%s\ngot:\n%s", want, untouchedSuperset)
	}
	if strings.Contains(string(untouchedSuperset), "req_a_v2_extra") || strings.Contains(string(untouchedSuperset), "[[req_a_v2]]") {
		t.Errorf("rename bled into the superset id: %s", untouchedSuperset)
	}

	if numUpdates != 1 {
		t.Errorf("expected exactly 1 file updated (req_a.atom.md only), got %d", numUpdates)
	}
}
