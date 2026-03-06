package cmd

import (
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "View data quality signals: contract checks and monitor alerts",
}

var resultsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List results across all jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		datasource, _ := cmd.Flags().GetString("datasource")
		dataset, _ := cmd.Flags().GetString("dataset")
		resType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")

		rows := mock.Results
		filtered := []map[string]string{}
		for _, r := range rows {
			if datasource != "" && r["datasource"] != datasource {
				continue
			}
			if dataset != "" && r["dataset"] != dataset {
				continue
			}
			if resType != "" && resType != "all" && r["type"] != resType {
				continue
			}
			if status != "" && status != "all" && r["status"] != status {
				continue
			}
			filtered = append(filtered, r)
		}

		cols := []string{"dataset", "type", "name", "status", "date"}
		output.Render(filtered, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

func init() {
	resultsListCmd.Flags().String("datasource", "", "Filter by datasource ID")
	resultsListCmd.Flags().String("dataset", "", "Filter by dataset ID")
	resultsListCmd.Flags().String("type", "all", "Filter by type: check|monitor|all")
	resultsListCmd.Flags().String("status", "all", "Filter by status: passing|failing|all")
	resultsListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	resultsListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")

	resultsCmd.AddCommand(resultsListCmd)
	rootCmd.AddCommand(resultsCmd)
}
