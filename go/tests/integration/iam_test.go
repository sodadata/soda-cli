//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestIAMUserList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("table", func(t *testing.T) {
		r := run(t, "iam", "user", "list")
		assertExitCode(t, r, 0)
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "iam", "user", "list", "--output", "json")
		assertExitCode(t, r, 0)
	})
}

func TestIAMGroupCRUD(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	// List
	t.Run("list", func(t *testing.T) {
		r := run(t, "iam", "group", "list")
		assertExitCode(t, r, 0)
	})

	t.Run("list_json", func(t *testing.T) {
		r := run(t, "iam", "group", "list", "--output", "json")
		assertExitCode(t, r, 0)
	})

	// Create
	r := run(t, "iam", "group", "create", "--name", "integration-test-group")
	if r.ExitCode != 0 {
		t.Fatalf("group create failed: %s", r.Output())
	}
	assertOutputContains(t, r, "created")

	// Extract group ID
	groupID := ""
	for _, line := range strings.Split(r.Output(), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(line), "id:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				groupID = strings.TrimRight(parts[len(parts)-1], ").,")
			}
		}
	}
	if groupID == "" {
		t.Fatal("could not extract group ID from create output")
	}
	t.Logf("Created group: %s", groupID)

	// Update name
	t.Run("update_name", func(t *testing.T) {
		r := run(t, "iam", "group", "update", groupID, "--name", "integration-test-renamed")
		assertExitCode(t, r, 0)
	})

	// Delete
	t.Run("delete", func(t *testing.T) {
		r := run(t, "iam", "group", "delete", groupID)
		assertExitCode(t, r, 0)
	})
}

func TestIAMGroupDeleteBadID(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "iam", "group", "delete", "bad-id")
	assertExitCode(t, r, 2)
}

func TestIAMRoleList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("table", func(t *testing.T) {
		r := run(t, "iam", "role", "list")
		assertExitCode(t, r, 0)
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "iam", "role", "list", "--output", "json")
		assertExitCode(t, r, 0)
	})

	t.Run("scope_dataset", func(t *testing.T) {
		r := run(t, "iam", "role", "list", "--scope", "dataset")
		assertExitCode(t, r, 0)
	})

	t.Run("scope_global", func(t *testing.T) {
		r := run(t, "iam", "role", "list", "--scope", "global")
		assertExitCode(t, r, 0)
	})
}
