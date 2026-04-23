package atom

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

// AtomData holds parsed atom metadata and content sections.
type AtomData struct {
	ID          string   `json:"id"`
	HumanName   string   `json:"human_name"`
	Type        string   `json:"type"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	Layer       string   `json:"layer"`
	Tags        []string `json:"tags"`
	Version     string   `json:"version"`
	Parents     []string `json:"parents"`
	Dependents  []string `json:"dependents"`
	Intent      string   `json:"intent"`      // Content of ## INTENT section
	Logic       string   `json:"logic"`       // Content of ## THE RULE / LOGIC section
	Interface   string   `json:"interface"`   // Content of ## TECHNICAL INTERFACE section
	Expectation string   `json:"expectation"` // Content of ## EXPECTATION section
	FilePath    string   `json:"filepath"`    // Original file path
	Implementations []string `json:"linked_codes"` // Linked source files
	HasTests    bool              `json:"has_tests"`    // Detection of @test-link tags
	Metadata    map[string]string `json:"metadata"`     // NEW: For project tagging
}

// GetProject returns the project name from metadata if present.
func (a *AtomData) GetProject() string {
	if a.Metadata != nil {
		return a.Metadata["project"]
	}
	return ""
}

// Parse reads a full .atom.md file and returns all metadata + content sections.
func Parse(path string) (AtomData, error) {
	f, err := os.Open(path)
	if err != nil {
		return AtomData{}, err
	}
	defer f.Close()

	var data AtomData
	data.FilePath = path
	scanner := bufio.NewScanner(f)
	mode := "header"
	inParents := false
	inDependents := false

	for scanner.Scan() {
		line := scanner.Text()

		// Metadata parsing
		if mode == "header" {
			trimmed := strings.TrimSpace(line)
			if trimmed == "---" {
				// We don't change mode here, just skip delimiters
				continue
			}

			if strings.HasPrefix(line, "id:") {
				data.ID = stripQuotes(strings.TrimPrefix(line, "id:"))
				inParents = false
				continue
			}
			if strings.HasPrefix(line, "human_name:") {
				data.HumanName = stripQuotes(strings.TrimPrefix(line, "human_name:"))
				inParents = false
				continue
			}
			if strings.HasPrefix(line, "type:") {
				data.Type = stripQuotes(strings.TrimPrefix(line, "type:"))
				inParents = false
				continue
			}
			if strings.HasPrefix(line, "layer:") {
				data.Layer = stripQuotes(strings.TrimPrefix(line, "layer:"))
				inParents = false
				continue
			}
			if strings.HasPrefix(line, "version:") {
				data.Version = stripQuotes(strings.TrimPrefix(line, "version:"))
				inParents = false
				continue
			}
			if strings.HasPrefix(line, "status:") {
				data.Status = stripQuotes(strings.TrimPrefix(line, "status:"))
				inParents = false
				continue
			}
			if strings.HasPrefix(line, "priority:") {
				data.Priority = stripQuotes(strings.TrimPrefix(line, "priority:"))
				inParents = false
				continue
			}
			if strings.HasPrefix(line, "tags:") {
				tagsStr := strings.TrimSpace(strings.TrimPrefix(line, "tags:"))
				tagsStr = strings.Trim(tagsStr, "[]")
				for _, t := range strings.Split(tagsStr, ",") {
					if tt := strings.TrimSpace(t); tt != "" {
						data.Tags = append(data.Tags, tt)
					}
				}
				inParents = false
				continue
			}
			if strings.HasPrefix(line, "parents:") {
				inParents = true
				inDependents = false
				inline := strings.TrimSpace(strings.TrimPrefix(line, "parents:"))
				if inline != "" && inline != "[]" {
					inline = strings.ReplaceAll(inline, "[", "")
					inline = strings.ReplaceAll(inline, "]", "")
					inline = strings.ReplaceAll(inline, "\"", "")
					for _, p := range strings.Split(inline, ",") {
						if t := strings.TrimSpace(p); t != "" {
							data.Parents = append(data.Parents, t)
						}
					}
				}
				continue
			}
			if strings.HasPrefix(line, "dependents:") {
				inParents = false
				inDependents = true
				inline := strings.TrimSpace(strings.TrimPrefix(line, "dependents:"))
				if inline != "" && inline != "[]" {
					inline = strings.ReplaceAll(inline, "[", "")
					inline = strings.ReplaceAll(inline, "]", "")
					inline = strings.ReplaceAll(inline, "\"", "")
					for _, d := range strings.Split(inline, ",") {
						if t := strings.TrimSpace(d); t != "" {
							data.Dependents = append(data.Dependents, t)
						}
					}
				}
				continue
			}
			if inParents || inDependents {
				trimmedLine := strings.TrimSpace(line)
				if strings.HasPrefix(trimmedLine, "- ") {
					entry := strings.TrimPrefix(trimmedLine, "- ")
					entry = strings.ReplaceAll(entry, "[", "")
					entry = strings.ReplaceAll(entry, "]", "")
					entry = strings.ReplaceAll(entry, "\"", "")
					if t := strings.TrimSpace(entry); t != "" {
						if inParents {
							data.Parents = append(data.Parents, t)
						} else {
							data.Dependents = append(data.Dependents, t)
						}
					}
					continue
				}
				inParents = false
				inDependents = false
			}

			// If we hit a header, switch to section parsing
			if strings.HasPrefix(line, "##") {
				mode = "sections"
			}
		}

		// Section parsing
		if mode == "sections" || mode == "intent" || mode == "logic" || mode == "interface" || mode == "expectation" {
			if strings.HasPrefix(line, "## INTENT") {
				mode = "intent"
				continue
			}
			if strings.HasPrefix(line, "## THE RULE") || strings.HasPrefix(line, "## LOGIC") {
				mode = "logic"
				continue
			}
			if strings.HasPrefix(line, "## TECHNICAL INTERFACE") {
				mode = "interface"
				continue
			}
			if strings.HasPrefix(line, "## EXPECTATION") {
				mode = "expectation"
				continue
			}
			if strings.HasPrefix(line, "##") {
				mode = "sections" // some other section
				continue
			}

			switch mode {
			case "intent":
				data.Intent += line + "\n"
			case "logic":
				data.Logic += line + "\n"
			case "interface":
				data.Interface += line + "\n"
			case "expectation":
				data.Expectation += line + "\n"
			}
		}
	}
	data.Intent = strings.TrimSpace(data.Intent)
	data.Logic = strings.TrimSpace(data.Logic)
	data.Interface = strings.TrimSpace(data.Interface)
	data.Expectation = strings.TrimSpace(data.Expectation)

	return data, scanner.Err()
}

// ParseMeta reads only the YAML frontmatter (fast, no body parsing).
func ParseMeta(path string) (id, humanName, atomType string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "id:") {
			id = stripQuotes(strings.TrimPrefix(line, "id:"))
		} else if strings.HasPrefix(line, "human_name:") {
			humanName = stripQuotes(strings.TrimPrefix(line, "human_name:"))
		} else if strings.HasPrefix(line, "type:") {
			atomType = stripQuotes(strings.TrimPrefix(line, "type:"))
		}
		if id != "" && humanName != "" && atomType != "" {
			break
		}
	}
	return
}

// BuildContent generates a complete .atom.md file from an AtomData struct.
func BuildContent(a AtomData) string {
	var parentsStr strings.Builder
	if len(a.Parents) == 0 {
		parentsStr.WriteString(" []")
	} else {
		sort.Strings(a.Parents)
		for _, p := range a.Parents {
			parentsStr.WriteString(fmt.Sprintf("  - [[%s]]\n", p))
		}
	}

	return fmt.Sprintf(`---
id: %s
human_name: %s
type: %s
layer: %s
version: %s
status: %s
priority: %s
tags: []
parents: %s
dependents: []
---

# %s

## INTENT
%s

## THE RULE / LOGIC
%s

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `+"`"+`@spec-link [[%s]]`+"`"+`

## EXPECTATION
%s
`, a.ID, a.HumanName, a.Type, a.Layer, a.Version, a.Status, a.Priority, strings.TrimRight(parentsStr.String(), "\n"), a.HumanName, a.Intent, a.Logic, a.ID, a.Expectation)
}

// BuildParentContent generates a MODULE parent .atom.md.
func BuildParentContent(id, humanName, intent, logic string) string {
	return fmt.Sprintf(`---
id: %s
human_name: %s
type: MODULE
version: 1.0
status: STABLE
priority: CORE
tags: []
parents: []
dependents: []
---

# %s

## INTENT
%s

## THE RULE / LOGIC
%s

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `+"`"+`@spec-link [[%s]]`+"`"+`
`, id, humanName, humanName, intent, logic, id)
}
