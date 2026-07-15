package atom

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"atd-tools/config"
	"sort"
)

// UpdateOptions holds the parameters for updating an atom.
type UpdateOptions struct {
	FilePath       string
	SetArgs        []string
	Intent         string
	Logic          string
	Interface      string
	Expectation    string
	SpecLinkID     string
	SpecLinkFile   string
	// Force bypasses the STABLE+BUSINESS governance guard (see ATD.md:329).
	// Required to modify an existing atom whose current on-disk status is
	// STABLE and whose layer is BUSINESS.
	Force bool
}

// Update performs the update on an atom file.
func Update(opts UpdateOptions) (string, error) {
	if opts.SpecLinkID != "" && opts.SpecLinkFile != "" {
		if err := ApplySpecLink(opts.SpecLinkID, opts.SpecLinkFile); err != nil {
			return "", err
		}
		if opts.FilePath == "" {
			return fmt.Sprintf("Injected spec-link for %s into %s", opts.SpecLinkID, opts.SpecLinkFile), nil
		}
	}

	if opts.FilePath == "" {
		return "", fmt.Errorf("file path is required for atom updates")
	}

	intentText := resolveArg(opts.Intent)
	logicText := resolveArg(opts.Logic)
	interfaceText := resolveArg(opts.Interface)
	expectationText := resolveArg(opts.Expectation)

	// Read original file
	fileExists := true
	content, err := os.ReadFile(opts.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fileExists = false
			content = []byte("---\nid: temp\nstatus: DRAFT\n---\n\n# New Atom\n\n## INTENT\n\n## THE RULE / LOGIC\n\n## TECHNICAL INTERFACE\n\n## EXPECTATION\n")
		} else {
			return "", fmt.Errorf("error reading file: %v", err)
		}
	}
	if len(content) == 0 {
		fileExists = false
		content = []byte("---\nid: temp\nstatus: DRAFT\n---\n\n# New Atom\n\n## INTENT\n\n## THE RULE / LOGIC\n\n## TECHNICAL INTERFACE\n\n## EXPECTATION\n")
	}

	lines := strings.Split(string(content), "\n")

	// 1. Process Frontmatter
	inFrontmatter := false
	frontmatterEnd := -1
	var frontmatterLines []string

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
		if inFrontmatter && i > 0 {
			frontmatterLines = append(frontmatterLines, line)
		}
	}

	if frontmatterEnd == -1 {
		return "", fmt.Errorf("could not find valid YAML frontmatter in %s", opts.FilePath)
	}

	// Governance guard: STABLE BUSINESS atoms require heavy human sign-off to
	// modify (ATD.md:329). Only applies to existing atoms whose CURRENT
	// on-disk state is already STABLE+BUSINESS; new-file creation and
	// DRAFT/REVIEW or non-BUSINESS atoms are unaffected.
	if fileExists && !opts.Force {
		currentStatus := frontmatterValue(frontmatterLines, "status")
		currentLayer := frontmatterValue(frontmatterLines, "layer")
		if strings.EqualFold(currentStatus, "STABLE") && strings.EqualFold(currentLayer, "BUSINESS") {
			currentID := frontmatterValue(frontmatterLines, "id")
			if currentID == "" {
				currentID = strings.TrimSuffix(filepath.Base(opts.FilePath), ".atom.md")
			}
			return "", fmt.Errorf("refusing to modify STABLE BUSINESS atom '%s' without confirmation: this atom requires human sign-off (ATD.md governance). Re-run with --force (CLI) or force:true (MCP) to override.", currentID)
		}
	}

	// Update frontmatter keys
	updates := make(map[string]string)
	var newID string

	for _, setArg := range opts.SetArgs {
		parts := strings.SplitN(setArg, "=", 2)
		if len(parts) == 2 {
			key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			updates[key] = val
			if key == "id" {
				newID = val
			}
		}
	}

	var newFrontmatter []string
	newFrontmatter = append(newFrontmatter, "---")
	
	// Enforce Naming Convention and Defaults
	atomType := updates["type"]
	if atomType == "" {
		for _, line := range frontmatterLines {
			if strings.HasPrefix(line, "type:") {
				atomType = strings.TrimSpace(strings.TrimPrefix(line, "type:"))
				break
			}
		}
	}

	if newID != "" && atomType != "" {
		prefix := strings.ToLower(atomType) + "_"
		if !strings.HasPrefix(newID, prefix) {
			parts := strings.FieldsFunc(newID, func(r rune) bool {
				return !unicode.IsLetter(r) && !unicode.IsNumber(r)
			})
			var snake string
			for i, p := range parts {
				if i > 0 {
					snake += "_"
				}
				snake += strings.ToLower(p)
			}
			newID = prefix + snake
			updates["id"] = newID
		}
	}

	// Mandatory metadata defaults
	if updates["version"] == "" {
		hasVersion := false
		for _, line := range frontmatterLines {
			if strings.HasPrefix(line, "version:") {
				hasVersion = true
				break
			}
		}
		if !hasVersion {
			updates["version"] = "1.0"
		}
	}
	if updates["status"] == "" {
		hasStatus := false
		for _, line := range frontmatterLines {
			if strings.HasPrefix(line, "status:") {
				hasStatus = true
				break
			}
		}
		if !hasStatus {
			updates["status"] = "DRAFT"
		}
	}
	
	hasParents := false
	hasDependents := false
	for _, line := range frontmatterLines {
		if strings.HasPrefix(line, "parents:") {
			hasParents = true
		}
		if strings.HasPrefix(line, "dependents:") {
			hasDependents = true
		}
	}
	if !hasParents && updates["parents"] == "" {
		updates["parents"] = "[]"
	}
	if !hasDependents && updates["dependents"] == "" {
		updates["dependents"] = "[]"
	}

	// Replace existing keys
	matchedKeys := make(map[string]bool)
	for i := 0; i < len(frontmatterLines); i++ {
		line := frontmatterLines[i]
		updated := false
		for k, v := range updates {
			prefix := k + ":"
			if strings.HasPrefix(line, prefix) {
				if k == "parents" || k == "dependents" {
					newFrontmatter = append(newFrontmatter, FormatYAMLList(k, v))
				} else {
					newFrontmatter = append(newFrontmatter, fmt.Sprintf("%s: %s", k, v))
				}
				matchedKeys[k] = true
				updated = true

				for i+1 < len(frontmatterLines) && strings.HasPrefix(frontmatterLines[i+1], "  -") {
					i++
				}
				break
			}
		}
		if !updated {
			newFrontmatter = append(newFrontmatter, line)
		}
	}

	// Append new keys
	for k, v := range updates {
		if !matchedKeys[k] {
			if k == "parents" || k == "dependents" {
				newFrontmatter = append(newFrontmatter, FormatYAMLList(k, v))
			} else {
				newFrontmatter = append(newFrontmatter, fmt.Sprintf("%s: %s", k, v))
			}
		}
	}
	newFrontmatter = append(newFrontmatter, "---")

	// 2. Process Body Sections
	var newBody []string
	currentSection := 0 // SecNone equivalent

	isTargetHeader := func(line string) int {
		if strings.HasPrefix(line, "## INTENT") { return 1 }
		if strings.HasPrefix(line, "## THE RULE / LOGIC") { return 2 }
		if strings.HasPrefix(line, "## TECHNICAL INTERFACE") { return 3 }
		if strings.HasPrefix(line, "## EXPECTATION") { return 4 }
		return 0
	}

	for i := frontmatterEnd + 1; i < len(lines); i++ {
		line := lines[i]
		if sec := isTargetHeader(line); sec != 0 {
			currentSection = sec
			newBody = append(newBody, line)
			switch sec {
			case 1: if intentText != "" { newBody = append(newBody, strings.TrimSpace(intentText), "") }
			case 2: if logicText != "" { newBody = append(newBody, strings.TrimSpace(logicText), "") }
			case 3: if interfaceText != "" { newBody = append(newBody, strings.TrimSpace(interfaceText), "") }
			case 4: if expectationText != "" { newBody = append(newBody, strings.TrimSpace(expectationText), "") }
			}
			continue
		} else if strings.HasPrefix(line, "## ") {
			currentSection = 0
		}

		skipLine := false
		switch currentSection {
		case 1: if intentText != "" { skipLine = true }
		case 2: if logicText != "" { skipLine = true }
		case 3: if interfaceText != "" { skipLine = true }
		case 4: if expectationText != "" { skipLine = true }
		}

		if !skipLine {
			if len(newBody) > 0 && newBody[len(newBody)-1] == "" && line == "" && currentSection != 0 {
				continue
			}
			newBody = append(newBody, line)
		}
	}

	finalOutput := strings.Join(append(newFrontmatter, newBody...), "\n")
	targetPath := opts.FilePath

	if newID != "" {
		dir := filepath.Dir(opts.FilePath)
		targetPath = filepath.Join(dir, fmt.Sprintf("%s.atom.md", newID))
	}

	if err := os.WriteFile(targetPath, []byte(finalOutput), 0644); err != nil {
		return "", fmt.Errorf("error writing file %s: %v", targetPath, err)
	}

	renamed := false
	if targetPath != opts.FilePath {
		os.Remove(opts.FilePath)
		renamed = true
	}

	logMsg := fmt.Sprintf("Updated %s", filepath.Base(targetPath))
	if len(updates) > 0 { logMsg += fmt.Sprintf(" | set: %d keys", len(updates)) }
	if intentText != "" || logicText != "" || interfaceText != "" || expectationText != "" {
		logMsg += " | updated body sections"
	}
	if renamed { logMsg += fmt.Sprintf(" | renamed from %s", filepath.Base(opts.FilePath)) }
	config.Log("atd-update", logMsg)

	if renamed && newID != "" {
		var oldID string
		for _, line := range frontmatterLines {
			if strings.HasPrefix(line, "id:") {
				oldID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
				break
			}
		}
		if oldID != "" && oldID != newID {
			docsPath := config.DocsDir()
			numUpdates := UpdateLinks(docsPath, oldID, newID)
			logMsg += fmt.Sprintf(" | Propagated ID change [[%s]] -> [[%s]] across %d files", oldID, newID, numUpdates)
		}
	}

	return "Success: " + logMsg, nil
}

