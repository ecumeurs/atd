package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var ActiveConfig Config

type DiscoveryMethod string

const (
	DiscoveryMethodWalk   = "walk"
	DiscoveryMethodGit    = "git-ls-files"
	DiscoveryMethodHybrid = "hybrid"
)

type Config struct {
	DocsDir                 string          `json:"docs_dir,omitempty"`
	CodePaths               []string        `json:"code_paths,omitempty"`
	SupportedExtensions     map[string]bool `json:"supported_extensions,omitempty"`
	DiscoveryMethod         DiscoveryMethod `json:"discovery_method,omitempty"`
	OrphanExcludedTypes     map[string]bool `json:"orphan_excluded_types,omitempty"`
	HierarchicalOrphanCheck bool            `json:"hierarchical_orphan_check,omitempty"`
	BusinessLayerException  bool            `json:"business_layer_exception,omitempty"`
	GitignorePatterns       []string        `json:"gitignore_patterns,omitempty"`
	MaxDepth                int             `json:"max_depth,omitempty"`

	// Legacy fields for backward compatibility
	DocsPath                string               `json:"docs_path,omitempty"`
	DiffSimilarityThreshold float64              `json:"diff_similarity_threshold"`
	BloatingFactor          BloatingFactorConfig `json:"bloating_factor"`
	Model                   string               `json:"model"` // kept for backward compat
	Logging                 LoggingConfig        `json:"logging"`
	LLM                     LLMConfig            `json:"llm"`
	WebUI                   WebUIConfig          `json:"webui"`
	Verify                  VerifyConfig         `json:"verify"`

	// Internal tracking
	loadedFromDir string
	// loadedFromFallback is true when no real .atd was found anywhere from
	// the search start dir up to root, and LoadFromDirLegacy fell back to
	// anchoring the config at the process cwd. Destructive, repo-wide
	// operations (rename propagation, fix, future bulk edits) must refuse
	// or loudly warn when the active config is anchored this way, since
	// "cwd" for a `go test` binary is the package source directory — see
	// test_atd_07_26.md §2.1 (incident I-1) and §3.6 (P-1).
	loadedFromFallback bool
	DocsDirOverride    string           `json:"docs_dir_override,omitempty"`
	Workspace          *WorkspaceConfig `json:"-"`
	ActiveProject      string           `json:"-"`
}

// configAlias has the same fields as Config (including the json tags above)
// but none of Config's methods, so it can be used as an UnmarshalJSON target
// without recursing back into Config.UnmarshalJSON.
type configAlias Config

// legacyPascalCaseAliases maps the historical, untagged Go field names this
// struct used to expose to the snake_case tag each now carries. Older .atd
// files — including every one `atd config model set-task-model`'s
// write-back path (runConfigUpdate) has ever produced, since it round-trips
// json.MarshalIndent(ActiveConfig, ...) verbatim — used these PascalCase
// keys because the fields had no json tag at all, so encoding/json fell
// back to matching the literal Go field name. That silently dropped any
// snake_case key a human wrote by hand for the same field (see
// atd_known_defects_backlog memory). UnmarshalJSON below accepts both
// casings so existing PascalCase files keep loading and hand-written
// snake_case files stop being silently ignored.
var legacyPascalCaseAliases = map[string]string{
	"DocsDir":                 "docs_dir",
	"CodePaths":               "code_paths",
	"SupportedExtensions":     "supported_extensions",
	"DiscoveryMethod":         "discovery_method",
	"OrphanExcludedTypes":     "orphan_excluded_types",
	"HierarchicalOrphanCheck": "hierarchical_orphan_check",
	"BusinessLayerException":  "business_layer_exception",
	"GitignorePatterns":       "gitignore_patterns",
	"MaxDepth":                "max_depth",
	"DocsDirOverride":         "docs_dir_override",
}

