package cmd

import (
	"fmt"
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
		datasetID, _   := cmd.Flags().GetString("dataset")
		datasetName, _ := cmd.Flags().GetString("dataset-name")
		status, _      := cmd.Flags().GetString("status")
		resType, _     := cmd.Flags().GetString("type")
		limit, _       := cmd.Flags().GetInt("limit")

		// Monitor results not yet available
		if resType == "monitor" {
			fmt.Println(output.Dim.Render("  Monitor results are not yet available in the public API."))
			return nil
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

		// Fetch more when client-side filters are active so they have enough data to work with
		fetchSize := limit
		if statusFilter != "" || datasetName != "" {
			fetchSize = 500
		}
		result, err := client.ListChecks(api.ListChecksParams{Size: fetchSize, DatasetID: datasetID})
		if err != nil {
			return err
		}

		checks := result.Content

		// Client-side: filter by status
		if statusFilter != "" {
			filtered := checks[:0]
			for _, c := range checks {
				if c.EvaluationStatus == statusFilter {
					filtered = append(filtered, c)
				}
			}
			checks = filtered
		}

		// Client-side: filter by dataset qualified name (substring match)
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

			// Column: use "dataset" for dataset-level checks, otherwise the column name
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

	resultsCmd.AddCommand(resultsListCmd)
	rootCmd.AddCommand(resultsCmd)
}
