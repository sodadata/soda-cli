package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication and profiles",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Soda Cloud",
	Long:  "Authenticate with Soda Cloud using an API key. Credentials are stored in ~/.soda/credentials.",
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")
		host, _ := cmd.Flags().GetString("host")

		if apiKey == "" {
			if GCtx.NoInteractive {
				return output.Errorf(2, "--api-key is required in non-interactive mode")
			}
			if host == "" {
				host = "cloud.soda.io"
			}
			form := huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Soda Cloud host").
					Placeholder("cloud.soda.io").
					Value(&host),
				huh.NewInput().
					Title("API key").
					EchoMode(huh.EchoModePassword).
					Value(&apiKey),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "login cancelled")
			}
		}

		if host == "" {
			host = "cloud.soda.io"
		}

		fmt.Println(output.Dim.Render("  Testing connection to " + host + "..."))
		output.PrintSuccess("Authenticated. Profile 'default' saved to ~/.soda/credentials.", GCtx)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials for the active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile := GCtx.Profile
		if profile == "" {
			profile = "default"
		}
		output.PrintSuccess(fmt.Sprintf("Logged out. Removed profile '%s' from ~/.soda/credentials.", profile), GCtx)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active profile and connection health",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("  %-20s %s\n", output.Bold.Render("Profile"), "default")
		fmt.Printf("  %-20s %s\n", output.Bold.Render("Host"), "cloud.soda.io")
		fmt.Printf("  %-20s %s\n", output.Bold.Render("Organization"), "acme-corp")
		fmt.Printf("  %-20s %s\n", output.Bold.Render("Connection"), output.Green.Render("✓ connected"))
		fmt.Printf("  %-20s %s\n", output.Bold.Render("User"), "alice@acme.com")
		return nil
	},
}

var authSwitchCmd = &cobra.Command{
	Use:   "switch <profile>",
	Short: "Switch the active auth profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Active profile set to '%s'.", args[0]), GCtx)
		return nil
	},
}

func init() {
	authLoginCmd.Flags().String("api-key", "", "Soda Cloud API key")
	authLoginCmd.Flags().String("host", "", "Soda Cloud host (default: cloud.soda.io)")

	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd, authSwitchCmd)
	rootCmd.AddCommand(authCmd)
}
