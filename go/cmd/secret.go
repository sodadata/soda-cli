package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets stored in Soda Cloud",
}

var secretListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		page, err := client.ListSecrets(0, 100)
		if err != nil {
			return err
		}
		if len(page.Content) == 0 {
			fmt.Println(output.Dim.Render("  No secrets found."))
			return nil
		}
		rows := make([]map[string]string, len(page.Content))
		for i, s := range page.Content {
			rows[i] = map[string]string{
				"id":      s.ID,
				"name":    s.Name,
				"created": s.Created,
				"updated": s.LastUpdated,
			}
		}
		cols := []string{"id", "name", "created", "updated"}
		output.Render(rows, cols, nil, GCtx)
		if !page.Last {
			fmt.Fprintf(cmd.ErrOrStderr(), output.Dim.Render("  Showing %d of %d secrets.\n"), len(page.Content), page.TotalElements)
		}
		return nil
	},
}

var secretGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show secret details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		s, err := client.GetSecret(args[0])
		if err != nil {
			return err
		}
		item := map[string]string{
			"id":      s.ID,
			"name":    s.Name,
			"created": s.Created,
			"updated": s.LastUpdated,
		}
		keys := []string{"id", "name", "created", "updated"}
		output.RenderOne(item, keys, GCtx)
		return nil
	},
}

var secretCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new secret",
	Long: `Create a new secret in Soda Cloud.

Secret values are encrypted client-side using AES-256-GCM + RSA-OAEP before
being sent to the API. Soda never sees the plaintext value — decryption
happens only during scan execution within the runner.

Use ${secret.NAME} in datasource configs to reference the secret.

The value can be provided via:
  --value flag         (visible in shell history — use for scripts/CI)
  interactive prompt   (masked input — default when --value is omitted)
  stdin pipe           (echo "val" | sodacli secret create --name X)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return output.Errorf(2, "--name is required")
		}

		value, err := readSecretValue(cmd, "Secret value")
		if err != nil {
			return err
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}
		s, err := client.CreateSecret(name, value)
		if err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Secret '%s' created (id: %s).", s.Name, s.ID), GCtx)
		fmt.Printf("  Reference it in configs as: ${secret.%s}\n", s.Name)
		return nil
	},
}

var secretUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update the value of a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value, err := readSecretValue(cmd, "New secret value")
		if err != nil {
			return err
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}
		s, err := client.UpdateSecret(args[0], value)
		if err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Secret '%s' updated.", s.Name), GCtx)
		return nil
	},
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		if err := client.DeleteSecret(args[0]); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Secret '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

// readSecretValue resolves the secret value from (in priority order):
//  1. --value flag (for scripts/CI — user accepts the shell-history trade-off)
//  2. stdin pipe  (echo "val" | sodacli secret create ...)
//  3. interactive masked prompt via huh
func readSecretValue(cmd *cobra.Command, prompt string) (string, error) {
	// 1. Explicit flag
	if cmd.Flags().Changed("value") {
		v, _ := cmd.Flags().GetString("value")
		if v == "" {
			return "", output.Errorf(2, "--value cannot be empty")
		}
		return v, nil
	}

	// 2. Piped stdin
	if stat, _ := os.Stdin.Stat(); stat.Mode()&os.ModeCharDevice == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			v := strings.TrimSpace(scanner.Text())
			if v == "" {
				return "", output.Errorf(2, "empty value read from stdin")
			}
			return v, nil
		}
		return "", output.Errorf(2, "no value provided on stdin")
	}

	// 3. Interactive masked prompt
	if GCtx.NoInteractive {
		return "", output.Errorf(2, "--value is required in non-interactive mode")
	}

	var value string
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(prompt).
			EchoMode(huh.EchoModePassword).
			Value(&value),
	))
	if err := form.Run(); err != nil {
		return "", err
	}
	if value == "" {
		return "", output.Errorf(2, "secret value cannot be empty")
	}
	return value, nil
}

func init() {
	secretCreateCmd.Flags().String("name", "", "Secret name (required, no whitespace)")
	secretCreateCmd.Flags().String("value", "", "Secret value (omit for masked prompt or pipe via stdin)")
	secretUpdateCmd.Flags().String("value", "", "New secret value (omit for masked prompt or pipe via stdin)")

	secretCmd.AddCommand(secretListCmd, secretGetCmd, secretCreateCmd, secretUpdateCmd, secretDeleteCmd)
}
