package exploration

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/workspace"
	"fmt"
	"os"
	"sort"
	"strings"
)

// @spec-link [[mechanic_atd_weave]]
// Weave performs bi-directional link synchronization.
//
// When a workspace can be loaded (either pre-attached on the explorer or
// discovered upward from cwd), Weave routes to the workspace-aware path so
// cross-project parents are resolved against the actual project that owns
// them. Otherwise it falls back to a single-project pass that no longer
// guesses prefixes from the host atom.
func (e *Explorer) Weave() (string, error) {
	if e.Workspace == nil {
		if cwd, err := os.Getwd(); err == nil {
			if ws, err := workspace.LoadWorkspace(cwd); err == nil {
				e.Workspace = ws
				if idx, err := ws.BuildIndex(); err == nil {
					e.Index = idx
					e.Resolver = workspace.NewResolver(ws, idx, "")
				}
			}
		}
	}

	if e.Workspace != nil && e.Index != nil {
		return e.weaveWorkspace()
	}
	return e.weaveSingleProject()
}

// weaveSingleProject is the legacy path: one project, no cross-project refs.
// References are kept verbatim — we never invent a project prefix from the
// host atom, since that broke cross-project links during migration.
func (e *Explorer) weaveSingleProject() (string, error) {
	if err := e.Load(false); err != nil {
		return "", err
	}

	parentToDependents := make(map[string][]string)

	for id, a := range e.Graph.Atoms {
		for _, p := range a.Parents {
			ref := strings.Trim(p, "[]")
			parentToDependents[ref] = append(parentToDependents[ref], id)
		}
	}

	edited := 0
	for id, a := range e.Graph.Atoms {
		if a.FilePath == "" {
			continue
		}
		// Look up dependents by raw id only — we are explicitly NOT prefixing
		// in single-project mode.
		deps := parentToDependents[id]
		sort.Strings(deps)
		changed, err := rewriteAtomLinks(a.FilePath, nil, deps, "" /* no project ctx */)
		if err != nil {
			return "", err
		}
		if changed {
			edited++
		}
	}

	// Refresh in-memory graph dependents.
	for id, a := range e.Graph.Atoms {
		a.Dependents = parentToDependents[id]
	}

	result := fmt.Sprintf("Link Weaving Complete. Updated %d files.", edited)
	config.Log("atd-weave", result)
	return result, nil
}

// weaveWorkspace runs a workspace-aware weave: parents and dependents are
// resolved through the workspace index/resolver, then written back with the
// canonical [[project:atom_id]] form (or [[shared:atom_id]] for atoms hosted
// in the workspace-root docs/ folder). Same-project dependents are written
// without a prefix to keep diffs minimal.
func (e *Explorer) weaveWorkspace() (string, error) {
	ws := e.Workspace
	idx := e.Index

	resolver := workspace.NewResolver(ws, idx, "")
	parentToDependents := make(map[string][]string) // canonical "[[proj:id]]" -> []canonical dep ids

	// 1. Build parent→dependents map across the workspace.
	type atomCtx struct {
		Loc     *workspace.AtomLocation
		AtomID  string
		Parsed  *atom.AtomData
	}
	atomsByCanonical := make(map[string]atomCtx)

	for atomID, loc := range idx.ByID {
		a, err := atom.Parse(loc.Path)
		if err != nil {
			continue
		}
		canonical := fmt.Sprintf("[[%s:%s]]", loc.Project, atomID)
		atomsByCanonical[canonical] = atomCtx{Loc: loc, AtomID: atomID, Parsed: &a}

		for _, pRef := range a.Parents {
			resolved, err := resolver.Resolve(pRef)
			if err != nil || resolved == nil {
				continue
			}
			parentCanonical := fmt.Sprintf("[[%s:%s]]", resolved.Project, resolved.AtomID)
			parentToDependents[parentCanonical] = append(parentToDependents[parentCanonical], canonical)
		}
	}

	// 2. Rewrite each atom file: canonicalize parents, refresh dependents.
	edited := 0
	for canonical, ctx := range atomsByCanonical {
		a := ctx.Parsed
		loc := ctx.Loc

		// Per-atom resolver so bare same-project parents stay canonical.
		projResolver := workspace.NewResolver(ws, idx, loc.Project)

		newParents := make([]string, 0, len(a.Parents))
		for _, pRef := range a.Parents {
			shouldUpdate, replacement := projResolver.ShouldUpdate(pRef)
			if shouldUpdate {
				newParents = append(newParents, replacement)
			} else {
				newParents = append(newParents, pRef)
			}
		}

		deps := parentToDependents[canonical]
		sort.Strings(deps)

		// Render dependents: strip project prefix when same-project as host.
		newDeps := make([]string, len(deps))
		for i, dep := range deps {
			parsed := workspace.ParseReference(dep)
			if parsed.Project == loc.Project {
				newDeps[i] = fmt.Sprintf("[[%s]]", parsed.AtomID)
			} else {
				newDeps[i] = dep
			}
		}

		changed, err := rewriteAtomLinks(loc.Path, newParents, newDeps, loc.Project)
		if err != nil {
			return "", err
		}
		if changed {
			edited++
		}
	}

	result := fmt.Sprintf("Workspace Link Weaving Complete. Updated %d files.", edited)
	config.Log("atd-weave", result)
	return result, nil
}

