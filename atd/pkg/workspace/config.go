package workspace

// Project represents a single project within an ATD workspace.
type Project struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	ConfigPath   string   `json:"config_path,omitempty"`
	DocsPath     string   `json:"docs_path,omitempty"`
	CodePaths    []string `json:"code_paths,omitempty"`
	FullDocsPath string   `json:"-"` // Resolved absolute path to docs
}

// WorkspaceConfig holds the configuration for a multi-project ATD workspace.
type WorkspaceConfig struct {
	WorkspaceName   string                 `json:"workspace_name"`
	WorkspaceRoot   string                 `json:"workspace_root"`
	Projects        []Project              `json:"projects"`
	SharedLibraries map[string]string      `json:"shared_libraries,omitempty"`
	CommonSettings  map[string]interface{} `json:"common_settings,omitempty"`
	LoadedFrom      string                 `json:"-"` // Path where .atd.workspace was found
}

// Workspace represents an active workspace context.
type Workspace struct {
	WorkspaceConfig
}
