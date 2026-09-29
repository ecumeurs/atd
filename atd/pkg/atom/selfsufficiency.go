package atom

import (
	"regexp"
	"strings"
)

// An atom must stand entirely on its own: a reader who has never opened any
// other file must be able to understand it from its own content. The only
// links an atom may carry are its structural graph edges — parents:,
// dependents:, and @spec-link/@test-link tags. Everything else that points a
// reader elsewhere (a hyperlink, a URL, a wiki-link to an unrelated atom, a
// citation of a design doc or report) is a narrative dependency on something
// that can move, rot, or never be read, and lets the atom itself stay vague.

var (
	mdLinkRe      = regexp.MustCompile(`!?\[[^\]]*\]\(\s*<?([^)\s>]+)>?[^)]*\)`)
	mdRefDefRe    = regexp.MustCompile(`^\s{0,3}\[[^\]]+\]:\s*<?(\S+?)>?(\s|$)`)
	urlRe         = regexp.MustCompile(`(?i)\b(?:https?|ftp|file)://[^\s<>()\[\]` + "`" + `"']+`)
	wikiLinkRe    = regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
	tagWikiLinkRe = regexp.MustCompile(`@(?:spec|test)-link\s+\[\[[^\[\]]+\]\]`)
	docExtRe      = regexp.MustCompile(`(?i)\.(?:md|markdown|mdx|rst|adoc|pdf|docx?|odt)$`)
	sectionCiteRe = regexp.MustCompile(`^\s*§`)
)

// SplitFrontmatter separates an atom file's YAML frontmatter lines from its
// markdown body. A file that does not open with a "---" line has no
// frontmatter and is all body.
func SplitFrontmatter(content string) (front []string, body string) {
	lines := strings.Split(content, "\n")
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	if start >= len(lines) || strings.TrimSpace(lines[start]) != "---" {
		return nil, content
	}
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return lines[start+1 : i], strings.Join(lines[i+1:], "\n")
		}
	}
	// Unterminated frontmatter: everything is frontmatter, nothing is body.
	return lines[start+1:], ""
}

// @spec-link [[rule_atd_atom_self_sufficiency]]
// FindOutsideReferences scans an atom file's content for references to
// documents outside the atom, returning one human-readable finding per
// distinct reference, in order of first appearance. allowed holds the
// atom's own id plus its parents:/dependents: entries; a prose [[id]]
// naming one of them is a restatement of a structural edge, not an outside
// link.
//
// Detected shapes:
//   - markdown links/images and reference-style link definitions (any target)
//   - bare or autolinked URLs in prose (a URL inside a code span or fenced
//     block is a literal value such as an endpoint, not a link)
//   - [[id]] wiki-links in prose to an atom not in allowed; @spec-link /
//     @test-link tags and code-span syntax examples like `[[id]]` are exempt
//   - document citations, in prose or code spans: a document-file path with a
//     directory component (reports/x.md) or a document cited by section
//     (GUIDE.md §1.4). Bare filenames (task_list.md) and placeholders/globs
//     (<atom.md>, *.atom.md) name artifacts, not cited documents.
//
// Fenced code blocks are skipped entirely: they hold schemas, commands and
// examples. The frontmatter description: value is scanned as prose; every
// other frontmatter key is structural and ignored.
func FindOutsideReferences(content string, allowed map[string]bool) []string {
	front, body := SplitFrontmatter(content)

	var findings []string
	seen := map[string]bool{}
	add := func(f string) {
		if !seen[f] {
			seen[f] = true
			findings = append(findings, f)
		}
	}

	var lines []string
	for _, l := range front {
		if strings.HasPrefix(l, "description:") {
			lines = append(lines, stripQuotes(strings.TrimPrefix(l, "description:")))
		}
	}

	fence := ""
	for _, l := range strings.Split(body, "\n") {
		t := strings.TrimSpace(l)
		if fence != "" {
			if strings.HasPrefix(t, fence) {
				fence = ""
			}
			continue
		}
		if strings.HasPrefix(t, "```") {
			fence = "```"
			continue
		}
		if strings.HasPrefix(t, "~~~") {
			fence = "~~~"
			continue
		}
		lines = append(lines, l)
	}

	for _, l := range lines {
		prose, code := splitCodeSpans(l)

		if m := mdRefDefRe.FindStringSubmatch(prose); m != nil {
			add("Markdown link definition to outside document: " + m[1])
			prose = ""
		}
		for _, m := range mdLinkRe.FindAllStringSubmatch(prose, -1) {
			add("Markdown link to outside document: " + m[1])
		}
		// Blank out markdown links so their URL targets aren't reported twice.
		proseNoLinks := mdLinkRe.ReplaceAllString(prose, " ")

		for _, u := range urlRe.FindAllString(proseNoLinks, -1) {
			add("URL to outside document: " + strings.TrimRight(u, ".,;:!?"))
		}

		for _, m := range wikiLinkRe.FindAllStringSubmatch(tagWikiLinkRe.ReplaceAllString(proseNoLinks, " "), -1) {
			ref := strings.TrimSpace(m[1])
			if !allowed[ref] {
				add("Wiki-link to an atom outside parents:/dependents: [[" + ref + "]]")
			}
		}

		for _, seg := range append([]string{proseNoLinks}, code...) {
			for _, c := range documentCitations(seg) {
				add("Citation of outside document: " + c)
			}
		}
	}
	return findings
}

// splitCodeSpans returns the line with inline code spans blanked out, plus
// the contents of those spans. A span opens with a run of N backticks and
// closes at the next run of exactly N; an unclosed run is literal text.
func splitCodeSpans(line string) (prose string, spans []string) {
	var b strings.Builder
	i := 0
	for i < len(line) {
		if line[i] != '`' {
			b.WriteByte(line[i])
			i++
			continue
		}
		n := 0
		for i+n < len(line) && line[i+n] == '`' {
			n++
		}
		open := i + n
		closeAt := -1
		for j := open; j < len(line); {
			if line[j] != '`' {
				j++
				continue
			}
			k := 0
			for j+k < len(line) && line[j+k] == '`' {
				k++
			}
			if k == n {
				closeAt = j
				break
			}
			j += k
		}
		if closeAt < 0 {
			b.WriteString(line[i:open])
			i = open
			continue
		}
		spans = append(spans, line[open:closeAt])
		b.WriteByte(' ')
		i = closeAt + n
	}
	return b.String(), spans
}

// documentCitations returns the document references in s: document-file
// tokens that carry a directory component, or that are immediately followed
// by a § section marker.
func documentCitations(s string) []string {
	var out []string
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '(' || r == ')' || r == '"' || r == '\'' || r == ','
	})
	for idx, f := range fields {
		tok := strings.TrimRight(f, ".;:!?")
		// Placeholders and globs name a shape of file, not a specific document.
		if strings.ContainsAny(tok, "<>*{}$[]") || !docExtRe.MatchString(tok) {
			continue
		}
		rest := ""
		if idx+1 < len(fields) {
			rest = fields[idx+1]
		}
		switch {
		case strings.Contains(tok, "/"):
			out = append(out, tok)
		case sectionCiteRe.MatchString(rest):
			out = append(out, tok+" "+strings.TrimRight(rest, ".;:!?"))
		}
	}
	return out
}
