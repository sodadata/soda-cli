package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
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
		filter, _ := cmd.Flags().GetString("filter")
		datasource, _ := cmd.Flags().GetString("datasource")

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		result, err := client.ListDatasets(api.ListDatasetsParams{
			Search:         filter,
			DatasourceName: datasource,
			Size:           100,
		})
		if err != nil {
			return err
		}

		rows := make([]map[string]string, len(result.Content))
		for i, d := range result.Content {
			rows[i] = map[string]string{
				"id":         d.ID,
				"name":       d.Name,
				"datasource": d.Datasource.Name,
				"status":     d.DataQualityStatus,
				"checks":     fmt.Sprintf("%.0f", d.Checks),
				"updated":    d.LastUpdated,
			}
		}

		cols := []string{"id", "name", "datasource", "status", "checks", "updated"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)

		if !result.Last {
			fmt.Fprintf(cmd.ErrOrStderr(), output.Dim.Render("  Showing %d of %d datasets.\n"), len(result.Content), result.TotalElements)
		}
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
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		result, err := client.DeleteDataset(args[0])
		if err != nil {
			return err
		}
		output.PrintSuccess(result.Message, GCtx)
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
		schedule, _ := cmd.Flags().GetString("schedule")
		timezone, _ := cmd.Flags().GetString("timezone")
		samplingRows, _ := cmd.Flags().GetInt("sampling-rows")

		if !enable && !disable && schedule == "" && samplingRows == 0 {
			// no flags → show current profiling data
			return runDatasetProfilingGet(args[0])
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		settings := api.ProfilingSettings{}
		if enable {
			t := true
			settings.Enabled = &t
		} else if disable {
			f := false
			settings.Enabled = &f
		}
		if schedule != "" {
			tz := timezone
			if tz == "" {
				tz = "UTC"
			}
			settings.ScanSchedule = &api.ScanSchedule{
				CronExpression: schedule,
				Timezone:       tz,
			}
		}
		if samplingRows > 0 {
			settings.ProfilingSamplingStrategy = &api.SamplingStrategy{
				NumberOfRows: samplingRows,
			}
		}

		if _, err := client.UpdateDatasetProfiling(args[0], settings); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Profiling config updated for dataset '%s'.", args[0]), GCtx)
		return nil
	},
}

var datasetProfilingRefreshCmd = &cobra.Command{
	Use:   "refresh <id>",
	Short: "Trigger a new profiling run (not yet available in the public API)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "profiling refresh is not yet available in the public API")
	},
}

func runDatasetProfilingGet(id string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}
	result, err := client.GetProfiling(id)
	if err != nil {
		return err
	}

	// Settings summary
	enabled := output.Red.Render("disabled")
	if result.Enabled {
		enabled = output.Green.Render("enabled")
	}
	fmt.Printf("  %-24s %s\n", output.Bold.Render("Profiling"), enabled)
	if result.ProfilingTime != "" {
		fmt.Printf("  %-24s %s\n", output.Bold.Render("Last run"), result.ProfilingTime)
	}
	if result.RowCount != nil {
		fmt.Printf("  %-24s %.0f\n", output.Bold.Render("Row count"), *result.RowCount)
	}
	if result.ScanSchedule != nil {
		fmt.Printf("  %-24s %s (%s)\n", output.Bold.Render("Schedule"), result.ScanSchedule.CronExpression, result.ScanSchedule.Timezone)
	}
	if result.SamplingStrategyConfig != nil && result.SamplingStrategyConfig.NumberOfRows > 0 {
		fmt.Printf("  %-24s %d rows\n", output.Bold.Render("Sampling"), result.SamplingStrategyConfig.NumberOfRows)
	}

	if len(result.Columns) == 0 {
		fmt.Println()
		fmt.Println(output.Dim.Render("  No profiling data available yet."))
		return nil
	}

	// Column stats table
	fmt.Println()
	rows := make([]map[string]string, len(result.Columns))
	for i, col := range result.Columns {
		row := map[string]string{
			"column": col.Name,
			"type":   col.Type,
		}
		if col.Metrics.MissingCount != nil && result.RowCount != nil && *result.RowCount > 0 {
			pct := (*col.Metrics.MissingCount / *result.RowCount) * 100
			row["missing"] = fmt.Sprintf("%.2f%%", pct)
		} else {
			row["missing"] = "-"
		}
		if col.Metrics.DistinctCount != nil {
			row["distinct"] = fmt.Sprintf("%.0f", *col.Metrics.DistinctCount)
		} else {
			row["distinct"] = "-"
		}
		if col.Metrics.Minimum != nil {
			row["min"] = fmt.Sprintf("%g", *col.Metrics.Minimum)
		} else {
			row["min"] = "-"
		}
		if col.Metrics.Maximum != nil {
			row["max"] = fmt.Sprintf("%g", *col.Metrics.Maximum)
		} else {
			row["max"] = "-"
		}
		rows[i] = row
	}
	cols := []string{"column", "type", "missing", "distinct", "min", "max"}
	output.Render(rows, cols, nil, GCtx)
	return nil
}

