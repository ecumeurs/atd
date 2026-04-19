package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"atd-tools/config"
	"github.com/spf13/cobra"
)

var Verbose bool

var rootCmd = &cobra.Command{
	Use:     "atd",
	Short:   "Atomic Traceable Documentation toolkit",
	Version: GetVersion(),
	Long: fmt.Sprintf(`atd is the unified CLI for managing Atomic Traceable Documentation.
It provides subcommands for dissecting documents, auditing atoms,
indexing codebases, searching semantically, and more.

Configuration is loaded from the .atd file found at the project root.

Revision: %s`, GetVersion()),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return config.Load()
	},
}

func GetVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	revision := ""
	modified := false
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			revision = setting.Value
		}
		if setting.Key == "vcs.modified" {
			modified = setting.Value == "true"
		}
	}
	if revision == "" {
		return "unknown"
	}
	if modified {
		revision += "*"
	}
	return revision
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
