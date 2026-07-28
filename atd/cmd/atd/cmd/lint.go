package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/exploration"
	"atd-tools/pkg/workspace"
	"github.com/spf13/cobra"
)

// canonicalAtomTypes is the union of every atom type ATD.md sanctions: the 11
// consolidated types from §1.3 (plus the USER_STORY synonyms USECASE/WORKFLOW)
// and the additional "typical types" named in the §1.5 layer sections
// (SERVICE/DATA/BUILD/USAGE/SPECIFICATION). Kept deliberately permissive so lint
// flags only genuine typos/garbage, not the SERVICE/USAGE-vs-§1.3 tension, which
// is a spec-reconciliation decision rather than a per-atom defect.
var canonicalAtomTypes = map[string]bool{
	"CONTRACT": true, "VISION": true, "REQUIREMENT": true, "USER_STORY": true,
	"USECASE": true, "WORKFLOW": true, "RULE": true, "DOMAIN": true,
	"MECHANIC": true, "MODULE": true, "ENTITY": true, "API": true, "UI": true,
	"SERVICE": true, "DATA": true, "BUILD": true, "USAGE": true, "SPECIFICATION": true,
}

// bareAtomID strips a leading "project:" cross-project disambiguation prefix
// (ATD.md workspace reference syntax) from a [[id]] reference, leaving the
// bare atom id for comparison against a same-project governance-atom set.
func bareAtomID(ref string) string {
	if idx := strings.Index(ref, ":"); idx >= 0 {
		return ref[idx+1:]
	}
	return ref
}

