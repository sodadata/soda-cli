package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Manage ML anomaly detection monitors",
}

var monitorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List monitors",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataset, _ := cmd.Flags().GetString("dataset")
		monType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")

		rows := mock.Monitors
		filtered := []map[string]string{}
		for _, m := range rows {
			if dataset != "" && m["dataset"] != dataset {
				continue
			}
			if monType != "" && m["type"] != monType {
				continue
			}
			if status != "" && status != "all" && m["status"] != status {
				continue
			}
			filtered = append(filtered, m)
		}

		cols := []string{"id", "dataset", "type", "metric", "status", "last_run"}
		output.Render(filtered, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

// monitor config — dataset-level MM settings

var monitorConfigCmd = &cobra.Command{
	Use:   "config <dataset-id>",
	Short: "View or update dataset-level monitor settings",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		schedule, _ := cmd.Flags().GetString("schedule")
		historical, _ := cmd.Flags().GetString("historical")
		historicalDays, _ := cmd.Flags().GetInt("historical-days")

		changed := enable || disable || schedule != "" || historical != "" || historicalDays != 0
		if !changed {
			// no flags → same as get
			return runMonitorConfigGet(args[0])
		}

		output.PrintSuccess(fmt.Sprintf("Monitor config updated for dataset '%s'.", args[0]), GCtx)
		return nil
	},
}

var monitorConfigGetCmd = &cobra.Command{
	Use:   "get <dataset-id>",
	Short: "View current monitor settings for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMonitorConfigGet(args[0])
	},
}

func runMonitorConfigGet(datasetID string) error {
	fmt.Printf("  %-22s %s\n", output.Bold.Render("Dataset"), datasetID)
	fmt.Printf("  %-22s %s\n", output.Bold.Render("Monitors enabled"), "yes")
	fmt.Printf("  %-22s %s\n", output.Bold.Render("Schedule"), "0 6 * * *")
	fmt.Printf("  %-22s %s\n", output.Bold.Render("Historical training"), "yes (90 days)")
	return nil
}

var monitorAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a monitor to a dataset",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataset, _ := cmd.Flags().GetString("dataset")
		monType, _ := cmd.Flags().GetString("type")
		if dataset == "" || monType == "" {
			return output.Errorf(2, "--dataset and --type are required")
		}
		output.PrintSuccess(fmt.Sprintf("Monitor (%s) added to dataset '%s'.", monType, dataset), GCtx)
		return nil
	},
}

var monitorUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a monitor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Monitor '%s' updated.", args[0]), GCtx)
		return nil
	},
}

var monitorDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a monitor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Monitor '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

func init() {
	monitorListCmd.Flags().String("dataset", "", "Filter by dataset ID")
	monitorListCmd.Flags().String("type", "", "Filter by type: dataset|column|group-by|custom")
	monitorListCmd.Flags().String("status", "", "Filter by status: enabled|disabled|all")

	monitorConfigCmd.Flags().Bool("enable", false, "Enable monitoring for this dataset")
	monitorConfigCmd.Flags().Bool("disable", false, "Disable monitoring for this dataset")
	monitorConfigCmd.Flags().String("schedule", "", "Cron schedule expression")
	monitorConfigCmd.Flags().String("historical", "", "Enable historical training: true|false")
	monitorConfigCmd.Flags().Int("historical-days", 0, "Number of historical days for training")
	monitorConfigCmd.AddCommand(monitorConfigGetCmd)

	monitorAddCmd.Flags().String("dataset", "", "Dataset ID (required)")
	monitorAddCmd.Flags().String("type", "", "Monitor type: dataset|column|group-by|custom (required)")
	monitorAddCmd.Flags().String("metric", "", "Metric to monitor")
	monitorAddCmd.Flags().String("column", "", "Column name (for column/group-by monitors)")
	monitorAddCmd.Flags().StringArray("group-by", nil, "Group-by column (repeatable)")
	monitorAddCmd.Flags().String("valid-min", "", "Lower bound for anomaly detection")
	monitorAddCmd.Flags().String("valid-max", "", "Upper bound for anomaly detection")
	monitorAddCmd.Flags().String("threshold", "", "Detection threshold: upper|lower|both")
	monitorAddCmd.Flags().StringArray("exclude-values", nil, "Values to exclude")
	monitorAddCmd.Flags().String("name", "", "Monitor name (for custom monitors)")
	monitorAddCmd.Flags().String("sql", "", "SQL query (for custom monitors)")
	monitorAddCmd.Flags().String("sql-file", "", "Path to SQL file (for custom monitors)")
	monitorAddCmd.Flags().String("result-metric", "", "Result metric column (for custom monitors)")
	monitorAddCmd.Flags().String("schedule", "", "Cron schedule (for custom monitors)")
	monitorAddCmd.Flags().StringArray("exclude-segments", nil, "Segments to exclude (for group-by monitors)")

	monitorCmd.AddCommand(monitorListCmd, monitorConfigCmd, monitorAddCmd, monitorUpdateCmd, monitorDeleteCmd)
	rootCmd.AddCommand(monitorCmd)
}
