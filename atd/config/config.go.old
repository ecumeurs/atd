package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

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
	Providers      []LLMProvider          `json:"providers"`
	Models         map[string]ModelConfig `json:"models"`
	FallbackModel  string                 `json:"fallback_model"`
	HealthTTLMs    int                    `json:"health_ttl_ms,omitempty"`
	ModelTTLMs     int                    `json:"model_ttl_ms,omitempty"`
}

type ATDConfig struct {
	DocsPath                string               `json:"docs_path"`
	DiffSimilarityThreshold float64              `json:"diff_similarity_threshold"`
	BloatingFactor          BloatingFactorConfig `json:"bloating_factor"`
	Model                   string               `json:"model"` // kept for backward compat
	Logging                 LoggingConfig        `json:"logging"`
	SupportedExtensions     map[string]bool      `json:"supported_extensions"`
	LLM                     LLMConfig            `json:"llm"`
	WebUI                   WebUIConfig          `json:"webui"`
	Verify                  VerifyConfig         `json:"verify"`
	loadedFromDir           string
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

var ActiveConfig ATDConfig

// Load looks for .atd in the current directory and up to the root.
func Load() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	return LoadFromDir(dir)
}

// LoadFromDir looks for .atd starting from dir and up to the root.
func LoadFromDir(dir string) error {
	// Defaults
	ActiveConfig = ATDConfig{
		DiffSimilarityThreshold: 0.85,
		BloatingFactor: BloatingFactorConfig{
			Default:       0.8,
			TypeOverrides: make(map[string]float64),
		},
		SupportedExtensions: map[string]bool{
			".go": true, ".py": true, ".ts": true, ".js": true,
			".rs": true, ".java": true, ".c": true, ".cpp": true,
			".h": true, ".hpp": true, ".cs": true, ".php": true,
			".rb": true, ".swift": true, ".kt": true, ".scala": true,
		},
		LLM: LLMConfig{
			HealthTTLMs: 300000, // 5 minutes
			ModelTTLMs:  300000, // 5 minutes
		},
	}

	var configPath string

	for {
		p := filepath.Join(dir, ".atd")
		if _, err := os.Stat(p); err == nil {
			configPath = p
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir || parent == "/" {
			break
		}
		dir = parent
	}

	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			json.Unmarshal(data, &ActiveConfig)
			abs, _ := filepath.Abs(filepath.Dir(configPath))
			ActiveConfig.loadedFromDir = abs
			// fmt.Printf("Config loaded from: %s\n", ActiveConfig.loadedFromDir)
		}
	} else {
		// No config found, default to CWD
		cwd, _ := os.Getwd()
		ActiveConfig.loadedFromDir = cwd
	}
	return nil
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

func DocsDir() string {
	p := ActiveConfig.DocsPath
	if p == "" {
		p = "docs/"
	}
	if !filepath.IsAbs(p) {
		return filepath.Join(ActiveConfig.loadedFromDir, p)
	}
	return p
}

// ModelForTask returns the model names assigned to a given task type,
// sorted by priority (if provided).
// Falls back to FallbackModel, then to the legacy "model" field, then "llama3.2".
func ModelForTask(taskType string) []string {
	type candidate struct {
		name     string
		priority int
	}
	var candidates []candidate

	for modelName, mc := range ActiveConfig.LLM.Models {
		for _, t := range mc.Tasks {
			if t == taskType {
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
