package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atd-tools/config"
	"atd-tools/pkg/atom"
	"atd-tools/pkg/exploration"
	"github.com/spf13/cobra"
)

// @spec-link [[mechanic_atd_lint]]
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

	knownAtoms := make(map[string]bool)
	var atoms []atom.AtomData

	for _, f := range files {
		a, err := atom.Parse(f)
		if err != nil {
			return "", fmt.Errorf("failed to parse %s: %w", f, err)
		}
		atoms = append(atoms, a)
		if a.ID != "" {
			knownAtoms[a.ID] = true
		}
	}

	var errors []string
	explorer := exploration.NewExplorer(config.ProjectRoot(), config.DocsDir())
	_ = explorer.Load(false)

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
			if !knownAtoms[strings.TrimSpace(p)] {
				atomErrors = append(atomErrors, fmt.Sprintf("Unresolved parent link: [[%s]]", p))
			}
		}
		for _, p := range a.Dependents {
			if !knownAtoms[strings.TrimSpace(p)] {
				atomErrors = append(atomErrors, fmt.Sprintf("Unresolved dependent link: [[%s]]", p))
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

	if len(errors) > 0 {
		return strings.Join(errors, "\n"), fmt.Errorf("linting failed")
	}

	return "", nil
}