// @spec-link [[service_atd_lint]]
var lintCmd = &cobra.Command{
	Use:   "lint [dir]",
	Short: "Structurally validate ATD atoms",
	Long: `lint performs a fast, deterministic check across all ATD atoms in the documentation directory.
It verifies mandatory fields, enums (Layer, Priority), section non-emptiness, and broken reference resolution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := config.DocsDir()
		if len(args) > 0 {
			dir = args[0]
		}

		out, err := runLint(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Lint Errors Found:")
			fmt.Print(out)
			os.Exit(1)
		}
		fmt.Println("All atoms passed structural validation.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}

func runLint(dir string) (string, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.atom.md"))
	if err != nil {
		return "", fmt.Errorf("failed to scan for atoms: %w", err)
	}

	var atoms []atom.AtomData

	for _, f := range files {
		a, err := atom.Parse(f)
		if err != nil {
			return "", fmt.Errorf("failed to parse %s: %w", f, err)
		}
		atoms = append(atoms, a)
	}

	var errors []string
	explorer := exploration.NewExplorer(config.ProjectRoot(), config.DocsDir())
	if err := explorer.Load(false); err != nil {
		return "", fmt.Errorf("failed to load explorer: %w", err)
	}

	// Governance atoms (CONTRACT/VISION, ATD.md §1.4) are read for governance,
	// never linked as structural ancestry — no atom may declare one as a
	// parents: target. Build the set of governance atom IDs before the main
	// validation loop so each atom's Parents can be checked against it.
	governanceAtomIDs := map[string]bool{}
	for _, a := range atoms {
		switch strings.ToUpper(strings.TrimSpace(a.Type)) {
		case "CONTRACT", "VISION":
			if a.ID != "" {
				governanceAtomIDs[a.ID] = true
			}
		}
	}

	for _, a := range atoms {
		var atomErrors []string

		// Mandatory Fields
		if a.ID == "" {
			atomErrors = append(atomErrors, "Missing mandatory field: id")
		}
		if a.HumanName == "" {
			atomErrors = append(atomErrors, "Missing mandatory field: human_name")
		}
		if a.Type == "" {
			atomErrors = append(atomErrors, "Missing mandatory field: type")
		} else if !canonicalAtomTypes[strings.ToUpper(strings.TrimSpace(a.Type))] {
			atomErrors = append(atomErrors, fmt.Sprintf("Non-canonical type: %s (see ATD.md §1.3/§1.5)", a.Type))
		}
		if a.Layer == "" {
			atomErrors = append(atomErrors, "Missing mandatory field: layer")
		} else {
			if a.Layer != "BUSINESS" && a.Layer != "ARCHITECTURE" && a.Layer != "IMPLEMENTATION" {
				atomErrors = append(atomErrors, fmt.Sprintf("Invalid layer enum: %s", a.Layer))
			}
		}

		if a.Version == "" {
			atomErrors = append(atomErrors, "Missing mandatory field: version")
		}
		if a.Status == "" {
			atomErrors = append(atomErrors, "Missing mandatory field: status")
		}
		if a.Priority == "" {
			atomErrors = append(atomErrors, "Missing mandatory field: priority")
		} else {
			if a.Priority != "1" && a.Priority != "2" && a.Priority != "3" && a.Priority != "4" && a.Priority != "5" && a.Priority != "CORE" {
				atomErrors = append(atomErrors, fmt.Sprintf("Invalid priority enum: %s", a.Priority))
			}
		}

		// Sections
		if strings.TrimSpace(a.Intent) == "" {
			atomErrors = append(atomErrors, "Missing mandatory section: ## INTENT")
		}
		if strings.TrimSpace(a.Logic) == "" {
			atomErrors = append(atomErrors, "Missing mandatory section: ## THE RULE / LOGIC")
		}
		if strings.TrimSpace(a.Interface) == "" {
			atomErrors = append(atomErrors, "Missing mandatory section: ## TECHNICAL INTERFACE")
		}
		if strings.TrimSpace(a.Expectation) == "" {
			atomErrors = append(atomErrors, "Missing mandatory section: ## EXPECTATION")
		}

		// Links
		for _, p := range a.Parents {
			cleanP := strings.TrimSpace(p)
			if _, err := explorer.ResolveAtom(cleanP); err != nil {
				if err == workspace.ErrUnknownProject {
					atomErrors = append(atomErrors, fmt.Sprintf("Unresolved parent link (unknown project): [[%s]]", p))
				} else {
					atomErrors = append(atomErrors, fmt.Sprintf("Unresolved parent link: [[%s]]", p))
				}
			}
			if governanceAtomIDs[bareAtomID(cleanP)] {
				atomErrors = append(atomErrors, fmt.Sprintf("CONTRACT/VISION referenced as parent: [[%s]] -- governance atoms are read for governance, never linked as structural ancestry (ATD.md §1.4)", p))
			}
		}
		for _, p := range a.Dependents {
			cleanP := strings.TrimSpace(p)
			if _, err := explorer.ResolveAtom(cleanP); err != nil {
				if err == workspace.ErrUnknownProject {
					atomErrors = append(atomErrors, fmt.Sprintf("Unresolved dependent link (unknown project): [[%s]]", p))
				} else {
					atomErrors = append(atomErrors, fmt.Sprintf("Unresolved dependent link: [[%s]]", p))
				}
			}
		}

		// Traceability (Missing Proof)
		hasImplementation := false
		for _, sl := range explorer.SpecLinks {
			if sl.AtomID == a.ID {
				hasImplementation = true
				break
			}
		}
		hasTest := false
		for _, tl := range explorer.TestLinks {
			if tl.AtomID == a.ID {
				hasTest = true
				break
			}
		}

		if hasImplementation && !hasTest && a.Layer == "IMPLEMENTATION" {
			atomErrors = append(atomErrors, "Traceability Gap: Found @spec-link in implementation but 0 @test-link (Missing Proof)")
		}

		if len(atomErrors) > 0 {
			id := a.ID
			if id == "" {
				id = filepath.Base(a.FilePath)
			}
			errors = append(errors, fmt.Sprintf("[%s]", id))
			for _, errStr := range atomErrors {
				errors = append(errors, "  - "+errStr)
			}
		}
	}

	// Project governance (§1.4): a project must have exactly ONE CONTRACT and one
	// VISION atom. Uniqueness (>1) is always a violation. Presence is required
	// only once the corpus contains at least one BUSINESS-layer atom, since
	// CONTRACT/VISION exist to gate BUSINESS-layer evolution — an architecture- or
	// implementation-only fixture is legitimately exempt.
	var contractCount, visionCount, businessCount int
	for _, a := range atoms {
		switch strings.ToUpper(strings.TrimSpace(a.Type)) {
		case "CONTRACT":
			contractCount++
		case "VISION":
			visionCount++
		}
		if strings.ToUpper(strings.TrimSpace(a.Layer)) == "BUSINESS" {
			businessCount++
		}
	}
	var govErrors []string
	if contractCount > 1 {
		govErrors = append(govErrors, fmt.Sprintf("Multiple CONTRACT atoms (%d); exactly one required (ATD.md §1.4)", contractCount))
	}
	if visionCount > 1 {
		govErrors = append(govErrors, fmt.Sprintf("Multiple VISION atoms (%d); exactly one required (ATD.md §1.4)", visionCount))
	}
	if businessCount > 0 {
		if contractCount == 0 {
			govErrors = append(govErrors, "Missing CONTRACT atom (ATD.md §1.4: exactly one required per project with BUSINESS atoms)")
		}
		if visionCount == 0 {
			govErrors = append(govErrors, "Missing VISION atom (ATD.md §1.4: exactly one required per project with BUSINESS atoms)")
		}
	}
	if len(govErrors) > 0 {
		errors = append(errors, "[PROJECT GOVERNANCE]")
		for _, e := range govErrors {
			errors = append(errors, "  - "+e)
		}
	}

	if len(errors) > 0 {
		return strings.Join(errors, "\n"), fmt.Errorf("linting failed")
	}

	return "", nil
}
