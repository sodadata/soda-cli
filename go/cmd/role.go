package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage roles and their permissions",
}

var roleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		scope, _ := cmd.Flags().GetString("scope")

		rows := mock.Roles
		if scope != "" {
			filtered := []map[string]string{}
			for _, r := range rows {
				if r["scope"] == scope {
					filtered = append(filtered, r)
				}
			}
			rows = filtered
		}

		cols := []string{"id", "name", "scope", "members"}
		output.Render(rows, cols, nil, GCtx)
		return nil
	},
}

var roleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new role",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		scope, _ := cmd.Flags().GetString("scope")

		if name == "" || scope == "" {
			return output.Errorf(2, "--name and --scope are required")
		}

		output.PrintSuccess(fmt.Sprintf("Role '%s' (scope: %s) created.", name, scope), GCtx)
		return nil
	},
}

var roleDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Role '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

var roleShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show permissions assigned to a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("  Permissions for role %s:\n\n", output.Bold.Render(args[0]))
		rows := []map[string]string{
			{"permission": "dataset.read", "scope": "global"},
			{"permission": "contract.verify", "scope": "global"},
			{"permission": "results.read", "scope": "global"},
		}
		cols := []string{"permission", "scope"}
		output.Render(rows, cols, nil, GCtx)
		return nil
	},
}

func init() {
	roleListCmd.Flags().String("scope", "", "Filter by scope: global|dataset")
	roleCreateCmd.Flags().String("name", "", "Role name (required)")
	roleCreateCmd.Flags().String("scope", "", "Role scope: global|dataset (required)")

	roleCmd.AddCommand(roleListCmd, roleCreateCmd, roleDeleteCmd, roleShowCmd)
	rootCmd.AddCommand(roleCmd)
}
