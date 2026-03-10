package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

// ── Metric type mapping ───────────────────────────────────────────────────────
// CLI uses kebab-case; API uses camelCase. Map in both directions.

var columnMetricToAPI = map[string]string{
	"missing-pct":    "missingPercentage",
	"duplicate-pct":  "duplicatePercentage",
	"avg":            "average",
	"sum":            "sum",
	"count":          "count",
	"distinct-count": "distinctValuesCount",
	"min":            "minimumValue",
	"max":            "maximumValue",
	"min-length":     "minimumLength",
	"max-length":     "maximumLength",
	"avg-length":     "averageLength",
	"std-dev":        "standardDeviation",
	"variance":       "variance",
	"q1":             "q1",
	"median":         "median",
	"q3":             "q3",
	"freshness":      "columnFreshness",
}

var columnMetricFromAPI = func() map[string]string {
	m := map[string]string{}
	for k, v := range columnMetricToAPI {
		m[v] = k
	}
	return m
}()

var columnMetricNames = func() []string {
	names := make([]string, 0, len(columnMetricToAPI))
	for k := range columnMetricToAPI {
		names = append(names, k)
	}
	return names
}()

func columnMetricAPIValue(cliName string) (string, bool) {
	v, ok := columnMetricToAPI[cliName]
	return v, ok
}

func columnMetricCLIName(apiValue string) string {
	if cli, ok := columnMetricFromAPI[apiValue]; ok {
		return cli
	}
	return apiValue // fall back to raw API value for unknown metrics
}

var monitorTypes = []string{"column", "custom"}

// ── monitor ───────────────────────────────────────────────────────────────────

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Manage ML anomaly detection monitors",
}

// ── monitor list ──────────────────────────────────────────────────────────────

var monitorListCmd = &cobra.Command{
	Use:   "list",
	Short: "List monitors for a dataset",
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetID, _ := cmd.Flags().GetString("dataset")
		monType, _ := cmd.Flags().GetString("type")

		if datasetID == "" {
			return output.Errorf(2, "--dataset is required (no global monitor list endpoint exists in the public API)")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}
		cfg, err := client.GetMetricMonitoring(datasetID)
		if err != nil {
			return err
		}

		rows := []map[string]string{}

		if monType == "" || monType == "column" {
			for _, m := range cfg.ColumnMetricMonitors {
				enabled := "disabled"
				if m.Configuration.IsEnabled {
					enabled = "enabled"
				}
				rows = append(rows, map[string]string{
					"id":      m.CheckID,
					"type":    "column",
					"column":  m.ColumnName,
					"metric":  columnMetricCLIName(m.MetricType),
					"enabled": enabled,
				})
			}
		}

		if monType == "" || monType == "custom" {
			for _, m := range cfg.CustomSqlMetricMonitors {
				enabled := "disabled"
				if m.Configuration.IsEnabled {
					enabled = "enabled"
				}
				rows = append(rows, map[string]string{
					"id":      m.CheckID,
					"type":    "custom",
					"column":  m.ColumnName,
					"metric":  m.MonitorName,
					"enabled": enabled,
				})
			}
		}

		if len(rows) == 0 {
			fmt.Println(output.Dim.Render("  No monitors found."))
			return nil
		}

		cols := []string{"id", "type", "column", "metric", "enabled"}
		output.Render(rows, cols, map[string]bool{"enabled": true}, GCtx)
		return nil
	},
}

// ── monitor config ────────────────────────────────────────────────────────────