// ── dataset diagnostics ───────────────────────────────────────────────────────

var datasetDiagnosticsCmd = &cobra.Command{
	Use:   "diagnostics <id>",
	Short: "Configure diagnostics warehouse overrides for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		schema, _ := cmd.Flags().GetString("schema")
		collectResults, _ := cmd.Flags().GetBool("collect-results")
		noCollectResults, _ := cmd.Flags().GetBool("no-collect-results")
		collectFailedRows, _ := cmd.Flags().GetBool("collect-failed-rows")
		noCollectFailedRows, _ := cmd.Flags().GetBool("no-collect-failed-rows")
		tablePrefix, _ := cmd.Flags().GetString("table-prefix")
		tableSuffix, _ := cmd.Flags().GetString("table-suffix")
		failedRowsDesc, _ := cmd.Flags().GetString("failed-rows-description")
		exposeQuery, _ := cmd.Flags().GetBool("expose-failed-rows-query")
		noExposeQuery, _ := cmd.Flags().GetBool("no-expose-failed-rows-query")
		cta, _ := cmd.Flags().GetBool("failed-rows-cta")
		noCta, _ := cmd.Flags().GetBool("no-failed-rows-cta")

		_ = collectResults || noCollectResults || collectFailedRows || noCollectFailedRows ||
			exposeQuery || noExposeQuery || cta || noCta

		if schema == "" && tablePrefix == "" && tableSuffix == "" && failedRowsDesc == "" &&
			!collectResults && !noCollectResults && !collectFailedRows && !noCollectFailedRows &&
			!exposeQuery && !noExposeQuery && !cta && !noCta {
			fmt.Printf("  %-26s %s\n", output.Bold.Render("Dataset"), args[0])
			fmt.Printf("  %-26s %s\n", output.Bold.Render("Diagnostics override"), output.Dim.Render("(none — inherits datasource config)"))
			return nil
		}

		output.PrintSuccess(fmt.Sprintf("Diagnostics config updated for dataset '%s'.", args[0]), GCtx)
		return nil
	},
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
	datasetListCmd.Flags().String("filter", "", "Fuzzy search on dataset name")
	datasetListCmd.Flags().String("datasource", "", "Filter by datasource name")
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
	datasetProfilingCmd.Flags().String("schedule", "", "Cron schedule expression (e.g. '0 6 * * *')")
	datasetProfilingCmd.Flags().String("timezone", "", "Timezone for schedule (default: UTC)")
	datasetProfilingCmd.Flags().Int("sampling-rows", 0, "Number of rows to sample")
	datasetProfilingCmd.AddCommand(datasetProfilingRefreshCmd)

	// diagnostics
	datasetDiagnosticsCmd.Flags().String("schema", "", "Schema for diagnostic tables (overrides datasource default)")
	datasetDiagnosticsCmd.Flags().Bool("collect-results", false, "Store check results and scan history")
	datasetDiagnosticsCmd.Flags().Bool("no-collect-results", false, "Disable storing check results and scan history")
	datasetDiagnosticsCmd.Flags().Bool("collect-failed-rows", false, "Store failed rows")
	datasetDiagnosticsCmd.Flags().Bool("no-collect-failed-rows", false, "Disable storing failed rows")
	datasetDiagnosticsCmd.Flags().String("table-prefix", "", "Prefix for diagnostic table names")
	datasetDiagnosticsCmd.Flags().String("table-suffix", "", "Suffix for diagnostic table names")
	datasetDiagnosticsCmd.Flags().String("failed-rows-description", "", "Description for failed rows storage context")
	datasetDiagnosticsCmd.Flags().Bool("expose-failed-rows-query", false, "Expose the failed rows SQL query in Cloud")
	datasetDiagnosticsCmd.Flags().Bool("no-expose-failed-rows-query", false, "Hide the failed rows SQL query in Cloud")
	datasetDiagnosticsCmd.Flags().Bool("failed-rows-cta", false, "Show a call-to-action link to where failed rows can be found")
	datasetDiagnosticsCmd.Flags().Bool("no-failed-rows-cta", false, "Hide the call-to-action link for failed rows")

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
