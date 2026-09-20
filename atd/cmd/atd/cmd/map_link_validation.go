package cmd

// Validation and salvage helpers for `atd map` discover mode's
// link-recommendation leg (runMapDiscover in map.go).
//
// Field reports: failures/20260901_atd_map_hallucinated_code_values.md,
// failures/20260901_atd_map_malformed_json_from_link_recommender.md,
// failures/20260902_atd_map_malformed_json_link_recommender_recurrence.md,
// failures/20260917_atd_map_hallucinated_links_and_malformed_json_recurrence.md.
//
// Two defects are addressed here:
//
//  1. Hallucinated recommendations. Whatever the model put in
//     `recommendations` was rendered verbatim as `- [[<entry>]]`, so
//     "Add comments to explain complex logic", a refusal sentence and a
//     fabricated GitHub URL were all presented as recommended atom links.
//     Every candidate now has to be atom-ID shaped AND resolve to a real atom
//     in this project's registry; anything else is dropped.
//
//  2. Truncated JSON. A response cut short by the provider's token budget used
//     to fail the whole command. salvageRecommendations recovers the ids from a
//     partial payload, which — combined with the schema reordering in
//     pkg/prompt/discover_links.go that now emits `recommendations` first —
//     turns the common truncation case into a complete answer minus its
//     rationale rather than a hard failure.

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"atd-tools/pkg/exploration"
)

// atomIDShape is this project's atom-ID naming convention (CLAUDE.md,
// "Naming Patterns"): lowercase snake_case `<type>_<slug>`, with an optional
// `project:` prefix for cross-project references. It deliberately rejects
// anything containing whitespace, uppercase letters, or punctuation, which is
// what every hallucinated entry in the field reports had in common.
var atomIDShape = regexp.MustCompile(`^([a-z][a-z0-9_]*:)?[a-z][a-z0-9]*(_[a-z0-9]+)+$`)

// maxAtomIDLen guards against a pathological "sentence_without_spaces" that
// would otherwise satisfy the shape regex. Real atom IDs are far shorter.
const maxAtomIDLen = 120

// isAtomIDShaped reports whether s could be an atom ID at all.
func isAtomIDShaped(s string) bool {
	return len(s) <= maxAtomIDLen && atomIDShape.MatchString(s)
}

// normalizeCandidate trims a raw model suggestion down to the bare id: models
// frequently echo the `[[...]]` wrapper used in the prompt's registry listing,
// and sometimes a trailing "- " bullet or stray whitespace.
func normalizeCandidate(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "- ")
	s = strings.TrimSpace(s)
	for strings.HasPrefix(s, "[[") && strings.HasSuffix(s, "]]") && len(s) > 4 {
		s = strings.TrimSpace(s[2 : len(s)-2])
	}
	return strings.TrimSpace(s)
}

// droppedCandidate is one rejected model suggestion, kept for the diagnostic
// note rendered under the recommendations (never as a `[[...]]` link).
type droppedCandidate struct {
	Value  string
	Reason string
}

// linkRegistry answers "is this a real atom in this project?" It reuses the
// existing exploration.Explorer resolution plumbing (the same lookup behind
// `atd query --field id` / `atd trace`) rather than building a second index,
// and it is lazy: the ids the semantic search already surfaced are known-good
// by construction, so the common case costs nothing.
type linkRegistry struct {
	docsDir  string
	known    map[string]bool
	explorer *exploration.Explorer
	loaded   bool
	loadErr  error
}

func newLinkRegistry(docsDir string, searchHits []string) *linkRegistry {
	known := make(map[string]bool, len(searchHits))
	for _, id := range searchHits {
		known[id] = true
	}
	return &linkRegistry{docsDir: docsDir, known: known}
}

// load builds the Explorer on first use. In a workspace the Resolver alone can
// answer id lookups; standalone projects need the atom graph loaded.
func (r *linkRegistry) load() (*exploration.Explorer, error) {
	if !r.loaded {
		r.loaded = true
		e := exploration.NewExplorer("", r.docsDir)
		if e.Resolver == nil {
			if err := e.Load(false); err != nil {
				r.loadErr = err
				return nil, err
			}
		}
		r.explorer = e
	}
	return r.explorer, r.loadErr
}

// exists reports whether id resolves to a real atom.
func (r *linkRegistry) exists(id string) (bool, error) {
	if r.known[id] {
		return true, nil
	}
	e, err := r.load()
	if err != nil {
		return false, err
	}
	if _, err := e.CanonicalAtomID(id); err != nil {
		return false, nil
	}
	return true, nil
}

