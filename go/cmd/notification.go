package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var notificationCmd = &cobra.Command{
	Use:   "notification",
	Short: "Manage notification rules and channels",
}

var notifListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		channel, _ := cmd.Flags().GetString("channel")
		dataset, _ := cmd.Flags().GetString("dataset")

		rows := mock.Notifications
		filtered := []map[string]string{}
		for _, n := range rows {
			if channel != "" && n["channel"] != channel {
				continue
			}
			if dataset != "" && n["dataset"] != dataset {
				continue
			}
			filtered = append(filtered, n)
		}

		cols := []string{"id", "channel", "trigger", "dataset", "status"}
		output.Render(filtered, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var notifAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a notification rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		channel, _ := cmd.Flags().GetString("channel")
		trigger, _ := cmd.Flags().GetString("trigger")
		dataset, _ := cmd.Flags().GetString("dataset")

		if channel == "" || trigger == "" {
			return output.Errorf(2, "--channel and --trigger are required")
		}

		scope := "(all datasets)"
		if dataset != "" {
			scope = dataset
		}
		output.PrintSuccess(fmt.Sprintf("Notification rule created: %s on %s → %s", trigger, scope, channel), GCtx)
		return nil
	},
}

var notifUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a notification rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Notification rule %s updated.", args[0]), GCtx)
		return nil
	},
}

var notifDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a notification rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Notification rule %s deleted.", args[0]), GCtx)
		return nil
	},
}

// notification channel sub-group
var notifChannelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Manage notification channels (Slack, Teams, webhooks)",
}

var channelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		cols := []string{"id", "name", "type", "status"}
		output.Render(mock.Channels, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var channelAddCmd = &cobra.Command{
	Use:       "add slack|teams|webhook",
	Short:     "Add a notification channel",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"slack", "teams", "webhook"},
	RunE: func(cmd *cobra.Command, args []string) error {
		chType := args[0]

		if GCtx.NoInteractive {
			return output.Errorf(2, "interactive setup required for channel configuration")
		}

		fmt.Printf("  Configuring %s channel...\n", chType)
		output.PrintSuccess(fmt.Sprintf("%s channel added.", chType), GCtx)
		return nil
	},
}

var channelDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a notification channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Channel %s deleted.", args[0]), GCtx)
		return nil
	},
}

var channelTestCmd = &cobra.Command{
	Use:   "test <id>",
	Short: "Send a test message to a channel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(output.Dim.Render("  Sending test message to channel " + args[0] + "..."))
		output.PrintSuccess("Test message sent successfully.", GCtx)
		return nil
	},
}

func init() {
	notifListCmd.Flags().String("channel", "", "Filter by channel ID")
	notifListCmd.Flags().String("dataset", "", "Filter by dataset ID")

	notifAddCmd.Flags().String("channel", "", "Channel ID (required)")
	notifAddCmd.Flags().String("trigger", "", "Trigger event (required): check-failure|incident-opened|incident-closed")
	notifAddCmd.Flags().String("dataset", "", "Scope to a specific dataset (optional)")

	notifChannelCmd.AddCommand(channelListCmd, channelAddCmd, channelDeleteCmd, channelTestCmd)

	notificationCmd.AddCommand(notifListCmd, notifAddCmd, notifUpdateCmd, notifDeleteCmd, notifChannelCmd)
	rootCmd.AddCommand(notificationCmd)
}
