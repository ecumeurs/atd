package config

import (
	"encoding/json"
	"testing"
)

// Regression test for the config-casing defect recorded in the project's
// atd_known_defects_backlog memory (2026-07-29): Config fields like
// CodePaths had no json tag at all, so json.Unmarshal only matched the
// literal PascalCase Go field name and silently dropped snake_case keys
// written by hand in a .atd file.

func TestConfigUnmarshalJSON_SnakeCase(t *testing.T) {
	data := []byte(`{"docs_path": "docs/", "code_paths": ["src/"]}`)
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cfg.CodePaths) != 1 || cfg.CodePaths[0] != "src/" {
		t.Fatalf("CodePaths = %#v, want [\"src/\"]", cfg.CodePaths)
	}
	if cfg.DocsPath != "docs/" {
		t.Fatalf("DocsPath = %q, want \"docs/\"", cfg.DocsPath)
	}
}

func TestConfigUnmarshalJSON_LegacyPascalCase(t *testing.T) {
	data := []byte(`{"CodePaths": ["legacy/"], "SupportedExtensions": {".go": true}, "DiscoveryMethod": "git-ls-files", "MaxDepth": 5}`)
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cfg.CodePaths) != 1 || cfg.CodePaths[0] != "legacy/" {
		t.Fatalf("CodePaths = %#v, want [\"legacy/\"]", cfg.CodePaths)
	}
	if !cfg.SupportedExtensions[".go"] {
		t.Fatalf("SupportedExtensions[.go] = false, want true")
	}
	if cfg.DiscoveryMethod != DiscoveryMethodGit {
		t.Fatalf("DiscoveryMethod = %q, want %q", cfg.DiscoveryMethod, DiscoveryMethodGit)
	}
	if cfg.MaxDepth != 5 {
		t.Fatalf("MaxDepth = %d, want 5", cfg.MaxDepth)
	}
}

func TestConfigUnmarshalJSON_SnakeCaseWinsOverPascalCase(t *testing.T) {
	data := []byte(`{"CodePaths": ["old/"], "code_paths": ["new/"]}`)
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cfg.CodePaths) != 1 || cfg.CodePaths[0] != "new/" {
		t.Fatalf("CodePaths = %#v, want [\"new/\"] (snake_case should win)", cfg.CodePaths)
	}
}

func TestConfigUnmarshalJSON_RoundTripThroughMarshal(t *testing.T) {
	// atd init and `atd config model set-task-model` both write .atd back
	// via json.MarshalIndent(ActiveConfig, ...); that output must still
	// unmarshal cleanly.
	orig := Config{
		CodePaths:           []string{"a/", "b/"},
		SupportedExtensions: map[string]bool{".go": true},
		DiscoveryMethod:     DiscoveryMethodHybrid,
		MaxDepth:            7,
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(cfg.CodePaths) != 2 || cfg.CodePaths[0] != "a/" || cfg.CodePaths[1] != "b/" {
		t.Fatalf("CodePaths = %#v, want [a/ b/]", cfg.CodePaths)
	}
	if cfg.MaxDepth != 7 {
		t.Fatalf("MaxDepth = %d, want 7", cfg.MaxDepth)
	}
}
