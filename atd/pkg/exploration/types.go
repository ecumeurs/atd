package exploration

import (
	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/workspace"
)

// @spec-link [[mechanic_atd_exploration_graph]]
// Explorer manages the ATD dependency graph and source code links with in-memory caching.
type Explorer struct {
	ProjectRoot string
	DocsDir     string
	Graph       *DependencyGraph
	SpecLinks   []SpecLink
	TestLinks   []TestLink
	Workspace   *workspace.Workspace
	Resolver    *workspace.Resolver
	Index       *workspace.AtomIndex
	Config      *config.Config
}

type DependencyGraph struct {
	Atoms map[string]*atom.AtomData `json:"atoms"`
}

type SpecLink struct {
	AtomID   string `json:"atom_id"`
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
}

type TestLink struct {
	AtomID   string `json:"atom_id"`
	TestFile string `json:"test_file"`
	Line     int    `json:"line"`
}

type TraceSnapshot struct {
	TargetID      string               `json:"target_id"`
	Layer         string               `json:"layer"`
	HealthSummary HealthSummary        `json:"health_summary"`
	Metrics       TraceMetrics         `json:"metrics"`
	GraphSlice    TraceGraphSlice      `json:"graph_slice"`
	Context       map[string]AtomBrief `json:"context"`
	Warnings      []string             `json:"warnings"`
	Summary       string               `json:"summary,omitempty"`
}

type AtomBrief struct {
	ID        string `json:"id"`
	HumanName string `json:"human_name"`
	Type      string `json:"type"`
	Layer     string `json:"layer"`
	Intent    string `json:"intent"`
	Logic     string `json:"logic"`
	FilePath  string `json:"file_path"`
}

type HealthSummary struct {
	AncestryComplete   bool    `json:"ancestry_complete"`
	HasBusinessOrigin  bool    `json:"has_business_origin"`
	ImplementationRate float64 `json:"implementation_rate"`
	TestCoverageRate   float64 `json:"test_coverage_rate"`
}

type TraceMetrics struct {
	TotalDependents       int `json:"total_dependents"`
	ImplementedDependents int `json:"implemented_dependents"`
	TotalCodeFiles        int `json:"total_code_files"`
	TotalTests            int `json:"total_tests"`
}

type TraceGraphSlice struct {
	Parents    []string `json:"parents"`
	Dependents []string `json:"dependents"`
	CodeLinks  []string `json:"code_links"`
	TestLinks  []string `json:"test_links"`
}