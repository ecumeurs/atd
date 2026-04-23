package exploration

import (
	"atd-tools/config"
	"fmt"
	"os"
	"sort"
	"strings"
)

// @spec-link [[mechanic_atd_weave]]
// Weave performs bi-directional link synchronization.
func (e *Explorer) Weave() (string, error) {
	if err := e.Load(false); err != nil {
		return "", err
	}

	parentToDependents := make(map[string][]string)

	// 1. Discover all declared parents
	for id, a := range e.Graph.Atoms {
		projectPrefix := ""
		if strings.Contains(id, ":") {
			projectPrefix = strings.Split(id, ":")[0]
		}

		for _, p := range a.Parents {
			normalizedParent := p
			if !strings.Contains(p, ":") && projectPrefix != "" {
				normalizedParent = projectPrefix + ":" + p
			}
			parentToDependents[normalizedParent] = append(parentToDependents[normalizedParent], id)
		}
	}

	// 2. Rewrite each file with updated dependents
	edited := 0

	for id, a := range e.Graph.Atoms {
		if a.FilePath == "" {
			continue
		}

		projectPrefix := ""
		if strings.Contains(id, ":") {
			projectPrefix = strings.Split(id, ":")[0]
		}

		content, err := os.ReadFile(a.FilePath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		var newLines []string
		inFrontmatter := false
		frontmatterEnd := -1
		parentsLineIdx := -1
		parentsEndIdx := -1
		dependentsLineIdx := -1
		dependentsEndIdx := -1

		// Find frontmatter and sections
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
			continue
		}

		// 3. Handle Parents sorting
		if parentsLineIdx != -1 {
			parents := a.Parents
			sort.Strings(parents)
			var newParentsLines []string
			if len(parents) == 0 {
				newParentsLines = append(newParentsLines, "parents: []")
			} else {
				newParentsLines = append(newParentsLines, "parents:")
				for _, p := range parents {
					newParentsLines = append(newParentsLines, fmt.Sprintf("  - [[%s]]", p))
				}
			}

			// Slice and replace parents
			tmpLines := append([]string{}, lines[:parentsLineIdx]...)
			tmpLines = append(tmpLines, newParentsLines...)
			tmpLines = append(tmpLines, lines[parentsEndIdx+1:]...)
			lines = tmpLines

			// Offset other indices if they were after parents
			diff := len(newParentsLines) - (parentsEndIdx - parentsLineIdx + 1)
			if dependentsLineIdx > parentsLineIdx {
				dependentsLineIdx += diff
				dependentsEndIdx += diff
			}
			frontmatterEnd += diff
		}

		// 4. Handle Dependents
		deps := parentToDependents[id]
		sort.Strings(deps)
		var newDepsLines []string
		if len(deps) == 0 {
			newDepsLines = append(newDepsLines, "dependents: []")
		} else {
			newDepsLines = append(newDepsLines, "dependents:")
			for _, dep := range deps {
				displayDep := dep
				if projectPrefix != "" && strings.HasPrefix(dep, projectPrefix+":") {
					displayDep = strings.TrimPrefix(dep, projectPrefix+":")
				}
				newDepsLines = append(newDepsLines, fmt.Sprintf("  - [[%s]]", displayDep))
			}
		}

		// Reconstruct file
		if dependentsLineIdx != -1 {
			// Replace existing dependents section
			newLines = append(newLines, lines[:dependentsLineIdx]...)
			newLines = append(newLines, newDepsLines...)
			newLines = append(newLines, lines[dependentsEndIdx+1:]...)
		} else {
			// Insert before the closing ---
			newLines = append(newLines, lines[:frontmatterEnd]...)
			newLines = append(newLines, newDepsLines...)
			newLines = append(newLines, lines[frontmatterEnd:]...)
		}

		newContent := strings.Join(newLines, "\n")
		if newContent != string(content) {
			if err := os.WriteFile(a.FilePath, []byte(newContent), 0644); err != nil {
				return "", fmt.Errorf("failed to write %s: %v", a.FilePath, err)
			}
			edited++
		}
	}

	result := fmt.Sprintf("Link Weaving Complete. Updated %d files.", edited)
	config.Log("atd-weave", result)

	// Since we edited files, we should probably invalidate the cache if we were to reload.
	// But for Weave, the IDs and Parents didn't change, only Dependents.
	// And Weave doesn't change what Explorer.Load() would find (it uses parents mostly).
	// However, to be safe:
	// e.Load(true) 
	// But e.Load(false) is fine because we just updated the files that we already have in memory.
	// We should update the in-memory graph too.
	for id, a := range e.Graph.Atoms {
		a.Dependents = parentToDependents[id]
	}

	return result, nil
}
