package cmd

import (
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
		return output.Errorf(2, "secret create is not yet available in the public API")
	},
}

var secretUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update the value of a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "secret update is not yet available in the public API")
	},
}

var secretDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "secret delete is not yet available in the public API")
	},
}

func init() {
	secretCreateCmd.Flags().String("name", "", "Secret name (required)")
	secretCreateCmd.Flags().String("value", "", "Secret value (required)")
	secretUpdateCmd.Flags().String("value", "", "New secret value (required)")

	secretCmd.AddCommand(secretCreateCmd, secretUpdateCmd, secretDeleteCmd)
}
