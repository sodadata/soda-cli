package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var notificationCmd = &cobra.Command{
	Use:   "notification",
	Short: "Manage notification rules and integrations",
}

// ── notification rule ─────────────────────────────────────────────────────────

var notifRuleCmd = &cobra.Command{
	Use:   "rule",
	Short: "Manage notification rules",
}

var notifRuleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		cols := []string{"id", "name", "source", "alert", "dataset", "notify"}
		output.Render(mock.NotificationRules, cols, nil, GCtx)
		return nil
	},
}

var notifRuleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a notification rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		source, _ := cmd.Flags().GetString("source")
		alert, _ := cmd.Flags().GetString("alert")
		notify, _ := cmd.Flags().GetStringArray("notify")

		if name == "" || source == "" || alert == "" || len(notify) == 0 {
			return output.Errorf(2, "--name, --source, --alert, and --notify are required")
		}
		output.PrintSuccess(fmt.Sprintf("Notification rule '%s' created.", name), GCtx)
		return nil
	},
}

var notifRuleUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a notification rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Notification rule '%s' updated.", args[0]), GCtx)
		return nil
	},
}

var notifRuleDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a notification rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Notification rule '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

// ── notification integration ──────────────────────────────────────────────────

var notifIntegrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "Manage notification integrations (Slack, Teams, webhooks)",
}

var notifIntegrationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification integrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cols := []string{"id", "name", "type", "status"}
		output.Render(mock.Integrations, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var notifIntegrationAddCmd = &cobra.Command{
	Use:       "add slack|teams|webhook",
	Short:     "Add a notification integration",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"slack", "teams", "webhook"},
	RunE: func(cmd *cobra.Command, args []string) error {
		intType := args[0]
		if GCtx.NoInteractive {
			return output.Errorf(2, "interactive setup required for integration configuration")
		}
		fmt.Printf("  Configuring %s integration...\n", intType)
		output.PrintSuccess(fmt.Sprintf("%s integration added.", intType), GCtx)
		return nil
	},
}

var notifIntegrationTestCmd = &cobra.Command{
	Use:   "test <id>",
	Short: "Send a test message to an integration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(output.Dim.Render("  Sending test message to integration " + args[0] + "..."))
		output.PrintSuccess("Test message sent successfully.", GCtx)
		return nil
	},
}

var notifIntegrationDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a notification integration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Integration '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

func init() {
	// rule flags
	notifRuleAddCmd.Flags().String("name", "", "Rule name (required)")
	notifRuleAddCmd.Flags().String("source", "", "Notification source: check|monitor|all (required)")
	notifRuleAddCmd.Flags().String("alert", "", "Alert condition: warn-fail|fail-only|anomaly (required)")
	notifRuleAddCmd.Flags().StringArray("notify", nil, "Recipient email or role ID (repeatable, required)")
	notifRuleAddCmd.Flags().String("datasource", "", "Scope to a datasource label")
	notifRuleAddCmd.Flags().String("dataset", "", "Scope to a dataset label")
	notifRuleAddCmd.Flags().String("dataset-owner", "", "Scope to datasets owned by email")
	notifRuleAddCmd.Flags().String("dataset-tag", "", "Scope to datasets with tag")
	notifRuleAddCmd.Flags().String("check-name", "", "Filter by check name (supports 'contains:value')")
	notifRuleAddCmd.Flags().String("check-owner", "", "Filter by check owner email")
	notifRuleAddCmd.Flags().String("monitor-type", "", "Filter by monitor type")
	notifRuleAddCmd.Flags().Bool("granular-results", false, "Send one notification per result (not a summary)")
	notifRuleAddCmd.Flags().String("message", "", "Custom notification message")

	notifRuleUpdateCmd.Flags().String("name", "", "Rule name")
	notifRuleUpdateCmd.Flags().String("source", "", "Notification source: check|monitor|all")
	notifRuleUpdateCmd.Flags().String("alert", "", "Alert condition: warn-fail|fail-only|anomaly")
	notifRuleUpdateCmd.Flags().StringArray("notify", nil, "Recipient email or role ID (repeatable)")
	notifRuleUpdateCmd.Flags().String("datasource", "", "Scope to a datasource label")
	notifRuleUpdateCmd.Flags().String("dataset", "", "Scope to a dataset label")
	notifRuleUpdateCmd.Flags().String("dataset-owner", "", "Scope to datasets owned by email")
	notifRuleUpdateCmd.Flags().String("dataset-tag", "", "Scope to datasets with tag")
	notifRuleUpdateCmd.Flags().String("check-name", "", "Filter by check name")
	notifRuleUpdateCmd.Flags().String("check-owner", "", "Filter by check owner email")
	notifRuleUpdateCmd.Flags().String("monitor-type", "", "Filter by monitor type")
	notifRuleUpdateCmd.Flags().Bool("granular-results", false, "Send one notification per result")
	notifRuleUpdateCmd.Flags().String("message", "", "Custom notification message")

	notifRuleCmd.AddCommand(notifRuleListCmd, notifRuleAddCmd, notifRuleUpdateCmd, notifRuleDeleteCmd)

	notifIntegrationCmd.AddCommand(notifIntegrationListCmd, notifIntegrationAddCmd, notifIntegrationTestCmd, notifIntegrationDeleteCmd)

	notificationCmd.AddCommand(notifRuleCmd, notifIntegrationCmd)
	rootCmd.AddCommand(notificationCmd)
}
