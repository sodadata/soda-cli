package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

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
		return output.Errorf(2, "job list is not yet available in the public API")
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

var jobStatusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Show the status of a scan/job",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		status, err := client.GetScanStatus(args[0])
		if err != nil {
			return err
		}

		if output.EffectiveFmt(GCtx) == "json" {
			item := map[string]string{
				"id":       status.ID,
				"state":    status.State,
				"started":  status.Started,
				"ended":    status.Ended,
				"checks":   fmt.Sprintf("%d", len(status.Checks)),
				"failures": fmt.Sprintf("%d", status.Failures),
				"warnings": fmt.Sprintf("%d", status.Warnings),
				"errors":   fmt.Sprintf("%d", status.Errors),
				"cloudUrl": status.CloudURL,
			}
			keys := []string{"id", "state", "started", "ended", "checks", "failures", "warnings", "errors", "cloudUrl"}
			output.RenderOne(item, keys, GCtx)
			return nil
		}

		fmt.Printf("  %-18s %s\n", output.Bold.Render("ID"), output.FmtID(status.ID))
		fmt.Printf("  %-18s %s\n", output.Bold.Render("State"), output.FmtStatus(status.State))
		if status.Started != "" {
			fmt.Printf("  %-18s %s\n", output.Bold.Render("Started"), output.FmtTime(status.Started))
		}
		if status.Ended != "" {
			fmt.Printf("  %-18s %s\n", output.Bold.Render("Ended"), output.FmtTime(status.Ended))
		}
		if status.Failures > 0 || status.Warnings > 0 || status.Errors > 0 {
			fmt.Printf("  %-18s %d failures, %d warnings, %d errors\n",
				output.Bold.Render("Results"), status.Failures, status.Warnings, status.Errors)
		}
		if status.CloudURL != "" {
			fmt.Printf("  %-18s %s\n", output.Bold.Render("Cloud URL"), status.CloudURL)
		}

		if len(status.Checks) > 0 {
			// The scan status API returns only id + evaluationStatus per check.
			// Show a summary breakdown rather than a table of empty names.
			counts := map[string]int{}
			for _, ch := range status.Checks {
				counts[ch.EvaluationStatus]++
			}
			fmt.Printf("  %-18s %d total", output.Bold.Render("Checks"), len(status.Checks))
			for _, s := range []string{"pass", "fail", "warn", "notEvaluated", "excluded"} {
				if c, ok := counts[s]; ok && c > 0 {
					fmt.Printf(", %d %s", c, s)
				}
			}
			fmt.Println()
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
	jobLogsCmd.Flags().Bool("follow", false, "Stream logs as they arrive")

	jobCmd.AddCommand(jobListCmd, jobStatusCmd, jobLogsCmd, jobCancelCmd)
}
