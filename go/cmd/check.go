package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "View check definitions from Soda Cloud",
}

var checkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List check definitions for a dataset or by check IDs",
	Long: `List check definitions from Soda Cloud.

  Requires either --dataset-id or --ids (mutually exclusive).

  Exit codes: 0=success, 2=error, 3=auth error`,
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetID, _ := cmd.Flags().GetString("dataset-id")
		ids, _       := cmd.Flags().GetString("ids")

		if datasetID == "" && ids == "" {
			return output.Errorf(2, "--dataset-id or --ids is required")
		}
		if datasetID != "" && ids != "" {
			return output.Errorf(2, "--dataset-id and --ids are mutually exclusive")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		result, err := client.ListChecks(api.ListChecksParams{
			DatasetID: datasetID,
			CheckIDs:  ids,
			Size:      500,
		})
		if err != nil {
			return err
		}

		if len(result.Content) == 0 {
			fmt.Println(output.Dim.Render("  No checks found."))
			return nil
		}

		rows := make([]map[string]string, len(result.Content))
		for i, c := range result.Content {
			value := ""
			if c.LastCheckResultValue != nil && c.LastCheckResultValue.Value != nil {
				value = strconv.FormatFloat(*c.LastCheckResultValue.Value, 'f', 2, 64)
			}
			rows[i] = map[string]string{
				"id":     c.ID,
				"column": c.Column,
				"type":   c.CheckType,
				"name":   c.Name,
				"value":  value,
				"status": fmtCheckStatus(c.EvaluationStatus),
			}
		}

		cols := []string{"id", "column", "type", "name", "value", "status"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

func init() {
	checkListCmd.Flags().String("dataset-id", "", "Filter checks by dataset ID")
	checkListCmd.Flags().String("ids", "", "Comma-separated list of check IDs to fetch")

	checkCmd.AddCommand(checkListCmd)
}
