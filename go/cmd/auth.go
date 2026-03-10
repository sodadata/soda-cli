package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/config"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication and profiles",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Soda Cloud",
	Long: "Authenticate with Soda Cloud using an API key ID and secret. Credentials are stored in ~/.soda/credentials.\n\nTo generate API keys: https://docs.soda.io/reference/generate-api-keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		host, _ := cmd.Flags().GetString("host")
		apiKeyID, _ := cmd.Flags().GetString("api-key-id")
		apiKeySecret, _ := cmd.Flags().GetString("api-key-secret")

		if apiKeyID == "" || apiKeySecret == "" {
			if GCtx.NoInteractive {
				return output.Errorf(2, "--api-key-id and --api-key-secret are required in non-interactive mode")
			}
			if host == "" {
				host = "cloud.soda.io"
			}
			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("Soda Cloud host").
					Description("Change to cloud.us.soda.io for the US region.").
					Options(
						huh.NewOption("cloud.soda.io (EU)", "cloud.soda.io"),
						huh.NewOption("cloud.us.soda.io (US)", "cloud.us.soda.io"),
					).
					Value(&host),
				huh.NewInput().
					Title("API key ID").
					Value(&apiKeyID),
				huh.NewInput().
					Title("API key secret").
					EchoMode(huh.EchoModePassword).
					Value(&apiKeySecret),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "login cancelled")
			}
		}

		if host == "" {
			host = "cloud.soda.io"
		}

		fmt.Println(output.Dim.Render("  Testing connection to " + host + "..."))

		// Test connection before saving
		testProfile := config.Profile{Host: host, APIKeyID: apiKeyID, APIKeySecret: apiKeySecret}
		if err := api.New(testProfile).Ping(); err != nil {
			return err
		}

		// Save credentials
		creds, err := config.LoadCredentials()
		if err != nil {
			creds = config.Credentials{}
		}
		profileName := GCtx.Profile
		if profileName == "" {
			profileName = "default"
		}
		creds[profileName] = testProfile
		if err := config.SaveCredentials(creds); err != nil {
			return output.Errorf(2, "could not save credentials: %v", err)
		}

		output.PrintSuccess(fmt.Sprintf("Authenticated. Profile '%s' saved to ~/.soda/credentials.", profileName), GCtx)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials for the active profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := GCtx.Profile
		if profileName == "" {
			profileName = "default"
		}
		creds, err := config.LoadCredentials()
		if err != nil {
			return output.Errorf(2, "could not read credentials: %v", err)
		}
		if _, ok := creds[profileName]; !ok {
			return output.Errorf(2, "profile '%s' not found in ~/.soda/credentials", profileName)
		}
		delete(creds, profileName)
		if err := config.SaveCredentials(creds); err != nil {
			return output.Errorf(2, "could not update credentials: %v", err)
		}
		output.PrintSuccess(fmt.Sprintf("Logged out. Removed profile '%s' from ~/.soda/credentials.", profileName), GCtx)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active profile and connection health",
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := GCtx.Profile
		if profileName == "" {
			profileName = "default"
		}
		creds, err := config.LoadCredentials()
		if err != nil {
			return output.Errorf(2, "could not read credentials: %v", err)
		}
		p, ok := creds[profileName]
		host := p.Host
		if host == "" {
			host = "cloud.soda.io"
		}

		fmt.Printf("  %-20s %s\n", output.Bold.Render("Profile"), profileName)
		fmt.Printf("  %-20s %s\n", output.Bold.Render("Host"), host)

		if !ok || p.APIKeyID == "" {
			fmt.Printf("  %-20s %s\n", output.Bold.Render("Connection"), output.Dim.Render("not configured — run `soda auth login`"))
			return nil
		}

		fmt.Printf("  %-20s %s\n", output.Bold.Render("API Key ID"), p.APIKeyID)

		if pingErr := api.New(p).Ping(); pingErr != nil {
			fmt.Printf("  %-20s %s\n", output.Bold.Render("Connection"), output.Red.Render("✗ failed — "+pingErr.Error()))
		} else {
			fmt.Printf("  %-20s %s\n", output.Bold.Render("Connection"), output.Green.Render("✓ connected"))
		}
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
	authLoginCmd.Flags().String("host", "", "Soda Cloud host (default: cloud.soda.io)")
	authLoginCmd.Flags().String("api-key-id", "", "Soda Cloud API key ID")
	authLoginCmd.Flags().String("api-key-secret", "", "Soda Cloud API key secret")

	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd, authSwitchCmd)
	rootCmd.AddCommand(authCmd)
}
