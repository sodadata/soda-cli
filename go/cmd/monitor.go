package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

// parseHistoryStartDate accepts:
//   - a relative form like "120d", "6m", "1y" (N days/months/years before now, UTC)
//   - a YYYY-MM-DD calendar date (interpreted as midnight UTC)
//   - a full RFC3339 timestamp (passed through, normalized to UTC)
//
// Returns an RFC3339 string suitable for sending to the API.
var relativeRE = regexp.MustCompile(`^\d+[dmy]$`)

func parseHistoryStartDate(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", fmt.Errorf("empty value")
	}
	if relativeRE.MatchString(s) {
		n, _ := strconv.Atoi(s[:len(s)-1])
		now := time.Now().UTC()
		var t time.Time
		switch s[len(s)-1] {
		case 'd':
			t = now.AddDate(0, 0, -n)
		case 'm':
			t = now.AddDate(0, -n, 0)
		case 'y':
			t = now.AddDate(-n, 0, 0)
		}
		return t.Format(time.RFC3339), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC().Format(time.RFC3339), nil
	}
	return "", fmt.Errorf("unrecognized date format %q (use YYYY-MM-DD, RFC3339, or relative like 30d / 6m / 1y)", input)
}

// parseExcludeValues parses repeated --exclude-values "column=v1,v2" entries
// into a per-column map. Each key must appear in groupByCols; empty values
// are dropped. Whitespace around values is trimmed.
func parseExcludeValues(entries, groupByCols []string) (map[string][]string, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	groupBySet := make(map[string]struct{}, len(groupByCols))
	for _, c := range groupByCols {
		groupBySet[c] = struct{}{}
	}
	result := make(map[string][]string)
	for _, e := range entries {
		key, rawVals, ok := strings.Cut(e, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return nil, output.Errorf(2, "--exclude-values must be in the form 'column=val1,val2' (got %q)", e)
		}
		if _, in := groupBySet[key]; !in {
			return nil, output.Errorf(2, "--exclude-values references column %q which is not in --group-by", key)
		}
		var vals []string
		for _, v := range strings.Split(rawVals, ",") {
			v = strings.TrimSpace(v)
			if v != "" {
				vals = append(vals, v)
			}
		}
		if len(vals) == 0 {
			return nil, output.Errorf(2, "--exclude-values for column %q has no values", key)
		}
		result[key] = append(result[key], vals...)
	}
	return result, nil
}

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

var datasetMetricToAPI = map[string]string{
	"row-count":        "rowCount",
	"freshness":        "freshness",
	"schema":           "schema",
	"rows-inserted":    "rowsInserted",
	"row-count-change": "totalRowCountChange",
	"timeliness":       "timeliness",
}

var datasetMetricFromAPI = func() map[string]string {
	m := map[string]string{}
	for k, v := range datasetMetricToAPI {
		m[v] = k
	}
	return m
}()

var datasetMetricNames = func() []string {
	names := make([]string, 0, len(datasetMetricToAPI))
	for k := range datasetMetricToAPI {
		names = append(names, k)
	}
	return names
}()

func datasetMetricCLIName(apiValue string) string {
	if cli, ok := datasetMetricFromAPI[apiValue]; ok {
		return cli
	}
	return apiValue
}

