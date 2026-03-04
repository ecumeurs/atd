package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"atd-tools/config"
)

// Section indices
const (
	SecNone = iota
	SecIntent
	SecLogic
	SecInterface
)

func main() {
	filePath := flag.String("file", "", "Path to the ATD file")
	intentArg := flag.String("intent", "", "New content for ## INTENT section (use '-' to read from stdin)")
	logicArg := flag.String("logic", "", "New content for ## THE RULE / LOGIC section (use '-' to read from stdin)")
	interfaceArg := flag.String("interface", "", "New content for ## TECHNICAL INTERFACE section (use '-' to read from stdin)")

	// Custom flag parsing for -set key=value
	var setArgs []string
	var args []string
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-set" {
			if i+1 < len(os.Args) {
				setArgs = append(setArgs, os.Args[i+1])
				i++
			}
		} else {
			args = append(args, os.Args[i])
		}
	}

	// Re-parse standard flags
	os.Args = append([]string{os.Args[0]}, args...)
	flag.Parse()

	if *filePath == "" {
		fmt.Println("Usage: atd-update -file <path> [-intent <content>] [-logic <content>] [-interface <content>] [-set key=value ...]")
		os.Exit(1)
	}

	err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}

	intentText := resolveArg(*intentArg)
	logicText := resolveArg(*logicArg)
	interfaceText := resolveArg(*interfaceArg)

	// Read original file
	content, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Error: Could not find valid YAML frontmatter in %s\n", *filePath)
		os.Exit(1)
	}

	// Update frontmatter keys
	updates := make(map[string]string)
	var newID, newType string

	for _, setArg := range setArgs {
		parts := strings.SplitN(setArg, "=", 2)
		if len(parts) == 2 {
			key, val := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			updates[key] = val
			if key == "id" {
				newID = val
			} else if key == "type" {
				newType = val
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

	// Helper to check if a line is a section header we care about
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

	// Helper to check if a line is any header `## ` (to know when a section ends)
	isAnyHeader := func(line string) bool {
		return strings.HasPrefix(line, "## ")
	}

	for i := frontmatterEnd + 1; i < len(lines); i++ {
		line := lines[i]

		// Check if we're entering a new section
		if sec := isTargetHeader(line); sec != SecNone {
			currentSection = sec
			newBody = append(newBody, line)

			// Inject new content immediately after header
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
			// We exited the target section
			currentSection = SecNone
		}

		// If we are IN a target section and we HAVE replacement text for it, skip the original lines
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
			// Avoid double blank lines that accumulate when skipping empty sections
			if len(newBody) > 0 && newBody[len(newBody)-1] == "" && line == "" && currentSection != SecNone {
				continue
			}
			newBody = append(newBody, line)
		}
	}

	// 3. Recombine and Save
	finalOutput := strings.Join(append(newFrontmatter, newBody...), "\n")

	// Handle Renaming
	targetPath := *filePath

	if newID != "" || newType != "" {
		dir := filepath.Dir(*filePath)

		// If ID wasn't explicitly changed, extract current ID to form the filename
		fileNameID := newID
		if fileNameID == "" {
			for _, line := range frontmatterLines {
				if strings.HasPrefix(line, "id:") {
					fileNameID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
					break
				}
			}
		}

		if fileNameID != "" {
			targetPath = filepath.Join(dir, fmt.Sprintf("%s.atom.md", fileNameID))
		}
	}

	err = os.WriteFile(targetPath, []byte(finalOutput), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	// If renamed, delete the old file
	renamed := false
	if targetPath != *filePath {
		os.Remove(*filePath)
		renamed = true
		fmt.Printf("Renamed %s -> %s\n", filepath.Base(*filePath), filepath.Base(targetPath))
	}

	// Log Action
	logMsg := fmt.Sprintf("Updated %s", filepath.Base(targetPath))
	if len(updates) > 0 {
		logMsg += fmt.Sprintf(" | set: %d keys", len(updates))
	}
	if intentText != "" || logicText != "" || interfaceText != "" {
		logMsg += " | updated body sections"
	}
	if renamed {
		logMsg += fmt.Sprintf(" | renamed from %s", filepath.Base(*filePath))
	}
	config.Log("atd-update", logMsg)

	fmt.Println("Success:", logMsg)

	// 4. Link Updates (If ID changed)
	if renamed && newID != "" {
		// We need to find the old ID to replace
		var oldID string
		for _, line := range frontmatterLines {
			if strings.HasPrefix(line, "id:") {
				oldID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
				break
			}
		}

		if oldID != "" && oldID != newID {
			docsPath := config.ActiveConfig.DocsPath
			if docsPath == "" {
				docsPath = filepath.Join(filepath.Dir(targetPath)) // Fallback to same folder
			}

			if !filepath.IsAbs(docsPath) {
				// Try to make it absolute if resolving from current directory
				cwd, _ := os.Getwd()

				// Keep going up until we find a docs dir, or use cwd/docsPath
				for {
					p := filepath.Join(cwd, docsPath)
					if _, err := os.Stat(p); err == nil {
						docsPath = p
						break
					}
					parent := filepath.Dir(cwd)
					if parent == cwd || parent == "/" {
						break
					}
					cwd = parent
				}
			}

			fmt.Printf("Updating links from [[%s]] to [[%s]] in %s...\n", oldID, newID, docsPath)
			linkUpdates := updateLinks(docsPath, oldID, newID)
			fmt.Printf("Updated %d files with new links.\n", linkUpdates)
			config.Log("atd-update", fmt.Sprintf("Propagated ID change [[%s]]->[[%s]] across %d files", oldID, newID, linkUpdates))
		}
	}
}

// resolveArg reads from stdin if arg is "-", otherwise returns arg
func resolveArg(arg string) string {
	if arg == "-" {
		bytes, _ := io.ReadAll(os.Stdin)
		return string(bytes)
	}
	return arg
}

// updateLinks crawls docs dir and replaces `[[oldID]]` with `[[newID]]`
func updateLinks(docsPath, oldID, newID string) int {
	updatedCount := 0

	oldLink := fmt.Sprintf("[[%s]]", oldID)
	newLink := fmt.Sprintf("[[%s]]", newID)

	// Create regexes to match the exact string without accidental substring matches
	// E.g. [[\boldID\b]]
	escapedOld := regexp.QuoteMeta(oldID)
	// Match both `parents: \n - [[id]]` and in-text `[[id]]`
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
