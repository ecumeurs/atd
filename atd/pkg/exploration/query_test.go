package exploration

// Query unit tests. Query is a small, pure, LLM-free part of this package
// (unlike Search/SemanticSearch, which need ollama/store) and its 0% prior
// coverage was easy headroom toward this package's WP-7 coverage floor
// (test_atd_07_26.md §6 WP-7: pkg/exploration >= 60%).

import (
	"sort"
	"testing"

	"atd-tools/pkg/atom"
)

func TestQuery(t *testing.T) {
	e := &Explorer{Graph: &DependencyGraph{Atoms: map[string]*atom.AtomData{
		"zz_query_alpha": {ID: "zz_query_alpha", HumanName: "Alpha Widget", Status: "STABLE", Layer: "BUSINESS", Type: "REQUIREMENT", Tags: []string{"zzfix", "widgets"}},
		"zz_query_beta":  {ID: "zz_query_beta", HumanName: "Beta Gadget", Status: "DRAFT", Layer: "ARCHITECTURE", Type: "API", Tags: []string{"zzfix"}},
	}}}

	idsOf := func(matches []*atom.AtomData) []string {
		var ids []string
		for _, m := range matches {
			ids = append(ids, m.ID)
		}
		sort.Strings(ids)
		return ids
	}

	t.Run("field_id_exact_substring", func(t *testing.T) {
		got := idsOf(e.Query("id", "zz_query_alpha"))
		if len(got) != 1 || got[0] != "zz_query_alpha" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("field_status_case_insensitive", func(t *testing.T) {
		got := idsOf(e.Query("status", "stable"))
		if len(got) != 1 || got[0] != "zz_query_alpha" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("field_type", func(t *testing.T) {
		got := idsOf(e.Query("type", "API"))
		if len(got) != 1 || got[0] != "zz_query_beta" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("field_tags_matches_joined_string", func(t *testing.T) {
		got := idsOf(e.Query("tags", "widgets"))
		if len(got) != 1 || got[0] != "zz_query_alpha" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("no_field_falls_back_to_id_name_or_tags", func(t *testing.T) {
		got := idsOf(e.Query("", "gadget"))
		if len(got) != 1 || got[0] != "zz_query_beta" {
			t.Errorf("got %v", got)
		}
		// A shared tag with no field specified matches both.
		got = idsOf(e.Query("", "zzfix"))
		if len(got) != 2 {
			t.Errorf("expected both atoms to match shared tag zzfix, got %v", got)
		}
	})

	t.Run("no_match_returns_empty_not_error", func(t *testing.T) {
		// Query has no error return at all -- a miss is always a silent
		// empty/nil slice (see cmd/atd/cmd/scenario_test.go's S1 "query"
		// subtest, which pins the same "silent empty" shape at the CLI
		// layer as a KNOWN DEFECT relative to check/trace's loud errors).
		got := e.Query("id", "does_not_exist_anywhere")
		if len(got) != 0 {
			t.Errorf("expected no matches, got %v", got)
		}
	})
}
