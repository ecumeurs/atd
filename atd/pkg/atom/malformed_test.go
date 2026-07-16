package atom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTemp writes content to path.atom.md under t.TempDir() and returns the
// path, without parsing it (the caller decides what to assert about Parse).
func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name+".atom.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	return path
}

// TestParseNoClosingDelimiter documents Parse's behavior on frontmatter that
// never closes with a second "---": Parse does not require a closing
// delimiter to make progress. Since header-vs-section mode is driven purely
// by "does this line start with '##'" (parse.go's `mode = "sections"`
// branch), a body that reaches its first "## " heading still transitions out
// of header parsing correctly, and every "key: value" line seen before that
// point is captured. This is not a crash or data-loss bug — it is a
// permissive parser — but it means a truncated/malformed file with no
// closing "---" silently parses "successfully" rather than erroring, which
// is worth pinning so a future stricter rewrite doesn't change this by
// accident without a conscious decision.
func TestParseNoClosingDelimiter(t *testing.T) {
	t.Parallel()
	content := `---
id: no_closing_delim
human_name: No Closing Delimiter
type: REQUIREMENT
status: DRAFT

## INTENT
Intent despite missing closing ---.
`
	path := writeTemp(t, "no_closing", content)
	data, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse returned an error for a missing closing delimiter (expected permissive success): %v", err)
	}
	if data.ID != "no_closing_delim" {
		t.Errorf("ID: got %q, want %q", data.ID, "no_closing_delim")
	}
	if data.Status != "DRAFT" {
		t.Errorf("Status: got %q, want %q", data.Status, "DRAFT")
	}
	if data.Intent != "Intent despite missing closing ---." {
		t.Errorf("Intent: got %q", data.Intent)
	}
}

// TestParseDuplicateKeys pins last-write-wins semantics for a frontmatter
// key repeated twice: Parse has no duplicate-key detection, it just
// overwrites data.ID (etc.) every time a matching prefix line is seen, so
// the final occurrence in the file wins.
func TestParseDuplicateKeys(t *testing.T) {
	t.Parallel()
	content := `---
id: first_id
id: second_id
status: DRAFT
status: STABLE
---

## INTENT
body
`
	path := writeTemp(t, "dup_keys", content)
	data, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if data.ID != "second_id" {
		t.Errorf("expected last-write-wins ID %q, got %q", "second_id", data.ID)
	}
	if data.Status != "STABLE" {
		t.Errorf("expected last-write-wins Status %q, got %q", "STABLE", data.Status)
	}
}

// TestParseCRLFLineEndings pins that bufio.Scanner's default split function
// (ScanLines) strips a trailing \r before the \n, so a frontmatter/body
// written with Windows line endings parses identically to LF-only content —
// no special-casing needed in Parse itself.
func TestParseCRLFLineEndings(t *testing.T) {
	t.Parallel()
	lfContent := "---\r\nid: crlf_atom\r\nhuman_name: CRLF Atom\r\ntype: REQUIREMENT\r\nstatus: DRAFT\r\n---\r\n\r\n## INTENT\r\nIntent over CRLF.\r\n\r\n## THE RULE / LOGIC\r\nLogic over CRLF.\r\n"
	path := writeTemp(t, "crlf", lfContent)
	data, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed on CRLF content: %v", err)
	}
	if data.ID != "crlf_atom" {
		t.Errorf("ID: got %q, want %q", data.ID, "crlf_atom")
	}
	if strings.Contains(data.ID, "\r") {
		t.Errorf("ID retained a stray \\r: %q", data.ID)
	}
	if data.Intent != "Intent over CRLF." {
		t.Errorf("Intent: got %q", data.Intent)
	}
	if strings.Contains(data.Intent, "\r") {
		t.Errorf("Intent retained a stray \\r: %q", data.Intent)
	}
	if data.Logic != "Logic over CRLF." {
		t.Errorf("Logic: got %q", data.Logic)
	}
}

// TestParseMissingRequiredFields pins that Parse never validates presence of
// "required" fields (id, human_name, type, status): a file missing all of
// them parses without error, simply leaving the corresponding AtomData
// fields at their zero value. Any required-field validation in this
// toolkit lives elsewhere (e.g. the CLI/MCP update path), not in Parse.
func TestParseMissingRequiredFields(t *testing.T) {
	t.Parallel()
	content := `---
priority: CORE
---

## INTENT
Orphaned content with no id/type/status.
`
	path := writeTemp(t, "missing_required", content)
	data, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse returned an error for missing required fields (expected no validation): %v", err)
	}
	if data.ID != "" {
		t.Errorf("expected empty ID, got %q", data.ID)
	}
	if data.HumanName != "" {
		t.Errorf("expected empty HumanName, got %q", data.HumanName)
	}
	if data.Type != "" {
		t.Errorf("expected empty Type, got %q", data.Type)
	}
	if data.Status != "" {
		t.Errorf("expected empty Status, got %q", data.Status)
	}
	if data.Priority != "CORE" {
		t.Errorf("Priority: got %q, want %q", data.Priority, "CORE")
	}
}

// TestParseUnknownFrontmatterFields pins that unrecognized frontmatter keys
// (fields Parse has no branch for) are silently ignored rather than
// rejected or accidentally captured into some other field.
func TestParseUnknownFrontmatterFields(t *testing.T) {
	t.Parallel()
	content := `---
id: unknown_fields_atom
human_name: Unknown Fields
type: REQUIREMENT
status: DRAFT
totally_unknown_key: some value
another_bogus_field: [1, 2, 3]
---

## INTENT
Body content.
`
	path := writeTemp(t, "unknown_fields", content)
	data, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed on unknown frontmatter fields: %v", err)
	}
	if data.ID != "unknown_fields_atom" {
		t.Errorf("ID: got %q, want %q", data.ID, "unknown_fields_atom")
	}
	if data.Intent != "Body content." {
		t.Errorf("Intent: got %q", data.Intent)
	}
	if data.Metadata != nil {
		t.Errorf("expected unknown fields to be ignored, not captured into Metadata: got %#v", data.Metadata)
	}
}

// TestParseEmptyFile pins Parse's behavior on a zero-byte file: no crash, a
// zero-value AtomData (aside from FilePath), no error.
func TestParseEmptyFile(t *testing.T) {
	t.Parallel()
	path := writeTemp(t, "empty", "")
	data, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed on empty file: %v", err)
	}
	if data.ID != "" || data.Intent != "" {
		t.Errorf("expected zero-value AtomData for empty file, got %#v", data)
	}
}

// TestParseNonexistentFile pins that Parse surfaces the underlying os.Open
// error for a missing file rather than returning a zero-value success.
func TestParseNonexistentFile(t *testing.T) {
	t.Parallel()
	_, err := Parse(filepath.Join(t.TempDir(), "does_not_exist.atom.md"))
	if err == nil {
		t.Fatal("expected an error for a nonexistent file, got nil")
	}
}
