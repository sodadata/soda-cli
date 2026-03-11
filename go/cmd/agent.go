package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage Soda Agents",
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered Soda Agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		result, err := client.ListAgents(100)
		if err != nil {
			return err
		}

		if len(result.Content) == 0 {
			fmt.Println(output.Dim.Render("  No agents found."))
			return nil
		}

		rows := make([]map[string]string, len(result.Content))
		for i, a := range result.Content {
			rows[i] = map[string]string{
				"id":        a.ID,
				"name":      a.Name,
				"type":      fmtAgentType(a.Type),
				"status":    fmtAgentStatus(a.IsOnline),
				"last seen": fmtAgentTime(a.LastSeenTimestamp),
				"version":   a.Versions.Agent,
			}
		}

		cols := []string{"id", "name", "type", "status", "last seen", "version"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var agentGetCmd = &cobra.Command{
	Use:   "get <agent-id>",
	Short: "Show details for a specific agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		a, err := client.GetAgent(args[0])
		if err != nil {
			return err
		}

		rows := []map[string]string{{
			"id":        a.ID,
			"name":      a.Name,
			"label":     a.Label,
			"type":      fmtAgentType(a.Type),
			"status":    fmtAgentStatus(a.IsOnline),
			"last seen": fmtAgentTime(a.LastSeenTimestamp),
			"agent version":  a.Versions.Agent,
			"library version": a.Versions.Library,
		}}

		output.RenderOne(rows[0], []string{"id", "name", "label", "type", "status", "last seen", "agent version", "library version"}, GCtx)
		return nil
	},
}

func fmtAgentType(t string) string {
	switch strings.ToUpper(t) {
	case "SELF_HOSTED":
		return "self-hosted"
	case "SODA_HOSTED":
		return "soda-hosted"
	default:
		return strings.ToLower(t)
	}
}

func fmtAgentStatus(online bool) string {
	if online {
		return "online"
	}
	return "offline"
}

func fmtAgentTime(s string) string {
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
	agentCmd.AddCommand(agentListCmd, agentGetCmd)
	rootCmd.AddCommand(agentCmd)
}
