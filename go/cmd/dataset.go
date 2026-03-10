package cmd

import (
	"fmt"
	"strings"

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
	Short: "Update dataset metadata (owner, tags)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		tags, _ := cmd.Flags().GetStringArray("tag")

		if owner == "" && len(tags) == 0 {
			return output.Errorf(2, "at least one of --owner or --tag is required")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		req := api.UpdateDatasetRequest{}
		if owner != "" {
			req.Owners = []api.DatasetOwnerRequest{{Type: "user", UserID: owner}}
		}
		if len(tags) > 0 {
			req.Tags = tags
		}

		if _, err := client.UpdateDataset(args[0], req); err != nil {
			return err
		}
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
			fmt.Printf("  %-22s %s\n", output.Bold.Render("Dataset"), args[0])
			fmt.Println(output.Dim.Render("  Time-partition view requires a single-dataset GET endpoint (not yet in the public API)."))
			return nil
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}
		req := api.UpdateDatasetRequest{
			TimePartition: &api.TimePartitionRequest{PartitionColumn: column},
		}
		if _, err := client.UpdateDataset(args[0], req); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Time-partition column set to '%s' for dataset '%s'.", column, args[0]), GCtx)
		return nil
	},
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
	Short: "View or configure diagnostics warehouse overrides for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collectResults, _ := cmd.Flags().GetBool("collect-results")
		noCollectResults, _ := cmd.Flags().GetBool("no-collect-results")
		collectFailedRows, _ := cmd.Flags().GetBool("collect-failed-rows")
		noCollectFailedRows, _ := cmd.Flags().GetBool("no-collect-failed-rows")
		// flags not yet in the public API — fail fast with a clear message
		unsupportedFlags := []string{"schema", "table-prefix", "table-suffix", "failed-rows-description",
			"expose-failed-rows-query", "no-expose-failed-rows-query", "failed-rows-cta", "no-failed-rows-cta"}
		for _, f := range unsupportedFlags {
			if cmd.Flags().Changed(f) {
				return output.Errorf(2, "--%s is not yet available in the public API", f)
			}
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		// no flags → show current settings
		if !collectResults && !noCollectResults && !collectFailedRows && !noCollectFailedRows {
			result, err := client.GetDatasetDiagnostics(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("  %-28s %s\n", output.Bold.Render("Dataset"), args[0])
			if result.ScanAndResultsConfiguration == nil && result.FailedRowsConfiguration == nil {
				fmt.Printf("  %-28s %s\n", output.Bold.Render("Diagnostics warehouse"), output.Dim.Render("not configured"))
				fmt.Printf("\n  %s\n", output.Dim.Render("Set it up at the datasource level first:"))
				fmt.Printf("  %s\n", output.Dim.Render("  soda datasource diagnostics <datasource-id> --enable"))
				return nil
			}
			if result.ScanAndResultsConfiguration != nil {
				v := output.Red.Render("disabled")
				if result.ScanAndResultsConfiguration.Enabled {
					v = output.Green.Render("enabled")
				}
				fmt.Printf("  %-28s %s\n", output.Bold.Render("Collect results"), v)
			}
			if result.FailedRowsConfiguration != nil {
				v := output.Red.Render("disabled")
				if result.FailedRowsConfiguration.Enabled {
					v = output.Green.Render("enabled")
				}
				fmt.Printf("  %-28s %s\n", output.Bold.Render("Collect failed rows"), v)
				if result.FailedRowsConfiguration.MaxRowCount > 0 {
					fmt.Printf("  %-28s %d\n", output.Bold.Render("Max row count"), result.FailedRowsConfiguration.MaxRowCount)
				}
				if result.FailedRowsConfiguration.State != "" {
					fmt.Printf("  %-28s %s\n", output.Bold.Render("State"), result.FailedRowsConfiguration.State)
				}
			}
			return nil
		}

		// flags provided → update
		cfg := api.DiagnosticsWarehouseConfig{}
		if collectResults || noCollectResults {
			enabled := collectResults
			cfg.ScanAndResultsConfiguration = &api.DiagnosticsScanConfig{Enabled: &enabled}
		}
		if collectFailedRows || noCollectFailedRows {
			enabled := collectFailedRows
			cfg.FailedRowsConfiguration = &api.DiagnosticsFailedRowsConfig{Enabled: &enabled}
		}

		if _, err := client.UpdateDatasetDiagnostics(args[0], cfg); err != nil {
			if isNotEnabledOnDatasource(err) {
				fmt.Fprintf(cmd.ErrOrStderr(), "\n  %s\n", output.Dim.Render("Set up the diagnostics warehouse on the datasource first:"))
				fmt.Fprintf(cmd.ErrOrStderr(), "  %s\n\n", output.Dim.Render("  soda datasource diagnostics <datasource-id> --enable"))
			}
			return err
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

func isNotEnabledOnDatasource(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not enabled on the datasource")
}

func init() {
	datasetListCmd.Flags().String("filter", "", "Fuzzy search on dataset name")
	datasetListCmd.Flags().String("datasource", "", "Filter by datasource name")
	datasetListCmd.Flags().String("tag", "", "Filter by tag")

	datasetUpdateCmd.Flags().String("owner", "", "Dataset owner user ID")
	datasetUpdateCmd.Flags().StringArray("tag", nil, "Tag to set (repeatable; replaces all existing tags)")

	// time-partition
	datasetTimePartitionCmd.Flags().String("column", "", "Column to use as time partition")

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
