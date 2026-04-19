package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the build revision of atd",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("atd revision: %s\n", rootCmd.Version)
	},
}
