package cmd

// Tool-set <-> docs equality (test_atd_07_26.md §3.3 #3, S11, the E3 drift
// killer). Unlike every other file in this package's test suite, this test
// reads the REAL repository docs/ directory (walked upward from the test
// binary's cwd), not a testdata/ fixture -- the whole point is comparing
// the registered tool set against ATD's own real corpus, the same "dogfood"
// principle §3.4 asks for elsewhere. It does not use testutil.Sandbox
// (nothing here is mutated, so there is nothing to isolate), but it does
// use testutil.Guard so a stray config.ActiveConfig load doesn't leak
// (atom.Parse doesn't touch config, but Guard is cheap insurance and this
// is the one test in the package that legitimately reads outside its own
// module tree).
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"atd-tools/pkg/atom"
	"atd-tools/pkg/mcp"
	"atd-tools/pkg/testutil"
)

// missingToolAtomsAllowlist is a SHRINKING allowlist (test_atd_07_26.md §6
// WP-4 AC: "either author [the 8 missing tool atoms] in this WP ... or
// start the equality test with an explicit, shrinking allowlist"). Every
// name below is a currently-registered MCP tool with NO corresponding
// api_atd_serve_<name> atom in docs/ yet (investigation_atd_07_26.md D1's
// "8 missing tool atoms" -- the count matches exactly). Remove an entry
// once its atom is authored; do NOT add entries here for new drift -- a
// newly registered tool with no atom should fail this test, not grow this
// list.
var missingToolAtomsAllowlist = map[string]bool{
	"atd_env":             true,
	"atd_config":          true,
	"atd_workspace_list":  true,
	"atd_workspace_use":   true,
	"atd_workspace_stats": true,
	"atd_heatmap":         true,
	"atd_heatmap_code":    true,
	"atd_heatmap_project": true,
}

// docParamViolationsAllowlist is the same shrinking-allowlist pattern
// applied to individual parameters: each entry names a param an EXISTING
// api_atd_serve_* atom documents in its "### Input Schema" JSON block that
// the actually-registered tool schema does not declare at all. Found by
// manually diffing every atom's Input Schema block against
// RegisterMCPTools' InputSchema in mcp_tools.go (2026-07-26/16) -- none of
// these are hypothetical:
//
//   - atd_audit:      registered schema has ZERO properties (empty object);
//     the atom documents docs/threshold/code/atom, none of which the
//     handler reads from args (docs/threshold come from config; the
//     code+atom "compliance mode" the atom's LOGIC section describes is
//     not wired into runFullAudit's signature at all).
//   - atd_crawl:       registered schema only declares gaps/workspace; the
//     atom also documents src/docs, but the handler hardcodes
//     src="." and docs=config.DocsDir(), never reading either from args.
//   - atd_dissect:     registered schema only declares file; the atom also
//     documents llm, but the MCP handler hardcodes useLLM=true
//     unconditionally (runDissect(file, true)), ignoring any llm arg.
//   - atd_index:       registered schema has ZERO properties; the atom
//     documents dir/db/mode, but the handler hardcodes ".",
//     config.IndexDBPath(...), and "all".
//   - atd_stats:       registered schema only declares workspace; the atom
//     also documents src/docs, neither read from args.
//   - atd_test_links:  registered schema only declares atom; the atom also
//     documents src/docs, neither read from args.
//
// Remove an entry (or a single param from its slice) once the atom or the
// schema is corrected to match -- this is drift being PINNED, not policy.
var docParamViolationsAllowlist = map[string][]string{
	"atd_audit":      {"docs", "threshold", "code", "atom"},
	"atd_crawl":      {"src", "docs"},
	"atd_dissect":    {"llm"},
	"atd_index":      {"dir", "db", "mode"},
	"atd_stats":      {"src", "docs"},
	"atd_test_links": {"src", "docs"},
}

// inputSchemaBlockPattern matches the fenced ```json block that follows a
// "### Input Schema" heading in every api_atd_serve_* atom that documents
// one this way (api_atd_serve_lint.atom.md documents its one param in
// prose instead -- extractDocumentedParams returns ok=false for it, and it
// is trivially excluded from the subset check as a result: no documented
// JSON-schema params to violate).
var inputSchemaBlockPattern = regexp.MustCompile("(?s)### Input Schema\\s*```json\\s*(.*?)```")

// extractDocumentedParams parses the atom file at path for its "### Input
// Schema" fenced JSON block and returns the sorted property names it
// declares. ok is false if the atom has no such block at all (prose-only
// docs, e.g. api_atd_serve_lint.atom.md).
func extractDocumentedParams(t *testing.T, path string) (params []string, ok bool) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	m := inputSchemaBlockPattern.FindSubmatch(content)
	if m == nil {
		return nil, false
	}
	var schema struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(m[1], &schema); err != nil {
		t.Fatalf("%s: Input Schema block is not valid JSON: %v\n%s", path, err, m[1])
	}
	for k := range schema.Properties {
		params = append(params, k)
	}
	sort.Strings(params)
	return params, true
}

