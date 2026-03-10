package cmd

import (
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var incidentCmd = &cobra.Command{
	Use:   "incident",
	Short: "Manage data quality incidents",
}

var incidentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List incidents",
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "incident list is not yet available in the public API")
	},
}

var incidentGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show incident details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "incident get is not yet available in the public API")
	},
}

var incidentUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an incident",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "incident update is not yet available in the public API")
	},
}

func init() {
	incidentListCmd.Flags().String("status", "", "Filter by status: reported|investigating|fixing|resolved")
	incidentListCmd.Flags().String("dataset", "", "Filter by dataset ID")

	incidentUpdateCmd.Flags().String("title", "", "New title")
	incidentUpdateCmd.Flags().String("severity", "", "Severity: minor|major|critical")
	incidentUpdateCmd.Flags().String("description", "", "Description")
	incidentUpdateCmd.Flags().String("assigned-to", "", "Assigned user email")
	incidentUpdateCmd.Flags().String("status", "", "Status: reported|investigating|fixing|resolved")

	incidentCmd.AddCommand(incidentListCmd, incidentGetCmd, incidentUpdateCmd)
	rootCmd.AddCommand(incidentCmd)
}
