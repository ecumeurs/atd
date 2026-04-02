package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"webui/parser"
)

var Atoms map[string]*parser.Atom

// @spec-link [[module_webui]]
func refreshAtoms() {
	atdDir := filepath.Join(AppConfig.ProjectPath, AppConfig.ATDPath)
	newAtoms, err := parser.ParseAtoms(atdDir)
	if err != nil {
		log.Printf("Warning: Failed to parse atoms: %v", err)
		return
	}
	parser.FindLinkedCode(AppConfig.ProjectPath, newAtoms)
	parser.CalculateStatuses(newAtoms)
	Atoms = newAtoms
	fmt.Printf("Refreshed %d atoms from %s\n", len(Atoms), atdDir)
}

// @spec-link [[module_webui]]
func extractIntent(content string) string {
	lines := strings.Split(content, "\n")
	inIntent := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## INTENT") {
			inIntent = true
			continue
		}
		if inIntent && strings.HasPrefix(trimmed, "## ") {
			break
		}
		if inIntent && trimmed != "" {
			return trimmed
		}
	}
	return ""
}
