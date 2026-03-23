package cmd

import (
	"fmt"
	"strings"
	"time"

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
		status, _ := cmd.Flags().GetString("status")
		idFilter, _ := cmd.Flags().GetString("id")
		limit, _ := cmd.Flags().GetInt("limit")
		fromStr, _ := cmd.Flags().GetString("from")
		untilStr, _ := cmd.Flags().GetString("until")

		// Validate status filter
		if status != "" {
			switch strings.ToLower(status) {
			case "onboarded", "not onboarded", "not-onboarded":
				// ok
			default:
				return output.Errorf(2, "invalid --status value '%s' — use: onboarded, not-onboarded", status)
			}
		}

		// Parse date filters
		var fromTime, untilTime time.Time
		if fromStr != "" {
			t, err := parseDate(fromStr)
			if err != nil {
				return output.Errorf(2, "invalid --from value '%s': use YYYY-MM-DD or ISO8601", fromStr)
			}
			fromTime = t
		}
		if untilStr != "" {
			t, err := parseDate(untilStr)
			if err != nil {
				return output.Errorf(2, "invalid --until value '%s': use YYYY-MM-DD or ISO8601", untilStr)
			}
			untilTime = t.Add(24*time.Hour - time.Second) // inclusive: end of day
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		result, err := client.ListDatasets(api.ListDatasetsParams{
			Search:         filter,
			DatasourceName: datasource,
			Size:           500,
		})
		if err != nil {
			return err
		}

		rows := make([]map[string]string, 0, len(result.Content))
		for _, d := range result.Content {
			rows = append(rows, map[string]string{
				"id":         d.ID,
				"name":       d.Name,
				"datasource": d.Datasource.Name,
				"status":     "onboarded",
				"checks":     fmt.Sprintf("%.0f", d.Checks),
				"monitors":   "-",
				"updated":    d.LastUpdated,
			})
		}

		// Append discovered-not-yet-onboarded datasets.
		dsPage, dsErr := client.ListDatasources(0, 500)
		if dsErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "  %s Could not fetch datasources: %v\n", output.Yellow.Render("⚠"), dsErr)
		} else {
			dsNameByID := map[string]string{}
			var dsIDs []string
			for _, ds := range dsPage.Content {
				dsNameByID[ds.ID] = ds.Name
				if datasource == "" || strings.EqualFold(ds.Name, datasource) {
					dsIDs = append(dsIDs, ds.ID)
				}
			}
			for _, dsID := range dsIDs {
				discPage, discErr := client.ListDiscoveredDatasets(dsID, 0, 500)
				if discErr != nil {
					continue
				}
				for _, d := range discPage.Content {
					if d.Onboarded || isInternalDataset(d.Name, d.QualifiedName) {
						continue
					}
					rows = append(rows, map[string]string{
						"id":         d.ID,
						"name":       d.Name,
						"datasource": dsNameByID[dsID],
						"status":     "not onboarded",
						"checks":     "-",
						"monitors":   "-",
						"updated":    d.CreatedAt,
					})
				}
			}
		}

		// Client-side filters
		if filter != "" {
			needle := strings.ToLower(filter)
			filtered := rows[:0]
			for _, r := range rows {
				if strings.Contains(strings.ToLower(r["name"]), needle) {
					filtered = append(filtered, r)
				}
			}
			rows = filtered
		}

		if status != "" {
			normalized := strings.ToLower(status)
			if normalized == "not-onboarded" {
				normalized = "not onboarded"
			}
			filtered := rows[:0]
			for _, r := range rows {
				if r["status"] == normalized {
					filtered = append(filtered, r)
				}
			}
			rows = filtered
		}

		if idFilter != "" {
			needle := strings.ToLower(idFilter)
			filtered := rows[:0]
			for _, r := range rows {
				if strings.Contains(strings.ToLower(r["id"]), needle) {
					filtered = append(filtered, r)
				}
			}
			rows = filtered
		}

		if !fromTime.IsZero() || !untilTime.IsZero() {
			filtered := rows[:0]
			for _, r := range rows {
				t, err := time.Parse(time.RFC3339, r["updated"])
				if err != nil {
					continue
				}
				if !fromTime.IsZero() && t.Before(fromTime) {
					continue
				}
				if !untilTime.IsZero() && t.After(untilTime) {
					continue
				}
				filtered = append(filtered, r)
			}
			rows = filtered
		}

		if len(rows) == 0 {
			fmt.Println(output.Dim.Render("  No datasets found."))
			return nil
		}

		// Apply limit
		total := len(rows)
		if limit > 0 && len(rows) > limit {
			rows = rows[:limit]
		}

		cols := []string{"id", "name", "datasource", "status", "checks", "monitors", "updated"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)

		if total > len(rows) || !result.Last {
			fmt.Fprintf(cmd.ErrOrStderr(), output.Dim.Render("  Showing %d of %d datasets.\n"), len(rows), total)
		}
		return nil
	},
}

var datasetGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show details for a single dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		ds, err := client.GetDataset(args[0])
		if err != nil {
			return err
		}

		if output.EffectiveFmt(GCtx) == "json" {
			item := map[string]string{
				"id":               ds.ID,
				"name":             ds.Name,
				"qualifiedName":    ds.QualifiedName,
				"datasource":       ds.Datasource.Name,
				"status":           ds.DataQualityStatus,
				"checks":           fmt.Sprintf("%.0f", ds.Checks),
				"incidents":        fmt.Sprintf("%.0f", ds.Incidents),
				"partitionColumn":  ds.PartitionColumn,
				"updated":          ds.LastUpdated,
				"cloudUrl":         ds.CloudURL,
			}
			keys := []string{"id", "name", "qualifiedName", "datasource", "status", "checks", "incidents", "partitionColumn", "updated", "cloudUrl"}
			output.RenderOne(item, keys, GCtx)
			return nil
		}

		fmt.Printf("  %-22s %s\n", output.Bold.Render("Name"), ds.Name)
		fmt.Printf("  %-22s %s\n", output.Bold.Render("ID"), output.FmtID(ds.ID))
		fmt.Printf("  %-22s %s\n", output.Bold.Render("Qualified name"), ds.QualifiedName)
		fmt.Printf("  %-22s %s\n", output.Bold.Render("Datasource"), ds.Datasource.Name)
		if ds.DataQualityStatus != "" {
			fmt.Printf("  %-22s %s\n", output.Bold.Render("DQ status"), output.FmtStatus(ds.DataQualityStatus))
		}
		fmt.Printf("  %-22s %.0f\n", output.Bold.Render("Checks"), ds.Checks)
		fmt.Printf("  %-22s %.0f\n", output.Bold.Render("Incidents"), ds.Incidents)
		if ds.PartitionColumn != "" {
			fmt.Printf("  %-22s %s\n", output.Bold.Render("Partition column"), ds.PartitionColumn)
		}
		if len(ds.Tags) > 0 {
			fmt.Printf("  %-22s %s\n", output.Bold.Render("Tags"), strings.Join(ds.Tags, ", "))
		}
		if ds.LastUpdated != "" {
			fmt.Printf("  %-22s %s\n", output.Bold.Render("Updated"), output.FmtTime(ds.LastUpdated))
		}
		if ds.CloudURL != "" {
			fmt.Printf("  %-22s %s\n", output.Bold.Render("Cloud URL"), ds.CloudURL)
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

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		if column == "" {
			// GET — show current time-partition
			ds, err := client.GetDataset(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("  %-22s %s\n", output.Bold.Render("Dataset"), ds.Name)
			fmt.Printf("  %-22s %s\n", output.Bold.Render("ID"), output.Dim.Render(ds.ID))
			if ds.PartitionColumn != "" {
				fmt.Printf("  %-22s %s\n", output.Bold.Render("Partition column"), ds.PartitionColumn)
			} else {
				fmt.Printf("  %-22s %s\n", output.Bold.Render("Partition column"), output.Dim.Render("not set"))
			}
			return nil
		}

		// SET — update time-partition column
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
				fmt.Printf("  %s\n", output.Dim.Render("  sodacli datasource diagnostics <datasource-id> --enable"))
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
				fmt.Fprintf(cmd.ErrOrStderr(), "  %s\n\n", output.Dim.Render("  sodacli datasource diagnostics <datasource-id> --enable"))
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
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		result, err := client.GetDatasetResponsibilities(args[0])
		if err != nil {
			return err
		}
		rows := make([]map[string]string, 0, len(result.Content))
		for _, r := range result.Content {
			principal := r.UserID
			if r.Type == "userGroup" {
				principal = r.UserGroupID
			}
			rows = append(rows, map[string]string{
				"principal": principal,
				"type":      r.Type,
				"role":      r.Role.Name,
				"role_id":   r.Role.ID,
			})
		}
		cols := []string{"principal", "type", "role", "role_id"}
		output.Render(rows, cols, nil, GCtx)
		return nil
	},
}

