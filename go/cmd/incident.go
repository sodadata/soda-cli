package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
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
		status, _ := cmd.Flags().GetString("status")
		dataset, _ := cmd.Flags().GetString("dataset")

		rows := mock.Incidents
		filtered := []map[string]string{}
		for _, i := range rows {
			if status != "" && status != "all" && i["status"] != status {
				continue
			}
			if dataset != "" && i["dataset"] != dataset {
				continue
			}
			filtered = append(filtered, i)
		}

		cols := []string{"id", "title", "dataset", "status", "created"}
		output.Render(filtered, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var incidentGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show incident details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, inc := range mock.Incidents {
			if inc["id"] == args[0] {
				keys := []string{"id", "title", "dataset", "status", "created"}
				output.RenderOne(inc, keys, GCtx)
				return nil
			}
		}
		return output.Errorf(2, "incident '%s' not found", args[0])
	},
}

var incidentUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update incident status or add a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		note, _ := cmd.Flags().GetString("note")

		if status != "" {
			fmt.Printf("  Status → %s\n", output.FmtStatus(status))
		}
		if note != "" {
			fmt.Printf("  Note added: %s\n", output.Dim.Render(note))
		}
		output.PrintSuccess(fmt.Sprintf("Incident %s updated.", args[0]), GCtx)
		return nil
	},
}

func init() {
	incidentListCmd.Flags().String("status", "open", "Filter by status: open|closed|all")
	incidentListCmd.Flags().String("dataset", "", "Filter by dataset ID")

	incidentUpdateCmd.Flags().String("status", "", "New status: open|closed")
	incidentUpdateCmd.Flags().String("note", "", "Add a note to the incident")

	incidentCmd.AddCommand(incidentListCmd, incidentGetCmd, incidentUpdateCmd)
	rootCmd.AddCommand(incidentCmd)
}