// findRepoDocsDir walks upward from the test binary's cwd (this package's
// own source directory under `go test` convention) looking for a directory
// literally named "docs" that also contains at least one *.atom.md file --
// distinguishing the real repo docs/ from any unrelated "docs" directory
// and, critically, from testdata/fixture_project/docs (which lives at a
// different relative path and is never matched by this walk).
func findRepoDocsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < 32; i++ {
		candidate := filepath.Join(dir, "docs")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			if matches, _ := filepath.Glob(filepath.Join(candidate, "*.atom.md")); len(matches) > 0 {
				return candidate, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not find a real docs/ directory (containing *.atom.md) by walking upward from %s", mustGetwdEquality())
}

func mustGetwdEquality() string {
	wd, _ := os.Getwd()
	return wd
}

// TestMCPContract_ToolSetDocsEquality pins test_atd_07_26.md §3.3 #3 / S11:
// the registered tool set must equal the api_atd_serve_* atoms in the
// REAL docs/ (modulo the shrinking missingToolAtomsAllowlist), and each
// existing atom's documented param names must be a subset of its tool's
// actually-registered schema properties (modulo docParamViolationsAllowlist).
func TestMCPContract_ToolSetDocsEquality(t *testing.T) {
	testutil.Guard(t)

	r := mcp.NewRegistry()
	RegisterMCPTools(r)

	registeredSchemas := make(map[string]map[string]any, len(r.List()))
	for _, tool := range r.List() {
		props, _ := tool.InputSchema["properties"].(map[string]any)
		registeredSchemas[tool.Name] = props
	}

	docsDir, err := findRepoDocsDir()
	if err != nil {
		t.Fatalf("locating the real repo docs/ dir: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(docsDir, "api_atd_serve_*.atom.md"))
	if err != nil {
		t.Fatalf("globbing api_atd_serve_* atoms in %s: %v", docsDir, err)
	}
	if len(matches) == 0 {
		t.Fatalf("found zero api_atd_serve_* atoms under %s -- findRepoDocsDir likely resolved the wrong directory", docsDir)
	}

	atomToolNames := make(map[string]bool, len(matches))
	for _, path := range matches {
		a, err := atom.Parse(path)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		toolName := "atd_" + strings.TrimPrefix(a.ID, "api_atd_serve_")
		atomToolNames[toolName] = true

		docParams, hasSchemaBlock := extractDocumentedParams(t, path)
		if !hasSchemaBlock {
			continue
		}
		regProps := registeredSchemas[toolName]
		for _, p := range docParams {
			if regProps != nil {
				if _, ok := regProps[p]; ok {
					continue
				}
			}
			allowed := false
			for _, a := range docParamViolationsAllowlist[toolName] {
				if a == p {
					allowed = true
					break
				}
			}
			if !allowed {
				t.Errorf("%s documents param %q which the registered %q tool's schema does not declare -- either the atom is stale, the schema regressed, or (if this is pre-existing drift being pinned rather than new) add it to docParamViolationsAllowlist",
					filepath.Base(path), p, toolName)
			}
		}
	}

	// Direction 1: every registered tool has a matching atom, or is on the
	// shrinking missing-atom allowlist.
	for name := range registeredSchemas {
		if atomToolNames[name] {
			continue
		}
		if !missingToolAtomsAllowlist[name] {
			t.Errorf("registered tool %q has no api_atd_serve_%s atom in docs/ and is not on missingToolAtomsAllowlist -- author the atom, or add it to the allowlist if this is pre-existing drift being pinned",
				name, strings.TrimPrefix(name, "atd_"))
		}
	}

	// Direction 2: no phantom atoms (E3) -- every api_atd_serve_* atom must
	// name a currently-registered tool.
	for name := range atomToolNames {
		if _, isRegistered := registeredSchemas[name]; !isRegistered {
			t.Errorf("docs/ has an api_atd_serve_%s atom but RegisterMCPTools does not register a tool named %q -- phantom atom (E3)",
				strings.TrimPrefix(name, "atd_"), name)
		}
	}

	// The allowlist itself must stay exact in both directions, or it has
	// silently stopped pinning what it claims to.
	for name := range missingToolAtomsAllowlist {
		if _, isRegistered := registeredSchemas[name]; !isRegistered {
			t.Errorf("missingToolAtomsAllowlist references %q which is not even a registered tool -- prune the allowlist", name)
			continue
		}
		if atomToolNames[name] {
			t.Errorf("missingToolAtomsAllowlist references %q but an atom now exists for it -- shrink the allowlist (remove this entry)", name)
		}
	}
}
