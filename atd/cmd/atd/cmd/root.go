package cmd

import (
    "fmt"
    "os"
    "atd-tools/config"
    "github.com/spf13/cobra"
)

var Verbose bool

var rootCmd = &cobra.Command{
    Use:   "atd",
    Short: "Atomic Traceable Documentation toolkit",
    Long: `atd is the unified CLI for managing Atomic Traceable Documentation.
It provides subcommands for dissecting documents, auditing atoms,
indexing codebases, searching semantically, and more.

Configuration is loaded from the .atd file found at the project root.`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return config.Load()
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}
