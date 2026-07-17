package exploration

// IsOrphan unit back-fill (test_atd_07_26.md §3.1 item 1, §6 WP-7): every
// short-circuit in orphan.go's IsOrphan, including the
// HierarchicalOrphanCheck dependent-rescue path and the
// BusinessLayerException.

import (
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/testutil"
)

func TestIsOrphan_BasicMatrix(t *testing.T) {
	testutil.SnapshotConfig(t)
	config.ActiveConfig.OrphanExcludedTypes = map[string]bool{"EXCLUDED_TYPE": true}
	config.ActiveConfig.HierarchicalOrphanCheck = false
	config.ActiveConfig.BusinessLayerException = false

	e := &Explorer{
		Graph:  &DependencyGraph{Atoms: make(map[string]*atom.AtomData)},
		Config: &config.ActiveConfig,
	}

	cases := []struct {
		name string
		node *atom.AtomData
		want bool
	}{
		{
			name: "stable_no_impl_is_orphan",
			node: &atom.AtomData{Status: "STABLE", Type: "MECHANIC", Layer: "IMPLEMENTATION"},
			want: true,
		},
		{
			name: "stable_with_impl_not_orphan",
			node: &atom.AtomData{Status: "STABLE", Type: "MECHANIC", Layer: "IMPLEMENTATION", Implementations: []string{"f.go:1"}},
			want: false,
		},
		{
			name: "non_stable_never_orphan",
			node: &atom.AtomData{Status: "DRAFT", Type: "MECHANIC", Layer: "IMPLEMENTATION"},
			want: false,
		},
		{
			name: "excluded_type_never_orphan",
			node: &atom.AtomData{Status: "STABLE", Type: "EXCLUDED_TYPE", Layer: "IMPLEMENTATION"},
			want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := e.IsOrphan(c.node); got != c.want {
				t.Errorf("IsOrphan(%+v) = %v, want %v", c.node, got, c.want)
			}
		})
	}
}

func TestIsOrphan_BusinessLayerException(t *testing.T) {
	testutil.SnapshotConfig(t)
	config.ActiveConfig.OrphanExcludedTypes = map[string]bool{}
	config.ActiveConfig.HierarchicalOrphanCheck = false
	config.ActiveConfig.BusinessLayerException = true

	e := &Explorer{Graph: &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}, Config: &config.ActiveConfig}

	node := &atom.AtomData{Status: "STABLE", Type: "REQUIREMENT", Layer: "BUSINESS"}
	if e.IsOrphan(node) {
		t.Error("expected BusinessLayerException to exempt a BUSINESS-layer atom from orphan detection")
	}

	// The exception is layer-specific: an otherwise-identical ARCHITECTURE
	// atom is not covered by it.
	nonBusiness := &atom.AtomData{Status: "STABLE", Type: "API", Layer: "ARCHITECTURE"}
	if !e.IsOrphan(nonBusiness) {
		t.Error("expected BusinessLayerException to NOT exempt a non-BUSINESS layer atom")
	}
}

// TestIsOrphan_HierarchicalCheck_DependentImplementedExempts pins the
// dependent-rescue branch: a STABLE parent with zero code of its own is not
// an orphan as long as HierarchicalOrphanCheck is on AND at least one of its
// dependents has implementation links -- and becomes an orphan the moment
// that stops being true.
func TestIsOrphan_HierarchicalCheck_DependentImplementedExempts(t *testing.T) {
	testutil.SnapshotConfig(t)
	config.ActiveConfig.OrphanExcludedTypes = map[string]bool{}
	config.ActiveConfig.HierarchicalOrphanCheck = true
	config.ActiveConfig.BusinessLayerException = false

	child := &atom.AtomData{ID: "child", Status: "STABLE", Implementations: []string{"f.go:1"}}
	parent := &atom.AtomData{ID: "parent", Status: "STABLE", Type: "MECHANIC", Layer: "IMPLEMENTATION", Dependents: []string{"child"}}

	e := &Explorer{
		Graph: &DependencyGraph{Atoms: map[string]*atom.AtomData{
			"child":  child,
			"parent": parent,
		}},
		Config: &config.ActiveConfig,
	}

	if e.IsOrphan(parent) {
		t.Error("expected HierarchicalOrphanCheck to exempt a parent whose dependent has code, even though the parent itself has none")
	}

	child.Implementations = nil
	if !e.IsOrphan(parent) {
		t.Error("expected the parent to become an orphan once its dependent also has no implementation")
	}
}

// TestIsOrphan_HierarchicalCheck_Disabled pins that HierarchicalOrphanCheck
// must be explicitly on for the dependent-rescue branch to run at all: with
// it off, an implemented dependent does NOT rescue an otherwise-orphaned
// parent.
func TestIsOrphan_HierarchicalCheck_Disabled(t *testing.T) {
	testutil.SnapshotConfig(t)
	config.ActiveConfig.OrphanExcludedTypes = map[string]bool{}
	config.ActiveConfig.HierarchicalOrphanCheck = false
	config.ActiveConfig.BusinessLayerException = false

	child := &atom.AtomData{ID: "child", Status: "STABLE", Implementations: []string{"f.go:1"}}
	parent := &atom.AtomData{ID: "parent", Status: "STABLE", Type: "MECHANIC", Layer: "IMPLEMENTATION", Dependents: []string{"child"}}

	e := &Explorer{
		Graph: &DependencyGraph{Atoms: map[string]*atom.AtomData{
			"child":  child,
			"parent": parent,
		}},
		Config: &config.ActiveConfig,
	}

	if !e.IsOrphan(parent) {
		t.Error("expected the parent to be an orphan when HierarchicalOrphanCheck is disabled, regardless of its dependent's implementation state")
	}
}
