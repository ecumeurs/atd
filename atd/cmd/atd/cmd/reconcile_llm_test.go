package cmd

// WP-6 (test_atd_07_26.md §3.5, §6, S12): parser-robustness tests for
// `atd reconcile`'s response handling, using pkg/testutil/fakeprovider.
// reconcileCmd's parsing lives inline in its RunE and prints via
// fmt.Println, so these tests execute the cobra command directly and
// capture os.Stdout (test-side only — no product refactor needed for a
// print-shaped command).

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atd-tools/pkg/testutil/fakeprovider"
)

// captureStdout runs fn while os.Stdout is redirected into a pipe and
// returns everything written.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	saved := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = saved }()

	fnErr := fn()
	w.Close()
	out, readErr := io.ReadAll(r)
	if readErr != nil {
		t.Fatal(readErr)
	}
	return string(out), fnErr
}

// runReconcileCmd invokes the reconcile command's RunE with temp --new and
// --store files and returns captured stdout + error.
func runReconcileCmd(t *testing.T) (string, error) {
	t.Helper()
	dir := t.TempDir()
	newPath := filepath.Join(dir, "inbound.md")
	storePath := filepath.Join(dir, "store.md")
	os.WriteFile(newPath, []byte("# inbound edits"), 0644)
	os.WriteFile(storePath, []byte("# existing atoms"), 0644)

	reconcileCmd.Flags().Set("new", newPath)
	reconcileCmd.Flags().Set("store", storePath)
	t.Cleanup(func() {
		reconcileCmd.Flags().Set("new", "")
		reconcileCmd.Flags().Set("store", "")
	})

	return captureStdout(t, func() error {
		return reconcileCmd.RunE(reconcileCmd, nil)
	})
}

// TestReconcile_ValidDiffsObject pins the happy path: {"diffs": [...]}
// prints the diffs array.
func TestReconcile_ValidDiffsObject(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`{"diffs": [{"proposed_id": "zzfix_a", "relationship": "NEW", "change_context": "added"}]}`)

	out, err := runReconcileCmd(t)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if !strings.Contains(out, `"proposed_id": "zzfix_a"`) {
		t.Errorf("expected the diffs array printed, got: %s", out)
	}
}

// TestReconcile_RawArraySalvage pins the documented salvage: a bare JSON
// array (schema says the output may be a top-level array) is re-indented
// and printed.
func TestReconcile_RawArraySalvage(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetJSON(`[{"proposed_id": "zzfix_b", "relationship": "UPDATE", "change_context": "changed"}]`)

	out, err := runReconcileCmd(t)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if !strings.Contains(out, `"proposed_id": "zzfix_b"`) {
		t.Errorf("expected the salvaged array printed, got: %s", out)
	}
}

// TestReconcile_MalformedResponses: syntactically-invalid JSON falls
// through both Unmarshal attempts and the raw response is printed verbatim
// — a loud salvage (the operator sees exactly what the model said), never
// an empty success. Pinned as-is.
func TestReconcile_MalformedResponses(t *testing.T) {
	cases := []struct {
		name string
		resp string
	}{
		{"markdown_fenced", "```json\n[{\"proposed_id\": \"x\"}]\n```"},
		{"trailing_prose", `Here are the diffs: [{"proposed_id": "x"}]`},
		{"truncated", `[{"proposed_id": "zzfix_c", "relationship": "NEW"`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fake := fakeprovider.InstallOllama(t)
			fake.SetJSON(c.resp)

			out, err := runReconcileCmd(t)
			if err != nil {
				t.Fatalf("reconcile must salvage malformed JSON rather than error, got: %v", err)
			}
			if strings.TrimSpace(out) == "" {
				t.Fatalf("malformed response (%s) must never produce empty output", c.name)
			}
			if !strings.Contains(out, strings.Split(c.resp, "\n")[0]) {
				t.Errorf("expected the raw response salvaged in output, got: %s", out)
			}
		})
	}
}

// TestReconcile_ValidObjectWithoutDiffsKey pins the fix for a genuine
// defect: any syntactically-valid JSON *object* that lacks a "diffs" key —
// including the field-report's real-world wrong-shape case — used to
// unmarshal into the anonymous struct with Diffs == nil and no error, so
// the command printed literally "null" and exited 0: a zero-value success
// that neither surfaced the model's actual (unusable) response nor
// signaled that reconciliation produced nothing parseable.
//
// FIXED (test_atd_07_26.md §8.3 #9b): reconcile.go now treats a
// missing/nil "diffs" key the same as the malformed-JSON case just above —
// the raw response is salvaged (printed verbatim) so the operator sees
// exactly what the model said instead of a misleadingly empty "null".
func TestReconcile_ValidObjectWithoutDiffsKey(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	// The real-world case from the field report: valid JSON, wrong shape.
	resp := `{"Confidence": 0, "Mismatches": "Yes, there are several problems"}`
	fake.SetJSON(resp)

	out, err := runReconcileCmd(t)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if strings.TrimSpace(out) == "null" {
		t.Fatal("a valid-JSON-object-without-diffs-key response must never silently print the zero-value \"null\"")
	}
	if !strings.Contains(out, `"Mismatches"`) {
		t.Errorf("expected the raw response salvaged in output so the operator sees the actual wrong-shape response, got: %q", out)
	}
}
