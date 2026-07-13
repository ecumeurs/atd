package workspace

import (
	"errors"
	"fmt"
	"strings"
)

// ReferenceType indicates if a reference is local or cross-project.
type ReferenceType int

const (
	ReferenceLocal ReferenceType = iota
	ReferenceCrossProject
	ReferenceUnresolved
)

var ErrAtomNotFound = errors.New("atom not found")
var ErrUnknownProject = errors.New("unknown project qualifier")

// knownIDPrefixes lists leading underscore-delimited tokens that are ATD
// type/layer words rather than part of an atom's actual id. When a bare id
// fails to resolve, Resolve retries once with such a leading token stripped
// (e.g. "requirement_req_ui_session_timeout" -> "req_ui_session_timeout").
var knownIDPrefixes = map[string]bool{
	"requirement": true,
	"rule":        true,
	"api":         true,
	"mech":        true,
	"mechanic":    true,
	"ui":          true,
	"entity":      true,
	"module":      true,
	"domain":      true,
	"us":          true,
	"usecase":     true,
	"user_story":  true,
	"vision":      true,
	"contract":    true,
	"workflow":    true,
}

// stripKnownPrefix removes a leading known type/layer token (e.g. "requirement_")
// from a bare atom id, returning the stripped id and true if a prefix was found.
func stripKnownPrefix(atomID string) (string, bool) {
	idx := strings.Index(atomID, "_")
	if idx <= 0 {
		return atomID, false
	}
	token := atomID[:idx]
	if !knownIDPrefixes[token] {
		return atomID, false
	}
	rest := atomID[idx+1:]
	if rest == "" {
		return atomID, false
	}
	return rest, true
}

// ParsedReference represents a parsed [[atom_id]] or [[project:atom_id]].
type ParsedReference struct {
	Original string
	Project  string // "" for local references
	AtomID   string
	Type     ReferenceType
	Location *AtomLocation
}

// Resolver resolves references across the workspace.
type Resolver struct {
	workspace      *Workspace
	index          *AtomIndex
	currentProject string // Context for local references
}

// NewResolver creates a new resolver.
func NewResolver(ws *Workspace, idx *AtomIndex, currentProject string) *Resolver {
	return &Resolver{
		workspace:      ws,
		index:          idx,
		currentProject: currentProject,
	}
}

// ParseReference parses a reference string like "[[atom_id]]" or "[[project:atom_id]]".
func ParseReference(ref string) *ParsedReference {
	content := strings.Trim(ref, "[]")

	if strings.Contains(content, ":") {
		parts := strings.SplitN(content, ":", 2)
		return &ParsedReference{
			Original: ref,
			Project:  parts[0],
			AtomID:   parts[1],
			Type:     ReferenceCrossProject,
		}
	}

	return &ParsedReference{
		Original: ref,
		Project:  "",
		AtomID:   content,
		Type:     ReferenceLocal,
	}
}

// Resolve resolves a reference to its canonical form with project prefix.
func (r *Resolver) Resolve(ref string) (*ParsedReference, error) {
	parsed := ParseReference(ref)

	// Handle special prefixes: parent and shared both point to workspace root docs/
	if parsed.Project == "parent" || parsed.Project == "shared" {
		loc := r.index.FindAtomInWorkspaceRoot(parsed.AtomID)
		if loc != nil {
			parsed.Project = "shared" // Normalize to shared
			parsed.Type = ReferenceCrossProject
			parsed.Location = loc
			return parsed, nil
		}
		return parsed, ErrAtomNotFound
	}

	// Already has a project prefix
	if parsed.Type == ReferenceCrossProject {
		loc := r.index.FindAtom(parsed.AtomID)
		// Verify project match
		if loc != nil && loc.Project == parsed.Project {
			parsed.Location = loc
			return parsed, nil
		}

		// Check if project actually exists in workspace
		projectExists := false
		if parsed.Project == "shared" || parsed.Project == "parent" {
			projectExists = true
		} else if r.workspace != nil {
			for _, p := range r.workspace.Projects {
				if p.Name == parsed.Project {
					projectExists = true
					break
				}
			}
		}

		if !projectExists {
			return parsed, ErrUnknownProject
		}

		return parsed, ErrAtomNotFound
	}

	// Local reference - check current project, shared root, then the rest of the workspace
	if loc, project, refType, ok := r.resolveBareID(parsed.AtomID); ok {
		parsed.Location = loc
		parsed.Type = refType
		parsed.Project = project
		return parsed, nil
	}

	// Retry once with a leading known type/layer token stripped, e.g.
	// "requirement_req_ui_session_timeout" -> "req_ui_session_timeout".
	if stripped, hasPrefix := stripKnownPrefix(parsed.AtomID); hasPrefix {
		if loc, project, refType, ok := r.resolveBareID(stripped); ok {
			parsed.AtomID = stripped
			parsed.Location = loc
			parsed.Type = refType
			parsed.Project = project
			return parsed, nil
		}
	}

	parsed.Type = ReferenceUnresolved
	return parsed, ErrAtomNotFound
}

// resolveBareID looks up a bare atom id across the current project, the
// workspace root (shared docs), and then every other project in the
// workspace. It returns the matched location, the project it classifies
// under ("" for local, "shared" or a project name for cross-project), the
// resulting reference type, and whether a match was found.
func (r *Resolver) resolveBareID(atomID string) (*AtomLocation, string, ReferenceType, bool) {
	if r.currentProject != "" {
		if loc := r.index.FindAtomInProject(r.currentProject, atomID); loc != nil {
			return loc, "", ReferenceLocal, true
		}
	}

	if loc := r.index.FindAtomInWorkspaceRoot(atomID); loc != nil {
		return loc, "shared", ReferenceCrossProject, true
	}

	if loc := r.index.FindAtom(atomID); loc != nil {
		return loc, loc.Project, ReferenceCrossProject, true
	}

	return nil, "", ReferenceUnresolved, false
}

// CanonicalForm returns the canonical [[project:atom_id]] format if it's cross-project.
func (r *Resolver) CanonicalForm(ref string) (string, error) {
	parsed, err := r.Resolve(ref)
	if err != nil {
		return ref, err
	}

	if parsed.Type == ReferenceLocal {
		return fmt.Sprintf("[[%s]]", parsed.AtomID), nil
	}
	return fmt.Sprintf("[[%s:%s]]", parsed.Project, parsed.AtomID), nil
}

// ShouldUpdate checks if a reference needs updating to include a project prefix.
func (r *Resolver) ShouldUpdate(ref string) (bool, string) {
	parsed := ParseReference(ref)

	// If it already has a project prefix, check if it's correct
	if parsed.Type == ReferenceCrossProject {
		loc := r.index.FindAtom(parsed.AtomID)
		if loc != nil && loc.Project != parsed.Project {
			// Found in a different project than specified? 
			// Usually we don't auto-fix this unless we are sure, 
			// but if it's not found in the specified project but found elsewhere:
			return true, fmt.Sprintf("[[%s:%s]]", loc.Project, parsed.AtomID)
		}
		return false, ref
	}

	// Local reference - check if it's actually in another project
	resolved, err := r.Resolve(ref)
	if err != nil {
		return false, ref
	}

	if resolved.Type == ReferenceCrossProject {
		return true, fmt.Sprintf("[[%s:%s]]", resolved.Project, resolved.AtomID)
	}

	return false, ref
}
