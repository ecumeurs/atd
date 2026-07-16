package atom

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzParse is the opportunistic §5.5 fuzz target: Parse must never panic on
// arbitrary file content, however malformed. It is seeded from the real
// fixture_project atom corpus (testdata/fixture_project/docs/*.atom.md,
// including mech_zzfix_gamma.atom.md — the ### H3-swallow fixture, E4's
// regression case) plus a handful of synthetic malformed shapes covering
// the same classes as malformed_test.go (no closing delimiter, CRLF,
// duplicate keys, empty file).
//
// Seeds only: this runs as a plain `go test` in the default lane (each
// f.Add'ed seed becomes one ordinary subtest) with no actual fuzzing —
// `go test -fuzz=FuzzParse` is opt-in and out of scope for CI per
// test_atd_07_26.md §5.5 ("no long fuzz run in the default lane").
func FuzzParse(f *testing.F) {
	fixtureDir := filepath.Join("..", "..", "testdata", "fixture_project", "docs")
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		f.Fatalf("reading fixture corpus dir %s: %v", fixtureDir, err)
	}
	seeded := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(fixtureDir, e.Name()))
		if err != nil {
			f.Fatalf("reading fixture %s: %v", e.Name(), err)
		}
		f.Add(string(content))
		seeded++
	}
	if seeded == 0 {
		f.Fatalf("no fixture atoms found under %s to seed FuzzParse", fixtureDir)
	}

	// Synthetic malformed shapes, mirroring malformed_test.go's classes.
	f.Add("")
	f.Add("---\nid: no_closing\nstatus: DRAFT\n\n## INTENT\nbody\n")
	f.Add("---\r\nid: crlf_seed\r\nstatus: DRAFT\r\n---\r\n\r\n## INTENT\r\nbody\r\n")
	f.Add("---\nid: dup\nid: dup2\n---\n## INTENT\nbody\n")
	f.Add("---\ntotally_unknown_key: [1, 2, 3\n---\n")
	f.Add("---\nparents:\n  - [[a]]\n  - [[b\n---\n")

	f.Fuzz(func(t *testing.T, content string) {
		path := filepath.Join(t.TempDir(), "fuzz.atom.md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("writing fuzz input: %v", err)
		}
		// Parse must never panic; an error return is fine (or no error --
		// Parse is deliberately permissive, see malformed_test.go).
		_, _ = Parse(path)
	})
}
