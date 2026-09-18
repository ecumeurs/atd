package exploration

import (
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
)

// gitignoreMatcher decides whether a project-root-relative path should be
// excluded from discovery because it matches the project's configured
// GitignorePatterns and/or a real .gitignore file at the project root.
//
// This is deliberately a subset of git's actual ignore semantics: patterns
// are matched independent of order and negation ("!pattern") lines are
// skipped rather than honored, since correctly reproducing git's
// last-match-wins/re-include behavior is out of scope for what discovery
// needs here (skip build artifacts and vendored code, not byte-for-byte
// parity with `git check-ignore`).
type gitignoreMatcher struct {
	patterns []string
}

// newGitignoreMatcher builds a matcher from config.GetGitignorePatterns()
// plus the lines of a .gitignore file directly under root, if one exists.
func newGitignoreMatcher(root string) *gitignoreMatcher {
	patterns := append([]string{}, config.GetGitignorePatterns()...)

	if data, err := os.ReadFile(filepath.Join(root, ".gitignore")); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
				continue
			}
			patterns = append(patterns, line)
		}
	}

	return &gitignoreMatcher{patterns: patterns}
}

// isIgnored reports whether rel (a project-root-relative, slash-separated
// path) matches any configured or .gitignore-sourced pattern.
func (m *gitignoreMatcher) isIgnored(rel string) bool {
	if m == nil {
		return false
	}
	rel = filepath.ToSlash(rel)

	for _, pat := range m.patterns {
		pat = strings.TrimSpace(pat)
		if pat == "" {
			continue
		}

		anchored := strings.HasPrefix(pat, "/")
		pat = strings.TrimPrefix(pat, "/")
		pat = strings.TrimSuffix(pat, "/")
		if pat == "" {
			continue
		}

		if anchored {
			if ok, _ := filepath.Match(pat, rel); ok {
				return true
			}
			continue
		}

		segs := strings.Split(rel, "/")
		for i := range segs {
			if ok, _ := filepath.Match(pat, segs[i]); ok {
				return true
			}
			if ok, _ := filepath.Match(pat, strings.Join(segs[i:], "/")); ok {
				return true
			}
		}
	}

	return false
}

// shouldSkipDiscoveredPath reports whether rel (relative to the walk root)
// should be excluded from atom/code discovery: dotfiles/dirs, common
// vendor/build directories, the project's docs directory when docsDir is
// non-empty (pass "" when the caller still needs to see docs, e.g. because
// it's parsing .atom.md files out of it), and anything the gitignore
// matcher flags.
func shouldSkipDiscoveredPath(rel string, docsDir string, ignore *gitignoreMatcher) bool {
	rel = filepath.ToSlash(rel)

	if strings.HasPrefix(rel, ".") ||
		strings.Contains(rel, "/.") ||
		strings.Contains(rel, "vendor/") ||
		strings.Contains(rel, "node_modules/") ||
		strings.Contains(rel, "dist/") ||
		strings.Contains(rel, "build/") {
		return true
	}

	if docsDir != "" {
		docsDir = strings.TrimSuffix(filepath.ToSlash(docsDir), "/")
		if docsDir != "" && (rel == docsDir || strings.HasPrefix(rel, docsDir+"/")) {
			return true
		}
	}

	return ignore.isIgnored(rel)
}
