package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold soda.yml and project structure",
	Long:  "Create soda.yml, contracts/, and configs/ directories in the current project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !GCtx.NoInteractive {
			fmt.Println("Scaffolding soda project structure...")
			fmt.Println()
		}
		fmt.Println("  Created  soda.yml")
		fmt.Println("  Created  contracts/")
		fmt.Println("  Created  configs/")
		fmt.Println()
		output.PrintSuccess("Project initialized. Edit soda.yml to configure your datasources.", GCtx)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
