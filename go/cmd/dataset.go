package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var datasetCmd = &cobra.Command{
	Use:   "dataset",
	Short: "Manage datasets registered in Soda Cloud",
}

var datasetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List datasets",
	RunE: func(cmd *cobra.Command, args []string) error {
		cols := []string{"id", "name", "datasource", "schema", "status", "checks", "updated"}
		output.Render(mock.Datasets, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var datasetUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update dataset metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Dataset '%s' updated.", args[0]), GCtx)
		return nil
	},
}

var datasetDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a dataset from Soda Cloud",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Dataset '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

var datasetProfilingCmd = &cobra.Command{
	Use:   "profiling <id>",
	Short: "View cached profiling data for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("  Profiling data for %s\n\n", output.Bold.Render(args[0]))
		rows := []map[string]string{
			{"column": "order_id", "type": "bigint", "nulls": "0%", "distinct": "100%", "min": "1", "max": "48231"},
			{"column": "customer_id", "type": "bigint", "nulls": "0.29%", "distinct": "34%", "min": "1001", "max": "9999"},
			{"column": "order_value", "type": "numeric", "nulls": "0%", "distinct": "72%", "min": "4.99", "max": "1249.00"},
			{"column": "created_at", "type": "timestamp", "nulls": "0%", "distinct": "100%", "min": "2025-01-01", "max": "2026-03-04"},
		}
		cols := []string{"column", "type", "nulls", "distinct", "min", "max"}
		output.Render(rows, cols, nil, GCtx)
		return nil
	},
}

var datasetProfilingRefreshCmd = &cobra.Command{
	Use:   "refresh <id>",
	Short: "Trigger a new profiling run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(output.Dim.Render("  Triggering profiling run for " + args[0] + "..."))
		output.PrintSuccess("Profiling job queued. Results will be available in ~2 minutes.", GCtx)
		return nil
	},
}

var datasetDiagnosticsCmd = &cobra.Command{
	Use:   "diagnostics <id>",
	Short: "View dataset-level diagnostics warehouse overrides",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("  %-24s %s\n", output.Bold.Render("Dataset"), args[0])
		fmt.Printf("  %-24s %s\n", output.Bold.Render("Diagnostics override"), output.Dim.Render("(none — inherits datasource config)"))
		return nil
	},
}

// dataset permissions sub-group
var datasetPermissionsCmd = &cobra.Command{
	Use:   "permissions",
	Short: "Manage dataset access permissions",
}

var datasetPermListCmd = &cobra.Command{
	Use:   "list <id>",
	Short: "List permissions for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rows := []map[string]string{
			{"principal": "alice@acme.com", "type": "user", "role": "Dataset Owner"},
			{"principal": "Data Engineering", "type": "group", "role": "Editor"},
			{"principal": "Analytics", "type": "group", "role": "Viewer"},
		}
		cols := []string{"principal", "type", "role"}
		output.Render(rows, cols, nil, GCtx)
		return nil
	},
}

var datasetPermSetCmd = &cobra.Command{
	Use:   "set <id>",
	Short: "Set permissions for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		user, _ := cmd.Flags().GetString("user")
		group, _ := cmd.Flags().GetString("group")

		principal := user
		if group != "" {
			principal = "group:" + group
		}
		output.PrintSuccess(fmt.Sprintf("Granted role '%s' to '%s' on dataset '%s'.", role, principal, args[0]), GCtx)
		return nil
	},
}

// dataset monitor sub-group
var datasetMonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Manage ML anomaly detection monitors for a dataset",
}

var datasetMonitorListCmd = &cobra.Command{
	Use:   "list <dataset-id>",
	Short: "List monitors for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rows := []map[string]string{
			{"id": "mon_001", "column": "daily_signups", "type": "anomaly", "status": "alert", "last_fired": "2026-03-05 06:45"},
			{"id": "mon_002", "column": "order_count", "type": "anomaly", "status": "passing", "last_fired": "2026-03-04 08:00"},
		}
		cols := []string{"id", "column", "type", "status", "last_fired"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var datasetMonitorAddCmd = &cobra.Command{
	Use:   "add <dataset-id>",
	Short: "Add a monitor to a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Monitor added to dataset '%s'.", args[0]), GCtx)
		return nil
	},
}

var datasetMonitorUpdateCmd = &cobra.Command{
	Use:   "update <dataset-id> <monitor-id>",
	Short: "Update a monitor",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Monitor '%s' updated.", args[1]), GCtx)
		return nil
	},
}

var datasetMonitorDeleteCmd = &cobra.Command{
	Use:   "delete <dataset-id> <monitor-id>",
	Short: "Delete a monitor",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Monitor '%s' deleted.", args[1]), GCtx)
		return nil
	},
}

func init() {
	datasetUpdateCmd.Flags().String("name", "", "New dataset name")
	datasetUpdateCmd.Flags().String("tag", "", "Add a tag")
	datasetUpdateCmd.Flags().StringArray("attr", nil, "Set attribute key=value")

	datasetListCmd.Flags().String("filter", "", "Filter datasets by query string")
	datasetListCmd.Flags().String("tag", "", "Filter by tag")

	datasetPermSetCmd.Flags().String("role", "", "Role ID to grant (required)")
	datasetPermSetCmd.Flags().String("user", "", "User email")
	datasetPermSetCmd.Flags().String("group", "", "Group ID")

	datasetPermissionsCmd.AddCommand(datasetPermListCmd, datasetPermSetCmd)

	datasetMonitorCmd.AddCommand(datasetMonitorListCmd, datasetMonitorAddCmd, datasetMonitorUpdateCmd, datasetMonitorDeleteCmd)

	datasetProfilingCmd.AddCommand(datasetProfilingRefreshCmd)

	datasetCmd.AddCommand(
		datasetListCmd,
		datasetUpdateCmd,
		datasetDeleteCmd,
		datasetProfilingCmd,
		datasetDiagnosticsCmd,
		datasetPermissionsCmd,
		datasetMonitorCmd,
	)
	rootCmd.AddCommand(datasetCmd)
}