var monitorTypes = []string{"column", "custom", "dataset"}

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

		if monType == "" || monType == "dataset" {
			for _, m := range cfg.DatasetMetricMonitorsConfiguration {
				enabled := "disabled"
				if m.Configuration.IsEnabled {
					enabled = "enabled"
				}
				rows = append(rows, map[string]string{
					"id":      "-",
					"type":    "dataset",
					"column":  "-",
					"metric":  datasetMetricCLIName(m.MetricType),
					"enabled": enabled,
				})
			}
		}

		if monType == "" || monType == "column" {
			for _, m := range cfg.ColumnMetricMonitors {
				enabled := "disabled"
				if m.Configuration.IsEnabled {
					enabled = "enabled"
				}
				groupBy := ""
				for i, g := range m.Configuration.GroupByColumns {
					if i > 0 {
						groupBy += ","
					}
					groupBy += g.ColumnName
					if len(g.ExcludedValues) > 0 {
						groupBy += "[excl: " + strings.Join(g.ExcludedValues, ",") + "]"
					}
				}
				col := m.ColumnName
				if groupBy != "" {
					col = m.ColumnName + " (group-by: " + groupBy + ")"
				}
				rows = append(rows, map[string]string{
					"id":      m.CheckID,
					"type":    "column",
					"column":  col,
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
	Long: `View or update dataset-level metric monitoring settings.

With no flags, prints the current configuration (enabled state, schedule,
historical backfill date, and total monitor count).

Update flags:
  --enable / --disable          Toggle monitoring on/off for the dataset
  --schedule <cron>             Cron expression (e.g. "0 6 * * *")
  --timezone <tz>               Timezone for the schedule (default: UTC)
  --history-start-date <date>   Historical backfill start (see formats below)

History start date formats:
  120d         Relative: 120 days ago (also Nm = months, Ny = years)
  2026-01-03   Calendar date (interpreted as midnight UTC)
  2026-01-03T00:00:00Z   RFC3339 timestamp

Updates fetch the current config first and merge — flags you don't pass
are preserved. The schedule and per-monitor configuration are always sent
back so the API doesn't reject the request.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		schedule, _ := cmd.Flags().GetString("schedule")
		timezone, _ := cmd.Flags().GetString("timezone")
		historyStartDate, _ := cmd.Flags().GetString("history-start-date")

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		anyUpdate := enable || disable || schedule != "" || timezone != "" || historyStartDate != ""

		if !anyUpdate {
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

		if enable && disable {
			return output.Errorf(2, "--enable and --disable are mutually exclusive")
		}

		// Fetch current state to merge — POST /api/v1/datasets/{id} with
		// metricMonitoring.enabled=true requires scanSchedule and
		// datasetMetricMonitorsConfiguration in the same payload.
		current, err := client.GetMetricMonitoring(args[0])
		if err != nil {
			return err
		}

		req := api.MetricMonitoringSettings{
			DatasetMetricMonitorsConfiguration:      current.DatasetMetricMonitorsConfiguration,
			HistoricalMetricCollectionScanStartDate: current.HistoricalMetricCollectionScanStartDate,
			ScanSchedule:                            current.ScanSchedule,
		}

		// enabled: user override wins; otherwise preserve current
		switch {
		case enable:
			t := true
			req.Enabled = &t
		case disable:
			f := false
			req.Enabled = &f
		default:
			cur := current.Enabled
			req.Enabled = &cur
		}

		// Merge schedule + timezone
		if schedule != "" || timezone != "" {
			cron := schedule
			tz := timezone
			if req.ScanSchedule != nil {
				if cron == "" {
					cron = req.ScanSchedule.CronExpression
				}
				if tz == "" {
					tz = req.ScanSchedule.Timezone
				}
			}
			if tz == "" {
				tz = "UTC"
			}
			if cron == "" {
				return output.Errorf(2, "--timezone requires --schedule or an existing schedule on the dataset")
			}
			req.ScanSchedule = &api.ScanSchedule{CronExpression: cron, Timezone: tz}
		}

		// History start date
		if historyStartDate != "" {
			parsed, err := parseHistoryStartDate(historyStartDate)
			if err != nil {
				return output.Errorf(2, "%v", err)
			}
			pt, _ := time.Parse(time.RFC3339, parsed)
			now := time.Now().UTC()
			if !pt.Before(now) {
				return output.Errorf(2, "--history-start-date must be in the past (got %s)", parsed)
			}
			if pt.Before(now.AddDate(-2, 0, 0)) {
				fmt.Fprintf(os.Stderr, "  %s --history-start-date %s is more than 2 years ago — Soda may not have historical data that far back.\n", output.Yellow.Render("⚠"), parsed)
			}
			req.HistoricalMetricCollectionScanStartDate = parsed
		}

		// Guard: if we're sending enabled=true, both scanSchedule and monitors
		// configuration must be non-empty (API constraint).
		if req.Enabled != nil && *req.Enabled {
			if req.ScanSchedule == nil || req.ScanSchedule.CronExpression == "" {
				return output.Errorf(2, "cannot enable monitoring without a schedule — pass --schedule, or run `sodacli dataset onboard %s --monitoring ...` to set defaults", args[0])
			}
			if len(req.DatasetMetricMonitorsConfiguration) == 0 {
				return output.Errorf(2, "cannot enable monitoring without per-monitor configuration — run `sodacli dataset onboard %s --monitoring ...` to set defaults first", args[0])
			}
		}

		updated, err := client.UpdateMetricMonitoring(args[0], req)
		if err != nil {
			return err
		}

		status := output.Red.Render("disabled")
		if updated.Enabled {
			status = output.Green.Render("enabled")
		}
		output.PrintSuccess(fmt.Sprintf("Monitor config updated for dataset '%s'. Monitors: %s", args[0], status), GCtx)
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
		case "dataset":
			return runMonitorAddDataset(cmd, client, datasetID)
		default:
			return output.Errorf(2, "unsupported type '%s' — use column, custom, or dataset", monType)
		}
	},
}

func runMonitorAddColumn(cmd *cobra.Command, client *api.Client, datasetID string) error {
	column, _ := cmd.Flags().GetString("column")
	metric, _ := cmd.Flags().GetString("metric")
	groupByCols, _ := cmd.Flags().GetStringArray("group-by")
	excludeValues, _ := cmd.Flags().GetStringArray("exclude-values")

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

	excludedByCol, err := parseExcludeValues(excludeValues, groupByCols)
	if err != nil {
		return err
	}

	cfg := api.ColumnMonitorConfig{IsEnabled: true}
	for _, col := range groupByCols {
		cfg.GroupByColumns = append(cfg.GroupByColumns, api.GroupByColumn{
			ColumnName:     col,
			ExcludedValues: excludedByCol[col],
		})
	}

	req := api.CreateColumnMonitorRequest{
		ColumnName: column,
		ColumnMetricMonitorConfiguration: api.ColumnMetricMonitorCfg{
			MetricType:    apiMetric,
			Configuration: cfg,
		},
	}
	result, err := client.CreateColumnMonitor(datasetID, req)
	if err != nil {
		return err
	}
	output.PrintSuccess(fmt.Sprintf("Column monitor created (id: %s).", result.CheckID), GCtx)
	return nil
}

func runMonitorAddDataset(_ *cobra.Command, _ *api.Client, datasetID string) error {
	// The public API only supports GET for dataset-level monitors — no write endpoint exists.
	// Dataset monitors (row count, freshness, schema, etc.) exist by default but must be
	// enabled/disabled from the Soda Cloud UI.
	return output.Errorf(2, "adding/enabling dataset-level monitors is not yet available in the public API.\n\n  Dataset monitors exist by default — enable them from the Soda Cloud UI.\n  To add column monitors:  sodacli monitor add --dataset %s --type column --column <col> --metric <metric>", datasetID)
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

// ── monitor update ────────────────────────────────────────────────────────────

var monitorUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a monitor (enable/disable, change SQL, etc.)",
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

		// Look up the monitor type by scanning the metric monitoring config
		cfg, err := client.GetMetricMonitoring(datasetID)
		if err != nil {
			return err
		}

		monitorID := args[0]

		// Check column monitors
		for _, m := range cfg.ColumnMetricMonitors {
			if m.CheckID == monitorID {
				return runMonitorUpdateColumn(cmd, client, datasetID, monitorID, m)
			}
		}

		// Check custom SQL monitors
		for _, m := range cfg.CustomSqlMetricMonitors {
			if m.CheckID == monitorID {
				return runMonitorUpdateCustom(cmd, client, datasetID, monitorID, m)
			}
		}

		return output.Errorf(2, "monitor '%s' not found on dataset '%s'.\n\n  Note: dataset-level monitors cannot be updated individually — use `sodacli monitor config %s --enable/--disable`", monitorID, datasetID, datasetID)
	},
}

func runMonitorUpdateColumn(cmd *cobra.Command, client *api.Client, datasetID, monitorID string, existing api.ColumnMonitor) error {
	enable, _ := cmd.Flags().GetBool("enable")
	disable, _ := cmd.Flags().GetBool("disable")

	if enable && disable {
		return output.Errorf(2, "--enable and --disable are mutually exclusive")
	}

	cfg := existing.Configuration
	if enable {
		cfg.IsEnabled = true
	} else if disable {
		cfg.IsEnabled = false
	}

	result, err := client.UpdateColumnMonitor(datasetID, monitorID, api.UpdateColumnMonitorRequest{
		Configuration: cfg,
	})
	if err != nil {
		return err
	}
	status := "disabled"
	if result.Configuration.IsEnabled {
		status = "enabled"
	}
	output.PrintSuccess(fmt.Sprintf("Column monitor '%s' updated (%s).", monitorID, status), GCtx)
	return nil
}

func runMonitorUpdateCustom(cmd *cobra.Command, client *api.Client, datasetID, monitorID string, existing api.CustomSqlMonitor) error {
	enable, _ := cmd.Flags().GetBool("enable")
	disable, _ := cmd.Flags().GetBool("disable")
	sqlQuery, _ := cmd.Flags().GetString("sql")
	name, _ := cmd.Flags().GetString("name")
	resultMetric, _ := cmd.Flags().GetString("result-metric")

	if enable && disable {
		return output.Errorf(2, "--enable and --disable are mutually exclusive")
	}

	cfg := existing.Configuration
	if enable {
		cfg.IsEnabled = true
	} else if disable {
		cfg.IsEnabled = false
	}
	if sqlQuery != "" {
		cfg.SQLQuery = sqlQuery
	}
	if resultMetric != "" {
		cfg.ResultMetric = resultMetric
	}

	reqName := existing.MonitorName
	if name != "" {
		reqName = name
	}

	result, err := client.UpdateCustomSqlMonitor(datasetID, monitorID, api.UpdateCustomSqlMonitorRequest{
		MonitorName:   reqName,
		Configuration: cfg,
		ColumnName:    existing.ColumnName,
	})
	if err != nil {
		return err
	}
	status := "disabled"
	if result.Configuration.IsEnabled {
		status = "enabled"
	}
	output.PrintSuccess(fmt.Sprintf("Custom SQL monitor '%s' updated (%s).", result.MonitorName, status), GCtx)
	return nil
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

func datasetMetricHelpList() string {
	ordered := []string{"row-count", "freshness", "schema", "rows-inserted", "row-count-change", "timeliness"}
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
	monitorConfigCmd.Flags().String("history-start-date", "", "Historical backfill start (YYYY-MM-DD, RFC3339, or relative like 30d / 6m / 1y)")

	monitorAddCmd.Flags().String("dataset", "", "Dataset ID (required)")
	monitorAddCmd.Flags().String("type", "", "Monitor type: column|custom|dataset (required)")
	monitorAddCmd.Flags().String("column", "", "Column name (required for type column)")
	monitorAddCmd.Flags().String("metric", "", fmt.Sprintf(
		"Metric type (required). Column: %s — Dataset: %s",
		columnMetricHelpList(), datasetMetricHelpList(),
	))
	monitorAddCmd.Flags().StringArray("group-by", nil, "Group-by column (repeatable, for type column)")
	monitorAddCmd.Flags().StringArray("exclude-values", nil, "Exclude values from a group-by column, format 'column=val1,val2' (repeatable)")
	monitorAddCmd.Flags().String("name", "", "Monitor name (required for type custom)")
	monitorAddCmd.Flags().String("sql", "", "SQL query (required for type custom, unless --sql-file)")
	monitorAddCmd.Flags().String("sql-file", "", "Path to SQL file (for type custom)")
	monitorAddCmd.Flags().String("result-metric", "", "Result metric column name (required for type custom)")

	_ = monitorAddCmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return monitorTypes, cobra.ShellCompDirectiveNoFileComp
	})
	_ = monitorAddCmd.RegisterFlagCompletionFunc("metric", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// suggest column or dataset metrics depending on --type flag value
		if t, _ := cmd.Flags().GetString("type"); t == "dataset" {
			return datasetMetricNames, cobra.ShellCompDirectiveNoFileComp
		}
		return columnMetricNames, cobra.ShellCompDirectiveNoFileComp
	})

	monitorUpdateCmd.Flags().String("dataset", "", "Dataset ID (required)")
	monitorUpdateCmd.Flags().Bool("enable", false, "Enable the monitor")
	monitorUpdateCmd.Flags().Bool("disable", false, "Disable the monitor")
	monitorUpdateCmd.Flags().String("sql", "", "Update SQL query (custom monitors only)")
	monitorUpdateCmd.Flags().String("name", "", "Update monitor name (custom monitors only)")
	monitorUpdateCmd.Flags().String("result-metric", "", "Update result metric (custom monitors only)")

	monitorDeleteCmd.Flags().String("dataset", "", "Dataset ID (required)")

	monitorCmd.AddCommand(monitorListCmd, monitorConfigCmd, monitorAddCmd, monitorUpdateCmd, monitorDeleteCmd)
}
