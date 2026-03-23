//go:build integration

package integration

import (
	"testing"
)

// TestBlockedCommands tests commands that are blocked by missing API endpoints.
// They should all fail gracefully with exit code 2 and a helpful message.

func TestIncident(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("list", func(t *testing.T) {
		r := run(t, "incident", "list")
		// API returns HTML — should fail gracefully
		t.Logf("incident list: exit=%d", r.ExitCode)
		if r.ExitCode == 0 {
			t.Log("incident list succeeded (API may have been fixed)")
		}
	})

	t.Run("get", func(t *testing.T) {
		r := run(t, "incident", "get", "fake-id")
		t.Logf("incident get: exit=%d", r.ExitCode)
	})

	t.Run("update", func(t *testing.T) {
		r := run(t, "incident", "update", "fake-id", "--title", "test")
		t.Logf("incident update: exit=%d", r.ExitCode)
	})
}

func TestNotification(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("rule_list", func(t *testing.T) {
		r := run(t, "notification", "rule", "list")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})

	t.Run("integration_list", func(t *testing.T) {
		r := run(t, "notification", "integration", "list")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})
}

func TestSecret(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("create", func(t *testing.T) {
		r := run(t, "secret", "create", "--name", "TEST", "--value", "abc")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})

	t.Run("update", func(t *testing.T) {
		r := run(t, "secret", "update", "fake-id", "--value", "new")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})

	t.Run("delete", func(t *testing.T) {
		r := run(t, "secret", "delete", "fake-id")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})
}

func TestDashboard(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "dashboard")
	// Dashboard may be implemented (exit 0) or blocked (exit 2)
	if r.ExitCode != 0 && r.ExitCode != 2 {
		t.Errorf("unexpected exit code %d", r.ExitCode)
	}
	t.Logf("dashboard: exit=%d", r.ExitCode)
}

func TestIAMBlocked(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("role_create", func(t *testing.T) {
		r := run(t, "iam", "role", "create", "--name", "test", "--scope", "dataset")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})

	t.Run("user_invite", func(t *testing.T) {
		r := run(t, "iam", "user", "invite", "--email", "test@example.com")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})

	t.Run("service_account_list", func(t *testing.T) {
		r := run(t, "iam", "service-account", "list")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})
}
