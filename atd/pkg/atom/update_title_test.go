package atom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeAtom(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readAtom(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// h1 returns the first level-1 heading of an atom body, or "" if absent.
func h1(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// @test-link [[service_atd_update]]
func TestUpdateNewAtomTitleFromHumanName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.atom.md")

	if _, err := Update(UpdateOptions{
		FilePath: path,
		SetArgs:  []string{"type=RULE", "human_name=Password Policy", "id=rule_password_policy"},
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got := readAtom(t, filepath.Join(dir, "rule_password_policy.atom.md"))
	if title := h1(got); title != "Password Policy" {
		t.Errorf("H1 = %q, want %q\n%s", title, "Password Policy", got)
	}
	if strings.Contains(got, "New Atom") {
		t.Errorf("placeholder title survived creation:\n%s", got)
	}
}

// A new atom created without human_name must still not keep the placeholder;
// it falls back to a humanized form of the id.
// @test-link [[service_atd_update]]
func TestUpdateNewAtomTitleFallsBackToID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mechanic_atd_weave.atom.md")

	if _, err := Update(UpdateOptions{FilePath: path, Intent: "x"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got := readAtom(t, path)
	if title := h1(got); title != "Mechanic Atd Weave" {
		t.Errorf("H1 = %q, want %q\n%s", title, "Mechanic Atd Weave", got)
	}
}

// An atom still carrying the placeholder from before this fix gets repaired on
// the next write.
// @test-link [[service_atd_update]]
func TestUpdateRepairsPlaceholderTitle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rule_legacy.atom.md")
	writeAtom(t, path, "---\nid: rule_legacy\nhuman_name: Legacy Rule\ntype: RULE\nstatus: DRAFT\n---\n\n# New Atom\n\n## INTENT\n\n")

	if _, err := Update(UpdateOptions{FilePath: path, Intent: "Some intent"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if title := h1(readAtom(t, path)); title != "Legacy Rule" {
		t.Errorf("H1 = %q, want %q", title, "Legacy Rule")
	}
}

// A hand-written title that deliberately diverges from human_name must survive
// an unrelated update.
// @test-link [[service_atd_update]]
func TestUpdatePreservesHandWrittenTitle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "api_serve_trace.atom.md")
	writeAtom(t, path, "---\nid: api_serve_trace\nhuman_name: \"MCP Tool: atd_trace\"\ntype: API\nstatus: DRAFT\n---\n\n# Trace, In Detail\n\n## INTENT\n\n")

	if _, err := Update(UpdateOptions{FilePath: path, Intent: "Some intent"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if title := h1(readAtom(t, path)); title != "Trace, In Detail" {
		t.Errorf("H1 = %q, want it preserved", title)
	}
}

// When the title tracked human_name, renaming human_name renames the title too.
// @test-link [[service_atd_update]]
func TestUpdateRetitlesWhenHumanNameChanges(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rule_thing.atom.md")
	writeAtom(t, path, "---\nid: rule_thing\nhuman_name: \"Old Name\"\ntype: RULE\nstatus: DRAFT\n---\n\n# Old Name\n\n## INTENT\n\n")

	if _, err := Update(UpdateOptions{FilePath: path, SetArgs: []string{"human_name=New Name"}}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if title := h1(readAtom(t, path)); title != "New Name" {
		t.Errorf("H1 = %q, want %q", title, "New Name")
	}
}
