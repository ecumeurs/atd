package atom

import (
	"os"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	content := `---
id: test_atom
human_name: Test Atom
type: REQUIREMENT
status: DRAFT
priority: CORE
tags: [tag1, tag2]
parents: [[p1]], [[p2]]
dependents: []
---

# Test Atom

## INTENT
This is the intent.
Multi-line.

## THE RULE / LOGIC
This is the logic.
With a rule.

## TECHNICAL INTERFACE
- Interface 1
`
	tmpfile, err := os.CreateTemp("", "test_atom*.atom.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	data, err := Parse(tmpfile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if data.ID != "test_atom" {
		t.Errorf("Expected ID test_atom, got %s", data.ID)
	}
	if data.HumanName != "Test Atom" {
		t.Errorf("Expected HumanName 'Test Atom', got '%s'", data.HumanName)
	}
	if data.Type != "REQUIREMENT" {
		t.Errorf("Expected Type REQUIREMENT, got %s", data.Type)
	}
	if !reflect.DeepEqual(data.Tags, []string{"tag1", "tag2"}) {
		t.Errorf("Expected tags [tag1, tag2], got %v", data.Tags)
	}
	if !reflect.DeepEqual(data.Parents, []string{"p1", "p2"}) {
		t.Errorf("Expected parents [p1, p2], got %v", data.Parents)
	}
	if data.Intent != "This is the intent.\nMulti-line." {
		t.Errorf("Intent mismatch. Got: %s", data.Intent)
	}
	if data.Logic != "This is the logic.\nWith a rule." {
		t.Errorf("Logic mismatch. Got: %s", data.Logic)
	}
	if data.Interface != "- Interface 1" {
		t.Errorf("Interface mismatch. Got: %s", data.Interface)
	}
}

func TestParseMultiLineParents(t *testing.T) {
	content := `---
id: multi_parent
parents:
  - [[p1]]
  - [[p2]]
---
## INTENT
test
`
	tmpfile, err := os.CreateTemp("", "test_multi*.atom.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	data, err := Parse(tmpfile.Name())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	expected := []string{"p1", "p2"}
	if !reflect.DeepEqual(data.Parents, expected) {
		t.Errorf("Expected parents %v, got %v", expected, data.Parents)
	}
}

func TestParseMeta(t *testing.T) {
	content := `---
id: meta_id
human_name: Meta Name
type: META_TYPE
---
`
	tmpfile, err := os.CreateTemp("", "test_meta*.atom.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	id, name, atomType, err := ParseMeta(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if id != "meta_id" || name != "Meta Name" || atomType != "META_TYPE" {
		t.Errorf("Meta mismatch: %s, %s, %s", id, name, atomType)
	}
}
