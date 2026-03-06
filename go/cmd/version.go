package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("soda version 0.1.0-dev")
		fmt.Println("commit:  (none)")
		fmt.Println("built:   2026-03-06")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
