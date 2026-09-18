package cmd

// Regression tests for the visibility requirement on llm.generate_timeout_ms
// (config.LLMConfig.GenerateTimeoutMs, config.DefaultGenerateTimeoutMs): the
// new bounded-timeout config field must be visible in the nearest .atd file
// at its default value, both on a fresh `atd init` and on `atd init
// --upgrade` against a pre-existing project -- not left as an invisible
// Go-side-only default that a project's checked-in .atd never actually
// shows.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"atd-tools/config"
)

func TestRunInit_Fresh_WritesGenerateTimeoutMsDefault(t *testing.T) {
	saved := config.Snapshot()
	defer config.Restore(saved)

	dir := t.TempDir()

	if _, err := runInit(dir, "docs/", "llama3.2", false, false); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".atd"))
	if err != nil {
		t.Fatalf("reading .atd: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing .atd: %v", err)
	}
	var llm map[string]json.RawMessage
	if err := json.Unmarshal(raw["llm"], &llm); err != nil {
		t.Fatalf("parsing .atd llm block: %v", err)
	}

	gt, ok := llm["generate_timeout_ms"]
	if !ok {
		t.Fatalf("expected llm.generate_timeout_ms to be visible in a fresh .atd, got: %s", data)
	}
	if string(gt) != "120000" {
		t.Errorf("expected llm.generate_timeout_ms=120000, got %s", gt)
	}
}

func TestRunInit_Upgrade_BackfillsMissingGenerateTimeoutMs(t *testing.T) {
	saved := config.Snapshot()
	defer config.Restore(saved)

	dir := t.TempDir()
	atdPath := filepath.Join(dir, ".atd")
	// A pre-existing .atd with an "llm" block that predates
	// generate_timeout_ms entirely, as a real project's checked-in .atd
	// would look before this feature existed.
	preexisting := `{
  "docs_path": "docs/",
  "llm": {
    "providers": [{"name": "local", "base_url": "http://localhost:11434", "timeout_ms": 500}],
    "models": {"llama3.2": {"tasks": ["*"]}},
    "fallback_model": "llama3.2"
  }
}`
	if err := os.WriteFile(atdPath, []byte(preexisting), 0644); err != nil {
		t.Fatalf("writing pre-existing .atd: %v", err)
	}

	if _, err := runInit(dir, "", "", false, true); err != nil {
		t.Fatalf("runInit --upgrade: %v", err)
	}

	data, err := os.ReadFile(atdPath)
	if err != nil {
		t.Fatalf("reading .atd after upgrade: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing .atd after upgrade: %v", err)
	}
	var llm map[string]json.RawMessage
	if err := json.Unmarshal(raw["llm"], &llm); err != nil {
		t.Fatalf("parsing .atd llm block after upgrade: %v", err)
	}

	gt, ok := llm["generate_timeout_ms"]
	if !ok {
		t.Fatalf("expected --upgrade to backfill llm.generate_timeout_ms into the pre-existing .atd, got: %s", data)
	}
	if string(gt) != "120000" {
		t.Errorf("expected backfilled llm.generate_timeout_ms=120000, got %s", gt)
	}

	// The existing provider config must survive the backfill untouched.
	var providers []map[string]json.RawMessage
	if err := json.Unmarshal(llm["providers"], &providers); err != nil {
		t.Fatalf("parsing llm.providers after upgrade: %v", err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected the pre-existing provider to survive the backfill, got: %s", llm["providers"])
	}
}

func TestRunInit_Upgrade_NoOpsWhenGenerateTimeoutMsAlreadyPresent(t *testing.T) {
	saved := config.Snapshot()
	defer config.Restore(saved)

	dir := t.TempDir()
	atdPath := filepath.Join(dir, ".atd")
	preexisting := `{
  "docs_path": "docs/",
  "llm": {
    "providers": [{"name": "local", "base_url": "http://localhost:11434", "timeout_ms": 500}],
    "models": {"llama3.2": {"tasks": ["*"]}},
    "fallback_model": "llama3.2",
    "generate_timeout_ms": 45000
  }
}`
	if err := os.WriteFile(atdPath, []byte(preexisting), 0644); err != nil {
		t.Fatalf("writing pre-existing .atd: %v", err)
	}

	if _, err := runInit(dir, "", "", false, true); err != nil {
		t.Fatalf("runInit --upgrade: %v", err)
	}

	data, err := os.ReadFile(atdPath)
	if err != nil {
		t.Fatalf("reading .atd after upgrade: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing .atd after upgrade: %v", err)
	}
	var llm map[string]json.RawMessage
	if err := json.Unmarshal(raw["llm"], &llm); err != nil {
		t.Fatalf("parsing .atd llm block after upgrade: %v", err)
	}

	if string(llm["generate_timeout_ms"]) != "45000" {
		t.Errorf("expected --upgrade to leave an already-present llm.generate_timeout_ms untouched, got %s", llm["generate_timeout_ms"])
	}
}

func TestRunInit_Upgrade_NeverFabricatesLLMBlock(t *testing.T) {
	saved := config.Snapshot()
	defer config.Restore(saved)

	dir := t.TempDir()
	atdPath := filepath.Join(dir, ".atd")
	// A project with no "llm" block configured at all -- --upgrade must
	// not fabricate one just to inject generate_timeout_ms.
	preexisting := `{"docs_path": "docs/"}`
	if err := os.WriteFile(atdPath, []byte(preexisting), 0644); err != nil {
		t.Fatalf("writing pre-existing .atd: %v", err)
	}

	if _, err := runInit(dir, "", "", false, true); err != nil {
		t.Fatalf("runInit --upgrade: %v", err)
	}

	data, err := os.ReadFile(atdPath)
	if err != nil {
		t.Fatalf("reading .atd after upgrade: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing .atd after upgrade: %v", err)
	}
	if _, ok := raw["llm"]; ok {
		t.Errorf("expected --upgrade to never fabricate an llm block for a project with none, got: %s", data)
	}
}
