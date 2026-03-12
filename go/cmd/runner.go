package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Manage Soda Runners",
}

var runnerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered Soda Runners",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		result, err := client.ListRunners(100)
		if err != nil {
			return err
		}

		if len(result.Content) == 0 {
			fmt.Println(output.Dim.Render("  No runners found."))
			return nil
		}

		rows := make([]map[string]string, len(result.Content))
		for i, a := range result.Content {
			rows[i] = map[string]string{
				"id":        a.ID,
				"name":      a.Name,
				"type":      fmtRunnerType(a.Type),
				"status":    fmtRunnerStatus(a.IsOnline),
				"last seen": fmtRunnerTime(a.LastSeenTimestamp),
				"version":   a.Versions.Agent,
			}
		}

		cols := []string{"id", "name", "type", "status", "last seen", "version"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var runnerGetCmd = &cobra.Command{
	Use:   "get <runner-id>",
	Short: "Show details for a specific runner",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		a, err := client.GetRunner(args[0])
		if err != nil {
			return err
		}

		rows := []map[string]string{{
			"id":              a.ID,
			"name":            a.Name,
			"label":           a.Label,
			"type":            fmtRunnerType(a.Type),
			"status":          fmtRunnerStatus(a.IsOnline),
			"last seen":       fmtRunnerTime(a.LastSeenTimestamp),
			"runner version":  a.Versions.Agent,
			"library version": a.Versions.Library,
		}}

		output.RenderOne(rows[0], []string{"id", "name", "label", "type", "status", "last seen", "runner version", "library version"}, GCtx)
		return nil
	},
}

func fmtRunnerType(t string) string {
	switch strings.ToUpper(t) {
	case "SELF_HOSTED":
		return "self-hosted"
	case "SODA_HOSTED":
		return "soda-hosted"
	default:
		return strings.ToLower(t)
	}
}

func fmtRunnerStatus(online bool) string {
	if online {
		return "online"
	}
	return "offline"
}

func fmtRunnerTime(s string) string {
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
	runnerCmd.AddCommand(runnerListCmd, runnerGetCmd)
	rootCmd.AddCommand(runnerCmd)
}
