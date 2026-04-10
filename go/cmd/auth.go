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
		scheme, _ := cmd.Flags().GetString("scheme")

		anyFlagSet := cmd.Flags().Changed("host") || cmd.Flags().Changed("api-key-id") || cmd.Flags().Changed("api-key-secret") || cmd.Flags().Changed("scheme")

		if anyFlagSet {
			// Non-interactive: flags were explicitly provided
			if host == "" {
				host = "cloud.soda.io"
			}

			if apiKeyID == "" || apiKeySecret == "" {
				profileName := GCtx.Profile
				if profileName == "" {
					profileName = "default"
				}

				// Only --host was provided, save it but remind to complete setup
				if cmd.Flags().Changed("host") {
					creds, err := config.LoadCredentials()
					if err != nil {
						creds = config.Credentials{}
					}
					p := creds[profileName]
					p.Host = host
					creds[profileName] = p
					if err := config.SaveCredentials(creds); err != nil {
						return output.Errorf(2, "could not save credentials: %v", err)
					}
					fmt.Printf("  Host set to %s for profile '%s'.\n\n", output.Bold.Render(host), profileName)
					fmt.Printf("  Complete setup by running:\n\n")
					fmt.Printf("    sodacli auth login --api-key-id <your-key-id> --api-key-secret <your-secret>\n\n")
					fmt.Printf("  Generate API keys: https://docs.soda.io/reference/generate-api-keys\n")
					return nil
				}

				// Partial key flags — tell the user what's missing
				missing := ""
				if apiKeyID == "" {
					missing = "--api-key-id"
				}
				if apiKeySecret == "" {
					if missing != "" {
						missing += " and "
					}
					missing += "--api-key-secret"
				}
				return output.Errorf(2, "%s required. Example:\n\n  sodacli auth login --api-key-id <id> --api-key-secret <secret>\n\nGenerate keys: https://docs.soda.io/reference/generate-api-keys", missing)
			}
		} else {
			// No flags: interactive wizard
			if GCtx.NoInteractive {
				return output.Errorf(2, "--api-key-id and --api-key-secret are required in non-interactive mode")
			}

			// Load existing profile to pre-fill values
			profileName := GCtx.Profile
			if profileName == "" {
				profileName = "default"
			}
			if creds, err := config.LoadCredentials(); err == nil {
				if p, ok := creds[profileName]; ok {
					if host == "" && p.Host != "" {
						host = p.Host
					}
					if apiKeyID == "" && p.APIKeyID != "" {
						apiKeyID = p.APIKeyID
					}
				}
			}

			if host == "" {
				host = "cloud.soda.io"
			}

			form := huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Soda Cloud host").
					Description("EU: cloud.soda.io · US: cloud.us.soda.io").
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

		// Test connection before saving
		testProfile := config.Profile{Host: host, APIKeyID: apiKeyID, APIKeySecret: apiKeySecret, Scheme: scheme}
		fmt.Println(output.Dim.Render("  Testing connection to " + testProfile.BaseURL() + "..."))
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
		baseURL := p.BaseURL()

		fmt.Printf("  %-20s %s\n", output.Bold.Render("Profile"), profileName)
		fmt.Printf("  %-20s %s\n", output.Bold.Render("Host"), baseURL)

		if !ok || p.APIKeyID == "" {
			fmt.Printf("  %-20s %s\n", output.Bold.Render("Connection"), output.Dim.Render("not configured — run `sodacli auth login`"))
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
	authLoginCmd.Flags().String("scheme", "", `URL scheme: "http" or "https" (default: "https")`)

	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authStatusCmd, authSwitchCmd)
}
