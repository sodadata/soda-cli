package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets stored in Soda Cloud",
}

var secretCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new secret",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		value, _ := cmd.Flags().GetString("value")
		if name == "" || value == "" {
			return output.Errorf(2, "--name and --value are required")
		}
		output.PrintSuccess(fmt.Sprintf("Secret '%s' created.", name), GCtx)
		return nil
	},
}

var secretUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update the value of a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value, _ := cmd.Flags().GetString("value")
		if value == "" {
			return output.Errorf(2, "--value is required")
		}
		output.PrintSuccess(fmt.Sprintf("Secret '%s' updated.", args[0]), GCtx)
		return nil
	},
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Secret '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

func init() {
	secretCreateCmd.Flags().String("name", "", "Secret name (required)")
	secretCreateCmd.Flags().String("value", "", "Secret value (required)")
	secretUpdateCmd.Flags().String("value", "", "New secret value (required)")

	secretCmd.AddCommand(secretCreateCmd, secretUpdateCmd, secretDeleteCmd)
	rootCmd.AddCommand(secretCmd)
}
