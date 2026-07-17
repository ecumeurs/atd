package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// AtomLocation describes where an atom is located.
type AtomLocation struct {
	Project string
	Path    string
	Exists  bool
}

// AtomIndex maps atom IDs to their location across the workspace.
type AtomIndex struct {
	sync.RWMutex
	ByID       map[string]*AtomLocation // "atom_id" -> location
	ByProject  map[string][]string      // "project_name" -> ["atom1_id", "atom2_id"]
	Duplicates map[string][]string      // "atom_id" -> every non-shared project that declares it, in registration order; len > 1 means ambiguous
}

// BuildIndex scans all projects in the workspace and builds the atom index.
func (ws *Workspace) BuildIndex() (*AtomIndex, error) {
	idx := &AtomIndex{
		ByID:       make(map[string]*AtomLocation),
		ByProject:  make(map[string][]string),
		Duplicates: make(map[string][]string),
	}

	// 1. Index shared atoms in workspace root docs/ folder (if exists)
	rootDocsPath := filepath.Join(ws.WorkspaceRoot, "docs")
	if _, err := os.Stat(rootDocsPath); err == nil {
		atoms, _ := listAtoms(rootDocsPath)
		if len(atoms) > 0 {
			idx.ByProject["shared"] = atoms
			for _, atomID := range atoms {
				idx.ByID[atomID] = &AtomLocation{
					Project: "shared",
					Path:    filepath.Join(rootDocsPath, atomID+".atom.md"),
					Exists:  true,
				}
			}
		}
	}

	// 2. Index all projects
	for _, project := range ws.Projects {
		atoms, err := listAtoms(project.FullDocsPath)
		if err != nil {
			continue // Skip projects with missing docs
		}

		idx.ByProject[project.Name] = atoms
		for _, atomID := range atoms {
			// Track every project that declares this id (not just the first),
			// so the resolver can tell a clean match from an ambiguous one
			// instead of silently picking whichever project registered first.
			idx.Duplicates[atomID] = append(idx.Duplicates[atomID], project.Name)

			// If already exists (e.g. in shared), local project takes precedence for that project context?
			// For now, first one wins; ambiguity across projects is surfaced via
			// AmbiguousProjects/Resolver instead.
			if _, exists := idx.ByID[atomID]; !exists {
				idx.ByID[atomID] = &AtomLocation{
					Project: project.Name,
					Path:    filepath.Join(project.FullDocsPath, atomID+".atom.md"),
					Exists:  true,
				}
			}
		}
	}

	return idx, nil
}

// FindAtom searches the workspace for an atom by ID.
func (idx *AtomIndex) FindAtom(atomID string) *AtomLocation {
	idx.RLock()
	defer idx.RUnlock()
	return idx.ByID[atomID]
}

// AmbiguousProjects returns every project that declares atomID, in
// registration order, when more than one does -- an empty/nil slice means the
// id is unique (or absent) across the workspace's projects.
func (idx *AtomIndex) AmbiguousProjects(atomID string) []string {
	idx.RLock()
	defer idx.RUnlock()
	if projects := idx.Duplicates[atomID]; len(projects) > 1 {
		return projects
	}
	return nil
}

// FindAtomInProject finds an atom in a specific project.
func (idx *AtomIndex) FindAtomInProject(projectName, atomID string) *AtomLocation {
	idx.RLock()
	defer idx.RUnlock()

	loc, ok := idx.ByID[atomID]
	if ok && loc.Project == projectName {
		return loc
	}

	// Also check if it's in this project's list (in case of collisions where ByID only stored one)
	// Actually, better to have a more robust ByID if collisions are allowed.
	// For now, let's keep it simple.
	return nil
}

// FindAtomInWorkspaceRoot finds an atom in the shared/shared root docs folder.
func (idx *AtomIndex) FindAtomInWorkspaceRoot(atomID string) *AtomLocation {
	return idx.FindAtomInProject("shared", atomID)
}

func listAtoms(dir string) ([]string, error) {
	var atoms []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".atom.md") {
			atomID := strings.TrimSuffix(info.Name(), ".atom.md")
			atoms = append(atoms, atomID)
		}
		return nil
	})
	return atoms, err
}
