package exploration

// Unit back-fill for extractLinks (test_atd_07_26.md §3.1 item 1, §6 WP-7):
// every comment style, multiple links per file, links inside string
// literals vs comments, and malformed link syntax. extractLinks is a pure
// regex scan over raw file text (specLinkRegex/testLinkRegex in explorer.go)
// with zero awareness of the host language's comment syntax -- these tests
// pin that reality rather than assuming comment-aware parsing exists.

import (
	"testing"
)

func TestExtractLinks_CommentStyles(t *testing.T) {
	styles := []struct {
		name    string
		content string
	}{
		{"go_line_comment", "package p\n\n// @spec-link [[zz_style_atom]]\nfunc F() {}\n"},
		{"shell_hash_comment", "#!/bin/sh\n# @spec-link [[zz_style_atom]]\necho hi\n"},
		{"html_comment", "<html>\n<!-- @spec-link [[zz_style_atom]] -->\n</html>\n"},
		{"c_block_comment", "/* @spec-link [[zz_style_atom]] */\nint main() {}\n"},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()
			e := newLinkTestExplorer("zz_style_atom")
			e.extractLinks("f.src", s.content)

			if len(e.SpecLinks) != 1 {
				t.Fatalf("expected 1 spec link regardless of comment style, got %d: %+v", len(e.SpecLinks), e.SpecLinks)
			}
			if e.SpecLinks[0].AtomID != "zz_style_atom" {
				t.Errorf("got atom id %q, want zz_style_atom", e.SpecLinks[0].AtomID)
			}
			node := e.Graph.Atoms["zz_style_atom"]
			if len(node.Implementations) != 1 {
				t.Errorf("expected node.Implementations to record 1 site, got %+v", node.Implementations)
			}
		})
	}
}

// TestExtractLinks_MultipleLinksPerFile pins that a single file may tag the
// same atom more than once (the fixture_project S2 dedup case, api_zzfix_beta
// in src/beta.go) and a different atom once more -- extractLinks itself does
// NOT dedup; every raw occurrence becomes its own SpecLink/Implementation
// entry. Dedup, where it happens at all, is a coverage-report concern
// (pkg/coverage), not extractLinks'.
func TestExtractLinks_MultipleLinksPerFile(t *testing.T) {
	e := newLinkTestExplorer("zz_multi_a", "zz_multi_b")
	content := "// @spec-link [[zz_multi_a]]\nfunc A() {}\n\n" +
		"// @spec-link [[zz_multi_b]]\nfunc B() {}\n\n" +
		"// @spec-link [[zz_multi_a]]\nfunc AAgain() {}\n"

	e.extractLinks("multi.go", content)

	if len(e.SpecLinks) != 3 {
		t.Fatalf("expected 3 raw spec links (a tagged twice, b once), got %d: %+v", len(e.SpecLinks), e.SpecLinks)
	}
	if got := len(e.Graph.Atoms["zz_multi_a"].Implementations); got != 2 {
		t.Errorf("expected zz_multi_a to have 2 implementation entries, got %d: %v", got, e.Graph.Atoms["zz_multi_a"].Implementations)
	}
	if got := len(e.Graph.Atoms["zz_multi_b"].Implementations); got != 1 {
		t.Errorf("expected zz_multi_b to have 1 implementation entry, got %d: %v", got, e.Graph.Atoms["zz_multi_b"].Implementations)
	}

	// Line numbers must reflect where each tag actually sits, not just count.
	want := map[string]int{"zz_multi_a": 1, "zz_multi_b": 4}
	seenFirst := map[string]bool{}
	for _, sl := range e.SpecLinks {
		if seenFirst[sl.AtomID] {
			continue
		}
		seenFirst[sl.AtomID] = true
		if sl.Line != want[sl.AtomID] {
			t.Errorf("first occurrence of %s: line = %d, want %d", sl.AtomID, sl.Line, want[sl.AtomID])
		}
	}
}

