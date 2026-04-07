package cmd

// @spec-link [[service_atd_trace]]

import (
	"atd-tools/pkg/exploration"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace <atom_id>",
	Short: "Trace an atom's dependencies and source implementations",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docsDir, _ := cmd.Flags().GetString("docs")
		srcPath, _ := cmd.Flags().GetString("src")

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
	traceCmd.Flags().String("src", "", "Override source directory (defaults to project root)")
}

func runTrace(targetID, docsDir, srcPath string) (string, error) {
	explorer := exploration.NewExplorer(srcPath, docsDir)
	if err := explorer.Load(false); err != nil {
		return "", err
	}

	snap, err := explorer.Trace(targetID)
	if err != nil {
		return "", err
	}

	out, _ := json.MarshalIndent(snap, "", "  ")
	return string(out), nil
}

