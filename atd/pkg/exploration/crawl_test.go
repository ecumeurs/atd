package exploration

// CrawlDocs/CrawlSrc/CrawlWorkspaceDocs unit back-fill (test_atd_07_26.md
// §3.1 item 1, §6 WP-7): crawling a project tree via testdata/fixture_project
// (and testdata/fixture_workspace for the workspace-aggregation path).

import (
	"testing"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/testutil"
)

func TestCrawlDocs_BuildsGraphFromFixture(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	graph := &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	if err := CrawlDocs(sb.DocsDir, graph); err != nil {
		t.Fatalf("CrawlDocs: %v", err)
	}

	// fixture_project/docs holds exactly these atoms (test_atd_07_26.md §3.2).
	want := []string{
		"api_zzfix_beta", "contract_zzfix", "mech_zzfix_gamma", "mech_zzfix_orphan",
		"req_zzfix_alpha", "req_zzfix_draft", "req_zzfix_tech_debt_backlog",
		"rule_zzfix_untested", "vision_zzfix",
	}
	for _, id := range want {
		if _, ok := graph.Atoms[id]; !ok {
			t.Errorf("expected CrawlDocs to have loaded atom %q", id)
		}
	}
	if len(graph.Atoms) != len(want) {
		t.Errorf("expected exactly %d atoms, got %d: %v", len(want), len(graph.Atoms), testutil.SortedKeys(graph.Atoms))
	}
}

func TestCrawlDocsWithTag_SetsProjectMetadata(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	graph := &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	if err := CrawlDocsWithTag(sb.DocsDir, graph, "myproj"); err != nil {
		t.Fatalf("CrawlDocsWithTag: %v", err)
	}

	node := graph.Atoms["api_zzfix_beta"]
	if node == nil {
		t.Fatal("expected api_zzfix_beta to be loaded")
	}
	if node.GetProject() != "myproj" {
		t.Errorf("expected project tag %q, got %q", "myproj", node.GetProject())
	}
}

// TestCrawlSrc_RecordsMultipleImplementationsAndTestLinks pins that CrawlSrc
// (like extractLinks) does not dedup: api_zzfix_beta is @spec-link-tagged
// twice in src/beta.go (Beta + BetaHelper, the S2 dedup fixture case) and
// both raw occurrences must show up as separate Implementations entries.
func TestCrawlSrc_RecordsMultipleImplementationsAndTestLinks(t *testing.T) {
	t.Parallel()
	sb := testutil.Sandbox(t, "fixture_project")

	graph := &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	if err := CrawlDocs(sb.DocsDir, graph); err != nil {
		t.Fatalf("CrawlDocs: %v", err)
	}
	if err := CrawlSrc(sb.SrcDir, graph); err != nil {
		t.Fatalf("CrawlSrc: %v", err)
	}

	beta := graph.Atoms["api_zzfix_beta"]
	if beta == nil {
		t.Fatal("api_zzfix_beta missing from graph")
	}
	if len(beta.Implementations) != 2 {
		t.Errorf("expected 2 raw implementation entries (CrawlSrc does not dedup), got %d: %v", len(beta.Implementations), beta.Implementations)
	}
	if !beta.HasTests {
		t.Error("expected api_zzfix_beta.HasTests to be set from beta_test.go's @test-link")
	}

	untested := graph.Atoms["rule_zzfix_untested"]
	if untested == nil {
		t.Fatal("rule_zzfix_untested missing from graph")
	}
	if len(untested.Implementations) != 1 {
		t.Errorf("expected 1 implementation entry, got %d: %v", len(untested.Implementations), untested.Implementations)
	}
	if untested.HasTests {
		t.Error("expected rule_zzfix_untested.HasTests to remain false (no @test-link anywhere in the fixture)")
	}

	orphan := graph.Atoms["mech_zzfix_orphan"]
	if orphan == nil {
		t.Fatal("mech_zzfix_orphan missing from graph")
	}
	if len(orphan.Implementations) != 0 {
		t.Errorf("expected mech_zzfix_orphan to have zero implementations (it is the deliberate orphan), got %v", orphan.Implementations)
	}
}

func TestCrawlWorkspaceDocsWithConfig_NoWorkspaceErrors(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}
	graph := &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}

	if err := CrawlWorkspaceDocsWithConfig(graph, cfg); err == nil {
		t.Fatal("expected an error when cfg.Workspace is nil")
	}
}

// TestCrawlWorkspaceDocs_AggregatesAcrossProjects covers the
// config.ActiveConfig-reading entry point (CrawlWorkspaceDocs), so it
// mutates the global directly and cannot run t.Parallel() alongside
// testutil.Sandbox-based tests whose Run() takes the same lock only around
// its own critical section.
func TestCrawlWorkspaceDocs_AggregatesAcrossProjects(t *testing.T) {
	testutil.SnapshotConfigLocked(t)
	ws := testutil.Sandbox(t, "fixture_workspace")

	config.ActiveConfig.Workspace = &config.WorkspaceConfig{
		WorkspaceName: "zzfix_workspace",
		LoadedFrom:    ws.Root,
		Projects: []config.ProjectConfig{
			{Name: "zzfix_a", Path: "./zzfix_a"},
			{Name: "zzfix_b", Path: "./zzfix_b"},
		},
	}

	graph := &DependencyGraph{Atoms: make(map[string]*atom.AtomData)}
	if err := CrawlWorkspaceDocs(graph); err != nil {
		t.Fatalf("CrawlWorkspaceDocs: %v", err)
	}

	a := graph.Atoms["req_zzfix_ws_alpha"]
	if a == nil || a.GetProject() != "zzfix_a" {
		t.Errorf("expected req_zzfix_ws_alpha tagged zzfix_a, got %+v", a)
	}
	b := graph.Atoms["api_zzfix_ws_beta"]
	if b == nil || b.GetProject() != "zzfix_b" {
		t.Errorf("expected api_zzfix_ws_beta tagged zzfix_b, got %+v", b)
	}
}
