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

	// Governance atoms (CONTRACT/VISION) sit outside the ancestry graph
	// (ATD.md §1.4). Weave is the repair pass for that rule: a parents: entry
	// naming one is dropped, and a governance atom's own parents: are emptied.
	governanceIDs := make(map[string]bool)
	for id, a := range e.Graph.Atoms {
		if atom.IsGovernanceType(a.Type) {
			governanceIDs[id] = true
		}
	}

	var stripped []string
	parentToDependents := make(map[string][]string)

	for id, a := range e.Graph.Atoms {
		for _, p := range a.Parents {
			ref := strings.Trim(p, "[]")
			if governanceIDs[id] || governanceIDs[atom.BareAtomID(ref)] {
				// Not an edge: excluded here so the target never regains a
				// dependents: entry for it either.
				continue
			}
			parentToDependents[ref] = append(parentToDependents[ref], id)
		}
	}

	edited := 0
	var preserved []string
	finalDeps := make(map[string][]string, len(e.Graph.Atoms))
	for id, a := range e.Graph.Atoms {
		if a.FilePath == "" {
			continue
		}
		// Look up dependents by raw id only — we are explicitly NOT prefixing
		// in single-project mode.
		deps := append([]string(nil), parentToDependents[id]...)

		// A pre-existing dependents: entry naming another project
		// ("proj:atom_id") is outside what single-project mode can verify —
		// it never resolves any parent ref beyond this project's own graph
		// (see the "legacy path" doc comment above). Recomputing from scratch
		// would silently drop it, so it's carried over verbatim instead of
		// being treated as dead weight, and named in the result text.
		for _, d := range a.Dependents {
			if strings.Contains(d, ":") {
				deps = append(deps, d)
				preserved = append(preserved, fmt.Sprintf("%s dependents: [[%s]] -- cross-project reference outside single-project weave's scope, left untouched", id, d))
			}
		}
		deps = dedupeStrings(deps)
		sort.Strings(deps)
		finalDeps[id] = deps

		// Pass nil (leave parents untouched) unless this atom actually holds a
		// forbidden link — otherwise every atom's parents: would be re-sorted
		// and re-rendered on every weave.
		newParents, removed := stripGovernanceParents(a, governanceIDs)
		stripped = append(stripped, removed...)

		changed, err := rewriteAtomLinks(a.FilePath, newParents, deps, "" /* no project ctx */)
		if err != nil {
			return "", err
		}
		if changed {
			edited++
		}
		if newParents != nil {
			a.Parents = newParents
		}
	}

	// Refresh in-memory graph dependents.
	for id, a := range e.Graph.Atoms {
		a.Dependents = finalDeps[id]
	}

	result := fmt.Sprintf("Link Weaving Complete. Updated %d files.", edited)
	result += formatStrippedGovernanceLinks(stripped)
	result += formatPreservedDependentsLinks(preserved)
	config.Log("atd-weave", result)
	return result, nil
}

// dedupeStrings returns refs with duplicate entries removed, preserving the
// first occurrence's order (callers sort afterward).
func dedupeStrings(refs []string) []string {
	seen := make(map[string]bool, len(refs))
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		if seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, r)
	}
	return out
}

// stripGovernanceParents returns the atom's parents: with every forbidden
// governance link removed (all of them, if the atom is itself a governance
// atom), plus a human-readable note per removal. It returns a nil slice when
// nothing is forbidden, which callers pass through to rewriteAtomLinks to mean
// "leave the parents block exactly as authored".
// @spec-link [[rule_atd_governance_graph_isolation]]
func stripGovernanceParents(a *atom.AtomData, governanceIDs map[string]bool) (parents []string, removed []string) {
	self := atom.IsGovernanceType(a.Type)
	kept := make([]string, 0, len(a.Parents))
	for _, p := range a.Parents {
		ref := strings.Trim(strings.TrimSpace(p), "[]")
		switch {
		case self:
			removed = append(removed, fmt.Sprintf("%s (%s) parents: [[%s]] -- a governance atom declares no parents", a.ID, strings.ToUpper(strings.TrimSpace(a.Type)), ref))
		case governanceIDs[atom.BareAtomID(ref)]:
			removed = append(removed, fmt.Sprintf("%s parents: [[%s]] -- governance atoms are never structural ancestry", a.ID, ref))
		default:
			kept = append(kept, p)
		}
	}
	if len(removed) == 0 {
		return nil, nil
	}
	return kept, removed
}

// formatStrippedGovernanceLinks renders the removal notes appended to weave's
// result text. Weave rewrites files in place, so every dropped link is named
// rather than silently discarded.
func formatStrippedGovernanceLinks(stripped []string) string {
	if len(stripped) == 0 {
		return ""
	}
	sort.Strings(stripped)
	var b strings.Builder
	fmt.Fprintf(&b, "\nRemoved %d forbidden governance link(s) (ATD.md §1.4):", len(stripped))
	for _, s := range stripped {
		fmt.Fprintf(&b, "\n  - %s", s)
	}
	return b.String()
}