// UnmarshalJSON accepts both the legacy PascalCase keys and the canonical
// snake_case keys for the fields in legacyPascalCaseAliases. If both are
// present for a given field, the snake_case key wins.
func (c *Config) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for pascal, snake := range legacyPascalCaseAliases {
		v, ok := raw[pascal]
		if !ok {
			continue
		}
		if _, hasSnake := raw[snake]; !hasSnake {
			raw[snake] = v
		}
		delete(raw, pascal)
	}

	normalized, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	return json.Unmarshal(normalized, (*configAlias)(c))
}

type ProjectConfig struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	ConfigPath string   `json:"config_path,omitempty"`
	DocsPath   string   `json:"docs_path,omitempty"`
	CodePaths  []string `json:"code_paths,omitempty"`
}

type WorkspaceConfig struct {
	WorkspaceName  string                 `json:"workspace_name"`
	WorkspaceRoot  string                 `json:"workspace_root"`
	Projects       []ProjectConfig        `json:"projects"`
	SharedLibs     map[string]string      `json:"shared_libraries,omitempty"`
	CommonSettings map[string]interface{} `json:"common_settings,omitempty"`

	// Internal tracking
	LoadedFrom string `json:"-"`
}

// Legacy types for backward compatibility
type BloatingFactorConfig struct {
	Default       float64            `json:"default"`
	TypeOverrides map[string]float64 `json:"type_overrides"`
}

type LoggingConfig struct {
	LogPath string `json:"log_path"`
}

type LLMProvider struct {
	Name      string `json:"name"`
	BaseURL   string `json:"base_url"`
	TimeoutMs int    `json:"timeout_ms"`
	Type      string `json:"type,omitempty"` // "passthrough" for IDE agent
}

type ModelConfig struct {
	Tasks    []string `json:"tasks"`
	Priority int      `json:"priority,omitempty"`
}

type LLMConfig struct {
	Providers     []LLMProvider          `json:"providers"`
	Models        map[string]ModelConfig `json:"models"`
	FallbackModel string                 `json:"fallback_model"`
	HealthTTLs    int                    `json:"health_ttl_ms,omitempty"`
	ModelTTLs     int                    `json:"model_ttl_ms,omitempty"`
}

type VerifyConfig struct {
	Command        string `json:"command"`         // e.g. "go test -v ./{{.Dir}}"
	TestPattern    string `json:"test_pattern"`    // e.g. "*_test.go"
	MaxParallelism int    `json:"max_parallelism"` // number of concurrent checks
}

type WebUIConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	ToolkitPath string `json:"toolkit_path"`
}

type LogEntry struct {
	Time string `json:"time"`
	Tool string `json:"tool"`
	Msg  string `json:"msg"`
}

func Load() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	return LoadFromDirLegacy(dir)
}

// LoadFromDir is an alias for LoadFromDirLegacy for backward compatibility
func LoadFromDir(dir string) error {
	return LoadFromDirLegacy(dir)
}

func findATDConfigFile() string {
	// Search upward from current directory for .atd file
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := cwd
	for {
		configPath := filepath.Join(dir, ".atd")
		if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
			return configPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding .atd
			break
		}
		dir = parent
	}

	return ""
}

func GetCodePaths() []string {
	if len(ActiveConfig.CodePaths) == 0 {
		// Default to current directory
		cwd, _ := os.Getwd()
		return []string{cwd}
	}
	return ActiveConfig.CodePaths
}

func DocsDir() string {
	return DocsDirLegacy()
}

func GetDiscoveryMethod() DiscoveryMethod {
	if ActiveConfig.DiscoveryMethod == "" {
		return DiscoveryMethodWalk
	}
	return ActiveConfig.DiscoveryMethod
}

func GetOrphanExcludedTypes() map[string]bool {
	if ActiveConfig.OrphanExcludedTypes == nil {
		return map[string]bool{
			"MODULE":      true,
			"REQUIREMENT": true,
			"USER_STORY":  true,
			"CONTRACT":    true,
			"VISION":      true,
			"API":         true,
		}
	}
	return ActiveConfig.OrphanExcludedTypes
}

func SetHierarchicalOrphanCheck(enabled bool) {
	ActiveConfig.HierarchicalOrphanCheck = enabled
}

