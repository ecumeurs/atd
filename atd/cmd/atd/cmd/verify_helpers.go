package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

func getSnippet(path string, line int, context int) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	start := line - context
	if start < 0 {
		start = 0
	}
	end := line + context
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n"), nil
}

func executeTestCommand(tpl, dir, file string) (string, error) {
	absDir, _ := filepath.Abs(dir)
	absFile, _ := filepath.Abs(file)
	modRoot := findGoModRoot(absDir)

	relDir := dir
	relFile := file
	if modRoot != "" {
		relDir, _ = filepath.Rel(modRoot, absDir)
		relFile, _ = filepath.Rel(modRoot, absFile)
	}

	t, err := template.New("cmd").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := struct {
		Dir  string
		File string
	}{
		Dir:  relDir,
		File: relFile,
	}

	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	cmdStr := buf.String()
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	if modRoot != "" {
		cmd.Dir = modRoot
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	return out.String(), err
}

func findGoModRoot(dir string) string {
	current := dir
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}
