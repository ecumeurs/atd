package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type BloatingFactorConfig struct {
	Default       float64            `json:"default"`
	TypeOverrides map[string]float64 `json:"type_overrides"`
}

type LoggingConfig struct {
	LogPath string `json:"log_path"`
}

type ATDConfig struct {
	DocsPath                string               `json:"docs_path"`
	BinPath                 string               `json:"bin_path"`
	DiffSimilarityThreshold float64              `json:"diff_similarity_threshold"`
	BloatingFactor          BloatingFactorConfig `json:"bloating_factor"`
	Model                   string               `json:"model"`
	Logging                 LoggingConfig        `json:"logging"`
	loadedFromDir           string               // internal tracking
}

type LogEntry struct {
	Time string `json:"time"`
	Tool string `json:"tool"`
	Msg  string `json:"msg"`
}

var ActiveConfig ATDConfig

// Load looks for .atd in the current directory and up to the root.
func Load() error {
	// Defaults
	ActiveConfig = ATDConfig{
		DiffSimilarityThreshold: 0.85,
		BloatingFactor: BloatingFactorConfig{
			Default:       0.8,
			TypeOverrides: make(map[string]float64),
		},
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
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
			ActiveConfig.loadedFromDir = filepath.Dir(configPath)
		}
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