var monitorConfigCmd = &cobra.Command{
	Use:   "config <dataset-id>",
	Short: "View or update dataset-level monitor settings",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		schedule, _ := cmd.Flags().GetString("schedule")
		timezone, _ := cmd.Flags().GetString("timezone")

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		if !enable && !disable && schedule == "" {
			cfg, err := client.GetMetricMonitoring(args[0])
			if err != nil {
				return err
			}
			enabled := output.Red.Render("disabled")
			if cfg.Enabled {
				enabled = output.Green.Render("enabled")
			}
			fmt.Printf("  %-24s %s\n", output.Bold.Render("Dataset"), args[0])
			fmt.Printf("  %-24s %s\n", output.Bold.Render("Monitors"), enabled)
			if cfg.ScanSchedule != nil {
				fmt.Printf("  %-24s %s (%s)\n", output.Bold.Render("Schedule"), cfg.ScanSchedule.CronExpression, cfg.ScanSchedule.Timezone)
			}
			if cfg.HistoricalMetricCollectionScanStartDate != "" {
				fmt.Printf("  %-24s %s\n", output.Bold.Render("Historical from"), cfg.HistoricalMetricCollectionScanStartDate)
			}
			total := len(cfg.ColumnMetricMonitors) + len(cfg.CustomSqlMetricMonitors)
			fmt.Printf("  %-24s %d\n", output.Bold.Render("Monitors total"), total)
			return nil
		}

		req := api.UpdateMetricMonitoringRequest{}
		if enable {
			t := true
			req.Enabled = &t
		} else if disable {
			f := false
			req.Enabled = &f
		}
		if schedule != "" {
			tz := timezone
			if tz == "" {
				tz = "UTC"
			}
			req.ScanSchedule = &api.ScanSchedule{CronExpression: schedule, Timezone: tz}
		}

		if _, err := client.UpdateMetricMonitoring(args[0], req); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Monitor config updated for dataset '%s'.", args[0]), GCtx)
		return nil
	},
}

// ── monitor add ───────────────────────────────────────────────────────────────

var monitorAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a monitor to a dataset",
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetID, _ := cmd.Flags().GetString("dataset")
		monType, _ := cmd.Flags().GetString("type")

		if datasetID == "" {
			return output.Errorf(2, "--dataset is required")
		}
		if monType == "" {
			return output.Errorf(2, "--type is required (column|custom)")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		switch monType {
		case "column":
			return runMonitorAddColumn(cmd, client, datasetID)
		case "custom":
			return runMonitorAddCustom(cmd, client, datasetID)
		default:
			return output.Errorf(2, "unsupported type '%s' — use column or custom (dataset and group-by types are not yet available in the public API)", monType)
		}
	},
}

func runMonitorAddColumn(cmd *cobra.Command, client *api.Client, datasetID string) error {
	column, _ := cmd.Flags().GetString("column")
	metric, _ := cmd.Flags().GetString("metric")

	if column == "" {
		return output.Errorf(2, "--column is required for type column")
	}
	if metric == "" {
		return output.Errorf(2, "--metric is required for type column")
	}

	apiMetric, ok := columnMetricAPIValue(metric)
	if !ok {
		return output.Errorf(2, "unknown metric '%s'\n\n  Valid values: %s", metric, columnMetricHelpList())
	}

	req := api.CreateColumnMonitorRequest{
		ColumnName: column,
		ColumnMetricMonitorConfiguration: api.ColumnMetricMonitorCfg{
			MetricType:    apiMetric,
			Configuration: api.ColumnMonitorConfig{IsEnabled: true},
		},
	}
	result, err := client.CreateColumnMonitor(datasetID, req)
	if err != nil {
		return err
	}
	output.PrintSuccess(fmt.Sprintf("Column monitor created (id: %s).", result.CheckID), GCtx)
	return nil
}

func runMonitorAddCustom(cmd *cobra.Command, client *api.Client, datasetID string) error {
	name, _ := cmd.Flags().GetString("name")
	sqlQuery, _ := cmd.Flags().GetString("sql")
	sqlFile, _ := cmd.Flags().GetString("sql-file")
	resultMetric, _ := cmd.Flags().GetString("result-metric")
	column, _ := cmd.Flags().GetString("column")

	if name == "" {
		return output.Errorf(2, "--name is required for type custom")
	}
	if sqlQuery == "" && sqlFile == "" {
		return output.Errorf(2, "--sql or --sql-file is required for type custom")
	}
	if resultMetric == "" {
		return output.Errorf(2, "--result-metric is required for type custom")
	}
	if sqlFile != "" && sqlQuery == "" {
		return output.Errorf(2, "--sql-file is not yet implemented; use --sql with the query inline")
	}

	req := api.CreateCustomSqlMonitorRequest{
		MonitorName: name,
		ColumnName:  column,
		Configuration: api.CustomSqlMonitorConfig{
			SQLQuery:     sqlQuery,
			ResultMetric: resultMetric,
			IsEnabled:    true,
		},
	}
	result, err := client.CreateCustomSqlMonitor(datasetID, req)
	if err != nil {
		return err
	}
	output.PrintSuccess(fmt.Sprintf("Custom SQL monitor '%s' created (id: %s).", result.MonitorName, result.CheckID), GCtx)
	return nil
}

