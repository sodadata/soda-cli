package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "Manage identity and access: roles, users, groups, service accounts",
}

// ── iam role ─────────────────────────────────────────────────────────────────

var iamRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage roles and their permissions",
}

var iamRoleListCmd = &cobra.Command{
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

var iamRoleCreateCmd = &cobra.Command{
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

var iamRoleDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Role '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

var iamRoleShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show permissions assigned to a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("  Permissions for role %s:\n\n", output.Bold.Render(args[0]))
		rows := []map[string]string{
			{"permission": "create-datasets", "granted": "yes"},
			{"permission": "manage-datasources", "granted": "yes"},
			{"permission": "manage-notification-rules", "granted": "no"},
		}
		cols := []string{"permission", "granted"}
		output.Render(rows, cols, nil, GCtx)
		return nil
	},
}

// ── iam user ──────────────────────────────────────────────────────────────────

var iamUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage organization users",
}

var iamUserListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users in the organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		result, err := client.ListUsers()
		if err != nil {
			return err
		}
		rows := make([]map[string]string, 0, len(result.Content))
		for _, u := range result.Content {
			name := u.FullName
			if name == "" {
				name = u.FirstName + " " + u.LastName
			}
			rows = append(rows, map[string]string{
				"id":    u.UserID,
				"email": u.Email,
				"name":  name,
			})
		}
		if len(rows) == 0 {
			fmt.Println(output.Dim.Render("  No users found."))
			return nil
		}
		cols := []string{"id", "email", "name"}
		output.Render(rows, cols, nil, GCtx)
		return nil
	},
}

var iamUserInviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a user to the organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		if email == "" {
			return output.Errorf(2, "--email is required")
		}
		output.PrintSuccess(fmt.Sprintf("Invitation sent to %s.", email), GCtx)
		return nil
	},
}

var iamUserRemoveCmd = &cobra.Command{
	Use:   "remove <user-id>",
	Short: "Remove a user from the organization",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("User '%s' removed from organization.", args[0]), GCtx)
		return nil
	},
}

var iamUserAssignCmd = &cobra.Command{
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

var iamUserRevokeCmd = &cobra.Command{
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

// ── iam group ────────────────────────────────────────────────────────────────

var iamGroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage user groups",
}

var iamGroupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		result, err := client.ListUserGroups()
		if err != nil {
			return err
		}
		rows := make([]map[string]string, 0, len(result.Content))
		for _, g := range result.Content {
			rows = append(rows, map[string]string{
				"id":      g.UserGroupID,
				"name":    g.Name,
				"members": fmt.Sprintf("%d", len(g.Users)),
			})
		}
		if len(rows) == 0 {
			fmt.Println(output.Dim.Render("  No groups found."))
			return nil
		}
		cols := []string{"id", "name", "members"}
		output.Render(rows, cols, nil, GCtx)
		return nil
	},
}

var iamGroupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		members, _ := cmd.Flags().GetStringArray("member")
		if name == "" {
			return output.Errorf(2, "--name is required")
		}
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		result, err := client.CreateUserGroup(api.CreateUserGroupRequest{
			Name:         name,
			MemberEmails: members,
		})
		if err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Group '%s' created (id: %s).", result.Name, result.UserGroupID), GCtx)
		return nil
	},
}

var iamGroupUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a group's name or members",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		newName, _ := cmd.Flags().GetString("name")
		addMembers, _ := cmd.Flags().GetStringArray("add-member")
		removeMembers, _ := cmd.Flags().GetStringArray("remove-member")

		req := api.UpdateUserGroupRequest{
			AddMemberEmails:    addMembers,
			RemoveMemberEmails: removeMembers,
		}
		if cmd.Flags().Changed("name") {
			req.Name = &newName
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}
		if _, err := client.UpdateUserGroup(args[0], req); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Group '%s' updated.", args[0]), GCtx)
		return nil
	},
}

var iamGroupDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		if err := client.DeleteUserGroup(args[0]); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Group '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

var iamGroupAssignCmd = &cobra.Command{
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

var iamGroupRevokeCmd = &cobra.Command{
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

// ── iam service-account ───────────────────────────────────────────────────────

var iamServiceAccountCmd = &cobra.Command{
	Use:   "service-account",
	Short: "Manage service accounts",
}

var iamSAListCmd = &cobra.Command{
	Use:   "list",
	Short: "List service accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		cols := []string{"id", "name", "email", "created"}
		output.Render(mock.ServiceAccounts, cols, nil, GCtx)
		return nil
	},
}

var iamSACreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a service account",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")
		if name == "" || email == "" {
			return output.Errorf(2, "--name and --email are required")
		}
		output.PrintSuccess(fmt.Sprintf("Service account '%s' (%s) created.", name, email), GCtx)
		return nil
	},
}

var iamSADeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a service account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintSuccess(fmt.Sprintf("Service account '%s' deleted.", args[0]), GCtx)
		return nil
	},
}

func init() {
	// role
	iamRoleListCmd.Flags().String("scope", "", "Filter by scope: global|dataset")
	iamRoleCreateCmd.Flags().String("name", "", "Role name (required)")
	iamRoleCreateCmd.Flags().String("scope", "", "Role scope: global|dataset (required)")
	iamRoleCreateCmd.Flags().String("description", "", "Role description")
	iamRoleCreateCmd.Flags().StringArray("permission", nil, "Permission to grant (repeatable)")
	iamRoleCmd.AddCommand(iamRoleListCmd, iamRoleCreateCmd, iamRoleDeleteCmd, iamRoleShowCmd)

	// user
	iamUserInviteCmd.Flags().String("email", "", "User email address (required)")
	iamUserAssignCmd.Flags().String("role", "", "Role ID (required)")
	iamUserRevokeCmd.Flags().String("role", "", "Role ID (required)")
	iamUserCmd.AddCommand(iamUserListCmd, iamUserInviteCmd, iamUserRemoveCmd, iamUserAssignCmd, iamUserRevokeCmd)

	// group
	iamGroupCreateCmd.Flags().String("name", "", "Group name (required)")
	iamGroupCreateCmd.Flags().StringArray("member", nil, "Initial member email (repeatable)")
	iamGroupUpdateCmd.Flags().String("name", "", "New group name")
	iamGroupUpdateCmd.Flags().StringArray("add-member", nil, "Email to add (repeatable)")
	iamGroupUpdateCmd.Flags().StringArray("remove-member", nil, "Email to remove (repeatable)")
	iamGroupAssignCmd.Flags().String("role", "", "Role ID (required)")
	iamGroupRevokeCmd.Flags().String("role", "", "Role ID (required)")
	iamGroupCmd.AddCommand(iamGroupListCmd, iamGroupCreateCmd, iamGroupUpdateCmd, iamGroupDeleteCmd, iamGroupAssignCmd, iamGroupRevokeCmd)

	// service-account
	iamSACreateCmd.Flags().String("name", "", "Service account name (required)")
	iamSACreateCmd.Flags().String("email", "", "Service account email (required)")
	iamServiceAccountCmd.AddCommand(iamSAListCmd, iamSACreateCmd, iamSADeleteCmd)

	iamCmd.AddCommand(iamRoleCmd, iamUserCmd, iamGroupCmd, iamServiceAccountCmd)
	rootCmd.AddCommand(iamCmd)
}