// filterRecommendedAtomIDs keeps only candidates that are atom-ID shaped and
// resolve to a real atom, preserving order and de-duplicating. Everything else
// comes back in dropped with the reason it was rejected.
func filterRecommendedAtomIDs(candidates []string, reg *linkRegistry) (valid []string, dropped []droppedCandidate) {
	seen := make(map[string]bool, len(candidates))
	for _, raw := range candidates {
		id := normalizeCandidate(raw)
		if id == "" {
			continue
		}
		if seen[id] {
			continue
		}
		seen[id] = true

		if !isAtomIDShaped(id) {
			dropped = append(dropped, droppedCandidate{Value: id, Reason: "not an atom ID (expected lowercase snake_case <type>_<slug>)"})
			continue
		}
		ok, err := reg.exists(id)
		switch {
		case err != nil:
			dropped = append(dropped, droppedCandidate{Value: id, Reason: fmt.Sprintf("atom registry unavailable, could not verify (%v)", err)})
		case !ok:
			dropped = append(dropped, droppedCandidate{Value: id, Reason: "no such atom in this project's registry"})
		default:
			valid = append(valid, id)
		}
	}
	return valid, dropped
}

// salvageRecommendations extracts the `recommendations` array out of a JSON
// payload that json.Unmarshal rejected — the truncated-response case from the
// malformed-JSON field reports. It scans for complete string literals only; an
// element cut off mid-token is discarded rather than guessed at, and a payload
// with no recoverable element returns nil so the caller can still fail loudly.
func salvageRecommendations(raw string) []string {
	const key = `"recommendations"`
	i := strings.Index(raw, key)
	if i < 0 {
		return nil
	}
	rest := raw[i+len(key):]

	open := strings.Index(rest, "[")
	if open < 0 {
		return nil
	}
	// Anything other than whitespace/':' between the key and the '[' means we
	// found the word somewhere else (e.g. inside the rationale prose).
	if strings.Trim(rest[:open], " \t\r\n:") != "" {
		return nil
	}
	rest = rest[open+1:]

	var out []string
	for {
		start := strings.IndexAny(rest, `"]`)
		if start < 0 || rest[start] == ']' {
			break
		}
		rest = rest[start+1:]

		end := -1
		for p := 0; p < len(rest); p++ {
			if rest[p] == '\\' {
				p++
				continue
			}
			if rest[p] == '"' {
				end = p
				break
			}
		}
		if end < 0 {
			// Truncated mid-element: drop the partial token.
			break
		}

		literal := rest[:end]
		var decoded string
		if err := json.Unmarshal([]byte(`"`+literal+`"`), &decoded); err != nil {
			decoded = literal
		}
		out = append(out, decoded)
		rest = rest[end+1:]
	}
	return out
}

// salvageRationale makes a best effort at the rationale from a partial payload
// (it is optional context, so a truncated one is simply trimmed).
func salvageRationale(raw string) string {
	const key = `"rationale"`
	i := strings.Index(raw, key)
	if i < 0 {
		return ""
	}
	rest := raw[i+len(key):]
	open := strings.Index(rest, `"`)
	if open < 0 {
		return ""
	}
	if strings.Trim(rest[:open], " \t\r\n:") != "" {
		return ""
	}
	rest = rest[open+1:]

	end := len(rest)
	for p := 0; p < len(rest); p++ {
		if rest[p] == '\\' {
			p++
			continue
		}
		if rest[p] == '"' {
			end = p
			break
		}
	}
	literal := rest[:end]
	var decoded string
	if err := json.Unmarshal([]byte(`"`+literal+`"`), &decoded); err != nil {
		// Truncated escape sequence or similar — fall back to the raw slice.
		decoded = strings.TrimRight(literal, `\`)
	}
	return strings.TrimSpace(decoded)
}

// renderLinkRecommendations builds the discover-mode report.
//
// When nothing survives validation the report says so explicitly and the
// rationale is suppressed: with no atom to anchor it, the rationale is exactly
// the free-form prose the field reports found to be confidently wrong
// (fabricated line values, a fabricated repo URL), so surfacing it adds risk
// and no traceability.
func renderLinkRecommendations(valid []string, dropped []droppedCandidate, rationale string, truncated bool) string {
	var b strings.Builder
	b.WriteString("### Recommended Atom Links\n")
	if len(valid) == 0 {
		b.WriteString("(none — no suggestion passed atom-ID validation against this project's registry)\n")
	}
	for _, id := range valid {
		b.WriteString(fmt.Sprintf("- [[%s]]\n", id))
	}

	if len(dropped) > 0 {
		b.WriteString(fmt.Sprintf("\n### Discarded Suggestions (%d)\n", len(dropped)))
		b.WriteString("Not rendered as links — the model returned these where atom IDs were expected:\n")
		for _, d := range dropped {
			b.WriteString(fmt.Sprintf("- %.80q — %s\n", d.Value, d.Reason))
		}
	}

	if truncated {
		b.WriteString("\n_Note: the model's response was truncated; the atom IDs above were recovered from the partial payload._\n")
	}

	if len(valid) > 0 && strings.TrimSpace(rationale) != "" {
		b.WriteString("\n### Rationale\n")
		b.WriteString(rationale)
	}
	return b.String()
}
