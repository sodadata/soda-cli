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
	Short: "Update dataset metadata (owner, tags, description)",
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

// ── dataset time-partition ────────────────────────────────────────────────────

var datasetTimePartitionCmd = &cobra.Command{
	Use:   "time-partition <id>",
	Short: "View or set the time-partition column for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		column, _ := cmd.Flags().GetString("column")
		if column == "" {
			// no flags → same as get
			return runDatasetTimePartitionGet(args[0])
		}
		output.PrintSuccess(fmt.Sprintf("Time-partition column set to '%s' for dataset '%s'.", column, args[0]), GCtx)
		return nil
	},
}

var datasetTimePartitionGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show the current time-partition column",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDatasetTimePartitionGet(args[0])
	},
}

func runDatasetTimePartitionGet(id string) error {
	fmt.Printf("  %-22s %s\n", output.Bold.Render("Dataset"), id)
	fmt.Printf("  %-22s %s\n", output.Bold.Render("Partition column"), "created_at")
	return nil
}

// ── dataset profiling ─────────────────────────────────────────────────────────

var datasetProfilingCmd = &cobra.Command{
	Use:   "profiling <id>",
	Short: "View or configure profiling for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		execution, _ := cmd.Flags().GetString("execution")
		schedule, _ := cmd.Flags().GetString("schedule")
		strategy, _ := cmd.Flags().GetString("strategy")
		samplingRows, _ := cmd.Flags().GetInt("sampling-rows")
		timeWindowDays, _ := cmd.Flags().GetInt("time-window-days")

		changed := enable || disable || execution != "" || schedule != "" || strategy != "" || samplingRows != 0 || timeWindowDays != 0
		if !changed {
			// no flags → same as get
			return runDatasetProfilingGet(args[0])
		}
		output.PrintSuccess(fmt.Sprintf("Profiling config updated for dataset '%s'.", args[0]), GCtx)
		return nil
	},
}

var datasetProfilingGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "View cached profiling data and current settings",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDatasetProfilingGet(args[0])
	},
}

func runDatasetProfilingGet(id string) error {
	fmt.Printf("  Profiling data for %s\n\n", output.Bold.Render(id))
	rows := []map[string]string{
		{"column": "order_id", "type": "bigint", "nulls": "0%", "distinct": "100%", "min": "1", "max": "48231"},
		{"column": "customer_id", "type": "bigint", "nulls": "0.29%", "distinct": "34%", "min": "1001", "max": "9999"},
		{"column": "order_value", "type": "numeric", "nulls": "0%", "distinct": "72%", "min": "4.99", "max": "1249.00"},
		{"column": "created_at", "type": "timestamp", "nulls": "0%", "distinct": "100%", "min": "2025-01-01", "max": "2026-03-04"},
	}
	cols := []string{"column", "type", "nulls", "distinct", "min", "max"}
	output.Render(rows, cols, nil, GCtx)
	return nil
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

// ── dataset diagnostics ───────────────────────────────────────────────────────

var datasetDiagnosticsCmd = &cobra.Command{
	Use:   "diagnostics <id>",
	Short: "View or configure diagnostics warehouse overrides for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schema, _ := cmd.Flags().GetString("schema")
		collectResults, _ := cmd.Flags().GetString("collect-results")
		collectFailedRows, _ := cmd.Flags().GetString("collect-failed-rows")

		changed := schema != "" || collectResults != "" || collectFailedRows != ""
		if !changed {
			return runDatasetDiagnosticsGet(args[0])
		}
		output.PrintSuccess(fmt.Sprintf("Diagnostics config updated for dataset '%s'.", args[0]), GCtx)
		return nil
	},
}

var datasetDiagnosticsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show current diagnostics settings for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDatasetDiagnosticsGet(args[0])
	},
}

func runDatasetDiagnosticsGet(id string) error {
	fmt.Printf("  %-26s %s\n", output.Bold.Render("Dataset"), id)
	fmt.Printf("  %-26s %s\n", output.Bold.Render("Diagnostics override"), output.Dim.Render("(none — inherits datasource config)"))
	return nil
}

// ── dataset permissions ───────────────────────────────────────────────────────

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

