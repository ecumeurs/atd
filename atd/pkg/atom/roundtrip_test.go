package atom

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// writeAndParse writes content to a fresh file under t.TempDir() and parses
// it back, failing the test on any I/O or parse error.
func writeAndParse(t *testing.T, content string) AtomData {
	t.Helper()
	path := filepath.Join(t.TempDir(), "roundtrip.atom.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	data, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	return data
}

// TestBuildContentParseRoundTrip is the core property from
// test_atd_07_26.md §3.1 item 3: Parse(BuildContent(a)) == a, checked across
// a broad table of field values (unicode, multiline logic, empty optional
// sections, and single/multi-element parent lists — the last of which used
// to be corrupted by the BuildContent bug fixed alongside this test; see
// parse.go's BuildContent doc comment on the parents-newline fix).
//
// Only the fields BuildContent actually threads through to the template are
// asserted here (ID, HumanName, Type, Layer, Version, Status, Priority,
// Bloating, HeatMap, Parents, Intent, Logic, Expectation). Tags, Dependents,
// Interface, and Metadata do NOT round-trip through BuildContent at all —
// that is a separate, pinned KNOWN DEFECT below
// (TestBuildContentParseRoundTrip_KnownDefects), not an oversight in this
// table.
func TestBuildContentParseRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   AtomData
	}{
		{
			name: "minimal single parent",
			in: AtomData{
				ID: "req_alpha", HumanName: "Alpha", Type: "REQUIREMENT", Layer: "BUSINESS",
				Version: "1.0", Status: "DRAFT", Priority: "CORE", Bloating: "on", HeatMap: "all",
				Parents:     []string{"req_root"},
				Intent:      "The intent.",
				Logic:       "The logic.",
				Expectation: "The expectation.",
			},
		},
		{
			name: "multiple parents, unsorted input",
			in: AtomData{
				ID: "mech_beta", HumanName: "Beta", Type: "MECHANIC", Layer: "IMPLEMENTATION",
				Version: "2.3", Status: "STABLE", Priority: "5", Bloating: "off", HeatMap: "no_dep",
				Parents:     []string{"req_zzz", "req_aaa", "req_mmm"},
				Intent:      "Intent beta.",
				Logic:       "Logic beta.",
				Expectation: "Expectation beta.",
			},
		},
		{
			name: "no parents",
			in: AtomData{
				ID: "vision_root", HumanName: "Root Vision", Type: "VISION", Layer: "BUSINESS",
				Version: "1.0", Status: "STABLE", Priority: "CORE", Bloating: "on", HeatMap: "all",
				Parents:     nil,
				Intent:      "Root intent.",
				Logic:       "Root logic.",
				Expectation: "Root expectation.",
			},
		},
		{
			name: "empty optional sections",
			in: AtomData{
				ID: "req_empty", HumanName: "Empty Sections", Type: "REQUIREMENT", Layer: "BUSINESS",
				Version: "1.0", Status: "DRAFT", Priority: "CORE", Bloating: "on", HeatMap: "all",
				Parents:     []string{"req_root"},
				Intent:      "",
				Logic:       "",
				Expectation: "",
			},
		},
		{
			name: "multiline logic with embedded ### subheading (E4 H3-swallow class)",
			in: AtomData{
				ID: "mech_gamma", HumanName: "Gamma", Type: "MECHANIC", Layer: "IMPLEMENTATION",
				Version: "1.0", Status: "DRAFT", Priority: "CORE", Bloating: "on", HeatMap: "all",
				Parents: []string{"api_zzfix_beta"},
				Intent:  "Intent line one.\nIntent line two.",
				Logic: "First paragraph.\n\n" +
					"### Sub-detail\n" +
					"This text lives inside a ### subheading nested in the same H2\n" +
					"section; Parse must not treat it as a section boundary and must\n" +
					"not drop anything after it.\n\n" +
					"### Another sub-detail\n" +
					"More content here.",
				Expectation: "Given X\nWhen Y\nThen Z.",
			},
		},
		{
			name: "unicode everywhere",
			in: AtomData{
				ID: "req_unicode", HumanName: "Résumé 日本語 🚀", Type: "REQUIREMENT", Layer: "BUSINESS",
				Version: "1.0", Status: "DRAFT", Priority: "CORE", Bloating: "on", HeatMap: "all",
				Parents:     []string{"req_root"},
				Intent:      "Iñtent with ümlauts and 中文字符 and emoji 🎉.",
				Logic:       "Logic with café, naïve, 東京, and — em-dash.",
				Expectation: "Expectation: “curly quotes” and ‘apostrophes’.",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			content := BuildContent(tc.in)
			got := writeAndParse(t, content)

			if got.ID != tc.in.ID {
				t.Errorf("ID: got %q, want %q", got.ID, tc.in.ID)
			}
			if got.HumanName != tc.in.HumanName {
				t.Errorf("HumanName: got %q, want %q", got.HumanName, tc.in.HumanName)
			}
			if got.Type != tc.in.Type {
				t.Errorf("Type: got %q, want %q", got.Type, tc.in.Type)
			}
			if got.Layer != tc.in.Layer {
				t.Errorf("Layer: got %q, want %q", got.Layer, tc.in.Layer)
			}
			if got.Version != tc.in.Version {
				t.Errorf("Version: got %q, want %q", got.Version, tc.in.Version)
			}
			if got.Status != tc.in.Status {
				t.Errorf("Status: got %q, want %q", got.Status, tc.in.Status)
			}
			if got.Priority != tc.in.Priority {
				t.Errorf("Priority: got %q, want %q", got.Priority, tc.in.Priority)
			}
			if got.Bloating != tc.in.Bloating {
				t.Errorf("Bloating: got %q, want %q", got.Bloating, tc.in.Bloating)
			}
			if got.HeatMap != tc.in.HeatMap {
				t.Errorf("HeatMap: got %q, want %q", got.HeatMap, tc.in.HeatMap)
			}

			wantParents := append([]string(nil), tc.in.Parents...)
			// BuildContent always sorts parents, so the round-tripped value
			// is the sorted form of the input, not necessarily input order.
			sort.Strings(wantParents)
			if len(wantParents) == 0 {
				wantParents = nil
			}
			if !reflect.DeepEqual(got.Parents, wantParents) {
				t.Errorf("Parents: got %#v, want %#v (sorted input)", got.Parents, wantParents)
			}

			if got.Intent != tc.in.Intent {
				t.Errorf("Intent: got %q, want %q", got.Intent, tc.in.Intent)
			}
			if got.Logic != tc.in.Logic {
				t.Errorf("Logic: got %q, want %q", got.Logic, tc.in.Logic)
			}
			if got.Expectation != tc.in.Expectation {
				t.Errorf("Expectation: got %q, want %q", got.Expectation, tc.in.Expectation)
			}
		})
	}
}