var datasetPermAssignCmd = &cobra.Command{
	Use:   "assign <id>",
	Short: "Grant a role to a user or group on a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		roleID, _ := cmd.Flags().GetString("role")
		user, _ := cmd.Flags().GetString("user")
		group, _ := cmd.Flags().GetString("group")

		if roleID == "" {
			return output.Errorf(2, "--role is required")
		}
		if user == "" && group == "" {
			return output.Errorf(2, "--user or --group is required")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		// read-modify-write: fetch current, append new, post back
		current, err := client.GetDatasetResponsibilities(args[0])
		if err != nil {
			return err
		}
		responsibilities := nonManagedResponsibilities(current.Content)
		newEntry := api.ResponsibilityRequest{RoleID: roleID, Type: "user", UserID: user}
		if group != "" {
			newEntry = api.ResponsibilityRequest{RoleID: roleID, Type: "userGroup", UserGroupID: group}
		}
		responsibilities = append(responsibilities, newEntry)

		if _, err := client.UpdateDatasetResponsibilities(args[0], api.UpdateResponsibilitiesRequest{Responsibilities: responsibilities}); err != nil {
			return err
		}
		principal := user
		if group != "" {
			principal = group
		}
		output.PrintSuccess(fmt.Sprintf("Granted role '%s' to '%s' on dataset '%s'.", roleID, principal, args[0]), GCtx)
		return nil
	},
}

var datasetPermRevokeCmd = &cobra.Command{
	Use:   "revoke <id>",
	Short: "Revoke a role from a user or group on a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		roleID, _ := cmd.Flags().GetString("role")
		user, _ := cmd.Flags().GetString("user")
		group, _ := cmd.Flags().GetString("group")

		if roleID == "" {
			return output.Errorf(2, "--role is required")
		}
		if user == "" && group == "" {
			return output.Errorf(2, "--user or --group is required")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		// read-modify-write: fetch current, remove matching entry, post back
		current, err := client.GetDatasetResponsibilities(args[0])
		if err != nil {
			return err
		}
		responsibilities := []api.ResponsibilityRequest{}
		for _, r := range nonManagedResponsibilities(current.Content) {
			if r.RoleID == roleID && ((user != "" && r.UserID == user) || (group != "" && r.UserGroupID == group)) {
				continue // drop this one
			}
			responsibilities = append(responsibilities, r)
		}

		if _, err := client.UpdateDatasetResponsibilities(args[0], api.UpdateResponsibilitiesRequest{Responsibilities: responsibilities}); err != nil {
			return err
		}
		principal := user
		if group != "" {
			principal = group
		}
		output.PrintSuccess(fmt.Sprintf("Revoked role '%s' from '%s' on dataset '%s'.", roleID, principal, args[0]), GCtx)
		return nil
	},
}

// nonManagedResponsibilities converts existing responsibilities to request format,
// excluding system-managed entries that cannot be modified.
func nonManagedResponsibilities(items []api.DatasetResponsibility) []api.ResponsibilityRequest {
	out := []api.ResponsibilityRequest{}
	for _, r := range items {
		if r.Managed {
			continue
		}
		out = append(out, api.ResponsibilityRequest{
			RoleID:      r.Role.ID,
			Type:        r.Type,
			UserID:      r.UserID,
			UserGroupID: r.UserGroupID,
		})
	}
	return out
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
	datasetListCmd.Flags().String("id", "", "Filter by dataset ID (substring match)")
	datasetListCmd.Flags().String("status", "", "Filter by status: onboarded|not-onboarded")
	datasetListCmd.Flags().Int("limit", 10, "Maximum number of datasets to show")
	datasetListCmd.Flags().String("from", "", "Show datasets updated on or after this date (YYYY-MM-DD or ISO8601)")
	datasetListCmd.Flags().String("until", "", "Show datasets updated on or before this date (YYYY-MM-DD or ISO8601)")
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
		datasetGetCmd,
		datasetUpdateCmd,
		datasetDeleteCmd,
		datasetTimePartitionCmd,
		datasetProfilingCmd,
		datasetDiagnosticsCmd,
		datasetPermissionsCmd,
	)
}
