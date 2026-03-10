package cmd

import (
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage Soda Agents",
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered Soda Agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "agent list is not yet available in the public API")
	},
}

func init() {
	agentCmd.AddCommand(agentListCmd)
	rootCmd.AddCommand(agentCmd)
}
