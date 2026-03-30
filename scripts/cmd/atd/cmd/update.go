package cmd
// @spec-link [[mechanic_atd_update]]

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"atd-tools/config"
	"github.com/spf13/cobra"
)

// Section indices
const (
	SecNone = iota
	SecIntent
	SecLogic
	SecInterface
	SecExpectation
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an ATD atom file's frontmatter or body sections",
	Long: `Update an ATD atom file's frontmatter or body sections.
Supports setting individual frontmatter keys, updating content sections (Intent, Rule/Logic, Technical Interface),
and prepending @spec-link tags to source files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		filterArg, _ := cmd.Flags().GetString("filter")
		intentArg, _ := cmd.Flags().GetString("intent")
		logicArg, _ := cmd.Flags().GetString("logic")
		interfaceArg, _ := cmd.Flags().GetString("interface")
		expectationArg, _ := cmd.Flags().GetString("expectation")
		setArgs, _ := cmd.Flags().GetStringSlice("set")
		specLinkArgs, _ := cmd.Flags().GetStringSlice("spec-link")

		var specLinkID, specLinkFile string
		if len(specLinkArgs) == 2 {
			specLinkID = specLinkArgs[0]
			specLinkFile = specLinkArgs[1]
		} else if len(specLinkArgs) > 0 {
			return fmt.Errorf("--spec-link requires exactly two arguments: <id> <file>")
		}

		if filterArg != "" {
			if filePath != "" {
				return fmt.Errorf("cannot use both --file and --filter")
			}
			text, err := runBatchUpdate(filterArg, setArgs, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile)
			if err != nil {
				return err
			}
			fmt.Println(text)
			return nil
		}

		text, err := runUpdate(filePath, setArgs, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile)
		if err != nil {
			return err
		}
		fmt.Println(text)
		return nil
	},
}

func runBatchUpdate(filter string, setArgs []string, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile string) (string, error) {
	filters := make(map[string]string)
	for _, part := range strings.Split(filter, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			filters[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	if len(filters) == 0 {
		return "", fmt.Errorf("invalid filter format, expected key=value,key2=value2")
	}

	docsDir := config.DocsDir()
	yamlRegex := regexp.MustCompile(`(?s)^---[\r\n]+(.*?)[\r\n]+---`)
	var targetFiles []string

	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".atom.md") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		yamlMatch := yamlRegex.FindStringSubmatch(string(content))
		if len(yamlMatch) > 1 {
			frontmatter := yamlMatch[1]
			matchesAll := true
			for k, v := range filters {
				fieldRegex := regexp.MustCompile(fmt.Sprintf(`(?m)^%s:\s*\[?(.*?)\]?$`, regexp.QuoteMeta(k)))
				fieldMatch := fieldRegex.FindStringSubmatch(frontmatter)
				if len(fieldMatch) > 1 {
					if strings.ToLower(strings.TrimSpace(fieldMatch[1])) != strings.ToLower(v) {
						matchesAll = false
						break
					}
				} else {
					matchesAll = false
					break
				}
			}
			if matchesAll {
				targetFiles = append(targetFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error finding matching files: %v", err)
	}

	if len(targetFiles) == 0 {
		return "No atoms matched the given filter.", nil
	}

	var results []string
	successCount := 0
	errorCount := 0

	for _, file := range targetFiles {
		res, err := runUpdate(file, setArgs, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile)
		if err != nil {
			results = append(results, fmt.Sprintf("Error updating %s: %v", file, err))
			errorCount++
		} else {
			results = append(results, res)
			successCount++
		}
	}

	summary := fmt.Sprintf("\nBatch complete: %d updated, %d errors out of %d matching atoms.", successCount, errorCount, len(targetFiles))
	return strings.Join(results, "\n") + summary, nil
}

func runUpdate(filePath string, setArgs []string, intentArg, logicArg, interfaceArg, expectationArg, specLinkID, specLinkFile string) (string, error) {
	if specLinkID != "" && specLinkFile != "" {
		if err := applySpecLink(specLinkID, specLinkFile); err != nil {
			return "", err
		}
		if filePath == "" {
			return fmt.Sprintf("Injected spec-link for %s into %s", specLinkID, specLinkFile), nil
		}
	}

	if filePath == "" {
		return "", fmt.Errorf("-file parameter is required for atom updates")
	}

	intentText := resolveArg(intentArg)
	logicText := resolveArg(logicArg)
	interfaceText := resolveArg(interfaceArg)
	expectationText := resolveArg(expectationArg)

	// Read original file
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			content = []byte("---\nid: temp\nstatus: DRAFT\n---\n\n# New Atom\n\n## INTENT\n\n## THE RULE / LOGIC\n\n## TECHNICAL INTERFACE\n\n## EXPECTATION\n")
		} else {
			return "", fmt.Errorf("error reading file: %v", err)
		}
	}
	if len(content) == 0 {
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
		return "", fmt.Errorf("could not find valid YAML frontmatter in %s", filePath)
	}

	// Update frontmatter keys
	updates := make(map[string]string)
	var newID string

	for _, setArg := range setArgs {
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
	// 1a. Enforce Naming Convention and Defaults
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
		// Enforce <type>_<snake_case_name>
		prefix := strings.ToLower(atomType) + "_"
		if !strings.HasPrefix(newID, prefix) {
			// Convert to snake_case
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
	// We handle parents/dependents if they are missing entirely
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
					newFrontmatter = append(newFrontmatter, formatYAMLList(k, v))
				} else {
					newFrontmatter = append(newFrontmatter, fmt.Sprintf("%s: %s", k, v))
				}
				matchedKeys[k] = true
				updated = true

				// Skip multi-line content (e.g. lists)
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
				newFrontmatter = append(newFrontmatter, formatYAMLList(k, v))
			} else {
				newFrontmatter = append(newFrontmatter, fmt.Sprintf("%s: %s", k, v))
			}
		}
	}
	newFrontmatter = append(newFrontmatter, "---")

	// 2. Process Body Sections
	var newBody []string
	currentSection := SecNone

	isTargetHeader := func(line string) int {
		if strings.HasPrefix(line, "## INTENT") {
			return SecIntent
		}
		if strings.HasPrefix(line, "## THE RULE / LOGIC") {
			return SecLogic
		}
		if strings.HasPrefix(line, "## EXPECTATION") {
			return SecExpectation
		}
		return SecNone
	}

	isAnyHeader := func(line string) bool {
		return strings.HasPrefix(line, "## ")
	}

	for i := frontmatterEnd + 1; i < len(lines); i++ {
		line := lines[i]

		if sec := isTargetHeader(line); sec != SecNone {
			currentSection = sec
			newBody = append(newBody, line)

			if sec == SecIntent && intentText != "" {
				newBody = append(newBody, strings.TrimSpace(intentText))
				newBody = append(newBody, "")
			} else if sec == SecLogic && logicText != "" {
				newBody = append(newBody, strings.TrimSpace(logicText))
				newBody = append(newBody, "")
			} else if sec == SecInterface && interfaceText != "" {
				newBody = append(newBody, strings.TrimSpace(interfaceText))
				newBody = append(newBody, "")
			} else if sec == SecExpectation && expectationText != "" {
				newBody = append(newBody, strings.TrimSpace(expectationText))
				newBody = append(newBody, "")
			}
			continue
		} else if isAnyHeader(line) {
			currentSection = SecNone
		}

		skipLine := false
		if currentSection == SecIntent && intentText != "" {
			skipLine = true
		}
		if currentSection == SecLogic && logicText != "" {
			skipLine = true
		}
		if currentSection == SecExpectation && expectationText != "" {
			skipLine = true
		}

		if !skipLine {
			if len(newBody) > 0 && newBody[len(newBody)-1] == "" && line == "" && currentSection != SecNone {
				continue
			}
			newBody = append(newBody, line)
		}
	}

	finalOutput := strings.Join(append(newFrontmatter, newBody...), "\n")
	targetPath := filePath

	if newID != "" {
		dir := filepath.Dir(filePath)
		targetPath = filepath.Join(dir, fmt.Sprintf("%s.atom.md", newID))
	}

	if err := os.WriteFile(targetPath, []byte(finalOutput), 0644); err != nil {
		return "", fmt.Errorf("error writing file %s: %v", targetPath, err)
	}

	renamed := false
	if targetPath != filePath {
		os.Remove(filePath)
		renamed = true
	}

	logMsg := fmt.Sprintf("Updated %s", filepath.Base(targetPath))
	if len(updates) > 0 {
		logMsg += fmt.Sprintf(" | set: %d keys", len(updates))
	}
	if intentText != "" || logicText != "" || interfaceText != "" || expectationText != "" {
		logMsg += " | updated body sections"
	}
	if renamed {
		logMsg += fmt.Sprintf(" | renamed from %s", filepath.Base(filePath))
	}
	config.Log("atd-update", logMsg)

	// Propagate ID change if renamed
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
			numUpdates := updateLinks(docsPath, oldID, newID)
			logMsg += fmt.Sprintf(" | Propagated ID change [[%s]] -> [[%s]] across %d files", oldID, newID, numUpdates)
			config.Log("atd-update", fmt.Sprintf("Propagated ID change across %d files", numUpdates))
		}
	}

	return "Success: " + logMsg, nil
}

func applySpecLink(id, file string) error {
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", file, err)
	}

	tag := fmt.Sprintf("// @spec-link [[%s]]\n", id)
	// Avoid double tagging
	if strings.Contains(string(content), tag) {
		return nil
	}

	updated := tag + string(content)
	if err := os.WriteFile(file, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %v", file, err)
	}
	fmt.Printf("Injected // @spec-link [[%s]] to %s\n", id, file)
	return nil
}

func resolveArg(arg string) string {
	if arg == "-" {
		bytes, _ := io.ReadAll(os.Stdin)
		return string(bytes)
	}
	// Interpret literal \n as actual newlines
	return strings.ReplaceAll(arg, "\\n", "\n")
}

func updateLinks(docsPath, oldID, newID string) int {
	updatedCount := 0
	oldLink := fmt.Sprintf("[[%s]]", oldID)
	newLink := fmt.Sprintf("[[%s]]", newID)
	escapedOld := regexp.QuoteMeta(oldID)
	re := regexp.MustCompile(`\[\[` + escapedOld + `\]\]`)

	replaceInFile := func(path string) {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return
		}
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

	// 1. Propagate through all .atom.md files in docs
	filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".atom.md") {
			return nil
		}
		replaceInFile(path)
		return nil
	})

	// 2. Propagate through source code @spec-link / @test-link tags
	projectRoot := config.ProjectRoot()
	if projectRoot != "" {
		filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			// Skip .atom.md files (already handled above) and binary/hidden dirs
			if strings.HasSuffix(path, ".atom.md") {
				return nil
			}
			relPath, _ := filepath.Rel(projectRoot, path)
			if strings.HasPrefix(relPath, ".git") || strings.HasPrefix(relPath, ".atd") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// Only scan text-like source files
			ext := filepath.Ext(path)
			switch ext {
			case ".go", ".py", ".js", ".ts", ".tsx", ".jsx", ".rs", ".rb", ".java",
				".c", ".cpp", ".h", ".hpp", ".cs", ".sh", ".yaml", ".yml", ".toml", ".md":
				replaceInFile(path)
			}
			return nil
		})
	}

	return updatedCount
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("file", "f", "", "Path to the ATD file")
	updateCmd.Flags().String("filter", "", "Filter atoms to update instead of a single file (e.g. 'status=DRAFT,type=RULE')")
	updateCmd.Flags().String("intent", "", "New content for ## INTENT section (use '-' for stdin)")
	updateCmd.Flags().String("logic", "", "New content for ## THE RULE / LOGIC section (use '-' for stdin)")
	updateCmd.Flags().String("interface", "", "New content for ## TECHNICAL INTERFACE section (use '-' for stdin)")
	updateCmd.Flags().String("expectation", "", "New content for ## EXPECTATION section (use '-' for stdin)")
	updateCmd.Flags().StringSlice("set", []string{}, "Set frontmatter key=value (can be used multiple times)")
	updateCmd.Flags().StringSlice("spec-link", []string{}, "Inbound ID and source file to tag: --spec-link <id> <file>")
}

func formatYAMLList(key, val string) string {
	val = strings.TrimSpace(val)
	if val == "" || val == "[]" {
		return fmt.Sprintf("%s: []", key)
	}

	// Split by comma or newline for lists
	var items []string
	if strings.Contains(val, "\n") {
		items = strings.Split(val, "\n")
	} else {
		// Strip outer brackets if present (e.g. from [a, b])
		val = strings.Trim(val, "[] ")
		items = strings.Split(val, ",")
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s:", key))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || item == "-" {
			continue
		}
		// Robustly strip ALL existing brackets and dashes
		item = strings.Trim(item, "[]- ")
		if item == "" {
			continue
		}
		result.WriteString(fmt.Sprintf("\n  - [[%s]]", item))
	}
	return result.String()
}