// formatPreservedDependentsLinks renders the notes for dependents: entries
// weave carried over verbatim instead of recomputing, because it could not
// verify them this run. Weave rebuilds dependents: from resolvable edges only;
// without this, anything it failed to re-derive would be dropped silently, so
// each kept-but-unverified entry is named instead.
func formatPreservedDependentsLinks(preserved []string) string {
	if len(preserved) == 0 {
		return ""
	}
	sort.Strings(preserved)
	var b strings.Builder
	fmt.Fprintf(&b, "\nKept %d dependents link(s) weave could not verify this run (left in place, not re-derived):", len(preserved))
	for _, s := range preserved {
		fmt.Fprintf(&b, "\n  - %s", s)
	}
	return b.String()
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
	// Canonical refs of governance atoms across the whole workspace: a
	// CONTRACT/VISION in any project is graph-isolated, including from
	// cross-project parents: entries (ATD.md §1.4).
	governanceCanonical := make(map[string]bool)

	for atomID, loc := range idx.ByID {
		a, err := atom.Parse(loc.Path)
		if err != nil {
			continue
		}
		canonical := fmt.Sprintf("[[%s:%s]]", loc.Project, atomID)
		atomsByCanonical[canonical] = atomCtx{Loc: loc, AtomID: atomID, Parsed: &a}
		if atom.IsGovernanceType(a.Type) {
			governanceCanonical[canonical] = true
		}
	}

	// Second pass: the parent→dependents map must skip forbidden governance
	// edges, which requires the full governance set from the pass above.
	for canonical, ctx := range atomsByCanonical {
		if governanceCanonical[canonical] {
			continue // its parents: are stripped below; it contributes no edges
		}
		for _, pRef := range ctx.Parsed.Parents {
			resolved, err := resolver.Resolve(pRef)
			if err != nil || resolved == nil {
				continue
			}
			parentCanonical := fmt.Sprintf("[[%s:%s]]", resolved.Project, resolved.AtomID)
			if governanceCanonical[parentCanonical] {
				continue
			}
			parentToDependents[parentCanonical] = append(parentToDependents[parentCanonical], canonical)
		}
	}

	// 2. Rewrite each atom file: canonicalize parents, refresh dependents.
	edited := 0
	var stripped []string
	var preserved []string
	for canonical, ctx := range atomsByCanonical {
		a := ctx.Parsed
		loc := ctx.Loc

		// Per-atom resolver so bare same-project parents stay canonical.
		projResolver := workspace.NewResolver(ws, idx, loc.Project)

		selfGovernance := governanceCanonical[canonical]
		newParents := make([]string, 0, len(a.Parents))
		for _, pRef := range a.Parents {
			// Governance graph isolation (ATD.md §1.4): drop the link rather
			// than canonicalize it. A governance atom loses all of its parents;
			// any atom loses the ones pointing at a governance atom.
			if selfGovernance {
				stripped = append(stripped, fmt.Sprintf("%s:%s (%s) parents: [[%s]] -- a governance atom declares no parents", loc.Project, ctx.AtomID, strings.ToUpper(strings.TrimSpace(a.Type)), strings.Trim(strings.TrimSpace(pRef), "[]")))
				continue
			}
			if resolved, err := resolver.Resolve(pRef); err == nil && resolved != nil {
				if governanceCanonical[fmt.Sprintf("[[%s:%s]]", resolved.Project, resolved.AtomID)] {
					stripped = append(stripped, fmt.Sprintf("%s:%s parents: [[%s]] -- governance atoms are never structural ancestry", loc.Project, ctx.AtomID, strings.Trim(strings.TrimSpace(pRef), "[]")))
					continue
				}
			}
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

		// A pre-existing dependents: entry that this run's resolver can't
		// resolve is NOT proof the edge is dead -- it may just mean the
		// declaring project isn't in scope this run, or the reference is
		// otherwise unverifiable right now. Recomputing from scratch would
		// silently drop it (the original tool-bug shape), so any such entry
		// is carried over verbatim and named in the result text instead.
		if selfGovernance {
			// A governance atom is graph-isolated (ATD.md §1.4): it never
			// legitimately holds dependents, so nothing here is worth
			// preserving -- an entry on it would itself be the bug, not a
			// resolve failure.
		} else {
			for _, d := range a.Dependents {
				if _, err := projResolver.Resolve(d); err == nil {
					continue // resolvable: the recomputed list above is authoritative for it
				}
				newDeps = append(newDeps, fmt.Sprintf("[[%s]]", d))
				preserved = append(preserved, fmt.Sprintf("%s:%s dependents: [[%s]] -- could not verify this run, left in place", loc.Project, ctx.AtomID, d))
			}
			newDeps = dedupeStrings(newDeps)
			sort.Strings(newDeps)
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
	result += formatStrippedGovernanceLinks(stripped)
	result += formatPreservedDependentsLinks(preserved)
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

