package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
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

var runnerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Soda Runner",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return output.Errorf(2, "--name is required")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		result, err := client.CreateRunner(api.CreateRunnerRequest{Name: name})
		if err != nil {
			return err
		}

		fmt.Printf("  %-20s %s\n", output.Bold.Render("Runner ID"), result.ID)
		fmt.Printf("  %-20s %s\n", output.Bold.Render("API Key ID"), result.APIKeyID)
		fmt.Printf("  %-20s %s\n", output.Bold.Render("API Key Secret"), result.APIKeySecret)
		fmt.Println()
		fmt.Println(output.Yellow.Render("  Save the API key secret now — it will not be shown again."))
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Runner '%s' created.", name), GCtx)
		fmt.Println()
		fmt.Println(output.Dim.Render("  Next steps:"))
		fmt.Println(output.Dim.Render("  To connect this runner, deploy the Soda Runner Helm chart on your Kubernetes cluster."))
		fmt.Println(output.Dim.Render("  Docs: https://go.soda.io/agent"))
		return nil
	},
}

var runnerDeleteCmd = &cobra.Command{
	Use:   "delete <runner-id>",
	Short: "Delete a Soda Runner",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		if err := client.DeleteRunner(args[0]); err != nil {
			return err
		}
		output.PrintSuccess("Runner deleted.", GCtx)
		return nil
	},
}

func init() {
	runnerCreateCmd.Flags().String("name", "", "Name for the new runner")

	runnerCmd.AddCommand(runnerListCmd, runnerGetCmd, runnerCreateCmd, runnerDeleteCmd)
}
