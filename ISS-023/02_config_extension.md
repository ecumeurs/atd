# Task 02 — Config Extension

**Depends on:** Task 01  
**Produces:** Extended `config.go` + updated `.atd`

## Context

Currently `config.go` has a flat structure with `Model`, `DocsPath`, `BinPath`. We need:
- Remove `BinPath` (no longer needed — single binary)
- Add `LLM` section with providers, models→tasks mapping, fallback
- All paths resolve relative to the directory containing `.atd`

## Steps

### 2.1 Add new types to `scripts/config/config.go`

Add before `ATDConfig`:

```go
type LLMProvider struct {
    Name      string `json:"name"`
    BaseURL   string `json:"base_url"`
    TimeoutMs int    `json:"timeout_ms"`
    Type      string `json:"type,omitempty"` // "passthrough" for IDE agent
}

type ModelConfig struct {
    Tasks []string `json:"tasks"`
}

type LLMConfig struct {
    Providers     []LLMProvider          `json:"providers"`
    Models        map[string]ModelConfig `json:"models"`
    FallbackModel string                `json:"fallback_model"`
}
```

### 2.2 Update `ATDConfig` struct

```go
type ATDConfig struct {
    DocsPath                string               `json:"docs_path"`
    DiffSimilarityThreshold float64              `json:"diff_similarity_threshold"`
    BloatingFactor          BloatingFactorConfig `json:"bloating_factor"`
    Model                   string               `json:"model"` // kept for backward compat
    Logging                 LoggingConfig        `json:"logging"`
    SupportedExtensions     map[string]bool      `json:"supported_extensions"`
    LLM                     LLMConfig            `json:"llm"`
    loadedFromDir           string
}
```

### 2.3 Add helper: `ProjectRoot()`

```go
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
```

### 2.4 Add helper: `ModelForTask(taskType string)`

```go
// ModelForTask returns the model name assigned to a given task type.
// Falls back to FallbackModel, then to the legacy "model" field, then "llama3.2".
func ModelForTask(taskType string) string {
    for modelName, mc := range ActiveConfig.LLM.Models {
        for _, t := range mc.Tasks {
            if t == taskType {
                return modelName
            }
        }
    }
    if ActiveConfig.LLM.FallbackModel != "" {
        return ActiveConfig.LLM.FallbackModel
    }
    if ActiveConfig.Model != "" {
        return ActiveConfig.Model
    }
    return "llama3.2"
}
```

### 2.5 Update root `.atd` file

Replace `/home/bastien/work/skill/.atd` with:

```json
{
  "docs_path": "docs/",
  "diff_similarity_threshold": 0.85,
  "bloating_factor": {
    "default": 0.8,
    "type_overrides": {
      "REQUIREMENT": 0.3,
      "SPECIFICATION": 0.3,
      "MODULE": 0.3,
      "USECASE": 0.1,
      "USER_STORY": 0.1,
      "API": 0.1
    }
  },
  "model": "llama3.2",
  "logging": {
    "log_path": ".agent/logs/atd_trace.log"
  },
  "llm": {
    "providers": [
      {"name": "remote", "base_url": "http://192.168.1.10:11434", "timeout_ms": 2000},
      {"name": "local",  "base_url": "http://localhost:11434",    "timeout_ms": 500},
      {"name": "ide_agent", "type": "passthrough"}
    ],
    "models": {
      "llama3.2":           {"tasks": ["audit_bloat", "intent_extract", "snapshot"]},
      "deepseek-r1:7b":     {"tasks": ["audit_code", "compare", "congruence", "reconcile", "fix_split"]},
      "qwen2.5-coder:14b":  {"tasks": ["dissect", "recon"]},
      "nomic-embed-text":   {"tasks": ["embed"]}
    },
    "fallback_model": "llama3.2"
  }
}
```

## Acceptance Criteria

- [ ] `config.Load()` parses the new `.atd` without error
- [ ] `config.ModelForTask("dissect")` returns `"qwen2.5-coder:14b"`
- [ ] `config.ModelForTask("embed")` returns `"nomic-embed-text"`
- [ ] `config.ModelForTask("unknown_task")` returns `"llama3.2"` (fallback)
- [ ] `config.DocsDir()` returns the absolute path to `docs/`
- [ ] `config.ProjectRoot()` returns the directory containing `.atd`
