package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var jobCmd = &cobra.Command{
	Use:     "job",
	Aliases: []string{"scan"},
	Short:   "View execution history and logs (read-only)",
}

var jobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		datasource, _ := cmd.Flags().GetString("datasource")
		dataset, _ := cmd.Flags().GetString("dataset")
		jobType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")

		rows := mock.Jobs
		filtered := []map[string]string{}
		for _, j := range rows {
			if datasource != "" && j["datasource"] != datasource {
				continue
			}
			if dataset != "" && j["dataset"] != dataset {
				continue
			}
			if jobType != "" && jobType != "all" && j["type"] != jobType {
				continue
			}
			if status != "" && j["status"] != status {
				continue
			}
			filtered = append(filtered, j)
		}

		cols := []string{"id", "datasource", "dataset", "type", "status", "date"}
		output.Render(filtered, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var jobLogsCmd = &cobra.Command{
	Use:   "logs <id>",
	Short: "Stream or display logs for a job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		result, err := client.GetScanLogs(args[0])
		if err != nil {
			return err
		}

		// API may return logs in Content or Logs field
		logs := result.Content
		if len(logs) == 0 {
			logs = result.Logs
		}

		if len(logs) == 0 {
			fmt.Println(output.Dim.Render("  No logs found for scan " + args[0] + "."))
			return nil
		}

		for _, entry := range logs {
			ts := entry.Timestamp
			if ts != "" {
				ts = output.Dim.Render(ts) + "  "
			}
			level := ""
			if entry.Level != "" {
				switch entry.Level {
				case "ERROR", "error":
					level = output.Red.Render(entry.Level) + "  "
				case "WARN", "warn", "WARNING":
					level = output.Yellow.Render(entry.Level) + "  "
				default:
					level = output.Dim.Render(entry.Level) + "  "
				}
			}
			fmt.Printf("  %s%s%s\n", ts, level, entry.Message)
		}
		return nil
	},
}

var jobCancelCmd = &cobra.Command{
	Use:   "cancel <id>",
	Short: "Cancel a running scan",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		if err := client.CancelScan(args[0]); err != nil {
			return err
		}
		output.PrintSuccess("Scan cancellation requested.", GCtx)
		return nil
	},
}

func init() {
	jobListCmd.Flags().String("datasource", "", "Filter by datasource ID")
	jobListCmd.Flags().String("dataset", "", "Filter by dataset ID")
	jobListCmd.Flags().String("type", "all", "Filter by type: contract|monitor|all")
	jobListCmd.Flags().String("status", "", "Filter by status: passing|failing|running|error")

	jobLogsCmd.Flags().Bool("follow", false, "Stream logs as they arrive")

	jobCmd.AddCommand(jobListCmd, jobLogsCmd, jobCancelCmd)
}