func SetBusinessLayerException(enabled bool) {
	ActiveConfig.BusinessLayerException = enabled
}

func GetGitignorePatterns() []string {
	return ActiveConfig.GitignorePatterns
}

func SetMaxDepth(depth int) error {
	if depth < 1 || depth > 50 {
		return fmt.Errorf("max_depth must be between 1 and 50")
	}
	ActiveConfig.MaxDepth = depth
	return nil
}

// Legacy functions for backward compatibility

// LoadLegacy loads configuration with backward compatibility
func LoadLegacy() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	return LoadFromDirLegacy(dir)
}

// LoadFromDirLegacy looks for .atd starting from dir and up to root.
func LoadFromDirLegacy(dir string) error {
	// Defaults
	ActiveConfig = Config{
		DiffSimilarityThreshold: 0.85,
		BloatingFactor: BloatingFactorConfig{
			Default:       0.8,
			TypeOverrides: make(map[string]float64),
		},
		SupportedExtensions: map[string]bool{
			".go": true, ".py": true, ".ts": true, ".js": true,
			".rs": true, ".java": true, ".c": true, ".cpp": true,
			".h": true, ".hpp": true, ".cs": true, ".php": true,
			".rb": true, ".swift": true, ".kt": true, ".scala": true, ".vue": true,
		},
		LLM: LLMConfig{
			HealthTTLs: 300000, // 5 minutes
			ModelTTLs:  300000, // 5 minutes
		},
		DiscoveryMethod: DiscoveryMethodWalk,
		OrphanExcludedTypes: map[string]bool{
			"MODULE":        true,
			"SPECIFICATION": true,
			"USECASE":       true,
			"USER_STORY":    true,
		},
		HierarchicalOrphanCheck: true,
		BusinessLayerException:  true,
		GitignorePatterns: []string{
			"node_modules/",
			".git/",
			"dist/",
			"build/",
			"target/",
			".vscode/",
		},
		MaxDepth: 10,
	}

	var configPath string
	searchDir := dir

	for {
		p := filepath.Join(searchDir, ".atd")
		if _, err := os.Stat(p); err == nil {
			configPath = p
			break
		}
		parent := filepath.Dir(searchDir)
		if parent == searchDir || parent == "/" {
			break
		}
		searchDir = parent
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			json.Unmarshal(data, &ActiveConfig)
			abs, _ := filepath.Abs(filepath.Dir(configPath))
			ActiveConfig.loadedFromDir = abs
			ActiveConfig.loadedFromFallback = false
		}
	} else {
		// No real .atd found anywhere above dir. Fall back to CWD, but mark
		// the config as fallback-anchored so destructive repo-wide
		// operations know not to trust ProjectRoot() as a genuine project
		// boundary (see loadedFromFallback doc comment above).
		cwd, _ := os.Getwd()
		ActiveConfig.loadedFromDir = cwd
		ActiveConfig.loadedFromFallback = true
	}

	// Always check for workspace context
	ws, err := LoadWorkspaceConfig(dir)
	if err != nil {
		// fmt.Printf("LoadWorkspaceConfig err: %v\n", err)
	}
	if err == nil && ws != nil {
		// fmt.Printf("Workspace found: %s\n", ws.WorkspaceName)
		ActiveConfig.Workspace = ws
		// If we don't have an active project yet (e.g. no .atd found or not in project dir), find it
		if ActiveConfig.ActiveProject == "" {
			// Pick the most specific (longest-path) project that contains dir, not
			// the first prefix match. Otherwise a project rooted at the workspace
			// root (path ".", e.g. "shared") prefixes every dir and would always
			// win, masking the real project dir lives in.
			var bestName string
			var bestProjPath string
			bestLen := -1
			for _, p := range ws.Projects {
				absProjPath := p.Path
				if !filepath.IsAbs(absProjPath) {
					absProjPath = filepath.Join(ws.LoadedFrom, p.Path)
				}
				absProjPath = filepath.Clean(absProjPath)

				rel, err := filepath.Rel(absProjPath, dir)
				if err == nil && (rel == "." || !strings.HasPrefix(rel, "..")) {
					if len(absProjPath) > bestLen {
						bestName = p.Name
						bestProjPath = absProjPath
						bestLen = len(absProjPath)
					}
				}
			}
			if bestName != "" {
				ActiveConfig.ActiveProject = bestName
				// If we haven't loaded a config yet, use this project as root
				if configPath == "" {
					ActiveConfig.loadedFromDir = bestProjPath
				}
			}
		}
	}

	return nil
}

