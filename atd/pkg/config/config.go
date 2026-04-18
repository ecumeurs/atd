package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var ActiveConfig Config

type DiscoveryMethod string

const (
	DiscoveryMethodWalk   = "walk"
	DiscoveryMethodGit    = "git-ls-files"
	DiscoveryMethodHybrid = "hybrid"
)

type Config struct {
	DocsDir          string
	CodePaths        []string
	SupportedExtensions map[string]bool
	DiscoveryMethod   DiscoveryMethod
	OrphanExcludedTypes map[string]bool
	HierarchicalOrphanCheck bool
	CustomerLayerException bool
	GitignorePatterns []string
	MaxDepth         int
}

func Load() (*Config, error) {
	// Load from .atd file, merge with environment defaults
	if config, err := loadFromATDFile(); err == nil && config != nil {
		ActiveConfig = *config
		return config, nil
	}
	return LoadFromEnv()
}

func LoadFromEnv() (*Config, error) {
	config := &Config{
		DocsDir:          "docs/",
		CodePaths:        []string{"."},
		SupportedExtensions: map[string]bool{
			".go":   true,
			".php":  true,
			".js":   true,
			".vue":  true,
			".ts":   true,
			".md":   false,
			".atom.md": false,
		},
		DiscoveryMethod:   DiscoveryMethodWalk,
		OrphanExcludedTypes: map[string]bool{
			"MODULE":         true,
			"SPECIFICATION":  true,
			"USECASE":        true,
			"USER_STORY":    true,
		},
		HierarchicalOrphanCheck: true,
		CustomerLayerException: true,
		GitignorePatterns: []string{
			"node_modules/",
			".git/",
			"dist/",
			"build/",
			"target/",
			".vscode/",
		},
		MaxDepth:         10,
	}
	ActiveConfig = *config
	return config, nil
}

func loadFromATDFile() (*Config, error) {
	configPath := findATDConfigFile()
	if configPath == "" {
		return nil, fmt.Errorf("no .atd configuration file found")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read .atd config: %v", err)
	}

	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse .atd config: %v", err)
	}

	ActiveConfig = *config
	return config, nil
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

func GetDiscoveryMethod() DiscoveryMethod {
	if ActiveConfig.DiscoveryMethod == "" {
		return DiscoveryMethodWalk
	}
	return ActiveConfig.DiscoveryMethod
}

func GetOrphanExcludedTypes() map[string]bool {
	if ActiveConfig.OrphanExcludedTypes == nil {
		return map[string]bool{
			"MODULE":         true,
			"SPECIFICATION":  true,
			"USECASE":        true,
			"USER_STORY":    true,
		}
	}
	return ActiveConfig.OrphanExcludedTypes
}

func SetHierarchicalOrphanCheck(enabled bool) {
	ActiveConfig.HierarchicalOrphanCheck = enabled
}

func SetCustomerLayerException(enabled bool) {
	ActiveConfig.CustomerLayerException = enabled
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
