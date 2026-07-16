package cmd

// WP-6 (test_atd_07_26.md §3.5, §6, S12): deterministic logic + parser-
// robustness tests for `atd dissect --llm`, using pkg/testutil/fakeprovider
// instead of a live Ollama. Named dissect_llm_test.go (rather than
// dissect_test.go) since these exercise the LLM branch specifically; there
// was no existing dissect test file to collide with.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/testutil/fakeprovider"
)

func writeTarget(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "sample.go")
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

// TestRunDissect_NoLLM_ReturnsPromptOnly pins the passthrough path: without
// --llm, runDissect must never touch the LLM seam at all.
func TestRunDissect_NoLLM_ReturnsPromptOnly(t *testing.T) {
	target := writeTarget(t, "package main\nfunc main() {}\n")

	out, err := runDissect(target, false)
	if err != nil {
		t.Fatalf("runDissect(useLLM=false): %v", err)
	}
	if !strings.Contains(out, "system_context") {
		t.Errorf("expected the raw prompt JSON to be returned, got: %s", out)
	}
}

// TestRunDissect_LLM_ValidJSON drives the --llm path with a schema-valid
// canned response and asserts it comes back re-indented but otherwise
// unchanged.
func TestRunDissect_LLM_ValidJSON(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"atoms":[{"id":"a1","responsibility":"does x","line_range":[1,2]}]}`)

	target := writeTarget(t, "package main\nfunc main() {}\n")
	out, err := runDissect(target, true)
	if err != nil {
		t.Fatalf("runDissect(useLLM=true): %v", err)
	}
	if !strings.Contains(out, `"responsibility": "does x"`) {
		t.Errorf("expected re-indented atoms JSON, got: %s", out)
	}
}

// TestRunDissect_LLM_MalformedJSON is the S12 case for dissect: prior to
// WP-6's surgical fix, cmd/atd/cmd/dissect.go discarded
// json.MarshalIndent's error (`out, _ := json.MarshalIndent(...)`), so any
// syntactically-invalid model response silently produced an EMPTY string
// with a nil error — a textbook zero-value success indistinguishable from
// "the file dissected into zero atoms". The fix (this WP) checks the error
// and salvages the raw response instead. This test would have failed before
// the fix (out == "", err == nil) and pins the corrected behavior now.
func TestRunDissect_LLM_MalformedJSON(t *testing.T) {
	cases := []struct {
		name string
		resp string
	}{
		{"markdown_fenced", "```json\n{\"atoms\":[]}\n```"},
		{"trailing_prose", `Here is the dissection: {"atoms":[]}`},
		{"truncated", `{"atoms":[{"id":"a1","responsibility":"x","line_range":[1,`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fake := fakeprovider.InstallOllama(t)
			fake.SetJSON(c.resp)

			target := writeTarget(t, "package main\nfunc main() {}\n")
			out, err := runDissect(target, true)
			if err != nil {
				t.Fatalf("runDissect must salvage malformed JSON rather than error, got err: %v", err)
			}
			if out == "" {
				t.Fatalf("KNOWN DEFECT expectation changed: malformed LLM response (%s) produced an empty string again -- this is exactly the zero-value-success bug WP-6 fixed; if this fires, the surgical fix in dissect.go regressed", c.name)
			}
			if out != c.resp {
				t.Errorf("expected the raw malformed response to be salvaged verbatim, got: %s", out)
			}
		})
	}
}