// SetProject overrides the active project and reloads its config if needed.
func SetProject(name string) error {
	if ActiveConfig.Workspace == nil {
		return fmt.Errorf("no workspace active")
	}

	for _, p := range ActiveConfig.Workspace.Projects {
		if p.Name == name {
			ActiveConfig.ActiveProject = name
			absProjPath := p.Path
			if !filepath.IsAbs(absProjPath) {
				absProjPath = filepath.Join(ActiveConfig.Workspace.LoadedFrom, p.Path)
			}

			projConfigPath := p.ConfigPath
			if projConfigPath == "" {
				projConfigPath = filepath.Join(absProjPath, ".atd")
			} else if !filepath.IsAbs(projConfigPath) {
				projConfigPath = filepath.Join(ActiveConfig.Workspace.LoadedFrom, projConfigPath)
			}

			if data, err := os.ReadFile(projConfigPath); err == nil {
				json.Unmarshal(data, &ActiveConfig)
				ActiveConfig.loadedFromDir = absProjPath
			} else {
				// Fallback to defaults in project dir
				ActiveConfig.loadedFromDir = absProjPath
			}
			return nil
		}
	}

	return fmt.Errorf("project '%s' not found in workspace", name)
}

// LoadWorkspaceConfig searches upward for .atd.workspace and parses it.
func LoadWorkspaceConfig(startDir string) (*WorkspaceConfig, error) {
	dir := startDir
	for {
		p := filepath.Join(dir, ".atd.workspace")
		if _, err := os.Stat(p); err == nil {
			data, err := os.ReadFile(p)
			if err != nil {
				return nil, err
			}
			var ws WorkspaceConfig
			if err := json.Unmarshal(data, &ws); err != nil {
				return nil, err
			}
			abs, _ := filepath.Abs(dir)
			ws.LoadedFrom = abs
			if ws.WorkspaceRoot == "" {
				ws.WorkspaceRoot = abs
			}
			return &ws, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir || parent == "/" {
			break
		}
		dir = parent
	}
	return nil, fmt.Errorf("workspace not found")
}

// Log writes a concise trace to the configured log_path.
func Log(toolName string, message string) {
	if ActiveConfig.Logging.LogPath == "" {
		return
	}

	// Resolve actual log path relative to where .atd was found, if not absolute
	logPath := ActiveConfig.Logging.LogPath
	if !filepath.IsAbs(logPath) && ActiveConfig.loadedFromDir != "" {
		logPath = filepath.Join(ActiveConfig.loadedFromDir, logPath)
	} else if !filepath.IsAbs(logPath) {
		cwd, _ := os.Getwd()
		logPath = filepath.Join(cwd, logPath)
	}

	err := os.MkdirAll(filepath.Dir(logPath), 0755)
	if err != nil {
		return
	}

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	entry := LogEntry{
		Time: now,
		Tool: toolName,
		Msg:  message,
	}

	data, err := json.Marshal(entry)
	if err == nil {
		file.Write(data)
		file.WriteString("\n")
	}
}

// GetBloatingStrictness returns either the type override or default
func GetBloatingStrictness(atomType string) float64 {
	if val, ok := ActiveConfig.BloatingFactor.TypeOverrides[atomType]; ok {
		return val
	}
	if ActiveConfig.BloatingFactor.Default > 0 {
		return ActiveConfig.BloatingFactor.Default
	}
	return 0.8
}

// GetVerifyDefaults returns seeded defaults for common languages.
func GetVerifyDefaults(ext string) (string, string) {
	switch ext {
	case ".go":
		return "go test -v ./{{.Dir}}", "*_test.go"
	case ".py":
		return "pytest {{.File}}", "test_*.py"
	case ".js", ".ts":
		return "npm test {{.File}}", "*.test.js,*.test.ts"
	case ".rs":
		return "cargo test", "*"
	default:
		return "", ""
	}
}

func ProjectRoot() string {
	return ActiveConfig.loadedFromDir
}

// LoadedFromFallback reports whether the active config was anchored by the
// cwd fallback (no real .atd found) rather than a genuine project .atd file.
// Callers performing destructive, repo-wide operations (rename propagation,
// fix, bulk edits) should refuse or loudly warn when this is true — see
// test_atd_07_26.md §3.6 (P-1).
func LoadedFromFallback() bool {
	return ActiveConfig.loadedFromFallback
}

// Snapshot returns a copy of the current ActiveConfig for later restoration
// via Restore. Every test that loads (Load/LoadFromDir/LoadFromDirLegacy) or
// otherwise mutates config.ActiveConfig MUST snapshot it beforehand and
// restore it afterwards (ideally via t.Cleanup), or global config state
// leaks into whichever test runs next in the same binary — this is exactly
// how incident I-1 (test_atd_07_26.md §2.1) caused the suite to rewrite its
// own source. The canonical idiom:
//
//	saved := config.Snapshot()
//	t.Cleanup(func() { config.Restore(saved) })
func Snapshot() Config {
	return ActiveConfig
}

// Restore sets ActiveConfig back to a value previously captured by Snapshot.
func Restore(snap Config) {
	ActiveConfig = snap
}

// IndexDBPath returns the on-disk location of the semantic search index
// database for the given docs directory. The index is stored under the
// user's cache directory (XDG_CACHE_HOME, or ~/.cache as a fallback) rather
// than inside the project's docs/ folder, so building or querying the index
// never drops an artifact into the git working tree. The path is keyed by
// the absolute docs directory so distinct projects (including per-project
// docs dirs in a workspace) get distinct, stable index files.
func IndexDBPath(docsDir string) string {
	key := docsDir
	if abs, err := filepath.Abs(docsDir); err == nil {
		key = abs
	}
	sum := sha256.Sum256([]byte(key))
	hash := hex.EncodeToString(sum[:])[:16]

	cacheRoot := os.Getenv("XDG_CACHE_HOME")
	if cacheRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		cacheRoot = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheRoot, "atd", hash, "index.db")
}

