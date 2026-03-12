package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var resultsCmd = &cobra.Command{
	Use:   "results",
	Short: "View data quality signals: contract checks and monitor alerts",
}

var resultsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List check results across datasets",
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetID, _  := cmd.Flags().GetString("dataset")
		datasetName, _ := cmd.Flags().GetString("dataset-name")
		status, _     := cmd.Flags().GetString("status")
		resType, _    := cmd.Flags().GetString("type")
		limit, _      := cmd.Flags().GetInt("limit")
		sortCol, _    := cmd.Flags().GetString("sort")
		order, _      := cmd.Flags().GetString("order")
		fromStr, _    := cmd.Flags().GetString("from")
		untilStr, _   := cmd.Flags().GetString("until")

		// Monitor results not yet available
		if resType == "monitor" {
			fmt.Println(output.Dim.Render("  Monitor results are not yet available in the public API."))
			return nil
		}

		// Validate sort column
		validSortCols := map[string]bool{
			"dataset": true, "name": true, "column": true, "status": true, "date": true,
		}
		if sortCol != "" && !validSortCols[sortCol] {
			return output.Errorf(2, "invalid --sort value '%s' — use: dataset, name, column, status, date", sortCol)
		}
		if sortCol == "" {
			sortCol = "date"
		}

		// Validate order
		order = strings.ToLower(order)
		if order != "asc" && order != "desc" {
			return output.Errorf(2, "invalid --order value '%s' — use: asc, desc", order)
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

		// Normalize status filter
		var statusFilter string
		switch strings.ToLower(status) {
		case "failing", "fail":
			statusFilter = "fail"
		case "passing", "pass":
			statusFilter = "pass"
		case "error":
			statusFilter = "error"
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		// Fetch more when client-side filters are active
		fetchSize := limit
		if statusFilter != "" || datasetName != "" || fromStr != "" || untilStr != "" {
			fetchSize = 500
		}
		result, err := client.ListChecks(api.ListChecksParams{Size: fetchSize, DatasetID: datasetID})
		if err != nil {
			return err
		}

		checks := result.Content

		// Client-side filters
		if statusFilter != "" {
			filtered := checks[:0]
			for _, c := range checks {
				if c.EvaluationStatus == statusFilter {
					filtered = append(filtered, c)
				}
			}
			checks = filtered
		}

		if datasetName != "" {
			filtered := checks[:0]
			needle := strings.ToLower(datasetName)
			for _, c := range checks {
				qn := ""
				if len(c.Datasets) > 0 {
					qn = strings.ToLower(c.Datasets[0].QualifiedName)
				}
				if strings.Contains(qn, needle) {
					filtered = append(filtered, c)
				}
			}
			checks = filtered
		}

		if !fromTime.IsZero() || !untilTime.IsZero() {
			filtered := checks[:0]
			for _, c := range checks {
				t, err := time.Parse(time.RFC3339, c.LastCheckRunTime)
				if err != nil {
					continue
				}
				if !fromTime.IsZero() && t.Before(fromTime) {
					continue
				}
				if !untilTime.IsZero() && t.After(untilTime) {
					continue
				}
				filtered = append(filtered, c)
			}
			checks = filtered
		}

		// Sort
		sort.SliceStable(checks, func(i, j int) bool {
			a, b := checks[i], checks[j]
			var less bool
			switch sortCol {
			case "date":
				ta, _ := time.Parse(time.RFC3339, a.LastCheckRunTime)
				tb, _ := time.Parse(time.RFC3339, b.LastCheckRunTime)
				less = ta.Before(tb)
			case "dataset":
				qa, qb := "", ""
				if len(a.Datasets) > 0 {
					qa = a.Datasets[0].QualifiedName
				}
				if len(b.Datasets) > 0 {
					qb = b.Datasets[0].QualifiedName
				}
				less = qa < qb
			case "name":
				less = a.Name < b.Name
			case "column":
				less = a.Column < b.Column
			case "status":
				less = a.EvaluationStatus < b.EvaluationStatus
			}
			if order == "desc" {
				return !less
			}
			return less
		})

		// Apply limit
		total := len(checks)
		if limit > 0 && len(checks) > limit {
			checks = checks[:limit]
		}

		if len(checks) == 0 {
			fmt.Println(output.Dim.Render("  No results found."))
			return nil
		}

		rows := make([]map[string]string, len(checks))
		for i, c := range checks {
			qualifiedName := ""
			if len(c.Datasets) > 0 {
				if c.Datasets[0].QualifiedName != "" {
					qualifiedName = c.Datasets[0].QualifiedName
				} else {
					qualifiedName = c.Datasets[0].Name
				}
			}

			col := c.Column
			if col == "" {
				col = "dataset"
			}

			rows[i] = map[string]string{
				"dataset": qualifiedName,
				"type":    "check",
				"name":    c.Name,
				"column":  col,
				"status":  fmtCheckStatus(c.EvaluationStatus),
				"date":    fmtCheckTime(c.LastCheckRunTime),
			}
		}

		cols := []string{"dataset", "type", "name", "column", "status", "date"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)

		if total > len(checks) || !result.Last {
			shown := len(checks)
			apiTotal := result.TotalElements
			fmt.Fprintf(cmd.ErrOrStderr(), output.Dim.Render("  Showing %d of %d results.\n"), shown, apiTotal)
		}
		return nil
	},
}

// parseDate accepts YYYY-MM-DD or any RFC3339 timestamp.
func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC(), nil
	}
	return time.Parse(time.RFC3339, s)
}

// fmtCheckStatus maps API evaluation status values to display values.
func fmtCheckStatus(s string) string {
	switch s {
	case "pass":
		return "passing"
	case "fail":
		return "failing"
	case "notEvaluated":
		return "n/a"
	default:
		return s
	}
}

// fmtCheckTime formats an ISO timestamp as "2006-01-02 15:04".
func fmtCheckTime(s string) string {
	if s == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.UTC().Format("2006-01-02 15:04")
}

func init() {
	resultsListCmd.Flags().String("dataset", "", "Filter by dataset ID")
	resultsListCmd.Flags().String("dataset-name", "", "Filter by dataset qualified name (substring match)")
	resultsListCmd.Flags().String("status", "", "Filter by status: passing|failing|error")
	resultsListCmd.Flags().String("type", "check", "Filter by type: check|monitor|all")
	resultsListCmd.Flags().Int("limit", 10, "Maximum number of results to show")
	resultsListCmd.Flags().String("sort", "date", "Sort by column: dataset|name|column|status|date")
	resultsListCmd.Flags().String("order", "desc", "Sort order: asc|desc")
	resultsListCmd.Flags().String("from", "", "Show results on or after this date (YYYY-MM-DD or ISO8601)")
	resultsListCmd.Flags().String("until", "", "Show results on or before this date (YYYY-MM-DD or ISO8601)")

	resultsCmd.AddCommand(resultsListCmd)
	rootCmd.AddCommand(resultsCmd)
}
