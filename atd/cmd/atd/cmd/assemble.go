package cmd

// @spec-link [[mechanic_atd_assemble]]

import (
	"fmt"

	"atd-tools/config"
	"atd-tools/pkg/exploration"
	"github.com/spf13/cobra"
)

var assembleCmd = &cobra.Command{
	Use:   "assemble",
	Short: "Stitch ATD atoms together into a cohesive document",
	Long: `Recursively gather ATD atoms starting from specified IDs and 
assemble their content into a single document. Supports structuring by layer and LLM summarization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		starts, _ := cmd.Flags().GetString("starts")
		intent, _ := cmd.Flags().GetString("intent")
		length, _ := cmd.Flags().GetString("length")
		structured, _ := cmd.Flags().GetBool("structured")
		asJSON, _ := cmd.Flags().GetBool("json")
		onlyParents, _ := cmd.Flags().GetBool("only-parents")
		onlyDependents, _ := cmd.Flags().GetBool("only-dependents")

		docsDir := config.DocsDir()

		text, err := runAssemble(starts, intent, length, structured, asJSON, onlyParents, onlyDependents, docsDir)
		if err != nil {
			return err
		}
		if text != "" {
			fmt.Println(text)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(assembleCmd)
	assembleCmd.Flags().String("starts", "", "Comma-separated list of Root Atom IDs")
	assembleCmd.Flags().String("intent", "Executive Summary", "The intent the LLM should focus on (e.g., summarize, executive summary)")
	assembleCmd.Flags().String("length", "default", "Length constraint (short, default, extended, long)")
	assembleCmd.Flags().Bool("structured", false, "Group atoms by layer and perform multi-pass summarization")
	assembleCmd.Flags().Bool("json", false, "Output results as JSON")
	assembleCmd.Flags().Bool("only-parents", false, "Restrict traversal to ancestry (upwards) only")
	assembleCmd.Flags().Bool("only-dependents", false, "Restrict traversal to descendants (downwards) only")
}

func runAssemble(starts, intent, length string, structured, asJSON, onlyParents, onlyDependents bool, docsDir string) (string, error) {
	opts := exploration.AssembleOptions{
		Starts:         starts,
		Intent:         intent,
		Length:         length,
		Structured:     structured,
		AsJSON:         asJSON,
		OnlyParents:    onlyParents,
		OnlyDependents: onlyDependents,
		DocsDir:        docsDir,
	}
	return exploration.Assemble(opts)
}