// ApplySpecLink injects a spec-link tag into a source file.
func ApplySpecLink(id, file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", file, err)
	}

	tag := fmt.Sprintf("// @spec-link [[%s]]\n", id)
	if strings.Contains(string(content), tag) {
		return nil
	}

	updated := tag + string(content)
	if err := os.WriteFile(file, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", file, err)
	}
	return nil
}

// frontmatterValue returns the trimmed value for a given top-level frontmatter
// key (e.g. "status", "layer", "id") as found in the raw frontmatter lines,
// or "" if the key is not present.
func frontmatterValue(frontmatterLines []string, key string) string {
	prefix := key + ":"
	for _, line := range frontmatterLines {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func resolveArg(arg string) string {
	if arg == "-" {
		bytes, _ := io.ReadAll(os.Stdin)
		return string(bytes)
	}
	return strings.ReplaceAll(arg, "\\n", "\n")
}

// UpdateLinks propagates ID changes across the project.
func UpdateLinks(docsPath, oldID, newID string) int {
	updatedCount := 0
	oldLink := fmt.Sprintf("[[%s]]", oldID)
	newLink := fmt.Sprintf("[[%s]]", newID)
	escapedOld := regexp.QuoteMeta(oldID)
	re := regexp.MustCompile(`\[\[` + escapedOld + `\]\]`)

	replaceInFile := func(path string) {
		content, readErr := os.ReadFile(path)
		if readErr != nil { return }
		strContent := string(content)
		if strings.Contains(strContent, oldLink) || re.MatchString(strContent) {
			newContent := strings.ReplaceAll(strContent, oldLink, newLink)
			newContent = re.ReplaceAllString(newContent, newLink)
			if newContent != strContent {
				os.WriteFile(path, []byte(newContent), 0644)
				updatedCount++
			}
		}
	}

	filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".atom.md") { return nil }
		replaceInFile(path)
		return nil
	})

	// P-1 scope guard (test_atd_07_26.md §2.1/§3.6): the second walk below
	// propagates a rename across *source files*, trusting config.ProjectRoot()
	// blindly used to let a rename in project X rewrite files anywhere the
	// process cwd happened to resolve to (incident I-1: a leaked, cwd-fallback
	// config caused UpdateLinks to rewrite the test suite's own source). Two
	// independent checks must both pass before the source walk runs:
	//
	//  1. The active config must not be fallback-anchored (no real .atd found;
	//     ProjectRoot() is just the process cwd, which for a `go test` binary
	//     is the package source directory).
	//  2. docsPath (the scope this rename was actually asked to operate on)
	//     must resolve inside ProjectRoot() — a rename in project X must never
	//     rewrite files outside X.
	projectRoot := config.ProjectRoot()
	if projectRoot != "" {
		if config.LoadedFromFallback() {
			fmt.Fprintf(os.Stderr, "atd: refusing source-file rename propagation for [[%s]] -> [[%s]]: active config is cwd-fallback-anchored (no .atd found), not a genuine project root; skipping to avoid rewriting unrelated files\n", oldID, newID)
		} else if !pathInside(docsPath, projectRoot) {
			fmt.Fprintf(os.Stderr, "atd: skipping source-file rename propagation for [[%s]] -> [[%s]]: docs path %q is outside project root %q\n", oldID, newID, docsPath, projectRoot)
		} else {
			filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
				if err != nil { return nil }
				relPath, _ := filepath.Rel(projectRoot, path)
				if strings.HasPrefix(relPath, ".git") || strings.HasPrefix(relPath, ".atd") {
					if info.IsDir() { return filepath.SkipDir }
					return nil
				}
				// P-2: fixture literals in tests and testdata are not real
				// spec-links; never propagate renames into them.
				if info.IsDir() {
					if info.Name() == "testdata" { return filepath.SkipDir }
					return nil
				}
				if strings.HasSuffix(path, "_test.go") { return nil }
				if strings.HasSuffix(path, ".atom.md") { return nil }
				ext := filepath.Ext(path)
				switch ext {
				case ".go", ".py", ".js", ".ts", ".tsx", ".jsx", ".rs", ".rb", ".java",
					".c", ".cpp", ".h", ".hpp", ".cs", ".sh", ".yaml", ".yml", ".toml", ".md":
					replaceInFile(path)
				}
				return nil
			})
		}
	}
	return updatedCount
}

// pathInside reports whether target resolves to projectRoot itself or to a
// path nested inside it, using absolute, cleaned paths for the comparison.
func pathInside(target, projectRoot string) bool {
	absTarget, errT := filepath.Abs(target)
	absRoot, errR := filepath.Abs(projectRoot)
	if errT != nil || errR != nil {
		return false
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// FormatYAMLList formats a YAML list for frontmatter.
func FormatYAMLList(key, val string) string {
	val = strings.TrimSpace(val)
	if val == "" || val == "[]" {
		return fmt.Sprintf("%s: []", key)
	}

	var items []string
	if strings.Contains(val, "\n") {
		items = strings.Split(val, "\n")
	} else {
		val = strings.Trim(val, "[] ")
		items = strings.Split(val, ",")
	}

	var processed []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || item == "-" {
			continue
		}
		item = strings.Trim(item, "[]- ")
		if item == "" {
			continue
		}
		processed = append(processed, item)
	}
	sort.Strings(processed)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s:", key))
	for _, item := range processed {
		result.WriteString(fmt.Sprintf("\n  - [[%s]]", item))
	}
	return result.String()
}