// rewriteAtomLinks updates the parents and/or dependents lists in an atom's
// frontmatter. Pass nil for parents to leave them untouched. The currentProject
// argument is unused at the file-rewrite layer but is kept for symmetry with
// callers that already strip project prefixes themselves.
func rewriteAtomLinks(path string, parents, dependents []string, _ string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(content), "\n")
	inFrontmatter := false
	frontmatterEnd := -1
	parentsLineIdx := -1
	parentsEndIdx := -1
	dependentsLineIdx := -1
	dependentsEndIdx := -1

	for i, line := range lines {
		if line == "---" {
			if !inFrontmatter && i == 0 {
				inFrontmatter = true
				continue
			}
			if inFrontmatter {
				inFrontmatter = false
				frontmatterEnd = i
				break
			}
		}
		if inFrontmatter {
			if strings.HasPrefix(line, "parents:") {
				parentsLineIdx = i
				j := i + 1
				for j < len(lines) && strings.HasPrefix(lines[j], "  -") {
					j++
				}
				parentsEndIdx = j - 1
			}
			if strings.HasPrefix(line, "dependents:") {
				dependentsLineIdx = i
				j := i + 1
				for j < len(lines) && strings.HasPrefix(lines[j], "  -") {
					j++
				}
				dependentsEndIdx = j - 1
			}
		}
	}

	if frontmatterEnd == -1 {
		return false, nil
	}

	// Update parents block if a list was provided and a parents line exists.
	if parents != nil && parentsLineIdx != -1 {
		sortedParents := append([]string(nil), parents...)
		sort.Strings(sortedParents)
		newParentsLines := renderRefBlock("parents", sortedParents)

		tmp := append([]string{}, lines[:parentsLineIdx]...)
		tmp = append(tmp, newParentsLines...)
		tmp = append(tmp, lines[parentsEndIdx+1:]...)

		diff := len(newParentsLines) - (parentsEndIdx - parentsLineIdx + 1)
		if dependentsLineIdx > parentsLineIdx {
			dependentsLineIdx += diff
			dependentsEndIdx += diff
		}
		frontmatterEnd += diff
		lines = tmp
	}

	// Always rewrite the dependents block.
	newDepsLines := renderRefBlock("dependents", dependents)
	var finalLines []string
	if dependentsLineIdx != -1 {
		finalLines = append(finalLines, lines[:dependentsLineIdx]...)
		finalLines = append(finalLines, newDepsLines...)
		finalLines = append(finalLines, lines[dependentsEndIdx+1:]...)
	} else {
		finalLines = append(finalLines, lines[:frontmatterEnd]...)
		finalLines = append(finalLines, newDepsLines...)
		finalLines = append(finalLines, lines[frontmatterEnd:]...)
	}

	newContent := strings.Join(finalLines, "\n")
	if newContent == string(content) {
		return false, nil
	}
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return false, fmt.Errorf("failed to write %s: %v", path, err)
	}
	return true, nil
}

// renderRefBlock emits a YAML list block for a frontmatter ref field.
// Each entry is wrapped in `[[…]]`. Bare ids (no brackets) are accepted and
// wrapped automatically.
func renderRefBlock(field string, refs []string) []string {
	if len(refs) == 0 {
		return []string{field + ": []"}
	}
	out := make([]string, 0, len(refs)+1)
	out = append(out, field+":")
	for _, r := range refs {
		body := strings.Trim(r, "[]")
		out = append(out, fmt.Sprintf("  - [[%s]]", body))
	}
	return out
}

