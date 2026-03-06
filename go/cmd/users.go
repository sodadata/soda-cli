package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users and groups",
}

var usersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users in the organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		cols := []string{"id", "email", "name", "role", "status"}
		output.Render(mock.Users, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var usersAssignCmd = &cobra.Command{
	Use:   "assign <user-id>",
	Short: "Assign a role to a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		if role == "" {
			return output.Errorf(2, "--role is required")
		}
		output.PrintSuccess(fmt.Sprintf("Role '%s' assigned to user '%s'.", role, args[0]), GCtx)
		return nil
	},
}

var usersRevokeCmd = &cobra.Command{
	Use:   "revoke <user-id>",
	Short: "Revoke a role from a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		if role == "" {
			return output.Errorf(2, "--role is required")
		}
		output.PrintSuccess(fmt.Sprintf("Role '%s' revoked from user '%s'.", role, args[0]), GCtx)
		return nil
	},
}

// users group sub-group
var usersGroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage user groups",
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		cols := []string{"id", "name", "members", "role"}
		output.Render(mock.Groups, cols, nil, GCtx)
		return nil
	},
}

var groupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return output.Errorf(2, "--name is required")
		}
		output.PrintSuccess(fmt.Sprintf("Group '%s' created.", name), GCtx)
		return nil
	},
}

var groupUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Group '%s' updated.", args[0]), GCtx)
		return nil
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Group '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

var groupAssignCmd = &cobra.Command{
	Use:   "assign <group-id>",
	Short: "Assign a role to a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		if role == "" {
			return output.Errorf(2, "--role is required")
		}
		output.PrintSuccess(fmt.Sprintf("Role '%s' assigned to group '%s'.", role, args[0]), GCtx)
		return nil
	},
}

var groupRevokeCmd = &cobra.Command{
	Use:   "revoke <group-id>",
	Short: "Revoke a role from a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		role, _ := cmd.Flags().GetString("role")
		if role == "" {
			return output.Errorf(2, "--role is required")
		}
		output.PrintSuccess(fmt.Sprintf("Role '%s' revoked from group '%s'.", role, args[0]), GCtx)
		return nil
	},
}

func init() {
	usersAssignCmd.Flags().String("role", "", "Role ID to assign (required)")
	usersRevokeCmd.Flags().String("role", "", "Role ID to revoke (required)")

	groupCreateCmd.Flags().String("name", "", "Group name (required)")
	groupCreateCmd.Flags().StringArray("members", nil, "Initial member emails")
	groupAssignCmd.Flags().String("role", "", "Role ID to assign (required)")
	groupRevokeCmd.Flags().String("role", "", "Role ID to revoke (required)")

	usersGroupCmd.AddCommand(groupListCmd, groupCreateCmd, groupUpdateCmd, groupDeleteCmd, groupAssignCmd, groupRevokeCmd)
	usersCmd.AddCommand(usersListCmd, usersAssignCmd, usersRevokeCmd, usersGroupCmd)
	rootCmd.AddCommand(usersCmd)
}
