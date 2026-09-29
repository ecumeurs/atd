package atom

import (
	"reflect"
	"testing"
)

// @test-link [[rule_atd_atom_self_sufficiency]]
func TestFindOutsideReferences(t *testing.T) {
	allowed := map[string]bool{"self_atom": true, "parent_atom": true, "proj:dep_atom": true}
	front := "---\nid: self_atom\nparents:\n  - [[parent_atom]]\ndependents:\n  - [[proj:dep_atom]]\n---\n"

	cases := []struct {
		name string
		body string
		want []string
	}{
		{"clean prose", "## INTENT\nPlain words only.\n", nil},
		{"markdown link", "See [the guide](https://example.com/guide) now.\n",
			[]string{"Markdown link to outside document: https://example.com/guide"}},
		{"markdown link to local file", "- **Impl:** [ext.js](file:///home/x/ext.js)\n",
			[]string{"Markdown link to outside document: file:///home/x/ext.js"}},
		{"reference-style definition", "[guide]: https://example.com/guide\n",
			[]string{"Markdown link definition to outside document: https://example.com/guide"}},
		{"bare url in prose", "Hosted at https://example.com/docs.\n",
			[]string{"URL to outside document: https://example.com/docs"}},
		{"url in code span is a literal", "Endpoint: `http://localhost:7474/mcp`\n", nil},
		{"url in fenced block is a literal", "```bash\ncurl http://localhost:7474/mcp\n```\n", nil},
		{"wiki-link to unrelated atom", "Handled by [[other_atom]] instead.\n",
			[]string{"Wiki-link to an atom outside parents:/dependents: [[other_atom]]"}},
		{"wiki-link to parent, dependent, self", "Built on [[parent_atom]], feeds [[proj:dep_atom]], is [[self_atom]].\n", nil},
		{"spec and test tags are exempt", "- **Code Tag:** @spec-link [[other_atom]] and @test-link [[other_atom]]\n", nil},
		{"wiki-link syntax example in code span", "Parents use `[[id]]` or `[[project:atom_id]]` form.\n", nil},
		{"doc path in prose", "Background lives in failures/2026_report.md now.\n",
			[]string{"Citation of outside document: failures/2026_report.md"}},
		{"doc path in code span", "Recorded in `failures/2026_report.md`.\n",
			[]string{"Citation of outside document: failures/2026_report.md"}},
		{"section citation", "Types come from ATD.md §1.3/§1.5.\n",
			[]string{"Citation of outside document: ATD.md §1.3/§1.5"}},
		{"section citation in parentheses", "Governance atoms (ATD.md §1.4) are isolated.\n",
			[]string{"Citation of outside document: ATD.md §1.4"}},
		{"bare artifact filename", "Writes a structured `task_list.md` then parses task_list.md.\n", nil},
		{"placeholders and globs", "Usage: `atd recon -atom <atom.md>`; scans `*.atom.md` and `<root>/<docs>/ID.atom.md`.\n", nil},
		{"source code path", "- **Binary:** `atd/cmd/atd/cmd/lint.go`\n", nil},
		{"duplicates reported once", "[[other_atom]] then [[other_atom]] again.\n",
			[]string{"Wiki-link to an atom outside parents:/dependents: [[other_atom]]"}},
		{"double-backtick span", "Example: `` `[[id]]` `` stays literal.\n", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindOutsideReferences(front+tc.body, allowed)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// @test-link [[rule_atd_atom_self_sufficiency]]
func TestFindOutsideReferencesScansDescriptionOnly(t *testing.T) {
	content := "---\nid: a\ndescription: \"Why this exists: see notes/why.md\"\nhuman_name: \"Per GUIDE.md §2\"\nparents:\n  - [[p]]\n---\n\n## INTENT\nFine.\n"
	got := FindOutsideReferences(content, map[string]bool{"a": true, "p": true})
	want := []string{"Citation of outside document: notes/why.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSplitFrontmatter(t *testing.T) {
	front, body := SplitFrontmatter("---\nid: x\n---\n# Title\n")
	if !reflect.DeepEqual(front, []string{"id: x"}) || body != "# Title\n" {
		t.Errorf("got front=%q body=%q", front, body)
	}
	front, body = SplitFrontmatter("# No frontmatter\n")
	if front != nil || body != "# No frontmatter\n" {
		t.Errorf("got front=%q body=%q", front, body)
	}
}
