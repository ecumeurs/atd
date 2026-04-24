package cmd
// @spec-link [[mechanic_atd_weave]]

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/workspace"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var weaveCmd = &cobra.Command{
	Use:   "weave",
	Short: "Bi-directionally link ATD atoms based on parent declarations",
	Long: `Crawl all ATD atoms to discover parent relationships and 
automatically update the 'dependents' field in each atom file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		wsFlag, _ := cmd.Flags().GetBool("workspace")

		// Load workspace if available
		ws, err := workspace.LoadWorkspace(".")
		isWorkspace := err == nil

		if !isWorkspace && wsFlag {
			return fmt.Errorf("no workspace found but --workspace flag used")
		}

		if isWorkspace {
			return runWorkspaceWeave(ws)
		}

		// Fallback to local weave
		if docsDir == "" {
			docsDir = config.DocsDir()
		}

		text, err := runWeave(docsDir, false)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func runWorkspaceWeave(ws *workspace.Workspace) error {
	idx, err := ws.BuildIndex()
	if err != nil {
		return err
	}

	// Unified weave across workspace
	resolver := workspace.NewResolver(ws, idx, "")
	
	parentToDependents := make(map[string][]string)
	
	// 1. Discover all declared parents across all projects
	for id, loc := range idx.ByID {
		a, err := atom.Parse(loc.Path)
		if err != nil {
			continue
		}
		
		// Fix references in this atom's parents
		for _, pRef := range a.Parents {
			resolved, err := resolver.Resolve(pRef)
			if err == nil {
				canonical := pRef
				if resolved.Type == workspace.ReferenceCrossProject {
					canonical = fmt.Sprintf("[[%s:%s]]", resolved.Project, resolved.AtomID)
				} else {
					// Even local parents should be canonicalized in the map to find them
					canonical = fmt.Sprintf("[[%s:%s]]", loc.Project, resolved.AtomID)
				}
				
				// Dependent ID should also be canonical
				depID := id
				if loc.Project != "" && loc.Project != "shared" {
					depID = fmt.Sprintf("%s:%s", loc.Project, id)
				}
				parentToDependents[canonical] = append(parentToDependents[canonical], depID)
			}
		}
	}

	// 2. Update each atom file with its discovered dependents and fixed parents
	editedTotal := 0
	for atomID, loc := range idx.ByID {
		a, err := atom.Parse(loc.Path)
		if err != nil {
			continue
		}

		projResolver := workspace.NewResolver(ws, idx, loc.Project)
		
		// Update parents with canonical forms if needed
		parentsChanged := false
		newParents := make([]string, 0, len(a.Parents))
		for _, pRef := range a.Parents {
			shouldUpdate, canonical := projResolver.ShouldUpdate(pRef)
			if shouldUpdate {
				newParents = append(newParents, canonical)
				parentsChanged = true
			} else {
				newParents = append(newParents, pRef)
			}
		}
		
		// Discovery dependents for this atom
		// We need to check both local ID and prefixed ID in parentToDependents
		myCanonical := fmt.Sprintf("[[%s:%s]]", loc.Project, atomID)
		if loc.Project == "shared" {
			myCanonical = fmt.Sprintf("[[shared:%s]]", atomID)
		}
		
		deps := parentToDependents[myCanonical]
		// Also check local if some project still has un-prefixed references
		// Actually, parentToDependents is built from resolved references, so it should be canonical.
		
		sort.Strings(deps)
		
		// Compare with current dependents
		dependentsChanged := false
		if len(deps) != len(a.Dependents) {
			dependentsChanged = true
		} else {
			sort.Strings(a.Dependents)
			for i := range deps {
				// Normalize current dependents for comparison
				currDep := a.Dependents[i]
				if !strings.Contains(currDep, ":") && loc.Project != "" && loc.Project != "shared" {
					// Local dep, should we prefix it for comparison?
					// For now, let's just see if the canonical forms match.
				}
				if deps[i] != a.Dependents[i] {
					dependentsChanged = true
					break
				}
			}
		}

		if parentsChanged || dependentsChanged {
			// Update the atom file
			// We'll use a simplified version of the update logic here for now, 
			// or better, use atom.Update if possible.
			// However, atom.Update is designed for CLI args.
			
			// Let's use the logic from exploration/weave.go to update the file
			if err := updateAtomFile(loc.Path, newParents, deps, loc.Project); err == nil {
				editedTotal++
			}
		}
	}

	fmt.Printf("Workspace Link Weaving Complete. Updated %d files.\n", editedTotal)
	return nil
}

func updateAtomFile(path string, parents, dependents []string, currentProject string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
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
			} else if inFrontmatter {
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
		return fmt.Errorf("no frontmatter")
	}

	// 1. Update Parents
	if parentsLineIdx != -1 {
		sort.Strings(parents)
		var newParentsLines []string
		if len(parents) == 0 {
			newParentsLines = append(newParentsLines, "parents: []")
		} else {
			newParentsLines = append(newParentsLines, "parents:")
			for _, p := range parents {
				newParentsLines = append(newParentsLines, fmt.Sprintf("  - [[%s]]", strings.Trim(p, "[]")))
			}
		}
		
		tmpLines := append([]string{}, lines[:parentsLineIdx]...)
		tmpLines = append(tmpLines, newParentsLines...)
		tmpLines = append(tmpLines, lines[parentsEndIdx+1:]...)
		
		diff := len(newParentsLines) - (parentsEndIdx - parentsLineIdx + 1)
		if dependentsLineIdx > parentsLineIdx {
			dependentsLineIdx += diff
			dependentsEndIdx += diff
		}
		frontmatterEnd += diff
		lines = tmpLines
	}

	// 2. Update Dependents
	var newDepsLines []string
	if len(dependents) == 0 {
		newDepsLines = append(newDepsLines, "dependents: []")
	} else {
		newDepsLines = append(newDepsLines, "dependents:")
		for _, dep := range dependents {
			displayDep := dep
			if currentProject != "" && currentProject != "shared" && strings.HasPrefix(dep, currentProject+":") {
				displayDep = strings.TrimPrefix(dep, currentProject+":")
			}
			newDepsLines = append(newDepsLines, fmt.Sprintf("  - [[%s]]", displayDep))
		}
	}

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

	return os.WriteFile(path, []byte(strings.Join(finalLines, "\n")), 0644)
}

func runWeave(docsDir string, workspace bool) (string, error) {
	explorer := exploration.NewExplorer(config.ProjectRoot(), docsDir)
	if workspace {
		if err := explorer.LoadWorkspace(false); err != nil {
			return "", err
		}
	}
	return explorer.Weave()
}

func init() {
	rootCmd.AddCommand(weaveCmd)
	weaveCmd.Flags().String("docs", "", "Override docs directory")
	weaveCmd.Flags().Bool("workspace", false, "Weave entire workspace")
}