// ── monitor delete ────────────────────────────────────────────────────────────

var monitorDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a monitor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetID, _ := cmd.Flags().GetString("dataset")
		if datasetID == "" {
			return output.Errorf(2, "--dataset is required")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		cfg, err := client.GetMetricMonitoring(datasetID)
		if err != nil {
			return err
		}

		monitorID := args[0]

		for _, m := range cfg.ColumnMetricMonitors {
			if m.CheckID == monitorID {
				result, err := client.DeleteColumnMonitor(datasetID, monitorID)
				if err != nil {
					return err
				}
				output.PrintSuccess(result.Message, GCtx)
				return nil
			}
		}

		for _, m := range cfg.CustomSqlMetricMonitors {
			if m.CheckID == monitorID {
				result, err := client.DeleteCustomSqlMonitor(datasetID, monitorID)
				if err != nil {
					return err
				}
				output.PrintSuccess(result.Message, GCtx)
				return nil
			}
		}

		return output.Errorf(2, "monitor '%s' not found on dataset '%s'", monitorID, datasetID)
	},
}

// ── monitor update (stub) ─────────────────────────────────────────────────────

var monitorUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a monitor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "monitor update is not yet implemented")
	},
}

// ── helpers ───────────────────────────────────────────────────────────────────

func columnMetricHelpList() string {
	ordered := []string{
		"count", "missing-pct", "duplicate-pct", "distinct-count",
		"min", "max", "avg", "sum", "std-dev", "variance",
		"q1", "median", "q3",
		"min-length", "max-length", "avg-length",
		"freshness",
	}
	out := ""
	for i, v := range ordered {
		if i > 0 {
			out += ", "
		}
		out += v
	}
	return out
}

func init() {
	monitorListCmd.Flags().String("dataset", "", "Dataset ID (required — no global monitor list endpoint exists)")
	monitorListCmd.Flags().String("type", "", "Filter by type: column|custom")
	_ = monitorListCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return monitorTypes, cobra.ShellCompDirectiveNoFileComp
	})

	monitorConfigCmd.Flags().Bool("enable", false, "Enable monitoring for this dataset")
	monitorConfigCmd.Flags().Bool("disable", false, "Disable monitoring for this dataset")
	monitorConfigCmd.Flags().String("schedule", "", "Cron schedule expression (e.g. '0 6 * * *')")
	monitorConfigCmd.Flags().String("timezone", "", "Timezone for schedule (default: UTC)")

	monitorAddCmd.Flags().String("dataset", "", "Dataset ID (required)")
	monitorAddCmd.Flags().String("type", "", "Monitor type: column|custom (required)")
	monitorAddCmd.Flags().String("column", "", "Column name (required for type column)")
	monitorAddCmd.Flags().String("metric", "", fmt.Sprintf(
		"Metric type for column monitor (required). Valid values: %s", columnMetricHelpList(),
	))
	monitorAddCmd.Flags().String("name", "", "Monitor name (required for type custom)")
	monitorAddCmd.Flags().String("sql", "", "SQL query (required for type custom, unless --sql-file)")
	monitorAddCmd.Flags().String("sql-file", "", "Path to SQL file (for type custom)")
	monitorAddCmd.Flags().String("result-metric", "", "Result metric column name (required for type custom)")

	_ = monitorAddCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return monitorTypes, cobra.ShellCompDirectiveNoFileComp
	})
	_ = monitorAddCmd.RegisterFlagCompletionFunc("metric", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return columnMetricNames, cobra.ShellCompDirectiveNoFileComp
	})

	monitorDeleteCmd.Flags().String("dataset", "", "Dataset ID (required)")

	monitorCmd.AddCommand(monitorListCmd, monitorConfigCmd, monitorAddCmd, monitorUpdateCmd, monitorDeleteCmd)
	rootCmd.AddCommand(monitorCmd)
}