// TestBuildContentParseRoundTrip_KnownDefects pins the parts of the
// BuildContent/Parse round-trip property that do NOT hold, per
// test_atd_07_26.md §3.1 item 3's instruction to record — rather than
// silently work around — any legitimate input for which the property fails.
//
// KNOWN DEFECT: BuildContent (pkg/atom/parse.go) drops four fields of
// AtomData entirely:
//   - Tags: the template hardcodes `tags: []` regardless of a.Tags.
//   - Dependents: the template hardcodes `dependents: []` regardless of
//     a.Dependents.
//   - Interface: the "## TECHNICAL INTERFACE" section is always the fixed
//     boilerplate "- **Code Tag:** `@spec-link [[id]]`" — a.Interface is
//     read by Parse but never written by BuildContent.
//   - Metadata: never emitted at all (there is no frontmatter key for it).
//
// This is distinct from the parents-corruption bug fixed alongside this
// test (see BuildContent's doc comment): that one was a formatting bug with
// an obviously-safe fix. These four are a design gap — BuildContent's only
// production caller (cmd/atd/cmd/fix.go, the atom-split path) never sets
// Tags/Dependents/Interface/Metadata on the AtomData it builds, so nothing
// currently depends on them round-tripping — but wiring them in would mean
// deciding a serialization format for each (matching FormatYAMLList's list
// style for Tags/Dependents, and deciding whether a caller-supplied
// Interface should override the autogenerated spec-link boilerplate), which
// is a product design decision, not a "small and obviously safe" fix. If
// this test ever starts failing (i.e. one of these fields DOES round-trip),
// treat it as a signal that BuildContent's contract changed and update/relax
// the pin accordingly.
func TestBuildContentParseRoundTrip_KnownDefects(t *testing.T) {
	t.Parallel()

	in := AtomData{
		ID: "req_lossy", HumanName: "Lossy", Type: "REQUIREMENT", Layer: "BUSINESS",
		Version: "1.0", Status: "DRAFT", Priority: "CORE", Bloating: "on", HeatMap: "all",
		Parents:    []string{"req_root"},
		Tags:       []string{"t1", "t2"},
		Dependents: []string{"req_dep1"},
		Interface:  "- **Custom Interface:** something specific",
		Metadata:   map[string]string{"project": "zzfix"},
		Intent:     "intent", Logic: "logic", Expectation: "expectation",
	}

	got := writeAndParse(t, BuildContent(in))

	if got.Tags != nil {
		t.Errorf("KNOWN DEFECT expectation changed: Tags now round-trips (got %#v) — BuildContent must have started serializing a.Tags; if intentional, relax this pin", got.Tags)
	}
	if got.Dependents != nil {
		t.Errorf("KNOWN DEFECT expectation changed: Dependents now round-trips (got %#v) — BuildContent must have started serializing a.Dependents; if intentional, relax this pin", got.Dependents)
	}
	if got.Interface == in.Interface {
		t.Errorf("KNOWN DEFECT expectation changed: Interface now round-trips — BuildContent must have started honoring a.Interface instead of the hardcoded spec-link boilerplate; if intentional, relax this pin")
	}
	if got.Metadata != nil {
		t.Errorf("KNOWN DEFECT expectation changed: Metadata now round-trips (got %#v) — BuildContent must have started emitting a metadata frontmatter key; if intentional, relax this pin", got.Metadata)
	}
}
