//go:build cli

package integration

import (
	"testing"
)

// TestBlockedCommands tests commands that are blocked by missing API endpoints.
// They should all fail gracefully with exit code 2 and a helpful message.
//
// The "not yet available" guard runs before the credential check, so these
// need no Cloud org.

func TestNotification(t *testing.T) {
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

func TestDashboard(t *testing.T) {
	r := run(t, "dashboard")
	// Dashboard may be implemented (exit 0) or blocked (exit 2)
	if r.ExitCode != 0 && r.ExitCode != 2 {
		t.Errorf("unexpected exit code %d", r.ExitCode)
	}
	t.Logf("dashboard: exit=%d", r.ExitCode)
}

func TestIAMBlocked(t *testing.T) {
	t.Run("role_create", func(t *testing.T) {
		r := run(t, "iam", "role", "create", "--name", "test", "--scope", "dataset")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})

	t.Run("service_account_list", func(t *testing.T) {
		r := run(t, "iam", "service-account", "list")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet")
	})
}
