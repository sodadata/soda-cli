package cmd

import (
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
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
		cols := []string{"id", "name", "status", "version", "last_seen"}
		output.Render(mock.Agents, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

func init() {
	agentCmd.AddCommand(agentListCmd)
	rootCmd.AddCommand(agentCmd)
}
