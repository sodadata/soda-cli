package cmd

import (
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "View check definitions from Soda Cloud",
}

var checkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List checks (alias for `sodacli results list`)",
	RunE:  resultsListCmd.RunE,
}

func init() {
	checkListCmd.Flags().String("dataset", "", "Filter by dataset ID")
	checkListCmd.Flags().String("ids", "", "Comma-separated list of check IDs to fetch (cannot combine with other filters)")
	checkListCmd.Flags().String("dataset-name", "", "Filter by dataset qualified name (substring match)")
	checkListCmd.Flags().String("status", "", "Filter by status: passing|failing|error")
	checkListCmd.Flags().String("type", "check", "Filter by type: check|monitor|all")
	checkListCmd.Flags().Int("limit", 10, "Maximum number of results to show")
	checkListCmd.Flags().String("sort", "date", "Sort by column: dataset|name|column|status|date")
	checkListCmd.Flags().String("order", "desc", "Sort order: asc|desc")
	checkListCmd.Flags().String("from", "", "Show results on or after this date (YYYY-MM-DD or ISO8601)")
	checkListCmd.Flags().String("until", "", "Show results on or before this date (YYYY-MM-DD or ISO8601)")

	checkCmd.AddCommand(checkListCmd)
}
