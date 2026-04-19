package indexer

import (
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
)

// DiscoverFiles discovers all source files in a directory tree,
// providing comprehensive coverage for indexing, not just git-tracked files.
func DiscoverFiles(root string, cfg *config.Config, mode string) ([]string, error) {
	var files []string
	var count int

	walker := func(path string, info os.FileInfo, err error) error {
		// Skip directories and special files
		if info.IsDir() || err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Determine if file should be indexed based on extension and mode
		ext := filepath.Ext(relPath)
		shouldIndex := false

		switch mode {
		case "code":
			shouldIndex = cfg.SupportedExtensions[ext]
		case "docs":
			shouldIndex = true
		case "all":
			shouldIndex = true
		}

		// Check gitignore if in code mode
		if mode == "code" {
			if isGitIgnored(relPath, root) {
				shouldIndex = false
			}
		}

		if shouldIndex {
			files = append(files, relPath)
			count++
		}

		return nil
	}

	// Execute the walk
	err := filepath.Walk(root, walker)
	return files, err
}

// isGitIgnored checks if a path should be excluded based on .gitignore
func isGitIgnored(relPath, root string) bool {
	// Check against configured ignore patterns or defaults
	ignorePatterns := []string{}
	if len(config.ActiveConfig.GitignorePatterns) > 0 {
		ignorePatterns = config.ActiveConfig.GitignorePatterns
	} else {
		ignorePatterns = []string{
			"vendor/",
			"node_modules/",
			".git/",
			"dist/",
			"build/",
			"target/",
			".vscode/",
			"*.test.go",
			"*_test.go",
		}
	}

	// Convert to absolute path for comparison
	absPath := filepath.Join(root, relPath)

	for _, pattern := range ignorePatterns {
		if strings.Contains(absPath, pattern) {
			return true
		}
	}

	return false
}

// GetFileList provides a list of discovered files for processing
func GetFileList(root string, cfg *config.Config, mode string) ([]string, int, error) {
	var files []string
	var count int

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// Skip directories and special files
		if info.IsDir() || err != nil {
			return err
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Determine if file should be indexed based on extension and mode
		ext := filepath.Ext(relPath)
		shouldIndex := false

		switch mode {
		case "code":
			shouldIndex = cfg.SupportedExtensions[ext]
		case "docs":
			shouldIndex = true
		case "all":
			shouldIndex = true
		}

		// Check gitignore if in code mode
		if mode == "code" {
			if isGitIgnored(relPath, root) {
				shouldIndex = false
			}
		}

		if shouldIndex {
			files = append(files, relPath)
			count++
		}

		return nil
	})

	return files, int(count), err
}