var datasetPermAssignCmd = &cobra.Command{
	Use:   "assign <id>",
	Short: "Grant a role to a user or group on a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		user, _ := cmd.Flags().GetString("user")
		group, _ := cmd.Flags().GetString("group")

		if role == "" {
			return output.Errorf(2, "--role is required")
		}
		if user == "" && group == "" {
			return output.Errorf(2, "--user or --group is required")
		}

		principal := user
		if group != "" {
			principal = "group:" + group
		}
		output.PrintSuccess(fmt.Sprintf("Granted role '%s' to '%s' on dataset '%s'.", role, principal, args[0]), GCtx)
		return nil
	},
}

var datasetPermRevokeCmd = &cobra.Command{
	Use:   "revoke <id>",
	Short: "Revoke a role from a user or group on a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		user, _ := cmd.Flags().GetString("user")
		group, _ := cmd.Flags().GetString("group")

		if role == "" {
			return output.Errorf(2, "--role is required")
		}
		if user == "" && group == "" {
			return output.Errorf(2, "--user or --group is required")
		}

		principal := user
		if group != "" {
			principal = "group:" + group
		}
		output.PrintSuccess(fmt.Sprintf("Revoked role '%s' from '%s' on dataset '%s'.", role, principal, args[0]), GCtx)
		return nil
	},
}

func init() {
	datasetListCmd.Flags().String("filter", "", "Filter datasets by query string")
	datasetListCmd.Flags().String("tag", "", "Filter by tag")

	datasetUpdateCmd.Flags().String("owner", "", "Dataset owner email")
	datasetUpdateCmd.Flags().String("tag", "", "Add a tag")
	datasetUpdateCmd.Flags().String("description", "", "Dataset description")

	// time-partition
	datasetTimePartitionCmd.Flags().String("column", "", "Column to use as time partition")
	datasetTimePartitionCmd.AddCommand(datasetTimePartitionGetCmd)

	// profiling
	datasetProfilingCmd.Flags().Bool("enable", false, "Enable profiling")
	datasetProfilingCmd.Flags().Bool("disable", false, "Disable profiling")
	datasetProfilingCmd.Flags().String("execution", "", "Execution mode: manual|scheduled")
	datasetProfilingCmd.Flags().String("schedule", "", "Cron schedule expression")
	datasetProfilingCmd.Flags().String("strategy", "", "Sampling strategy: sampling|time-window")
	datasetProfilingCmd.Flags().Int("sampling-rows", 0, "Number of rows to sample")
	datasetProfilingCmd.Flags().Int("time-window-days", 0, "Number of days in time window")
	datasetProfilingCmd.AddCommand(datasetProfilingGetCmd, datasetProfilingRefreshCmd)

	// diagnostics
	datasetDiagnosticsCmd.Flags().String("schema", "", "Diagnostics schema override")
	datasetDiagnosticsCmd.Flags().String("collect-results", "", "Collect results: true|false")
	datasetDiagnosticsCmd.Flags().String("collect-failed-rows", "", "Collect failed rows: true|false")
	datasetDiagnosticsCmd.AddCommand(datasetDiagnosticsGetCmd)

	// permissions
	datasetPermAssignCmd.Flags().String("role", "", "Role ID (required)")
	datasetPermAssignCmd.Flags().String("user", "", "User email")
	datasetPermAssignCmd.Flags().String("group", "", "Group ID")
	datasetPermRevokeCmd.Flags().String("role", "", "Role ID (required)")
	datasetPermRevokeCmd.Flags().String("user", "", "User email")
	datasetPermRevokeCmd.Flags().String("group", "", "Group ID")
	datasetPermissionsCmd.AddCommand(datasetPermListCmd, datasetPermAssignCmd, datasetPermRevokeCmd)

	datasetCmd.AddCommand(
		datasetListCmd,
		datasetUpdateCmd,
		datasetDeleteCmd,
		datasetTimePartitionCmd,
		datasetProfilingCmd,
		datasetDiagnosticsCmd,
		datasetPermissionsCmd,
	)
	rootCmd.AddCommand(datasetCmd)
}
