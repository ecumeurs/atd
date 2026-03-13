package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
	"github.com/spf13/cobra"
)

// Section indices
const (
	SecNone = iota
	SecIntent
	SecLogic
	SecInterface
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an ATD atom file's frontmatter or body sections",
	Long: `Update an ATD atom file's frontmatter or body sections.
Supports setting individual frontmatter keys, updating content sections (Intent, Rule/Logic, Technical Interface),
and prepending @spec-link tags to source files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath, _ := cmd.Flags().GetString("file")
		intentArg, _ := cmd.Flags().GetString("intent")
		logicArg, _ := cmd.Flags().GetString("logic")
		interfaceArg, _ := cmd.Flags().GetString("interface")
		setArgs, _ := cmd.Flags().GetStringSlice("set")
		specLinkArgs, _ := cmd.Flags().GetStringSlice("spec-link")

		// Handle --spec-link <id> <file>
		// Note: Cobra StringSlice captures ["id", "file", "id2", "file2"] if used multiple times
		if len(specLinkArgs) > 0 {
			if len(specLinkArgs)%2 != 0 {
				return fmt.Errorf("--spec-link requires exactly two arguments: <id> <file>")
			}
			for i := 0; i < len(specLinkArgs); i += 2 {
				id := specLinkArgs[i]
				targetFile := specLinkArgs[i+1]
				if err := applySpecLink(id, targetFile); err != nil {
					return err
				}
			}
			if filePath == "" {
				return nil // Finished if only spec-link was requested
			}
		}

		if filePath == "" {
			return fmt.Errorf("-file parameter is required for atom updates")
		}

		intentText := resolveArg(intentArg)
		logicText := resolveArg(logicArg)
		interfaceText := resolveArg(interfaceArg)

		// Read original file
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file: %v", err)
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
			return fmt.Errorf("could not find valid YAML frontmatter in %s", filePath)
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

		// Replace existing keys
		matchedKeys := make(map[string]bool)
		for _, line := range frontmatterLines {
			updated := false
			for k, v := range updates {
				prefix := k + ":"
				if strings.HasPrefix(line, prefix) {
					newFrontmatter = append(newFrontmatter, fmt.Sprintf("%s: %s", k, v))
					matchedKeys[k] = true
					updated = true
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
				newFrontmatter = append(newFrontmatter, fmt.Sprintf("%s: %s", k, v))
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
			if strings.HasPrefix(line, "## TECHNICAL INTERFACE") {
				return SecInterface
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
			if currentSection == SecInterface && interfaceText != "" {
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
			return fmt.Errorf("error writing file %s: %v", targetPath, err)
		}

		renamed := false
		if targetPath != filePath {
			os.Remove(filePath)
			renamed = true
			fmt.Printf("Renamed %s -> %s\n", filepath.Base(filePath), filepath.Base(targetPath))
		}

		logMsg := fmt.Sprintf("Updated %s", filepath.Base(targetPath))
		if len(updates) > 0 {
			logMsg += fmt.Sprintf(" | set: %d keys", len(updates))
		}
		if intentText != "" || logicText != "" || interfaceText != "" {
			logMsg += " | updated body sections"
		}
		if renamed {
			logMsg += fmt.Sprintf(" | renamed from %s", filepath.Base(filePath))
		}
		config.Log("atd-update", logMsg)
		fmt.Println("Success:", logMsg)

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
				fmt.Printf("Propagated ID change [[%s]] -> [[%s]] across %d files\n", oldID, newID, numUpdates)
				config.Log("atd-update", fmt.Sprintf("Propagated ID change across %d files", numUpdates))
			}
		}

		return nil
	},
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
	return arg
}

func updateLinks(docsPath, oldID, newID string) int {
	updatedCount := 0
	oldLink := fmt.Sprintf("[[%s]]", oldID)
	newLink := fmt.Sprintf("[[%s]]", newID)
	escapedOld := regexp.QuoteMeta(oldID)
	re := regexp.MustCompile(`\[\[` + escapedOld + `\]\]`)

	filepath.Walk(docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".atom.md") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
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
		return nil
	})
	return updatedCount
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringP("file", "f", "", "Path to the ATD file")
	updateCmd.Flags().String("intent", "", "New content for ## INTENT section (use '-' for stdin)")
	updateCmd.Flags().String("logic", "", "New content for ## THE RULE / LOGIC section (use '-' for stdin)")
	updateCmd.Flags().String("interface", "", "New content for ## TECHNICAL INTERFACE section (use '-' for stdin)")
	updateCmd.Flags().StringSlice("set", []string{}, "Set frontmatter key=value (can be used multiple times)")
	updateCmd.Flags().StringSlice("spec-link", []string{}, "Inbound ID and source file to tag: --spec-link <id> <file>")
}