func DocsDirLegacy() string {
	p := ActiveConfig.DocsDir
	if p == "" {
		p = ActiveConfig.DocsPath // Fall back to legacy field
		if p == "" {
			p = "docs/"
		}
	}
	if ActiveConfig.DocsDirOverride != "" {
		return ActiveConfig.DocsDirOverride
	}
	if !filepath.IsAbs(p) {
		return filepath.Join(ActiveConfig.loadedFromDir, p)
	}
	return p
}

func SrcDir() string {
	paths := GetCodePaths()
	if len(paths) > 0 {
		return paths[0]
	}
	return ActiveConfig.loadedFromDir
}

// ModelForTask returns model names assigned to a given task type,
// sorted by priority (if provided).
// Falls back to FallbackModel, then to legacy "model" field, then "llama3.2".
func ModelForTask(taskType string) []string {
	type candidate struct {
		name     string
		priority int
	}
	var candidates []candidate

	for modelName, mc := range ActiveConfig.LLM.Models {
		for _, t := range mc.Tasks {
			if t == taskType || t == "*" {
				candidates = append(candidates, candidate{modelName, mc.Priority})
			}
		}
	}

	if len(candidates) > 0 {
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].priority > candidates[j].priority
		})
		var names []string
		for _, c := range candidates {
			names = append(names, c.name)
		}
		return names
	}

	fallback := "llama3.2"
	if ActiveConfig.LLM.FallbackModel != "" {
		fallback = ActiveConfig.LLM.FallbackModel
	} else if ActiveConfig.Model != "" {
		fallback = ActiveConfig.Model
	}
	return []string{fallback}
}
