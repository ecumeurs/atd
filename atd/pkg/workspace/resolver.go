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
		return parsed, ErrAtomNotFound
	}

	// Local reference - check current project first
	if r.currentProject != "" {
		if loc := r.index.FindAtomInProject(r.currentProject, parsed.AtomID); loc != nil {
			parsed.Location = loc
			parsed.Type = ReferenceLocal
			return parsed, nil
		}
	}

	// Check shared/workspace root
	if loc := r.index.FindAtomInWorkspaceRoot(parsed.AtomID); loc != nil {
		parsed.Project = "shared"
		parsed.Type = ReferenceCrossProject
		parsed.Location = loc
		return parsed, nil
	}

	// Search other projects in workspace
	if loc := r.index.FindAtom(parsed.AtomID); loc != nil {
		parsed.Project = loc.Project
		parsed.Type = ReferenceCrossProject
		parsed.Location = loc
		return parsed, nil
	}

	parsed.Type = ReferenceUnresolved
	return parsed, ErrAtomNotFound
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
