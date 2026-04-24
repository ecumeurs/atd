package workspace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	workspaceCache = make(map[string]*Workspace)
	cacheMutex     sync.RWMutex
	ErrNoWorkspace = errors.New("no workspace found")
)

type projectMinimalConfig struct {
	DocsPath string `json:"docs_path"`
}

// LoadWorkspace loads and caches a workspace configuration.
func LoadWorkspace(root string) (*Workspace, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	cacheMutex.RLock()
	if ws, ok := workspaceCache[absRoot]; ok {
		cacheMutex.RUnlock()
		return ws, nil
	}
	cacheMutex.RUnlock()

	// Find .atd.workspace upward
	wsPath, loadedFrom, err := findWorkspaceConfig(absRoot)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(wsPath)
	if err != nil {
		return nil, err
	}

	var config WorkspaceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	config.LoadedFrom = loadedFrom
	if config.WorkspaceRoot == "" {
		config.WorkspaceRoot = loadedFrom
	} else if !filepath.IsAbs(config.WorkspaceRoot) {
		config.WorkspaceRoot = filepath.Join(loadedFrom, config.WorkspaceRoot)
	}

	ws := &Workspace{WorkspaceConfig: config}

	// Auto-detect project paths and configs
	for i := range ws.Projects {
		projPath := ws.Projects[i].Path
		if !filepath.IsAbs(projPath) {
			projPath = filepath.Join(ws.WorkspaceRoot, projPath)
		}

		configPath := ws.Projects[i].ConfigPath
		if configPath == "" {
			configPath = filepath.Join(projPath, ".atd")
		} else if !filepath.IsAbs(configPath) {
			configPath = filepath.Join(ws.WorkspaceRoot, configPath)
		}

		ws.Projects[i].ConfigPath = configPath

		// Load project config to get docs_path
		if data, err := os.ReadFile(configPath); err == nil {
			var pc projectMinimalConfig
			if err := json.Unmarshal(data, &pc); err == nil {
				if pc.DocsPath != "" {
					ws.Projects[i].DocsPath = pc.DocsPath
				}
			}
		}

		if ws.Projects[i].DocsPath == "" {
			ws.Projects[i].DocsPath = "docs"
		}

		ws.Projects[i].FullDocsPath = filepath.Join(projPath, ws.Projects[i].DocsPath)
	}

	cacheMutex.Lock()
	workspaceCache[absRoot] = ws
	workspaceCache[loadedFrom] = ws
	cacheMutex.Unlock()

	return ws, nil
}

func findWorkspaceConfig(startDir string) (string, string, error) {
	dir := startDir
	for {
		p := filepath.Join(dir, ".atd.workspace")
		if _, err := os.Stat(p); err == nil {
			abs, _ := filepath.Abs(dir)
			return p, abs, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir || parent == "/" {
			break
		}
		dir = parent
	}
	return "", "", ErrNoWorkspace
}

// FindProjectByCWD finds the active project based on current working directory.
func (ws *Workspace) FindProjectByCWD(cwd string) *Project {
	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		return nil
	}

	for i := range ws.Projects {
		projPath := ws.Projects[i].Path
		if !filepath.IsAbs(projPath) {
			projPath = filepath.Join(ws.WorkspaceRoot, projPath)
		}

		if strings.HasPrefix(absCWD, projPath) {
			return &ws.Projects[i]
		}
	}
	return nil
}
