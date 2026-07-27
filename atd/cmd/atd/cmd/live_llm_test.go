package cmd

// WP-6 opt-in live-LLM smoke suite (test_atd_07_26.md §3.5 point 3, §6
// WP-6 deliverable 5, WP-8). These tests hit a REAL provider and are
// therefore double-gated, matching the conventions the WP-5 CI lane
// (.github/workflows/live.yml) already assumes:
//
//   - name prefix TestLive* (the manual workflow runs `go test -run TestLive`)
//   - env gate: skipped unless ATD_LIVE_LLM=1 is set
//
// They are NEVER part of the default path: without ATD_LIVE_LLM they skip
// immediately, before any config/provider resolution, so they pass the
// WP-6 "no network" AC (run with http_proxy pointing at a dead port).
//
// Requirements when actually run live: a reachable Ollama-compatible
// provider configured in the active .atd (or reachable at the default
// config), with a generation model for code_analysis/text_analysis and an
// embedding model (e.g. nomic-embed-text). Failures here mean "the live
// provider setup is broken or the model regressed", not "the product code
// is wrong" — deterministic logic is covered by the fakeprovider suites.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/indexer"
	"atd-tools/pkg/testutil"
)

func skipUnlessLive(t *testing.T) {
	t.Helper()
	if os.Getenv("ATD_LIVE_LLM") == "" {
		t.Skip("live LLM smoke test: set ATD_LIVE_LLM=1 (and have a provider running) to enable")
	}
}

// liveSandbox anchors config at a temp project so nothing a live run does
// (index db paths, pipeline output) can land in the repo. The semantic
// index db itself goes to the user cache dir via config.IndexDBPath, keyed
// by the sandbox docs path, so it is unique per run.
func liveSandbox(t *testing.T) (root, docsDir string) {
	t.Helper()
	testutil.SnapshotConfig(t)

	root = t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".atd"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	docsDir = filepath.Join(root, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := config.LoadFromDir(root); err != nil {
		t.Fatalf("config.LoadFromDir: %v", err)
	}
	return root, docsDir
}

func writeLiveAtom(t *testing.T, docsDir, id, intent string) {
	t.Helper()
	content := `---
id: ` + id + `
type: RULE
status: DRAFT
parents: []
dependents: []
layer: BUSINESS
---

## INTENT
` + intent + `

## THE RULE / LOGIC
` + intent + `

## EXPECTATION
n/a
`
	if err := os.WriteFile(filepath.Join(docsDir, id+".atom.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// TestLiveIndexAndSemanticSearch indexes a couple of small atom files with
// the real embedding model, then semantic-searches them and asserts the
// obviously-relevant atom ranks first.
func TestLiveIndexAndSemanticSearch(t *testing.T) {
	skipUnlessLive(t)
	_, docsDir := liveSandbox(t)

	writeLiveAtom(t, docsDir, "rule_zzlive_auth", "Users must authenticate with a password before accessing the system.")
	writeLiveAtom(t, docsDir, "rule_zzlive_billing", "Invoices are generated on the first day of each month.")

	dbPath := filepath.Join(t.TempDir(), "live_index.db")
	if _, err := indexer.Index(docsDir, dbPath, "all"); err != nil {
		t.Fatalf("live index: %v", err)
	}

	results, err := exploration.SemanticSearch("password login authentication", dbPath, 2, "docs", "")
	if err != nil {
		t.Fatalf("live semantic search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("live semantic search returned no results over a freshly built index")
	}
	if !strings.Contains(results[0].FilePath, "rule_zzlive_auth") {
		t.Errorf("expected the auth atom to rank first for an auth query, got: %s (sim %.4f)", results[0].FilePath, results[0].Similarity)
	}
}

// TestLiveMapConfirm runs one real recon (map --atom) round-trip and
// asserts the response parses into the ReconResult schema (any confidence
// value is acceptable — this smokes the model's schema compliance, not its
// judgment).
func TestLiveMapConfirm(t *testing.T) {
	skipUnlessLive(t)
	root, docsDir := liveSandbox(t)

	writeLiveAtom(t, docsDir, "rule_zzlive_greeting", "The Greet function must return the string 'hello' followed by the given name.")
	code := "package zzlive\n\nfunc Greet(name string) string {\n\treturn \"hello \" + name\n}\n"
	if err := os.WriteFile(filepath.Join(root, "greet.go"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	out, err := runMapConfirm("greet.go", code, "rule_zzlive_greeting")
	if err != nil {
		t.Fatalf("live map confirm: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("live map confirm returned empty output")
	}
	// Either a parsed verdict (pretty JSON with Confidence) or the explicit
	// inconclusive message are acceptable; raw unparsed text is a schema
	// violation worth failing loudly on in a live smoke.
	if !strings.Contains(out, "\"Confidence\"") && !strings.Contains(out, "inconclusive") {
		t.Errorf("live recon response did not honor the ReconResult schema; got: %s", out)
	}
}