// TestExtractLinks_LinksInStringsAlsoMatch pins that extractLinks has no
// lexical awareness: a tag sitting inside a Go string literal (not a
// comment at all) is picked up exactly like one inside a // comment. This
// is documented current behavior, not something this back-fill fixes.
func TestExtractLinks_LinksInStringsAlsoMatch(t *testing.T) {
	e := newLinkTestExplorer("zz_string_atom")
	content := "package p\n\nconst msg = \"@spec-link [[zz_string_atom]]\"\n"

	e.extractLinks("strings.go", content)

	if len(e.SpecLinks) != 1 {
		t.Fatalf("expected the tag inside the string literal to be picked up (extractLinks has no comment/string distinction), got %d links", len(e.SpecLinks))
	}
}

// TestExtractLinks_MalformedSyntax pins specLinkRegex's actual tolerance
// (explorer.go: `@spec-link\s+\[?\[?([a-zA-Z0-9_\-\.\:]+)\]?\]?`) for
// various malformed [[...]] shapes. Both directions are real: some
// "malformed" input still resolves (brackets are fully optional), while
// other shapes silently produce no link at all.
func TestExtractLinks_MalformedSyntax(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    int // expected number of recorded SpecLinks
	}{
		{
			name:    "missing_closing_brackets_still_resolves",
			content: "// @spec-link [[zz_malformed_atom\nfunc F() {}\n",
			want:    1,
		},
		{
			name:    "single_closing_bracket_still_resolves",
			content: "// @spec-link [[zz_malformed_atom]\nfunc F() {}\n",
			want:    1,
		},
		{
			name:    "triple_opening_bracket_no_match",
			content: "// @spec-link [[[zz_malformed_atom]]]\n",
			want:    0,
		},
		{
			name:    "empty_double_brackets_no_match",
			content: "// @spec-link [[]]\n",
			want:    0,
		},
		{
			name:    "missing_whitespace_before_brackets_no_match",
			content: "// @spec-link[[zz_malformed_atom]]\n",
			want:    0,
		},
		{
			name:    "bare_id_no_brackets_resolves",
			content: "// @spec-link zz_malformed_atom\n",
			want:    1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			e := newLinkTestExplorer("zz_malformed_atom")
			e.extractLinks("f.go", c.content)
			if len(e.SpecLinks) != c.want {
				t.Errorf("content %q: expected %d spec links, got %d: %+v", c.content, c.want, len(e.SpecLinks), e.SpecLinks)
			}
		})
	}
}

// TestExtractLinks_TestLinkSetsHasTests exercises the parallel @test-link
// path (testLinkRegex), confirming it records a TestLink and flips
// HasTests, mirroring the @spec-link assertions above.
func TestExtractLinks_TestLinkSetsHasTests(t *testing.T) {
	e := newLinkTestExplorer("zz_test_atom")
	e.extractLinks("f_test.go", "// @test-link [[zz_test_atom]]\nfunc TestF(t *testing.T) {}\n")

	if len(e.TestLinks) != 1 {
		t.Fatalf("expected 1 test link, got %d: %+v", len(e.TestLinks), e.TestLinks)
	}
	if !e.Graph.Atoms["zz_test_atom"].HasTests {
		t.Error("expected HasTests to be set on the target atom")
	}
}

// TestExtractLinks_UnknownAtomIDSilentlyIgnored pins that a @spec-link
// pointing at an atom id absent from the graph produces no SpecLink and no
// error -- extractLinks degrades silently for an unresolvable tag rather
// than surfacing it (that surfacing, where it happens, is `atd lint`'s job).
func TestExtractLinks_UnknownAtomIDSilentlyIgnored(t *testing.T) {
	e := newLinkTestExplorer() // empty graph
	e.extractLinks("f.go", "// @spec-link [[zz_does_not_exist]]\n")

	if len(e.SpecLinks) != 0 {
		t.Errorf("expected no spec link recorded for an atom id absent from the graph, got %+v", e.SpecLinks)
	}
}
