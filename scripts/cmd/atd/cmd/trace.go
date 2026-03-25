package cmd

// @spec-link [[service_atd_trace]]

import (
	"atd-tools/config"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type TraceSnapshot struct {
	TargetID      string          `json:"target_id"`
	Layer         string          `json:"layer"`
	HealthSummary HealthSummary   `json:"health_summary"`
	Metrics       TraceMetrics    `json:"metrics"`
	GraphSlice    TraceGraphSlice `json:"graph_slice"`
	Warnings      []string        `json:"warnings"`
}

type HealthSummary struct {
	AncestryComplete   bool    `json:"ancestry_complete"`
	HasCustomerOrigin  bool    `json:"has_customer_origin"`
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

var traceCmd = &cobra.Command{
	Use:   "trace <atom_id>",
	Short: "Trace an atom's dependencies and source implementations",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		srcPath, _ := cmd.Flags().GetString("src")
		if docsDir == "" {
			docsDir = config.DocsDir()
		}
		if srcPath == "" {
			srcPath = "."
		}

		out, err := runTrace(args[0], docsDir, srcPath)
		if err != nil {
			return err
		}
		fmt.Println(out)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(traceCmd)
	traceCmd.Flags().String("docs", "", "Override docs directory")
	traceCmd.Flags().String("src", ".", "Path to source code")
}

func runTrace(targetID, docsDir, srcPath string) (string, error) {
	// 1. Build the dependency graph and scan source code for @spec-link
	graph := &DependencyGraph{Atoms: make(map[string]*AtomNode)}
	if err := crawlDocs(docsDir, graph); err != nil {
		return "", err
	}
	if err := crawlSrc(srcPath, graph); err != nil {
		return "", err
	}

	target, exists := graph.Atoms[targetID]
	if !exists {
		return "", fmt.Errorf("atom not found: %s", targetID)
	}

	// 2. We also need @test-link information. We leverage runTestLinks.
	testLinksJSON, err := runTestLinks(srcPath, "", docsDir)
	var allTestLinks []TestLink
	if err == nil {
		_ = json.Unmarshal([]byte(testLinksJSON), &allTestLinks)
	}

	testMap := make(map[string][]string) // atom -> slice of test files
	for _, tl := range allTestLinks {
		testMap[tl.AtomID] = append(testMap[tl.AtomID], tl.TestFile)
	}

	snap := TraceSnapshot{
		TargetID: targetID,
		Layer:    target.Layer,
		GraphSlice: TraceGraphSlice{
			Parents:    []string{},
			Dependents: []string{},
			CodeLinks:  []string{},
			TestLinks:  []string{},
		},
		Warnings: []string{},
	}

	// 3. Walk UP
	ancestryComplete := true
	visitedUp := make(map[string]bool)
	var walkUp func(string)
	walkUp = func(id string) {
		if visitedUp[id] {
			return
		}
		visitedUp[id] = true
		node, ok := graph.Atoms[id]
		if !ok {
			return
		}
		if id != targetID {
			snap.GraphSlice.Parents = append(snap.GraphSlice.Parents, id)
			if node.Status != "STABLE" {
				ancestryComplete = false
			}
		}
		for _, p := range node.Parents {
			walkUp(p)
		}
	}
	walkUp(targetID)

	// 4. Walk DOWN
	visitedDown := make(map[string]bool)
	var walkDown func(string)
	walkDown = func(id string) {
		if visitedDown[id] {
			return
		}
		visitedDown[id] = true
		node, ok := graph.Atoms[id]
		if !ok {
			return
		}

		if id != targetID {
			snap.GraphSlice.Dependents = append(snap.GraphSlice.Dependents, id)
		}
		for _, d := range node.Dependents {
			walkDown(d)
		}
	}
	walkDown(targetID)

	// 5. Calculate implementations and test links over the dependents
	codeFiles := make(map[string]bool)
	testFiles := make(map[string]bool)

	implementedCount := 0
	testedCount := 0
	uniqueCodeForTargetOrDescendants := make(map[string]bool)

	// Helper to strip line number from implementation "file.go:12"
	getFile := func(impl string) string {
		parts := strings.Split(impl, ":")
		if len(parts) > 0 {
			return parts[0]
		}
		return impl
	}

	for _, id := range snap.GraphSlice.Dependents {
		node := graph.Atoms[id]
		if node == nil {
			continue
		}

		isImplemented := len(node.Implementations) > 0
		isTested := len(testMap[id]) > 0

		if isImplemented {
			implementedCount++
			for _, impl := range node.Implementations {
				f := getFile(impl)
				codeFiles[f] = true
				uniqueCodeForTargetOrDescendants[f] = true
			}
		}
		if isTested {
			for _, tf := range testMap[id] {
				testFiles[tf] = true
				uniqueCodeForTargetOrDescendants[tf] = true
			}
			if isImplemented {
				testedCount++
			}
		}
	}

	targetImplemented := len(target.Implementations) > 0
	if targetImplemented {
		for _, impl := range target.Implementations {
			uniqueCodeForTargetOrDescendants[getFile(impl)] = true
		}
	}

	// 6. Layer Metrics and Warnings
	foundArchDesc := false
	foundImplDesc := false
	for _, d := range snap.GraphSlice.Dependents {
		layer := graph.Atoms[d].Layer
		if layer == "ARCHITECTURE" {
			foundArchDesc = true
		}
		if layer == "IMPLEMENTATION" {
			foundImplDesc = true
		}
	}

	foundArchAnc := false
	foundCustAnc := false
	for _, p := range snap.GraphSlice.Parents {
		layer := graph.Atoms[p].Layer
		if layer == "ARCHITECTURE" {
			foundArchAnc = true
		}
		if layer == "CUSTOMER" {
			foundCustAnc = true
		}
	}

	switch target.Layer {
	case "CUSTOMER":
		if !foundArchDesc {
			snap.Warnings = append(snap.Warnings, "Customer atom has no Architecture dependents")
		}
		if !foundImplDesc {
			snap.Warnings = append(snap.Warnings, "Customer atom has no Implementation dependents")
		}
	case "ARCHITECTURE":
		if !foundImplDesc {
			snap.Warnings = append(snap.Warnings, "Architecture atom has no Implementation dependents")
		}
		if !foundCustAnc {
			snap.Warnings = append(snap.Warnings, "Architecture atom has no Customer origin")
		}
	case "IMPLEMENTATION":
		if !foundArchAnc {
			snap.Warnings = append(snap.Warnings, "Implementation atom has no Architecture origin")
		}
		if len(uniqueCodeForTargetOrDescendants) == 0 {
			snap.Warnings = append(snap.Warnings, "Implementation atom has no linked code")
		}
	}

	for f := range codeFiles {
		snap.GraphSlice.CodeLinks = append(snap.GraphSlice.CodeLinks, f)
	}
	for f := range testFiles {
		snap.GraphSlice.TestLinks = append(snap.GraphSlice.TestLinks, f)
	}

	snap.HealthSummary.AncestryComplete = ancestryComplete
	snap.HealthSummary.HasCustomerOrigin = foundCustAnc

	if len(snap.GraphSlice.Dependents) > 0 {
		snap.HealthSummary.ImplementationRate = float64(implementedCount) / float64(len(snap.GraphSlice.Dependents))
	} else {
		snap.HealthSummary.ImplementationRate = 1.0 // Vacuously true
	}

	if implementedCount > 0 {
		snap.HealthSummary.TestCoverageRate = float64(testedCount) / float64(implementedCount)
	} else {
		snap.HealthSummary.TestCoverageRate = 1.0
	}

	snap.Metrics.TotalDependents = len(snap.GraphSlice.Dependents)
	snap.Metrics.ImplementedDependents = implementedCount
	snap.Metrics.TotalCodeFiles = len(codeFiles)
	snap.Metrics.TotalTests = len(testFiles)

	out, _ := json.MarshalIndent(snap, "", "  ")
	return string(out), nil
}
